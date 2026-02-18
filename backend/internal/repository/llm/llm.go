// Package llm provides a unified interface for interacting with various LLM providers.
package llm

import (
	"context"
	"errors"
	"fmt"
)

// ProviderType represents the type of LLM provider
type ProviderType int

const (
	ProviderGemini      ProviderType = iota + 1 // 1 = gemini
	ProviderOpenAI                              // 2 = openai
	ProviderOllama                              // 3 = ollama
	ProviderOtherOpenAI                         // 4 = other (openai compatible)
	ProviderOther                               // 5 = other
)

// Message represents a chat message
type Message struct {
	Role    string // "user", "assistant", "system"
	Content string
}

// ChatRequest represents a request to the LLM
type ChatRequest struct {
	Model        string
	Messages     []Message
	MaxTokens    int
	Temperature  float64
	SystemPrompt string
}

// ChatResponse represents a response from the LLM
type ChatResponse struct {
	Content      string
	Model        string
	FinishReason string
	TokensUsed   int
}

// Provider defines the interface for LLM providers
type Provider interface {
	// Chat sends a chat request and returns the response
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	// Name returns the provider name
	Name() string
	// Models returns available models for this provider
	Models() []string
}

// Config holds the configuration for an LLM provider
type Config struct {
	APIKey  string
	BaseURL string
	Type    ProviderType
	Timeout int // seconds
}

// Client is the main LLM client that wraps different providers
type Client struct {
	provider Provider
	config   Config
}

// NewClient creates a new LLM client based on the provider type
func NewClient(cfg Config) (*Client, error) {
	var provider Provider
	var err error

	switch cfg.Type {
	case ProviderGemini:
		provider, err = NewGeminiProvider(cfg)
	case ProviderOpenAI, ProviderOtherOpenAI:
		provider, err = NewOpenAIProvider(cfg)
	case ProviderOllama, ProviderOther:
		provider, err = NewOllamaProvider(cfg)
	default:
		return nil, errors.New("unsupported provider type")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return &Client{
		provider: provider,
		config:   cfg,
	}, nil
}

// Chat sends a chat request to the LLM provider
func (c *Client) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	return c.provider.Chat(ctx, req)
}

// SimpleChat sends a simple text prompt and returns the response
func (c *Client) SimpleChat(ctx context.Context, model, prompt string) (string, error) {
	req := ChatRequest{
		Model: model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	resp, err := c.Chat(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// ChatWithHistory sends a chat request with conversation history
func (c *Client) ChatWithHistory(ctx context.Context, model string, history []Message, userMessage string) (*ChatResponse, error) {
	messages := append(history, Message{Role: "user", Content: userMessage})

	req := ChatRequest{
		Model:    model,
		Messages: messages,
	}

	return c.Chat(ctx, req)
}

// ProviderName returns the name of the current provider
func (c *Client) ProviderName() string {
	return c.provider.Name()
}

// AvailableModels returns the available models for the current provider
func (c *Client) AvailableModels() []string {
	return c.provider.Models()
}
