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

// AvailableModels returns the available models for the specified provider
func (s *Service) AvailableModels(providerID int) ([]string, error) {
	client, err := s.GetClient(providerID)
	if err != nil {
		return nil, err
	}

	return client.AvailableModels(), nil
}

// FetchModels retrieves the latest list of models from the specified provider's API
func (s *Service) FetchModels(ctx context.Context, providerID int) ([]string, error) {
	client, err := s.GetClient(providerID)
	if err != nil {
		return nil, err
	}

	return client.FetchModels(ctx)
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
