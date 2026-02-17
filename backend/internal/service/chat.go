package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"nagare/internal/model"
	"nagare/internal/repository"
	"nagare/internal/repository/llm"
)

// ChatReq represents a chat request
type ChatReq struct {
	ProviderID uint   `json:"provider_id" binding:"required"`
	Model      string `json:"model"`
	Content    string `json:"content" binding:"required"`
	Mode       string `json:"mode,omitempty"`
	UseTools   *bool  `json:"use_tools,omitempty"`
	Privileges int    `json:"-"`
}

// ChatRes represents a chat response
type ChatRes struct {
	ID         uint   `json:"id"`
	ProviderID uint   `json:"provider_id" binding:"required"`
	Role       string `json:"role" binding:"required"`
	Model      string `json:"model"`
	Content    string `json:"content" binding:"required"`
}

// GetAllChatsServ retrieves chat history (limited to 10 items)
func GetAllChatsServ() ([]model.Chat, error) {
	return repository.GetAllChatsDAO()
}

// GetChatsWithPaginationServ retrieves chat history with pagination
func GetChatsWithPaginationServ(limit, offset int) ([]model.Chat, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return repository.GetChatsWithLimitDAO(limit, offset)
}

// SearchChatsServ retrieves chat history by filter
func SearchChatsServ(filter model.ChatFilter) ([]model.Chat, error) {
	return repository.SearchChatsDAO(filter)
}

// SendChatServ sends a normal chat message (no special context)
func SendChatServ(req ChatReq) (ChatRes, error) {
	if isNetworkStatusQuery(req.Content) {
		return analyzeNetworkStatus(req)
	}
	personaPrompt := resolveChatPersonaPrompt(req.Mode)
	useTools := req.Privileges >= 2
	if req.UseTools != nil {
		useTools = *req.UseTools && req.Privileges >= 2
	}
	if useTools {
		return sendChatWithTools(req, personaPrompt)
	}
	return sendChatPlain(req, personaPrompt)
}

func analyzeNetworkStatus(req ChatReq) (ChatRes, error) {
	client, llmModel, err := createLLMClient(req.ProviderID, req.Model)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(req.ProviderID, 2)
		return ChatRes{}, err
	}

	status := 0
	limit := 50
	alertFilter := model.AlertFilter{Status: &status, Limit: limit}
	alerts, err := SearchAlertsServ(alertFilter)
	if err != nil {
		alerts = nil
	}

	metrics, err := GetNetworkMetricsServ("", 50)
	if err != nil {
		metrics = nil
	}

	health, err := GetHealthScoreServ()
	if err != nil {
		health = HealthScore{}
	}

	context := buildNetworkStatusContext(alerts, metrics, health)

	ctx, cancel := aiAnalysisContext()
	defer cancel()

	start := time.Now()
	resp, err := client.Chat(ctx, llm.ChatRequest{
		Model:        llmModel,
		SystemPrompt: networkStatusPrompt(),
		Messages: []llm.Message{
			{Role: "user", Content: context},
		},
	})
	logLLMRequest("network_status", req.ProviderID, llmModel, time.Since(start), err)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(req.ProviderID, 2)
		return ChatRes{}, fmt.Errorf("failed to analyze network status: %w", err)
	}

	_ = repository.UpdateProviderStatusDAO(req.ProviderID, 1)
	return ChatRes{Content: resp.Content, ProviderID: req.ProviderID, Role: "assistant", Model: llmModel}, nil
}

