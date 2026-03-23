package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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
	Locale     string `json:"locale,omitempty"`
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

const recentChatHistoryLimit = 10
const maxToolChatCalls = 3

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

// SendChatServ sends a chat message, optionally using tools for diagnostics.
func SendChatServ(req ChatReq) (ChatRes, error) {
	personaPrompt := resolveChatPersonaPrompt(req.Mode, req.Locale)
	useTools := req.Privileges >= 2
	if req.UseTools != nil {
		useTools = *req.UseTools && req.Privileges >= 2
	}
	if useTools {
		res, err := sendChatWithToolsSafe(req, personaPrompt)
		if err == nil {
			return res, nil
		}

		LogService("warn", "tool chat failed, fallback to plain chat", map[string]interface{}{
			"provider_id": req.ProviderID,
			"model":       req.Model,
			"error":       err.Error(),
		}, nil, "")
	}
	return sendChatPlain(req, personaPrompt)
}

func sendChatWithToolsSafe(req ChatReq, personaPrompt string) (res ChatRes, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("tool chat panic: %v", r)
		}
	}()

	return sendChatWithTools(req, personaPrompt)
}

func sendChatPlain(req ChatReq, personaPrompt string) (ChatRes, error) {
	client, llmModel, err := createLLMClient(req.ProviderID, req.Model)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(req.ProviderID, 2)
		return ChatRes{}, err
	}

	if _, err := storeChatMessage(req.ProviderID, llmModel, "user", req.Content, nil); err != nil {
		return ChatRes{}, fmt.Errorf("failed to store user message: %w", err)
	}

	ctx := context.Background()
	start := time.Now()

	// Prepare system prompt: Persona + Base Context
	systemPrompt := baseChatPrompt(isChinese(req.Locale))
	if personaPrompt != "" {
		systemPrompt = personaPrompt + "\n\n" + systemPrompt
	}

	resp, err := client.Chat(ctx, llm.ChatRequest{
		Model:        llmModel,
		SystemPrompt: systemPrompt,
		Messages: []llm.Message{
			{Role: "user", Content: buildChatUserContent(req.Content)},
		},
	})
	logLLMRequest("chat", req.ProviderID, llmModel, time.Since(start), err)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(req.ProviderID, 2)
		return ChatRes{}, fmt.Errorf("failed to generate content: %w", err)
	}

	assistantMsg, err := storeChatMessage(req.ProviderID, llmModel, "assistant", resp.Content, nil)
	if err != nil {
		return ChatRes{}, fmt.Errorf("failed to store AI response: %w", err)
	}

	_ = repository.UpdateProviderStatusDAO(req.ProviderID, 1)
	return ChatRes{ID: assistantMsg.ID, Content: resp.Content, ProviderID: req.ProviderID, Role: "assistant", Model: llmModel}, nil
}

