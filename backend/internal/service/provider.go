package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"nagare/internal/model"
	"nagare/internal/repository"
	"nagare/internal/repository/llm"
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
func SearchProvidersServ(filter model.ProviderFilter) ([]ProviderRes, error) {
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
func CountProvidersServ(filter model.ProviderFilter) (int64, error) {
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
	return repository.AddProviderDAO(model.Provider{
		Name:         req.Name,
		URL:          req.URL,
		APIKey:       req.APIKey,
		DefaultModel: req.DefaultModel,
		Models:       req.Models,
		Type:         req.Type,
		Description:  req.Description,
		Enabled:      req.Enabled,
		Status:       0, // Default to inactive on creation
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
	updated := model.Provider{
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
	if err := repository.UpdateProviderDAO(id, updated); err != nil {
		return err
	}
	return nil
}

// FetchProviderModelsServ fetches available models from the provider's API and updates the DB
func FetchProviderModelsServ(id uint) ([]string, error) {
	provider, err := repository.GetProviderByIDDAO(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	client, _, err := createLLMClient(id, "")
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	models, err := client.FetchModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models from API: %w", err)
	}

	// Update DB with fresh models
	provider.Models = models
	if err := repository.UpdateProviderDAO(id, provider); err != nil {
		return nil, fmt.Errorf("failed to save fetched models: %w", err)
	}

	return models, nil
}

// FetchModelsDirectServ fetches available models using provided config without saving to DB
func FetchModelsDirectServ(req ProviderReq) ([]string, error) {
	if req.APIKey == "" {
		return nil, errors.New("API key is required")
	}

	var providerType llm.ProviderType
	switch req.Type {
	case 1:
		providerType = llm.ProviderGemini
	case 2:
		providerType = llm.ProviderOpenAI
	default:
		providerType = llm.ProviderGemini
	}

	client, err := llm.NewClient(llm.Config{
		APIKey:  req.APIKey,
		BaseURL: req.URL,
		Type:    providerType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return client.FetchModels(ctx)
}

// providerToRes converts a domain Provider to ProviderRes
func providerToRes(p model.Provider) ProviderRes {
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