func sendChatPlain(req ChatReq, personaPrompt string) (ChatRes, error) {
	client, llmModel, err := createLLMClient(req.ProviderID, req.Model)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(req.ProviderID, 2)
		return ChatRes{}, err
	}

	if err := repository.AddChatDAO(model.Chat{
		UserID:     1,
		ProviderID: req.ProviderID,
		LLMModel:   llmModel,
		Role:       "user",
		Content:    req.Content,
	}); err != nil {
		return ChatRes{}, fmt.Errorf("failed to store user message: %w", err)
	}

	ctx := context.Background()
	start := time.Now()
	responseText := ""
	if personaPrompt != "" {
		resp, err := client.Chat(ctx, llm.ChatRequest{
			Model:        llmModel,
			SystemPrompt: personaPrompt,
			Messages: []llm.Message{
				{Role: "user", Content: req.Content},
			},
		})
		logLLMRequest("persona_chat", req.ProviderID, llmModel, time.Since(start), err)
		if err != nil {
			_ = repository.UpdateProviderStatusDAO(req.ProviderID, 2)
			return ChatRes{}, fmt.Errorf("failed to generate content: %w", err)
		}
		responseText = resp.Content
	} else {
		responseText, err = client.SimpleChat(ctx, llmModel, req.Content)
		logLLMRequest("simple_chat", req.ProviderID, llmModel, time.Since(start), err)
		if err != nil {
			_ = repository.UpdateProviderStatusDAO(req.ProviderID, 2)
			return ChatRes{}, fmt.Errorf("failed to generate content: %w", err)
		}
	}

	if err := repository.AddChatDAO(model.Chat{
		UserID:     1,
		ProviderID: req.ProviderID,
		LLMModel:   llmModel,
		Role:       "assistant",
		Content:    responseText,
	}); err != nil {
		return ChatRes{}, fmt.Errorf("failed to store AI response: %w", err)
	}

	_ = repository.UpdateProviderStatusDAO(req.ProviderID, 1)
	return ChatRes{Content: responseText, ProviderID: req.ProviderID, Role: "assistant", Model: llmModel}, nil
}

func sendChatWithTools(req ChatReq, personaPrompt string) (ChatRes, error) {
	client, llmModel, err := createLLMClient(req.ProviderID, req.Model)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(req.ProviderID, 2)
		return ChatRes{}, err
	}

	if err := repository.AddChatDAO(model.Chat{
		UserID:     1,
		ProviderID: req.ProviderID,
		LLMModel:   llmModel,
		Role:       "user",
		Content:    req.Content,
	}); err != nil {
		return ChatRes{}, fmt.Errorf("failed to store user message: %w", err)
	}

	tools := ListTools()
	ctx := context.Background()
	start := time.Now()
	initialResp, err := client.Chat(ctx, llm.ChatRequest{
		Model:        llmModel,
		SystemPrompt: buildToolSystemPrompt(tools, personaPrompt),
		Messages: []llm.Message{
			{Role: "user", Content: req.Content},
		},
	})
	logLLMRequest("tool_chat", req.ProviderID, llmModel, time.Since(start), err)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(req.ProviderID, 2)
		return ChatRes{}, fmt.Errorf("failed to generate content: %w", err)
	}

	toolCall, ok := parseToolCall(initialResp.Content)
	finalText := initialResp.Content
	if ok {
		toolResult, err := CallTool(toolCall.Name, toolCall.Arguments)
		if err != nil {
			toolResult = map[string]string{"error": err.Error()}
		}
		resultJSON, _ := json.Marshal(toolResult)
		toolResultText := fmt.Sprintf("Tool result for %s: %s", toolCall.Name, string(resultJSON))
		start := time.Now()
		finalResp, err := client.Chat(ctx, llm.ChatRequest{
			Model:        llmModel,
			SystemPrompt: toolAnswerPrompt(personaPrompt),
			Messages: []llm.Message{
				{Role: "user", Content: req.Content},
				{Role: "user", Content: toolResultText},
			},
		})
		logLLMRequest("tool_answer", req.ProviderID, llmModel, time.Since(start), err)
		if err != nil {
			_ = repository.UpdateProviderStatusDAO(req.ProviderID, 2)
			return ChatRes{}, fmt.Errorf("failed to generate content: %w", err)
		}
		finalText = finalResp.Content
	}

	if err := repository.AddChatDAO(model.Chat{
		UserID:     1,
		ProviderID: req.ProviderID,
		LLMModel:   llmModel,
		Role:       "assistant",
		Content:    finalText,
	}); err != nil {
		return ChatRes{}, fmt.Errorf("failed to store AI response: %w", err)
	}

	_ = repository.UpdateProviderStatusDAO(req.ProviderID, 1)
	return ChatRes{Content: finalText, ProviderID: req.ProviderID, Role: "assistant", Model: llmModel}, nil
}

type toolCall struct {
	Name      string          `json:"tool"`
	AltName   string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func parseToolCall(content string) (toolCall, bool) {
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, "{") || !strings.HasSuffix(trimmed, "}") {
		return toolCall{}, false
	}
	var call toolCall
	if err := json.Unmarshal([]byte(trimmed), &call); err != nil {
		return toolCall{}, false
	}
	if call.Name == "" {
		call.Name = call.AltName
	}
	if call.Name == "" {
		return toolCall{}, false
	}
	if len(call.Arguments) == 0 {
		call.Arguments = json.RawMessage("{}")
	}
	return call, true
}

