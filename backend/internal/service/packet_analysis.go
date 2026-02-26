package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"nagare/internal/model"
	"nagare/internal/repository"
	"nagare/internal/repository/llm"
)

type PacketAnalysisReq struct {
	Name       string `json:"name"`
	ProviderID uint   `json:"provider_id"`
	Model      string `json:"model"`
	RawContent string `json:"raw_content"`
}

func AddPacketAnalysisServ(req PacketAnalysisReq, userID uint, fileName string) (model.PacketAnalysis, error) {
	pa := model.PacketAnalysis{
		Name:       req.Name,
		AIModel:    req.Model,
		RawContent: req.RawContent,
		Status:     0, // Pending
		FilePath:   fileName,
	}
	if req.ProviderID > 0 {
		pID := req.ProviderID
		pa.ProviderID = &pID
	}
	if userID > 0 {
		uID := userID
		pa.UserID = &uID
	}

	if err := repository.AddPacketAnalysisDAO(&pa); err != nil {
		return pa, err
	}

	return pa, nil
}

func StartPacketAnalysisServ(id uint) error {
	pa, err := repository.GetPacketAnalysisByIDDAO(id)
	if err != nil {
		return err
	}

	pa.Status = 1 // Analyzing
	_ = repository.UpdatePacketAnalysisDAO(&pa)

	go analyzePacketAsync(pa)
	return nil
}

func analyzePacketAsync(pa model.PacketAnalysis) {
	fmt.Printf(">>> Starting AI analysis for packet '%s' (ID: %d)\n", pa.Name, pa.ID)
	
	var pID uint = 0
	if pa.ProviderID != nil {
		pID = *pa.ProviderID
	}
	
	client, resolvedModel, err := createLLMClient(pID, pa.AIModel)
	if err != nil {
		fmt.Printf(">>> Failed to create AI client: %v\n", err)
		updatePacketAnalysisStatus(pa.ID, 3, "Failed to create AI client: "+err.Error(), "error")
		return
	}

	// Prepare data for AI
	var contentToAnalyze string
	if pa.RawContent != "" {
		contentToAnalyze = "RAW DATA SNIPPET:\n" + pa.RawContent
	} else if pa.FilePath != "" {
		fmt.Printf(">>> Extracting summary from file: %s\n", pa.FilePath)
		contentToAnalyze = extractPacketSummary(pa.FilePath)
	}

	if contentToAnalyze == "" {
		contentToAnalyze = "No packet data could be extracted for analysis. Please check if the file is a valid capture or provides enough information."
	}

	fmt.Printf(">>> Sending content to AI (%d chars). Model: %s\n", len(contentToAnalyze), resolvedModel)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	systemPrompt := packetAnalysisPrompt()
	start := time.Now()

	resp, err := client.Chat(ctx, llm.ChatRequest{
		Model:        resolvedModel,
		SystemPrompt: systemPrompt,
		MaxTokens:    4096,
		Messages: []llm.Message{
			{Role: "user", Content: contentToAnalyze},
		},
	})

	duration := time.Since(start)
	
	var logPID uint = 0
	if pa.ProviderID != nil {
		logPID = *pa.ProviderID
	}
	logLLMRequest("packet_analysis", logPID, resolvedModel, duration, err)

	if err != nil {
		fmt.Printf(">>> AI Analysis failed after %v: %v\n", duration, err)
		updatePacketAnalysisStatus(pa.ID, 3, "AI Analysis failed: "+err.Error(), "error")
		return
	}

	analysis := resp.Content
	if analysis == "" {
		fmt.Printf(">>> AI returned empty response for packet ID %d\n", pa.ID)
		analysis = "AI returned an empty analysis result. This may happen if the input was too cryptic or the model refused to analyze it."
	}

	riskLevel := parseRiskLevel(analysis)
	fmt.Printf(">>> Analysis completed in %v. Risk: %s\n", duration, riskLevel)

	pa.Status = 2 // Completed
	pa.Analysis = analysis
	pa.RiskLevel = riskLevel
	if err := repository.UpdatePacketAnalysisDAO(&pa); err != nil {
		fmt.Printf(">>> Failed to update database for packet ID %d: %v\n", pa.ID, err)
	}

	// Notify if notable or malicious
	if riskLevel == "notable" || riskLevel == "malicious" {
		title := "Notable Packet Detected"
		if riskLevel == "malicious" {
			title = "Malicious Packet Detected!"
		}
		_ = CreateSiteMessageServ(title, fmt.Sprintf("AI analyzed packet '%s' and determined it is %s. Details: %s", pa.Name, riskLevel, pa.Name), "alert", 3, pa.UserID)
	}
}

