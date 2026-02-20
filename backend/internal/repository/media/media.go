package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Provider defines a media delivery provider
type Provider interface {
	SendMessage(ctx context.Context, target, message string) error
}

// Service manages multiple media providers
type Service struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

// NewService creates a new media service
func NewService() *Service {
	return &Service{providers: make(map[string]Provider)}
}

var globalService *Service
var serviceOnce sync.Once

// GetService returns the global media service instance
func GetService() *Service {
	serviceOnce.Do(func() {
		globalService = NewService()
		globalService.RegisterProvider("qq", NewQQProvider(DefaultQQBaseURL))
		globalService.RegisterProvider("webhook", NewWebhookProvider())
		globalService.RegisterProvider("wechat", NewWebhookProvider())
		globalService.RegisterProvider("gmail", NewGmailProvider())
	})
	return globalService
}

// RegisterProvider registers a media provider
func (s *Service) RegisterProvider(key string, provider Provider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providers[key] = provider
}

// GetProvider returns a provider by key
func (s *Service) GetProvider(key string) (Provider, error) {
	s.mu.RLock()
	provider, ok := s.providers[key]
	s.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("media provider not found: %s", key)
	}
	return provider, nil
}

// ListProviders returns all provider keys
func (s *Service) ListProviders() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.providers))
	for k := range s.providers {
		keys = append(keys, k)
	}
	return keys
}

// RemoveProvider removes a provider by key
func (s *Service) RemoveProvider(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.providers, key)
}

// SendMessage sends a message via the specified media type
func (s *Service) SendMessage(ctx context.Context, mediaType, target, message string) error {
	provider, err := s.GetProvider(mediaType)
	if err != nil {
		return err
	}
	return provider.SendMessage(ctx, target, message)
}

// NoopProvider is a placeholder provider that does nothing
type NoopProvider string

func (p NoopProvider) SendMessage(ctx context.Context, target, message string) error {
	return nil
}

// WebhookProvider posts a JSON payload to the target URL
type WebhookProvider struct {
	Client *http.Client
}

// NewWebhookProvider creates a webhook provider
func NewWebhookProvider() *WebhookProvider {
	return &WebhookProvider{Client: &http.Client{Timeout: 5 * time.Second}}
}

// SendMessage sends a webhook payload
func (p *WebhookProvider) SendMessage(ctx context.Context, target, message string) error {
	if target == "" {
		return fmt.Errorf("webhook target is empty")
	}
	payload := map[string]string{"message": message}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook status %d", resp.StatusCode)
	}
	return nil
}
