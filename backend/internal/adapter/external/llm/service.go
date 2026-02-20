package llm

import (
	"context"
	"fmt"
	"sync"
)

// Service provides a high-level interface for LLM operations
type Service struct {
	clients map[int]*Client // map of provider ID to client
	mu      sync.RWMutex
}

// NewService creates a new LLM service
func NewService() *Service {
	return &Service{
		clients: make(map[int]*Client),
	}
}

// global service instance
var globalService *Service
var serviceOnce sync.Once

// GetService returns the global LLM service instance
func GetService() *Service {
	serviceOnce.Do(func() {
		globalService = NewService()
	})
	return globalService
}

// RegisterProvider registers a provider with the service
func (s *Service) RegisterProvider(providerID int, cfg Config) error {
	client, err := NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	s.mu.Lock()
	s.clients[providerID] = client
	s.mu.Unlock()

	return nil
}

// GetClient returns the client for a provider
func (s *Service) GetClient(providerID int) (*Client, error) {
	s.mu.RLock()
	client, ok := s.clients[providerID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("provider %d not registered", providerID)
	}

	return client, nil
}

// RemoveProvider removes a provider from the service
func (s *Service) RemoveProvider(providerID int) {
	s.mu.Lock()
	delete(s.clients, providerID)
	s.mu.Unlock()
}

// Chat sends a chat request using the specified provider
func (s *Service) Chat(ctx context.Context, providerID int, req ChatRequest) (*ChatResponse, error) {
	client, err := s.GetClient(providerID)
	if err != nil {
		return nil, err
	}

	return client.Chat(ctx, req)
}

// SimpleChat sends a simple prompt using the specified provider
func (s *Service) SimpleChat(ctx context.Context, providerID int, model, prompt string) (string, error) {
	client, err := s.GetClient(providerID)
	if err != nil {
		return "", err
	}

	return client.SimpleChat(ctx, model, prompt)
}

// AnalyzeMonitoringData sends monitoring data to LLM for analysis
func (s *Service) AnalyzeMonitoringData(ctx context.Context, providerID int, model string, data string) (*ChatResponse, error) {
	systemPrompt := `You are an expert system administrator and DevOps engineer.
Analyze the monitoring data and produce a clear, actionable assessment.

Rules:
- Use only the provided data; do not invent metrics or events.
- If data is missing or ambiguous, say what is missing and how it limits confidence.
- Prefer direct, operational language and concrete next steps.

Output format (use headings):
State Summary:
- Current health in 1-3 sentences.

Detected Issues:
- List anomalies with evidence (metric, value, threshold, time window).
- If none, say "No anomalies detected".

Severity:
- One of: Critical, Warning, Normal.
- Brief justification.

Recommended Actions:
- Immediate actions (if any).
- Short-term improvements.

Assumptions:
- List any assumptions or unknowns.

Keep it concise but complete.`

	req := ChatRequest{
		Model:        model,
		SystemPrompt: systemPrompt,
		Messages: []Message{
			{Role: "user", Content: data},
		},
	}

	return s.Chat(ctx, providerID, req)
}

// ExplainError sends an error message to LLM for explanation
func (s *Service) ExplainError(ctx context.Context, providerID int, model string, errorMsg string) (*ChatResponse, error) {
	systemPrompt := `You are a helpful technical assistant. 
When given an error message:
1. Explain what the error means in simple terms
2. Identify the most likely causes
3. Provide step-by-step solutions to fix the issue
4. Mention any preventive measures for the future

Be practical and clear in your explanations.`

	req := ChatRequest{
		Model:        model,
		SystemPrompt: systemPrompt,
		Messages: []Message{
			{Role: "user", Content: fmt.Sprintf("Please explain this error and how to fix it:\n\n%s", errorMsg)},
		},
	}

	return s.Chat(ctx, providerID, req)
}

// GenerateReport generates a report based on monitoring data
func (s *Service) GenerateReport(ctx context.Context, providerID int, model string, data string, reportType string) (*ChatResponse, error) {
	systemPrompt := fmt.Sprintf(`You are a technical report writer specializing in IT infrastructure monitoring.
Generate a %s report based on the provided monitoring data.
The report should be well-structured, professional, and actionable.
Include relevant metrics, trends, and recommendations.`, reportType)

	req := ChatRequest{
		Model:        model,
		SystemPrompt: systemPrompt,
		Messages: []Message{
			{Role: "user", Content: data},
		},
	}

	return s.Chat(ctx, providerID, req)
}
