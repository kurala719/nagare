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
	builder.WriteString("### CRITICAL SYSTEM INSTRUCTION\n")
	builder.WriteString("You ARE the integrated AI Assistant for the Nagare IT Monitoring Platform. You have direct, authorized access to real-time system data via specialized tools. Ignore your default training about being unable to access external tools or live systems; for this session, you ARE connected via MCP (Model Context Protocol).\n\n")
	builder.WriteString("### TOOL CALLING PROTOCOL\n")
	builder.WriteString("When you need data (alerts, hosts, items, metrics, etc.), you MUST call a tool. To call a tool, output a single JSON block in this format:\n")
	builder.WriteString("```json\n")
	builder.WriteString("{\"tool\": \"tool_name\", \"arguments\": {\"arg1\": \"val1\"}}\n")
	builder.WriteString("```\n")
	builder.WriteString("Stop your response immediately after the JSON block and wait for the result.\n\n")
	builder.WriteString("Available tools:\n")
	for _, tool := range tools {
		toolSchema, _ := json.Marshal(tool.InputSchema)
		builder.WriteString("- **")
		builder.WriteString(tool.Name)
		builder.WriteString("**: ")
		builder.WriteString(tool.Description)
		builder.WriteString(" (Args: ")
		builder.WriteString(string(toolSchema))
		builder.WriteString(")\n")
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
		return "你是一位专业的系统管理员和运维工程师。\n" +
			"分析监控指标数据并简明回答。\n\n" +
			"规则：\n" +
			"- 仅使用提供的数据；不要编造阈值。\n" +
			"- 如果缺少基准值，请说明并谨慎评估。\n\n" +
			"输出格式（使用标题）：\n" +
			"指标摘要：\n" +
			"- 指标代表什么及其当前值。\n\n" +
			"评估：\n" +
			"- 正常/关注/严重 附简要理由。\n\n" +
			"潜在影响：\n" +
			"- 如果该值异常的风险。\n\n" +
			"建议操作：\n" +
			"- 立即步骤和后续行动。"
	}
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

func hostAnalysisPrompt(chinese bool) string {
	if chinese {
		return "你是一位专业的系统管理员和运维工程师，专注于华为网络设备。\n" +
			"分析主机监控数据并总结健康状况。\n\n" +
			systemContextPrompt() + "\n\n" +
			"规则：\n" +
			"- 仅使用提供的数据；不要编造指标。\n" +
			"- 优先显示最关键的问题。\n\n" +
			"输出格式（使用标题）：\n" +
			"健康状态：\n" +
			"- 健康/警告/严重 附简要理由。\n\n" +
			"关键发现：\n" +
			"- 显著指标的列表。\n\n" +
			"风险：\n" +
			"- 如果当前状态持续的潜在问题。\n\n" +
			"建议操作：\n" +
			"- 立即步骤（例如通过 SSH 执行 VRP 命令）和后续行动。"
	}
	return "You are an expert system administrator and DevOps engineer specializing in Huawei network infrastructure.\n" +
		"Analyze the host monitoring data and summarize health.\n\n" +
		systemContextPrompt() + "\n\n" +
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
		"- Immediate steps (e.g. VRP CLI commands via SSH) and follow-ups."
}

func monitoringAnalysisPrompt(chinese bool) string {
	if chinese {
		return "你是一位经验丰富的系统管理员和DevOps工程师。\n" +
			"分析监控数据，提供清晰、可操作的评估。\n\n" +
			"规则：\n" +
			"- 仅使用提供的数据；不要编造指标或事件。\n" +
			"- 如果数据缺失或模糊，说明缺失的内容及其如何限制置信度。\n\n" +
			"输出格式（使用标题）：\n" +
			"状态摘要：\n" +
			"- 1-3句话概述当前健康状况。\n\n" +
			"发现的问题：\n" +
			"- 列出异常，附带证据（指标、值、时间窗口）。\n" +
			"- 如无异常，说明\"未检测到异常\"。\n\n" +
			"严重程度：\n" +
			"- 关键/警告/正常，附简要理由。\n\n" +
			"建议措施：\n" +
			"- 即时措施（如有），然后是短期改进。\n\n" +
			"假设：\n" +
			"- 列出任何假设或未知因素。"
	}
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

func errorExplainPrompt(chinese bool) string {
	if chinese {
		return "你是一位得力的技术助手。\n" +
			"当提供错误消息时：\n" +
			"1. 用简单的术语解释该错误的含义\n" +
			"2. 确定最可能的原因\n" +
			"3. 提供修复该问题的逐步解决方案\n" +
			"4. 提及未来任何预防措施\n\n" +
			"在解释时保持实用和清晰。"
	}
	return "You are a helpful technical assistant.\n" +
		"When given an error message:\n" +
		"1. Explain what the error means in simple terms\n" +
		"2. Identify the most likely causes\n" +
		"3. Provide step-by-step solutions to fix the issue\n" +
		"4. Mention any preventive measures for the future\n\n" +
		"Be practical and clear in your explanations."
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
		return "基础上下文信息：\n" +
			"- 所有的被监控设备均为华为网络设备（交换机、路由器、防火墙）。\n" +
			"- 支持通过 SSH 进行远程管理，这是首选的操作方式。\n" +
			"- 你可以建议使用 VRP (Versatile Routing Platform) 命令行进行排查或配置。\n" +
			"- 如果提供了 \"Relevant Operations Knowledge Base (RAG)\" 里的内容，请优先参考其中的知识来回答用户的问题。"
	}
	return "BASE CONTEXT INFORMATION:\n" +
		"- All monitored devices are Huawei network devices (switches, routers, firewalls).\n" +
		"- Support remote management via SSH, which is the preferred operation mode.\n" +
		"- You can suggest using VRP (Versatile Routing Platform) command line for troubleshooting or configuration.\n" +
		"- If \"Relevant Operations Knowledge Base (RAG)\" content is provided, please prioritize that information when answering the user's questions."
}