func buildToolSystemPrompt(tools []ToolDefinition, personaPrompt string) string {
	var builder strings.Builder
	if personaPrompt != "" {
		builder.WriteString(personaPrompt)
		builder.WriteString("\n\n")
	}
	builder.WriteString("You are an assistant that can call server tools when needed.\n")
	builder.WriteString("If a tool is required, respond ONLY with JSON: {\"tool\":\"name\",\"arguments\":{...}}.\n")
	builder.WriteString("If no tool is required, respond normally.\n\n")
	builder.WriteString("Available tools:\n")
	for _, tool := range tools {
		toolSchema, _ := json.Marshal(tool.InputSchema)
		builder.WriteString("- ")
		builder.WriteString(tool.Name)
		builder.WriteString(": ")
		builder.WriteString(tool.Description)
		builder.WriteString(" Args schema: ")
		builder.WriteString(string(toolSchema))
		builder.WriteString("\n")
	}
	return builder.String()
}

func toolAnswerPrompt(personaPrompt string) string {
	base := "Use the tool result to answer the user. Summarize with counts and key fields. If the list is long, show the top 10 and mention there are more. Do not call tools."
	if personaPrompt == "" {
		return base
	}
	return personaPrompt + "\n\n" + base
}

func resolveChatPersonaPrompt(mode string) string {
	mode = strings.TrimSpace(strings.ToLower(mode))
	if mode != "roast" {
		return ""
	}
	return "You are a witty, slightly sarcastic senior SRE. Keep it playful and professional.\n" +
		"Rules:\n" +
		"- No profanity, slurs, or personal attacks.\n" +
		"- Critique configurations and metrics, not people.\n" +
		"- Always include at least one concrete, actionable fix.\n" +
		"- Keep responses concise (3-6 sentences)."
}

// ConsultAlertServ consults AI about a specific alert
func ConsultAlertServ(providerID uint, model string, alertID int) (ChatRes, error) {
	// Get alert data
	alert, err := repository.GetAlertByIDDAO(alertID)
	if err != nil {
		return ChatRes{}, fmt.Errorf("failed to get alert: %w", err)
	}

	client, resolvedModel, err := createLLMClient(providerID, model)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(providerID, 2)
		return ChatRes{}, err
	}

	ctx, cancel := aiAnalysisContext()
	defer cancel()
	systemPrompt := alertAnalysisPrompt()
	start := time.Now()

	alertData := fmt.Sprintf("Alert ID: %d\nHost ID: %d\nSeverity: %d\nMessage: %s\nStatus: %d",
		alert.ID, alert.HostID, alert.Severity, sanitizeSensitiveText(alert.Message), alert.Status)

	resp, err := client.Chat(ctx, llm.ChatRequest{
		Model:        resolvedModel,
		SystemPrompt: systemPrompt,
		Messages: []llm.Message{
			{Role: "user", Content: alertData},
		},
	})
	logLLMRequest("alert_consult", providerID, resolvedModel, time.Since(start), err)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(providerID, 2)
		return ChatRes{}, fmt.Errorf("failed to analyze alert: %w", err)
	}

	_ = repository.UpdateProviderStatusDAO(providerID, 1)
	return ChatRes{Content: resp.Content, ProviderID: providerID, Role: "assistant", Model: resolvedModel}, nil
}

// ConsultItemServ consults AI about a specific monitoring item
func ConsultItemServ(itemID uint) (ChatRes, error) {
	// Get item data
	item, err := repository.GetItemByIDDAO(itemID)
	if err != nil {
		return ChatRes{}, fmt.Errorf("failed to get item: %w", err)
	}

	// Get host data for context
	host, err := repository.GetHostByIDDAO(item.HID)
	if err != nil {
		return ChatRes{}, fmt.Errorf("failed to get host: %w", err)
	}

	client, resolvedModel, err := createLLMClient(1, "") // Default to provider ID 1
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(1, 2)
		return ChatRes{}, err
	}

	ctx, cancel := aiAnalysisContext()
	defer cancel()
	systemPrompt := itemAnalysisPrompt()

	itemData := fmt.Sprintf("Host: %s\nItem Name: %s\nItem ID: %s\nCurrent Value: %s\nUnits: %s",
		sanitizeSensitiveText(host.Name), sanitizeSensitiveText(item.Name), item.ItemID, sanitizeSensitiveText(item.LastValue), sanitizeSensitiveText(item.Units))

	start := time.Now()
	resp, err := client.Chat(ctx, llm.ChatRequest{
		Model:        resolvedModel,
		SystemPrompt: systemPrompt,
		Messages: []llm.Message{
			{Role: "user", Content: itemData},
		},
	})
	logLLMRequest("item_consult", 1, resolvedModel, time.Since(start), err)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(1, 2)
		return ChatRes{}, fmt.Errorf("failed to analyze item: %w", err)
	}

	_ = repository.UpdateProviderStatusDAO(1, 1)
	return ChatRes{Content: resp.Content, ProviderID: 1, Role: "assistant", Model: resolvedModel}, nil
}

