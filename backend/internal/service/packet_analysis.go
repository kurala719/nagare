package service

import (
	"context"
	"fmt"
	"os"
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
		// Read small part of file if it's text-based, or just use name for now
		// In a real scenario, we might use a library to parse PCAP
		contentToAnalyze = fmt.Sprintf("Analyze packet capture file: %s", pa.FilePath)
		// For now, let's assume we want AI to look at the name and whatever info we have
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
