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
		ProviderID: req.ProviderID,
		AIModel:    req.Model,
		RawContent: req.RawContent,
		UserID:     userID,
		Status:     0, // Pending
		FilePath:   fileName,
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
	client, resolvedModel, err := createLLMClient(pa.ProviderID, pa.AIModel)
	if err != nil {
		updatePacketAnalysisStatus(pa.ID, 3, "Failed to create AI client: "+err.Error(), "error")
		return
	}

	// Prepare data for AI
	var contentToAnalyze string
	if pa.RawContent != "" {
		contentToAnalyze = pa.RawContent
	} else if pa.FilePath != "" {
		contentToAnalyze = extractPacketSummary(pa.FilePath)
	}

	if contentToAnalyze == "" {
		contentToAnalyze = "No packet data could be extracted for analysis."
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	systemPrompt := packetAnalysisPrompt()
	start := time.Now()

	resp, err := client.Chat(ctx, llm.ChatRequest{
		Model:        resolvedModel,
		SystemPrompt: systemPrompt,
		Messages: []llm.Message{
			{Role: "user", Content: contentToAnalyze},
		},
	})

	logLLMRequest("packet_analysis", pa.ProviderID, resolvedModel, time.Since(start), err)

	if err != nil {
		updatePacketAnalysisStatus(pa.ID, 3, "AI Analysis failed: "+err.Error(), "error")
		return
	}

	analysis := resp.Content
	riskLevel := parseRiskLevel(analysis)

	pa.Status = 2 // Completed
	pa.Analysis = analysis
	pa.RiskLevel = riskLevel
	_ = repository.UpdatePacketAnalysisDAO(&pa)

	// Notify if notable or malicious
	if riskLevel == "notable" || riskLevel == "malicious" {
		title := "Notable Packet Detected"
		if riskLevel == "malicious" {
			title = "Malicious Packet Detected!"
		}
		_ = CreateSiteMessageServ(title, fmt.Sprintf("AI analyzed packet '%s' and determined it is %s. Details: %s", pa.Name, riskLevel, pa.Name), "alert", 3, &pa.UserID)
	}
}

func extractPacketSummary(fileName string) string {
	fullPath := filepath.Join("public/uploads/packets", fileName)
	
	// Check if tshark is available
	if _, err := exec.LookPath("tshark"); err == nil {
		// Extract top conversations and protocol hierarchy as a summary
		// -c 100: limit to first 100 packets for analysis speed and token limits
		cmd := exec.Command("tshark", "-r", fullPath, "-c", "100", "-T", "fields", "-e", "frame.number", "-e", "_ws.col.Protocol", "-e", "ip.src", "-e", "ip.dst", "-e", "tcp.dstport", "-e", "udp.dstport")
		out, err := cmd.CombinedOutput()
		if err == nil && len(out) > 0 {
			return "Extracted first 100 packets summary:\n" + string(out)
		}
	}

	// Fallback to tcpdump if available
	if _, err := exec.LookPath("tcpdump"); err == nil {
		cmd := exec.Command("tcpdump", "-nn", "-r", fullPath, "-c", "50")
		out, err := cmd.CombinedOutput()
		if err == nil && len(out) > 0 {
			return "Extracted packet headers (tcpdump):\n" + string(out)
		}
	}

	// Last fallback: Read the file as text/hex (first 4KB)
	file, err := os.Open(fullPath)
	if err != nil {
		return fmt.Sprintf("Error opening file for analysis: %v", err)
	}
	defer file.Close()

	buf := make([]byte, 4096)
	n, _ := file.Read(buf)
	if n > 0 {
		return fmt.Sprintf("Raw file content (first %d bytes):\n%x", n, buf[:n])
	}

	return "Could not extract any data from the file."
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
	if strings.Contains(lower, "malicious") || strings.Contains(lower, "attack") || strings.Contains(lower, "danger") {
		return "malicious"
	}
	if strings.Contains(lower, "notable") || strings.Contains(lower, "suspicious") || strings.Contains(lower, "warning") {
		return "notable"
	}
	return "clean"
}

func packetAnalysisPrompt() string {
	return `You are a senior network security analyst and protocol expert.
Your task is to analyze the provided packet capture data (hex, text, or flow summary).

Determine:
1. What protocol/service this is.
2. The source and destination if apparent.
3. If the packet/flow is "clean", "notable" (suspicious or unusual), or "malicious" (part of an attack, exploit, or malware).

Output Format (use headings):
Summary:
- Brief description of the traffic.

Findings:
- Key attributes (Port, Protocol, Flags).

Risk Assessment:
- RISK: [CLEAN / NOTABLE / MALICIOUS]
- Reasoning for this classification.

Recommendations:
- What the administrator should do.

Keep your analysis technical and objective.`
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
