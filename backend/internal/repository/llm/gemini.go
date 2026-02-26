package llm

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"google.golang.org/genai"
)

// GeminiProvider implements the Provider interface for Google Gemini
type GeminiProvider struct {
	client *genai.Client
	apiKey string
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(cfg Config) (*GeminiProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required for Gemini provider")
	}

	ctx := context.Background()
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 60
	}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.APIKey,
		Backend: genai.BackendGeminiAPI,
		HTTPClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiProvider{
		client: client,
		apiKey: cfg.APIKey,
	}, nil
}

// Chat implements the Provider interface
func (p *GeminiProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Build the prompt from messages
	var prompt string
	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			prompt += fmt.Sprintf("System: %s\n\n", msg.Content)
		case "user":
			prompt += fmt.Sprintf("User: %s\n\n", msg.Content)
		case "assistant":
			prompt += fmt.Sprintf("Assistant: %s\n\n", msg.Content)
		}
	}

	// Add system prompt if provided
	if req.SystemPrompt != "" {
		prompt = fmt.Sprintf("System: %s\n\n%s", req.SystemPrompt, prompt)
	}

	result, err := p.client.Models.GenerateContent(ctx, req.Model, genai.Text(prompt), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	return &ChatResponse{
		Content:      fmt.Sprint(result.Text()),
		Model:        req.Model,
		FinishReason: "stop",
	}, nil
}

// Name returns the provider name
func (p *GeminiProvider) Name() string {
	return "gemini"
}

// Models returns available Gemini models
func (p *GeminiProvider) Models() []string {
	return []string{
		"gemini-2.0-flash",
		"gemini-2.0-flash-lite",
		"gemini-1.5-flash",
		"gemini-1.5-pro",
		"gemini-1.0-pro",
	}
}

// FetchModels retrieves models from Gemini API
func (p *GeminiProvider) FetchModels(ctx context.Context) ([]string, error) {
	it, err := p.client.Models.List(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	var models []string
	for {
		m, err := it.Next(ctx)
		if err != nil {
			break
		}
		// m.Name typically starts with "models/"
		name := strings.TrimPrefix(m.Name, "models/")
		models = append(models, name)
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("no Gemini models found")
	}

	return models, nil
}