// ConsultHostServ consults AI about a host's status based on all its items
func ConsultHostServ(providerID uint, model string, hostID uint) (ChatRes, error) {
	// Get host data
	host, err := repository.GetHostByIDDAO(hostID)
	if err != nil {
		return ChatRes{}, fmt.Errorf("failed to get host: %w", err)
	}

	// Get all items for this host
	items, err := repository.GetItemsByHIDDAO(hostID)
	if err != nil {
		return ChatRes{}, fmt.Errorf("failed to get host items: %w", err)
	}

	client, resolvedModel, err := createLLMClient(providerID, model)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(providerID, 2)
		return ChatRes{}, err
	}

	ctx, cancel := aiAnalysisContext()
	defer cancel()
	systemPrompt := hostAnalysisPrompt()

	// Build items data string
	var itemsData string
	if len(items) == 0 {
		itemsData = "No monitoring metrics available for this host.\n"
	} else {
		for _, item := range items {
			itemsData += fmt.Sprintf("- %s: %s %s\n", sanitizeSensitiveText(item.Name), sanitizeSensitiveText(item.LastValue), sanitizeSensitiveText(item.Units))
		}
	}

	hostData := fmt.Sprintf("Host: %s\nIP Address: %s\nStatus: %d\nDescription: %s\n\nMonitoring Metrics:\n%s",
		sanitizeSensitiveText(host.Name), sanitizeSensitiveText(host.IPAddr), host.Status, sanitizeSensitiveText(host.Description), itemsData)

	start := time.Now()
	resp, err := client.Chat(ctx, llm.ChatRequest{
		Model:        resolvedModel,
		SystemPrompt: systemPrompt,
		Messages: []llm.Message{
			{Role: "user", Content: hostData},
		},
	})
	logLLMRequest("host_consult", providerID, resolvedModel, time.Since(start), err)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(providerID, 2)
		return ChatRes{}, fmt.Errorf("failed to analyze host: %w", err)
	}

	_ = repository.UpdateProviderStatusDAO(providerID, 1)
	return ChatRes{Content: resp.Content, ProviderID: providerID, Role: "assistant", Model: resolvedModel}, nil
}

// createLLMClient creates an LLM client for the given provider
func createLLMClient(providerID uint, model string) (*llm.Client, string, error) {
	provider, err := repository.GetProviderByIDDAO(providerID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get provider: %w", err)
	}

	if provider.APIKey == "" {
		return nil, "", errors.New("provider API key is not configured")
	}

	var providerType llm.ProviderType
	switch provider.Type {
	case 1:
		providerType = llm.ProviderGemini
	case 2:
		providerType = llm.ProviderOpenAI
	case 3:
		providerType = llm.ProviderOther
	default:
		providerType = llm.ProviderGemini
	}

	client, err := llm.NewClient(llm.Config{
		APIKey:  provider.APIKey,
		BaseURL: provider.URL,
		Type:    providerType,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Use provider's default model if not specified
	resolvedModel := model
	if resolvedModel == "" {
		resolvedModel = provider.DefaultModel
	}
	if resolvedModel == "" && provider.Type == 1 {
		resolvedModel = "gemini-2.5-flash"
	}

	return client, resolvedModel, nil
}

// ConsultServ consults the AI provider for general analysis (legacy support)
func ConsultServ(req ChatReq) (ChatRes, error) {
	client, resolvedModel, err := createLLMClient(req.ProviderID, req.Model)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(req.ProviderID, 2)
		return ChatRes{}, err
	}

	ctx := context.Background()
	responseText, err := client.SimpleChat(ctx, resolvedModel, req.Content)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(req.ProviderID, 2)
		return ChatRes{}, fmt.Errorf("failed to get AI response: %w", err)
	}

	return ChatRes{Content: responseText, ProviderID: req.ProviderID, Role: "assistant", Model: resolvedModel}, nil
}

