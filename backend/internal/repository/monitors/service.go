package monitors

import (
	"context"
	"fmt"
	"sync"
)

// Service manages multiple monitoring providers
type Service struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

// NewService creates a new monitoring service
func NewService() *Service {
	return &Service{
		providers: make(map[string]Provider),
	}
}

// RegisterProvider registers a new monitoring provider
func (s *Service) RegisterProvider(name string, provider Provider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providers[name] = provider
}

// GetProvider returns a provider by name
func (s *Service) GetProvider(name string) (Provider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	provider, ok := s.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return provider, nil
}

// ListProviders returns all registered provider names
func (s *Service) ListProviders() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.providers))
	for name := range s.providers {
		names = append(names, name)
	}
	return names
}

// RemoveProvider removes a provider by name
func (s *Service) RemoveProvider(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.providers, name)
}

// GetAllHosts returns hosts from all registered providers
func (s *Service) GetAllHosts(ctx context.Context) (map[string][]Host, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string][]Host)
	var errors []error

	for name, provider := range s.providers {
		hosts, err := provider.GetHosts(ctx)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", name, err))
			continue
		}
		result[name] = hosts
	}

	if len(errors) > 0 && len(result) == 0 {
		return nil, fmt.Errorf("all providers failed: %v", errors)
	}

	return result, nil
}

// GetAllAlerts returns alerts from all registered providers
func (s *Service) GetAllAlerts(ctx context.Context) (map[string][]Alert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string][]Alert)
	var errors []error

	for name, provider := range s.providers {
		alerts, err := provider.GetAlerts(ctx)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", name, err))
			continue
		}
		result[name] = alerts
	}

	if len(errors) > 0 && len(result) == 0 {
		return nil, fmt.Errorf("all providers failed: %v", errors)
	}

	return result, nil
}

// GetAllTriggers returns triggers from all registered providers
func (s *Service) GetAllTriggers(ctx context.Context) (map[string][]Trigger, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string][]Trigger)
	var errors []error

	for name, provider := range s.providers {
		triggers, err := provider.GetTriggers(ctx)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", name, err))
			continue
		}
		result[name] = triggers
	}

	if len(errors) > 0 && len(result) == 0 {
		return nil, fmt.Errorf("all providers failed: %v", errors)
	}

	return result, nil
}

// HealthCheck performs a health check on all providers
func (s *Service) HealthCheck(ctx context.Context) map[string]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]bool)

	for name, provider := range s.providers {
		// Try to authenticate as a health check
		err := provider.Authenticate(ctx)
		result[name] = err == nil
	}

	return result
}

// CreateProviderFromConfig creates a provider from configuration
func CreateProviderFromConfig(cfg Config) (Provider, error) {
	switch cfg.Type {
	case MonitorZabbix:
		return NewZabbixProvider(cfg)
	case MonitorPrometheus:
		return NewPrometheusProvider(cfg)
	case MonitorOther:
		return nil, fmt.Errorf("other provider not implemented yet")
	default:
		return nil, fmt.Errorf("unknown monitor type: %s", cfg.Type)
	}
}

// InitializeService creates and initializes a monitoring service with providers from config
func InitializeService(ctx context.Context, configs []Config) (*Service, error) {
	service := NewService()

	for _, cfg := range configs {
		provider, err := CreateProviderFromConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create provider %s: %w", cfg.Name, err)
		}

		// Authenticate the provider
		if err := provider.Authenticate(ctx); err != nil {
			return nil, fmt.Errorf("failed to authenticate provider %s: %w", cfg.Name, err)
		}

		service.RegisterProvider(cfg.Name, provider)
	}

	return service, nil
}

// MonitorStats represents statistics from monitoring providers
type MonitorStats struct {
	TotalHosts    int            `json:"total_hosts"`
	TotalAlerts   int            `json:"total_alerts"`
	TotalTriggers int            `json:"total_triggers"`
	ByProvider    map[string]int `json:"by_provider"`
}

// GetStats returns statistics from all providers
func (s *Service) GetStats(ctx context.Context) (*MonitorStats, error) {
	stats := &MonitorStats{
		ByProvider: make(map[string]int),
	}

	hosts, err := s.GetAllHosts(ctx)
	if err == nil {
		for provider, h := range hosts {
			stats.TotalHosts += len(h)
			stats.ByProvider[provider+"_hosts"] = len(h)
		}
	}

	alerts, err := s.GetAllAlerts(ctx)
	if err == nil {
		for provider, a := range alerts {
			stats.TotalAlerts += len(a)
			stats.ByProvider[provider+"_alerts"] = len(a)
		}
	}

	triggers, err := s.GetAllTriggers(ctx)
	if err == nil {
		for provider, t := range triggers {
			stats.TotalTriggers += len(t)
			stats.ByProvider[provider+"_triggers"] = len(t)
		}
	}

	return stats, nil
}
