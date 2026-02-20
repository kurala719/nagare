package repository

import (
	"errors"

	"nagare/internal/core/domain"
	"nagare/internal/database"

	"gorm.io/gorm"
)

// CreateRegisterApplicationDAO creates a new registration application
func CreateRegisterApplicationDAO(app domain.RegisterApplication) error {
	return database.DB.Create(&app).Error
}

// GetRegisterApplicationByIDDAO retrieves a registration application by ID
func GetRegisterApplicationByIDDAO(id uint) (domain.RegisterApplication, error) {
	var app domain.RegisterApplication
	err := database.DB.First(&app, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return app, domain.ErrNotFound
	}
	return app, err
}

// GetRegisterApplicationByUsernameDAO retrieves a registration application by username
func GetRegisterApplicationByUsernameDAO(username string) (domain.RegisterApplication, error) {
	var app domain.RegisterApplication
	err := database.DB.Where("username = ?", username).Order("id DESC").First(&app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return app, domain.ErrNotFound
	}
	return app, err
}

// SearchRegisterApplicationsDAO retrieves registration applications by filter
func SearchRegisterApplicationsDAO(filter domain.RegisterApplicationFilter) ([]domain.RegisterApplication, error) {
	query := database.DB.Model(&domain.RegisterApplication{})
	if filter.Query != "" {
		query = query.Where("username LIKE ?", "%"+filter.Query+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	query = applySort(query, filter.SortBy, filter.SortOrder, map[string]string{
		"name":        "username",
		"username":    "username",
		"status":      "status",
		"approved_by": "approved_by",
		"created_at":  "created_at",
		"updated_at":  "updated_at",
		"id":          "id",
	}, "id desc")
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}
	var apps []domain.RegisterApplication
	if err := query.Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

// CountRegisterApplicationsDAO returns total count for register applications by filter
func CountRegisterApplicationsDAO(filter domain.RegisterApplicationFilter) (int64, error) {
	query := database.DB.Model(&domain.RegisterApplication{})
	if filter.Query != "" {
		query = query.Where("username LIKE ?", "%"+filter.Query+"%")
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

// UpdateRegisterApplicationStatusDAO updates status and approval info
func UpdateRegisterApplicationStatusDAO(id uint, status int, approvedBy *uint, reason string) error {
	updates := map[string]interface{}{
		"status": status,
		"reason": reason,
	}
	if approvedBy != nil {
		updates["approved_by"] = *approvedBy
	}
	result := database.DB.Model(&domain.RegisterApplication{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
