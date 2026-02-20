package repository

import (
	"errors"

	"nagare/internal/core/domain"
	"nagare/internal/database"

	"gorm.io/gorm"
)

// CreatePasswordResetApplicationDAO creates a new password reset application
func CreatePasswordResetApplicationDAO(app domain.PasswordResetApplication) error {
	return database.DB.Create(&app).Error
}

// GetPasswordResetApplicationByIDDAO retrieves a password reset application by ID
func GetPasswordResetApplicationByIDDAO(id uint) (domain.PasswordResetApplication, error) {
	var app domain.PasswordResetApplication
	err := database.DB.First(&app, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return app, domain.ErrNotFound
	}
	return app, err
}

// SearchPasswordResetApplicationsDAO retrieves password reset applications by filter
func SearchPasswordResetApplicationsDAO(filter domain.RegisterApplicationFilter) ([]domain.PasswordResetApplication, error) {
	query := database.DB.Model(&domain.PasswordResetApplication{})
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
	var apps []domain.PasswordResetApplication
	if err := query.Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

// CountPasswordResetApplicationsDAO returns total count for reset applications by filter
func CountPasswordResetApplicationsDAO(filter domain.RegisterApplicationFilter) (int64, error) {
	query := database.DB.Model(&domain.PasswordResetApplication{})
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

// UpdatePasswordResetApplicationStatusDAO updates status and approval info
func UpdatePasswordResetApplicationStatusDAO(id uint, status int, approvedBy *uint, reason string) error {
	updates := map[string]interface{}{
		"status": status,
		"reason": reason,
	}
	if approvedBy != nil {
		updates["approved_by"] = *approvedBy
	}
	result := database.DB.Model(&domain.PasswordResetApplication{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