// AnalyzeMonitoringDataServ analyzes monitoring data using LLM
func AnalyzeMonitoringDataServ(providerID uint, model string, data string) (ChatRes, error) {
	client, resolvedModel, err := createLLMClient(providerID, model)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(providerID, 2)
		return ChatRes{}, err
	}

	ctx, cancel := aiAnalysisContext()
	defer cancel()
	systemPrompt := monitoringAnalysisPrompt()

	start := time.Now()
	resp, err := client.Chat(ctx, llm.ChatRequest{
		Model:        resolvedModel,
		SystemPrompt: systemPrompt,
		Messages: []llm.Message{
			{Role: "user", Content: sanitizeSensitiveText(data)},
		},
	})
	logLLMRequest("monitoring_analysis", providerID, resolvedModel, time.Since(start), err)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(providerID, 2)
		return ChatRes{}, fmt.Errorf("failed to analyze data: %w", err)
	}

	_ = repository.UpdateProviderStatusDAO(providerID, 1)
	return ChatRes{Content: resp.Content, ProviderID: providerID, Role: "assistant", Model: resolvedModel}, nil
}

// ExplainErrorServ explains an error message using LLM
func ExplainErrorServ(providerID uint, model string, errorMsg string) (ChatRes, error) {
	client, resolvedModel, err := createLLMClient(providerID, model)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(providerID, 2)
		return ChatRes{}, err
	}

	ctx := context.Background()
	systemPrompt := errorExplainPrompt()

	start := time.Now()
	resp, err := client.Chat(ctx, llm.ChatRequest{
		Model:        resolvedModel,
		SystemPrompt: systemPrompt,
		Messages: []llm.Message{
			{Role: "user", Content: fmt.Sprintf("Please explain this error and how to fix it:\n\n%s", sanitizeSensitiveText(errorMsg))},
		},
	})
	logLLMRequest("error_explain", providerID, resolvedModel, time.Since(start), err)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(providerID, 2)
		return ChatRes{}, fmt.Errorf("failed to explain error: %w", err)
	}

	_ = repository.UpdateProviderStatusDAO(providerID, 1)
	return ChatRes{Content: resp.Content, ProviderID: providerID, Role: "assistant", Model: resolvedModel}, nil
}

func itemAnalysisPrompt() string {
	return "You are an expert system administrator and DevOps engineer.\n" +
		"Analyze the monitoring item data and respond concisely.\n\n" +
		"Rules:\n" +
		"- Use only the provided data; do not invent thresholds.\n" +
		"- If a baseline is missing, say so and provide a cautious assessment.\n\n" +
		"Output format (use headings):\n" +
		"Metric Summary:\n" +
		"- What the metric represents and its current value.\n\n" +
		"Assessment:\n" +
		"- Normal/Concerning/Critical with brief reasoning.\n\n" +
		"Potential Impact:\n" +
		"- Risks if the value is abnormal.\n\n" +
		"Recommended Actions:\n" +
		"- Immediate steps and follow-ups."
}

func hostAnalysisPrompt() string {
	return "You are an expert system administrator and DevOps engineer.\n" +
		"Analyze the host monitoring data and summarize health.\n\n" +
		"Rules:\n" +
		"- Use only the provided data; do not invent metrics.\n" +
		"- Highlight the most critical issues first.\n\n" +
		"Output format (use headings):\n" +
		"Health Status:\n" +
		"- Healthy/Warning/Critical with brief justification.\n\n" +
		"Key Findings:\n" +
		"- Bullet list of notable metrics.\n\n" +
		"Risks:\n" +
		"- Potential issues if the current state persists.\n\n" +
		"Recommended Actions:\n" +
		"- Immediate steps and follow-ups."
}

func monitoringAnalysisPrompt() string {
	return "You are an expert system administrator and DevOps engineer.\n" +
		"Analyze the monitoring data and produce a clear, actionable assessment.\n\n" +
		"Rules:\n" +
		"- Use only the provided data; do not invent metrics or events.\n" +
		"- If data is missing or ambiguous, say what is missing and how it limits confidence.\n\n" +
		"Output format (use headings):\n" +
		"State Summary:\n" +
		"- Current health in 1-3 sentences.\n\n" +
		"Detected Issues:\n" +
		"- List anomalies with evidence (metric, value, time window).\n" +
		"- If none, say \"No anomalies detected\".\n\n" +
		"Severity:\n" +
		"- Critical/Warning/Normal with brief justification.\n\n" +
		"Recommended Actions:\n" +
		"- Immediate actions (if any), then short-term improvements.\n\n" +
		"Assumptions:\n" +
		"- List any assumptions or unknowns."
}