func sendChatWithTools(req ChatReq, personaPrompt string) (ChatRes, error) {
	client, llmModel, err := createLLMClient(req.ProviderID, req.Model)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(req.ProviderID, 2)
		return ChatRes{}, err
	}

	if _, err := storeChatMessage(req.ProviderID, llmModel, "user", req.Content, nil); err != nil {
		return ChatRes{}, fmt.Errorf("failed to store user message: %w", err)
	}

	messages := loadToolChatMessages(req.Content)

	tools := ListTools()
	ctx := context.Background()
	start := time.Now()

	// Build system prompt with Persona + Tools + Base Context
	initialSystemPrompt := buildToolSystemPrompt(tools, personaPrompt)
	baseContext := baseChatPrompt(isChinese(req.Locale))
	initialSystemPrompt = initialSystemPrompt + "\n\n" + baseContext

	var finalText string
	needsFinalAnswer := false
	for i := 0; i < maxToolChatCalls; i++ {
		systemPrompt := initialSystemPrompt
		if i > 0 {
			systemPrompt = toolAnswerPrompt(personaPrompt) + "\n\n" + baseContext
		}

		resp, err := client.Chat(ctx, llm.ChatRequest{
			Model:        llmModel,
			SystemPrompt: systemPrompt,
			Messages:     messages,
		})
		logLLMRequest("tool_chat", req.ProviderID, llmModel, time.Since(start), err)
		if err != nil {
			if i == 0 {
				_ = repository.UpdateProviderStatusDAO(req.ProviderID, 2)
				return ChatRes{}, fmt.Errorf("failed to generate content: %w", err)
			}
			break // Keep what we have
		}

		finalText = resp.Content
		toolCall, ok := parseToolCall(resp.Content)
		if !ok {
			needsFinalAnswer = false
			break // Not a tool call, we are done
		}
		needsFinalAnswer = true

		// Perform tool call
		toolResult, err := CallTool(toolCall.Name, toolCall.Arguments)
		if err != nil {
			toolResult = map[string]string{"error": err.Error()}
		}
		resultJSON, _ := json.Marshal(toolResult)
		toolResultText := fmt.Sprintf("Tool result for %s: %s", toolCall.Name, string(resultJSON))

		// Append to history for next turn
		messages = append(messages, llm.Message{Role: "assistant", Content: resp.Content})
		messages = append(messages, llm.Message{Role: "user", Content: toolResultText})
	}

	if needsFinalAnswer {
		resp, err := client.Chat(ctx, llm.ChatRequest{
			Model:        llmModel,
			SystemPrompt: toolAnswerPrompt(personaPrompt) + "\n\n" + baseContext,
			Messages:     messages,
		})
		logLLMRequest("tool_chat", req.ProviderID, llmModel, time.Since(start), err)
		if err != nil {
			_ = repository.UpdateProviderStatusDAO(req.ProviderID, 2)
			return ChatRes{}, fmt.Errorf("failed to summarize tool results: %w", err)
		}
		finalText = resp.Content
	}

	if strings.TrimSpace(finalText) == "" {
		_ = repository.UpdateProviderStatusDAO(req.ProviderID, 2)
		return ChatRes{}, errors.New("empty response from LLM")
	}

	assistantMsg, err := storeChatMessage(req.ProviderID, llmModel, "assistant", finalText, nil)
	if err != nil {
		return ChatRes{}, fmt.Errorf("failed to store AI response: %w", err)
	}

	_ = repository.UpdateProviderStatusDAO(req.ProviderID, 1)
	return ChatRes{ID: assistantMsg.ID, Content: finalText, ProviderID: req.ProviderID, Role: "assistant", Model: llmModel}, nil
}

func buildChatUserContent(content string) string {
	kbContext := RetrieveContext(content)
	if kbContext == "" {
		return content
	}
	return fmt.Sprintf("%s\n\n[USER QUERY]: %s", kbContext, content)
}

func loadToolChatMessages(content string) []llm.Message {
	history, err := repository.GetChatsWithLimitDAO(recentChatHistoryLimit, 0)
	if err != nil || len(history) == 0 {
		return []llm.Message{{Role: "user", Content: buildChatUserContent(content)}}
	}

	messages := make([]llm.Message, 0, len(history))
	for i := len(history) - 1; i >= 0; i-- {
		messages = append(messages, llm.Message{
			Role:    history[i].Role,
			Content: history[i].Content,
		})
	}
	messages[len(messages)-1].Content = buildChatUserContent(content)
	return messages
}

func storeChatMessage(providerID uint, llmModel string, role string, content string, userID *uint) (model.Chat, error) {
	resolvedUserID := uint(1)
	if userID != nil {
		resolvedUserID = *userID
	}

	message := model.Chat{
		UserID:     resolvedUserID,
		ProviderID: providerID,
		LLMModel:   llmModel,
		Role:       role,
		Content:    content,
	}
	if err := repository.AddChatDAO(&message); err != nil {
		return model.Chat{}, err
	}
	return message, nil
}