func extractPacketSummary(fileName string) string {
	fullPath := filepath.Join("public/uploads/packets", fileName)
	fmt.Printf(">>> Checking for file at: %s\n", fullPath)
	
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Sprintf("Error: File %s does not exist on server.", fileName)
	}

	// Helper to find tool in path or common Windows locations
	findTool := func(name string) string {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
		// Common Windows paths
		var paths []string
		if name == "tshark" {
			paths = []string{
				`C:\Program Files\Wireshark\tshark.exe`,
				`C:\Program Files (x86)\Wireshark\tshark.exe`,
			}
		} else if name == "tcpdump" {
			paths = []string{
				`C:\Program Files\tcpdump\tcpdump.exe`,
				`C:\wpcap\tcpdump.exe`,
			}
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
		return ""
	}

	// Try tshark
	if tool := findTool("tshark"); tool != "" {
		fmt.Printf(">>> Using tshark at %s for extraction...\n", tool)
		// Extract summary: frame number, protocol, source, destination, ports
		cmd := exec.Command(tool, "-r", fullPath, "-c", "100", "-T", "fields", "-e", "frame.number", "-e", "_ws.col.Protocol", "-e", "ip.src", "-e", "ip.dst", "-e", "tcp.dstport", "-e", "udp.dstport")
		out, err := cmd.CombinedOutput()
		if err == nil && len(out) > 0 {
			return "EXTRACTED PACKET SUMMARY (TSHARK):\n" + string(out)
		}
		fmt.Printf(">>> tshark failed: %v\n", err)
	}

	// Try tcpdump
	if tool := findTool("tcpdump"); tool != "" {
		fmt.Printf(">>> Using tcpdump at %s for extraction...\n", tool)
		cmd := exec.Command(tool, "-nn", "-r", fullPath, "-c", "50")
		out, err := cmd.CombinedOutput()
		if err == nil && len(out) > 0 {
			return "EXTRACTED HEADERS (TCPDUMP):\n" + string(out)
		}
		fmt.Printf(">>> tcpdump failed: %v\n", err)
	}

	// Last fallback: Read the file as text/hex
	fmt.Println(">>> Using raw hex fallback (first 8KB)...")
	file, err := os.Open(fullPath)
	if err != nil {
		return fmt.Sprintf("Error opening file for analysis: %v", err)
	}
	defer file.Close()

	// Send a bit more data in fallback, up to 8KB hex
	buf := make([]byte, 4096)
	n, _ := file.Read(buf)
	if n > 0 {
		return fmt.Sprintf("RAW FILE CONTENT (HEX DUMP - first %d bytes):\n%x\n\n[End of Hex Dump]\nNote to AI: This is a binary packet capture file. Please analyze the magic numbers and structural patterns if possible.", n, buf[:n])
	}

	return "Could not extract any data from the file. Please ensure it is a valid capture file."
}

func updatePacketAnalysisStatus(id uint, status int, analysis string, risk string) {
	pa, err := repository.GetPacketAnalysisByIDDAO(id)
	if err == nil {
		pa.Status = status
		pa.Analysis = analysis
		pa.RiskLevel = risk
		_ = repository.UpdatePacketAnalysisDAO(&pa)
	}
}

func parseRiskLevel(analysis string) string {
	lower := strings.ToLower(analysis)
	
	// Check for explicit markers first to avoid false positives
	if strings.Contains(lower, "risk: malicious") || strings.Contains(lower, "risk: [malicious]") {
		return "malicious"
	}
	if strings.Contains(lower, "risk: notable") || strings.Contains(lower, "risk: [notable]") || strings.Contains(lower, "risk: suspicious") {
		return "notable"
	}
	if strings.Contains(lower, "risk: clean") || strings.Contains(lower, "risk: [clean]") {
		return "clean"
	}

	// Fallback to keyword search
	if strings.Contains(lower, "malicious") || strings.Contains(lower, "attack") || strings.Contains(lower, "danger") || strings.Contains(lower, "exploit") {
		return "malicious"
	}
	if strings.Contains(lower, "notable") || strings.Contains(lower, "suspicious") || strings.Contains(lower, "warning") || strings.Contains(lower, "unusual") {
		return "notable"
	}
	return "clean"
}

func packetAnalysisPrompt() string {
	return `You are a world-class network security expert specializing in Huawei infrastructure. Your task is to analyze network traffic data provided to you.

` + systemContextPrompt() + `

DATA FORMAT:
The input may be a TSHARK field summary (Frame, Protocol, Source IP, Destination IP, Port), TCPDUMP headers, or a RAW HEX DUMP of a capture file.

INSTRUCTIONS:
1. Identify the primary protocols involved.
2. Search for security risks:
   - Port scanning patterns (many destination ports from one source).
   - Cleartext credentials in payloads (if ASCII is visible).
   - Unusual protocol usage or non-standard ports.
   - Signs of tunneling or data exfiltration.
   - Known attack patterns (e.g., SYN floods, SQLi, XSS in HTTP).
3. If provided with a HEX DUMP, look for magic numbers at the start (e.g., "d4c3b2a1" for libpcap, "0a0d0d0a" for pcapng) and try to infer the capture environment.

REQUIRED OUTPUT STRUCTURE:
Summary:
- High-level technical overview.

Findings:
- Key technical details (IPs, Ports, Protocols, Anomalies).

Risk Assessment:
- RISK: [CLEAN / NOTABLE / MALICIOUS]
- Explicit reasoning.

Recommendations:
- Concrete next steps (e.g. VRP CLI commands via SSH to block or isolate).

Be decisive. If you see high-risk patterns, mark it as MALICIOUS immediately.`
}

func GetAllPacketAnalysesServ(userID uint) ([]model.PacketAnalysis, error) {
	return repository.GetAllPacketAnalysesDAO(userID)
}

func DeletePacketAnalysisServ(id uint) error {
	pa, err := repository.GetPacketAnalysisByIDDAO(id)
	if err != nil {
		return err
	}

	// Delete file if exists
	if pa.FilePath != "" {
		uploadDir := "public/uploads/packets"
		_ = os.Remove(filepath.Join(uploadDir, pa.FilePath))
	}

	return repository.DeletePacketAnalysisDAO(id)
}
