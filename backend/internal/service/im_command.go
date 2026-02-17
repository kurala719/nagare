package service

import (
	"context"
	"fmt"
	"strings"

	"nagare/internal/model"
	mediaSvc "nagare/internal/repository/media"
)

type IMCommandResult struct {
	Reply string `json:"reply"`
}

// IMCommandContext contains context for executing IM commands
type IMCommandContext struct {
	MediaType  string
	UserID     string
	GroupID    string
	ProviderID uint
}

// HandleIMCommand processes incoming IM commands
func HandleIMCommand(message string) (IMCommandResult, error) {
	trimmed := strings.TrimSpace(message)
	lower := strings.ToLower(trimmed)

	// Status command
	if strings.HasPrefix(lower, "/status") || lower == "status" {
		score, err := GetHealthScoreServ()
		if err != nil {
			return IMCommandResult{}, err
		}
		reply := fmt.Sprintf("Health Score: %d (monitors %d/%d, hosts %d/%d, items %d/%d)",
			score.Score,
			score.MonitorActive, score.MonitorTotal,
			score.HostActive, score.HostTotal,
			score.ItemActive, score.ItemTotal,
		)
		return IMCommandResult{Reply: reply}, nil
	}

	// Get alerts command
	if strings.HasPrefix(lower, "/get_alert") {
		return handleGetAlerts()
	}

	// Chat command
	if strings.HasPrefix(lower, "/chat") {
		content := strings.TrimSpace(trimmed[5:])
		if content == "" {
			return IMCommandResult{Reply: "Usage: /chat <message>"}, nil
		}
		return handleChatCommand(content)
	}

	return IMCommandResult{Reply: "Unsupported command. Try /status, /get_alert, or /chat <message>."}, nil
}

// handleGetAlerts retrieves active alerts
func handleGetAlerts() (IMCommandResult, error) {
	status := 0 // 0 = active
	limit := 10
	filter := model.AlertFilter{
		Status: &status,
		Limit:  limit,
		Offset: 0,
	}

	alerts, err := SearchAlertsServ(filter)
	if err != nil {
		return IMCommandResult{Reply: fmt.Sprintf("Error retrieving alerts: %v", err)}, nil
	}

	if len(alerts) == 0 {
		return IMCommandResult{Reply: "No active alerts."}, nil
	}

	var reply strings.Builder
	reply.WriteString(fmt.Sprintf("Active Alerts (%d):\n", len(alerts)))
	for i, alert := range alerts {
		reply.WriteString(fmt.Sprintf("[%d] %s (Severity: %d)\n", i+1, alert.Message, alert.Severity))
	}

	return IMCommandResult{Reply: reply.String()}, nil
}

// handleChatCommand processes chat messages with LLM
func handleChatCommand(content string) (IMCommandResult, error) {
	// Get the default provider (first one)
	providers, err := GetAllProvidersServ()
	if err != nil || len(providers) == 0 {
		return IMCommandResult{Reply: "No LLM provider configured."}, nil
	}

	// Use the first available provider
	providerID := uint(providers[0].ID)

	chatReq := ChatReq{
		ProviderID: providerID,
		Content:    content,
		UseTools:   nil, // No tools for IM chat
		Privileges: 1,
	}

	// Get the model from provider
	if providers[0].DefaultModel != "" {
		chatReq.Model = providers[0].DefaultModel
	}

	resp, err := SendChatServ(chatReq)
	if err != nil {
		return IMCommandResult{Reply: fmt.Sprintf("Chat error: %v", err)}, nil
	}

	return IMCommandResult{Reply: resp.Content}, nil
}

// HandleIMCommandWithContext processes IM commands with media context
func HandleIMCommandWithContext(message string, ctx IMCommandContext) (IMCommandResult, error) {
	return HandleIMCommand(message)
}

func SendIMReply(mediaType, target, message string) error {
	if strings.TrimSpace(mediaType) == "" || strings.TrimSpace(target) == "" {
		return nil
	}
	return mediaSvc.GetService().SendMessage(context.Background(), strings.ToLower(mediaType), target, message)
}