func networkStatusPrompt() string {
	return "You are an expert network operations analyst.\n" +
		"Use the provided alerts, health score, and metrics to assess current network status.\n\n" +
		"Rules:\n" +
		"- Use only the provided data; do not invent alerts or metrics.\n" +
		"- If no alerts are present, explicitly say so and rely on metrics/health score.\n\n" +
		"Output format (use headings):\n" +
		"Network Status Summary:\n" +
		"- 1-3 sentence overview.\n\n" +
		"Active Alerts:\n" +
		"- Count and top items (if any).\n\n" +
		"Key Metrics:\n" +
		"- Highlight notable metrics.\n\n" +
		"Recommendations:\n" +
		"- Immediate steps and follow-ups.\n\n" +
		"Confidence:\n" +
		"- Note any missing data or limits."
}

func buildNetworkStatusContext(alerts []AlertRes, metrics []MetricSnapshot, health HealthScore) string {
	var builder strings.Builder
	builder.WriteString("Health Score: ")
	builder.WriteString(fmt.Sprintf("%d (monitors %d/%d, hosts %d/%d, items %d/%d)\n\n",
		health.Score,
		health.MonitorActive, health.MonitorTotal,
		health.HostActive, health.HostTotal,
		health.ItemActive, health.ItemTotal,
	))

	builder.WriteString("Active Alerts:\n")
	if len(alerts) == 0 {
		builder.WriteString("- None\n")
	} else {
		max := len(alerts)
		if max > 10 {
			max = 10
		}
		for i := 0; i < max; i++ {
			alert := alerts[i]
			builder.WriteString(fmt.Sprintf("- #%d severity=%d host_id=%d item_id=%d message=%s\n",
				alert.ID,
				alert.Severity,
				alert.HostID,
				alert.ItemID,
				sanitizeSensitiveText(alert.Message),
			))
		}
		if len(alerts) > max {
			builder.WriteString(fmt.Sprintf("- ...and %d more\n", len(alerts)-max))
		}
	}

	builder.WriteString("\nNetwork Metrics:\n")
	if len(metrics) == 0 {
		builder.WriteString("- None\n")
	} else {
		max := len(metrics)
		if max > 20 {
			max = 20
		}
		for i := 0; i < max; i++ {
			metric := metrics[i]
			builder.WriteString(fmt.Sprintf("- host=%s item=%s value=%s %s status=%d updated=%s\n",
				sanitizeSensitiveText(metric.HostName),
				sanitizeSensitiveText(metric.ItemName),
				sanitizeSensitiveText(metric.Value),
				sanitizeSensitiveText(metric.Units),
				metric.Status,
				metric.UpdatedAt.Format(time.RFC3339),
			))
		}
		if len(metrics) > max {
			builder.WriteString(fmt.Sprintf("- ...and %d more\n", len(metrics)-max))
		}
	}

	return builder.String()
}

func isNetworkStatusQuery(content string) bool {
	lower := strings.ToLower(strings.TrimSpace(content))
	if lower == "" {
		return false
	}
	// Check English keywords
	keywords := []string{"network", "net", "status", "health", "alert"}
	for _, k := range keywords {
		if strings.Contains(lower, k) {
			return true
		}
	}
	// Check Chinese keywords
	cnKeywords := []string{"网络", "状态", "状况", "健康", "告警"}
	for _, k := range cnKeywords {
		if strings.Contains(content, k) {
			return true
		}
	}
	return false
}

func errorExplainPrompt() string {
	return "You are a helpful technical assistant.\n" +
		"When given an error message:\n" +
		"1. Explain what the error means in simple terms\n" +
		"2. Identify the most likely causes\n" +
		"3. Provide step-by-step solutions to fix the issue\n" +
		"4. Mention any preventive measures for the future\n\n" +
		"Be practical and clear in your explanations."
}

func logLLMRequest(operation string, providerID uint, model string, duration time.Duration, err error) {
	level := "info"
	context := map[string]interface{}{
		"operation":   operation,
		"provider_id": providerID,
		"model":       model,
		"duration_ms": duration.Milliseconds(),
	}
	if err != nil {
		level = "error"
		context["error"] = err.Error()
	}
	LogService(level, "llm request", context, nil, "")
}
