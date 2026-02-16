package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaProvider implements the Provider interface for Ollama (local LLM)
type OllamaProvider struct {
	baseURL string
	client  *http.Client
}

// Ollama API request/response structures
type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Options  *ollamaOptions  `json:"options,omitempty"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
}

type ollamaResponse struct {
	Model     string        `json:"model"`
	CreatedAt string        `json:"created_at"`
	Message   ollamaMessage `json:"message"`
	Done      bool          `json:"done"`
	Error     string        `json:"error,omitempty"`
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(cfg Config) (*OllamaProvider, error) {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 120 // Ollama can be slow for large models
	}

	return &OllamaProvider{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}, nil
}

// Chat implements the Provider interface
func (p *OllamaProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Convert messages to Ollama format
	messages := make([]ollamaMessage, 0, len(req.Messages)+1)

	// Add system prompt if provided
	if req.SystemPrompt != "" {
		messages = append(messages, ollamaMessage{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}

	for _, msg := range req.Messages {
		messages = append(messages, ollamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	ollamaReq := ollamaRequest{
		Model:    req.Model,
		Messages: messages,
		Stream:   false,
	}

	if req.Temperature > 0 || req.MaxTokens > 0 {
		ollamaReq.Options = &ollamaOptions{
			Temperature: req.Temperature,
			NumPredict:  req.MaxTokens,
		}
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if ollamaResp.Error != "" {
		return nil, fmt.Errorf("Ollama API error: %s", ollamaResp.Error)
	}

	return &ChatResponse{
		Content:      ollamaResp.Message.Content,
		Model:        ollamaResp.Model,
		FinishReason: "stop",
	}, nil
}

// Name returns the provider name
func (p *OllamaProvider) Name() string {
	return "ollama"
}

// Models returns available Ollama models (common ones)
func (p *OllamaProvider) Models() []string {
	return []string{
		"llama3.2",
		"llama3.1",
		"llama3",
		"mistral",
		"mixtral",
		"codellama",
		"phi3",
		"qwen2.5",
		"deepseek-coder-v2",
	}
}

// ListModels queries Ollama for installed models
func (p *OllamaProvider) ListModels(ctx context.Context) ([]string, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var tagsResp struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]string, 0, len(tagsResp.Models))
	for _, m := range tagsResp.Models {
		models = append(models, m.Name)
	}

	return models, nil
}
