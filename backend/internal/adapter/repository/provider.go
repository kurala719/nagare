package repository

import (
	"errors"

	"nagare/internal/core/domain"
	"nagare/internal/database"

	"gorm.io/gorm"
)

// GetAllProvidersDAO retrieves all providers from the database
func GetAllProvidersDAO() ([]domain.Provider, error) {
	var providers []domain.Provider
	if err := database.DB.Find(&providers).Error; err != nil {
		return nil, err
	}
	return providers, nil
}

// SearchProvidersDAO retrieves providers by filter
func SearchProvidersDAO(filter domain.ProviderFilter) ([]domain.Provider, error) {
	query := database.DB.Model(&domain.Provider{})
	if filter.Query != "" {
		query = query.Where("name LIKE ? OR url LIKE ? OR description LIKE ? OR default_model LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%", "%"+filter.Query+"%", "%"+filter.Query+"%")
	}
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	query = applySort(query, filter.SortBy, filter.SortOrder, map[string]string{
		"name":       "name",
		"status":     "status",
		"type":       "type",
		"created_at": "created_at",
		"updated_at": "updated_at",
		"id":         "id",
	}, "id desc")
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}
	var providers []domain.Provider
	if err := query.Find(&providers).Error; err != nil {
		return nil, err
	}
	return providers, nil
}

// CountProvidersDAO returns total count for providers by filter
func CountProvidersDAO(filter domain.ProviderFilter) (int64, error) {
	query := database.DB.Model(&domain.Provider{})
	if filter.Query != "" {
		query = query.Where("name LIKE ? OR url LIKE ? OR description LIKE ? OR default_model LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%", "%"+filter.Query+"%", "%"+filter.Query+"%")
	}
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// GetProviderByIDDAO retrieves a provider by ID
func GetProviderByIDDAO(id uint) (domain.Provider, error) {
	var provider domain.Provider
	err := database.DB.First(&provider, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return provider, domain.ErrNotFound
	}
	return provider, err
}

// AddProviderDAO creates a new provider
func AddProviderDAO(p domain.Provider) error {
	return database.DB.Create(&p).Error
}

// DeleteProviderByIDDAO deletes a provider by ID
func DeleteProviderByIDDAO(id uint) error {
	return database.DB.Delete(&domain.Provider{}, id).Error
}

// UpdateProviderDAO updates a provider by ID
func UpdateProviderDAO(id uint, p domain.Provider) error {
	return database.DB.Model(&domain.Provider{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":          p.Name,
		"url":           p.URL,
		"api_key":       p.APIKey,
		"default_model": p.DefaultModel,
		"models":        p.Models,
		"type":          p.Type,
		"description":   p.Description,
		"enabled":       p.Enabled,
		"status":        p.Status,
	}).Error
}

// UpdateProviderStatusDAO updates only the status for a provider
func UpdateProviderStatusDAO(id uint, status int) error {
	return database.DB.Model(&domain.Provider{}).Where("id = ?", id).Update("status", status).Error
}
