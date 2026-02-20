package service

import (
	"fmt"

	"nagare/internal/adapter/repository"
	"nagare/internal/core/domain"
)

// ProviderReq represents a provider request
type ProviderReq struct {
	Name         string   `json:"name" binding:"required"`
	URL          string   `json:"url"`
	APIKey       string   `json:"api_key" binding:"required"`
	DefaultModel string   `json:"default_model"`
	Models       []string `json:"models"`
	Type         int      `json:"type" binding:"required,oneof=1 2 3 4 5"`
	Description  string   `json:"description"`
	Enabled      int      `json:"enabled"`
}

// ProviderRes represents a provider response
type ProviderRes struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	URL          string   `json:"url"`
	APIKey       string   `json:"api_key"`
	DefaultModel string   `json:"default_model"`
	Models       []string `json:"models"`
	Type         int      `json:"type"`
	Description  string   `json:"description"`
	Enabled      int      `json:"enabled"`
	Status       int      `json:"status"`
}

// GetAllProvidersServ retrieves all providers
func GetAllProvidersServ() ([]ProviderRes, error) {
	providers, err := repository.GetAllProvidersDAO()
	if err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}

	result := make([]ProviderRes, 0, len(providers))
	for _, p := range providers {
		result = append(result, providerToRes(p))
	}
	return result, nil
}

// SearchProvidersServ retrieves providers by filter
func SearchProvidersServ(filter domain.ProviderFilter) ([]ProviderRes, error) {
	providers, err := repository.SearchProvidersDAO(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search providers: %w", err)
	}
	result := make([]ProviderRes, 0, len(providers))
	for _, p := range providers {
		result = append(result, providerToRes(p))
	}
	return result, nil
}

// CountProvidersServ returns total count for providers by filter
func CountProvidersServ(filter domain.ProviderFilter) (int64, error) {
	return repository.CountProvidersDAO(filter)
}

// GetProviderByIDServ retrieves a provider by ID
func GetProviderByIDServ(id uint) (ProviderRes, error) {
	p, err := repository.GetProviderByIDDAO(id)
	if err != nil {
		return ProviderRes{}, fmt.Errorf("failed to get provider: %w", err)
	}
	return providerToRes(p), nil
}

// AddProviderServ creates a new provider
func AddProviderServ(req ProviderReq) error {
	return repository.AddProviderDAO(domain.Provider{
		Name:         req.Name,
		URL:          req.URL,
		APIKey:       req.APIKey,
		DefaultModel: req.DefaultModel,
		Models:       req.Models,
		Type:         req.Type,
		Description:  req.Description,
		Enabled:      req.Enabled,
		Status:       determineProviderStatus(domain.Provider{Enabled: req.Enabled, APIKey: req.APIKey}),
	})
}

// DeleteProviderByIDServ deletes a provider by ID
func DeleteProviderByIDServ(id uint) error {
	return repository.DeleteProviderByIDDAO(id)
}

// UpdateProviderServ updates an existing provider
func UpdateProviderServ(id uint, req ProviderReq) error {
	existing, err := repository.GetProviderByIDDAO(id)
	if err != nil {
		return err
	}
	updated := domain.Provider{
		Name:         req.Name,
		URL:          req.URL,
		APIKey:       req.APIKey,
		DefaultModel: req.DefaultModel,
		Models:       req.Models,
		Type:         req.Type,
		Description:  req.Description,
		Enabled:      req.Enabled,
		Status:       existing.Status,
	}
	// Preserve status unless enabled state or API key changed
	if req.Enabled != existing.Enabled || req.APIKey != existing.APIKey {
		updated.Status = determineProviderStatus(domain.Provider{Enabled: req.Enabled, APIKey: req.APIKey})
	}
	if err := repository.UpdateProviderDAO(id, updated); err != nil {
		return err
	}
	_, _ = recomputeProviderStatus(id)
	return nil
}

// providerToRes converts a domain Provider to ProviderRes
func providerToRes(p domain.Provider) ProviderRes {
	return ProviderRes{
		ID:           int(p.ID),
		Name:         p.Name,
		URL:          p.URL,
		APIKey:       p.APIKey,
		DefaultModel: p.DefaultModel,
		Models:       p.Models,
		Type:         p.Type,
		Description:  p.Description,
		Enabled:      p.Enabled,
		Status:       p.Status,
	}
}