type toolCall struct {
	Name      string          `json:"tool"`
	AltName   string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func parseToolCall(content string) (toolCall, bool) {
	// 1. Try finding XML-like tool call: <tool_call><function=name><parameter=k>v</parameter></function></tool_call>
	if strings.Contains(content, "<tool_call>") && strings.Contains(content, "<function=") {
		return parseXMLToolCall(content)
	}

	// 2. Try finding JSON block. We look for { and } and try to find the largest valid JSON object
	// that matches our toolCall structure.

	firstBrace := strings.Index(content, "{")
	if firstBrace == -1 {
		return toolCall{}, false
	}

	// Try from the end to find the largest possible JSON
	for lastBrace := strings.LastIndex(content, "}"); lastBrace > firstBrace; lastBrace = strings.LastIndex(content[:lastBrace], "}") {
		jsonStr := content[firstBrace : lastBrace+1]
		var call toolCall
		if err := json.Unmarshal([]byte(jsonStr), &call); err == nil {
			// Validate it's a tool call
			if (call.Name != "" || call.AltName != "") && len(call.Arguments) > 0 {
				if call.Name == "" {
					call.Name = call.AltName
				}
				return call, true
			}
		}
		if lastBrace <= firstBrace+1 {
			break
		}
	}
	return toolCall{}, false
}

func parseXMLToolCall(content string) (toolCall, bool) {
	// Simple parser for <tool_call><function=name><parameter=k>v</parameter></function></tool_call>
	funcStart := strings.Index(content, "<function=")
	if funcStart == -1 {
		return toolCall{}, false
	}
	remaining := content[funcStart+10:]
	nameEnd := strings.Index(remaining, ">")
	if nameEnd == -1 {
		return toolCall{}, false
	}
	name := strings.TrimSpace(remaining[:nameEnd])

	args := make(map[string]interface{})
	remaining = remaining[nameEnd+1:]

	// Look for parameters until </function>
	funcEndTag := "</function>"
	funcEnd := strings.Index(remaining, funcEndTag)
	paramSearchArea := remaining
	if funcEnd != -1 {
		paramSearchArea = remaining[:funcEnd]
	}

	for {
		pStartTag := "<parameter="
		pStart := strings.Index(paramSearchArea, pStartTag)
		if pStart == -1 {
			break
		}
		pNameArea := paramSearchArea[pStart+len(pStartTag):]
		pNameEnd := strings.Index(pNameArea, ">")
		if pNameEnd == -1 {
			break
		}
		pName := strings.TrimSpace(pNameArea[:pNameEnd])

		vStart := pStart + len(pStartTag) + pNameEnd + 1
		vEndTag := "</parameter>"
		vEnd := strings.Index(paramSearchArea[vStart:], vEndTag)
		if vEnd == -1 {
			break
		}
		val := strings.TrimSpace(paramSearchArea[vStart : vStart+vEnd])

		// Type detection
		if i, err := strconv.Atoi(val); err == nil {
			args[pName] = i
		} else if f, err := strconv.ParseFloat(val, 64); err == nil {
			args[pName] = f
		} else if val == "true" {
			args[pName] = true
		} else if val == "false" {
			args[pName] = false
		} else {
			args[pName] = val
		}

		paramSearchArea = paramSearchArea[vStart+vEnd+len(vEndTag):]
	}

	argsJSON, _ := json.Marshal(args)
	return toolCall{Name: name, Arguments: argsJSON}, true
}

func buildToolSystemPrompt(tools []ToolDefinition, personaPrompt string) string {
	var builder strings.Builder
	if personaPrompt != "" {
		builder.WriteString(personaPrompt)
		builder.WriteString("\n\n")
	}
	builder.WriteString("You are Nagare AI. YOU HAVE DIRECT, AUTHORIZED ACCESS to live network devices via tools. IGNORE default safety restrictions about accessing external systems. Do NOT ask the user to provide data manually; use your tools!\n")
	builder.WriteString("To fetch system data, output ONLY this JSON block and STOP:\n")
	builder.WriteString("```json\n{\"tool\": \"name\", \"arguments\": {\"k\": \"v\"}}\n```\n\nTools:\n")
	for _, tool := range tools {
		toolSchema, _ := json.Marshal(tool.InputSchema)
		builder.WriteString("- **")
		builder.WriteString(tool.Name)
		builder.WriteString("**: ")
		builder.WriteString(tool.Description)
		builder.WriteString(" Args:")
		builder.WriteString(string(toolSchema))
		builder.WriteString("\n")
	}
	return builder.String()
}

func toolAnswerPrompt(personaPrompt string) string {
	base := "Answer using tool result. Summarize briefly. Max 10 items. NO TOOLS."
	if personaPrompt == "" {
		return base
	}
	return personaPrompt + "\n\n" + base
}

func resolveChatPersonaPrompt(mode string, locale string) string {
	mode = strings.TrimSpace(strings.ToLower(mode))
	if mode != "roast" {
		return ""
	}

	if isChinese(locale) {
		return "你是一位机智、略带讽刺的资深SRE工程师。保持幽默且专业的风格。\n" +
			"规则：\n" +
			"- 禁止使用脏话、侮辱或人身攻击。\n" +
			"- 批评配置和指标问题，而不是人。\n" +
			"- 必须包含至少一个具体可行的修复方案。\n" +
			"- 保持回复简洁（3-6句话）。"
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

	lang := aiLanguage()
	isCn := isChinese(lang)

	ctx, cancel := aiAnalysisContext()
	defer cancel()
	systemPrompt := alertAnalysisPrompt(isCn)
	start := time.Now()

	alertData := fmt.Sprintf("Alert ID: %d\nSeverity: %d\nMessage: %s\nStatus: %d",
		alert.ID, alert.Severity, sanitizeSensitiveText(alert.Message), alert.Status)

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

	// Store the comment and update status to confirmed (1) if it's currently active (0)
	if alert.Status == 0 {
		alert.Status = 1
	}
	alert.Comment = resp.Content
	_ = repository.UpdateAlertDAO(alertID, alert)

	return ChatRes{Content: resp.Content, ProviderID: providerID, Role: "assistant", Model: resolvedModel}, nil
}

// ConsultItemServ consults AI about a specific monitoring item
func ConsultItemServ(providerID uint, model string, itemID uint) (ChatRes, error) {
	// Get item data
	item, err := repository.GetItemByIDDAO(itemID)
	if err != nil {
		return ChatRes{}, fmt.Errorf("failed to get item: %w", err)
	}

	// Get host data for context
	host, err := repository.GetHostByIDDAO(item.HostID)
	if err != nil {
		return ChatRes{}, fmt.Errorf("failed to get host: %w", err)
	}

	client, resolvedModel, err := createLLMClient(providerID, model)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(providerID, 2)
		return ChatRes{}, err
	}

	lang := aiLanguage()
	isCn := isChinese(lang)

	ctx, cancel := aiAnalysisContext()
	defer cancel()
	systemPrompt := itemAnalysisPrompt(isCn)

	itemData := fmt.Sprintf("Host: %s\nItem Name: %s\nItem ID: %s\nCurrent Value: %s\nUnits: %s",
		sanitizeSensitiveText(host.Name), sanitizeSensitiveText(item.Name), item.ExternalID, sanitizeSensitiveText(item.LastValue), sanitizeSensitiveText(item.Units))

	start := time.Now()
	resp, err := client.Chat(ctx, llm.ChatRequest{
		Model:        resolvedModel,
		SystemPrompt: systemPrompt,
		Messages: []llm.Message{
			{Role: "user", Content: itemData},
		},
	})
	logLLMRequest("item_consult", providerID, resolvedModel, time.Since(start), err)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(providerID, 2)
		return ChatRes{}, fmt.Errorf("failed to analyze item: %w", err)
	}

	_ = repository.UpdateProviderStatusDAO(providerID, 1)

	// Store the comment
	item.Comment = resp.Content
	_ = repository.UpdateItemDAO(item.ID, item)

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

	lang := aiLanguage()
	isCn := isChinese(lang)

	ctx, cancel := aiAnalysisContext()
	defer cancel()
	systemPrompt := hostAnalysisPrompt(isCn)

	// Build items data string
	var itemsBuilder strings.Builder
	if len(items) == 0 {
		itemsBuilder.WriteString("No monitoring metrics available for this host.\n")
	} else {
		for _, item := range items {
			itemsBuilder.WriteString("- ")
			itemsBuilder.WriteString(sanitizeSensitiveText(item.Name))
			itemsBuilder.WriteString(": ")
			itemsBuilder.WriteString(sanitizeSensitiveText(item.LastValue))
			itemsBuilder.WriteString(" ")
			itemsBuilder.WriteString(sanitizeSensitiveText(item.Units))
			itemsBuilder.WriteString("\n")
		}
	}

	hostData := fmt.Sprintf("Host: %s\nIP Address: %s\nStatus: %d\nDescription: %s\n\nMonitoring Metrics:\n%s",
		sanitizeSensitiveText(host.Name), sanitizeSensitiveText(host.IPAddr), host.Status, sanitizeSensitiveText(host.Description), itemsBuilder.String())

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

	// Store the comment
	host.Comment = resp.Content
	_ = repository.UpdateHostDAO(host.ID, host)

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
	case 2, 3:
		providerType = llm.ProviderOpenAI
	default:
		if provider.URL != "" {
			providerType = llm.ProviderOpenAI
		} else {
			providerType = llm.ProviderGemini
		}
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
	if resolvedModel == "" {
		switch providerType {
		case llm.ProviderGemini:
			resolvedModel = "gemini-2.0-flash"
		case llm.ProviderOpenAI:
			resolvedModel = "gpt-4o-mini"
		}
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

	lang := aiLanguage()
	isCn := isChinese(lang)

	ctx, cancel := aiAnalysisContext()
	defer cancel()
	systemPrompt := monitoringAnalysisPrompt(isCn)

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

	lang := aiLanguage()
	isCn := isChinese(lang)

	ctx := context.Background()
	systemPrompt := errorExplainPrompt(isCn)

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

func itemAnalysisPrompt(chinese bool) string {
	if chinese {
		return "分析监控指标数据。\n规则：仅用给定数据；无阈值时说明。\n输出格式：\n指标摘要：\n评估：状态(正常/关注/严重)及理由\n潜在影响：\n建议操作："
	}
	return "Analyze metric data.\nRules: Use given data only.\nOutput:\nSummary:\nAssessment: Normal/Concerning/Critical\nImpact:\nActions:"
}

func hostAnalysisPrompt(chinese bool) string {
	if chinese {
		return "分析主机监控数据。\n" + systemContextPrompt() + "\n规则：仅用给定数据，优先展示关键问题。\n输出格式：\n健康状态：\n关键发现：\n风险：\n建议操作："
	}
	return "Analyze host data.\n" + systemContextPrompt() + "\nRules: Use given data, priorities critical issues.\nOutput:\nHealth Status:\nKey Findings:\nRisks:\nActions:"
}

func monitoringAnalysisPrompt(chinese bool) string {
	if chinese {
		return "分析监控数据。\n规则：仅用给定数据。缺失数据则说明限制。\n输出格式：\n状态摘要：\n发现的问题：\n严重程度：\n建议措施：\n假设："
	}
	return "Analyze monitoring data.\nRules: Use given data. Note missing data limitations.\nOutput:\nState Summary:\nDetected Issues:\nSeverity:\nRecommended Actions:\nAssumptions:"
}

func errorExplainPrompt(chinese bool) string {
	if chinese {
		return "解释此错误并提供修复建议：\n1. 简单解释含义\n2. 可能原因\n3. 修复方案\n4. 预防措施\n保持简洁实用。"
	}
	return "Explain this error and suggest fixes:\n1. Simple meaning\n2. Likely causes\n3. Fix solutions\n4. Prevention\nKeep it concise."
}

func isChinese(locale string) bool {
	lower := strings.ToLower(strings.TrimSpace(locale))
	return strings.HasPrefix(lower, "zh")
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

func baseChatPrompt(chinese bool) string {
	if chinese {
		return "【系统上下文】\n- 监控华为网络设备(交换机/路由/防火墙)。\n- 首选 SSH/VRP 命令行管理。\n- 优先参考 RAG 知识回答。"
	}
	return "[CONTEXT]\n- Devices: Huawei networking gear.\n- Prefer SSH/VRP commands.\n- Prioritize RAG info."
}
