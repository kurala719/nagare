package repository

import (
	"errors"

	"gorm.io/gorm"
	"nagare/internal/database"
	"nagare/internal/model"
)

// CreatePasswordResetApplicationDAO creates a new password reset application
func CreatePasswordResetApplicationDAO(app model.PasswordResetApplication) error {
	return database.DB.Create(&app).Error
}

// GetPasswordResetApplicationByIDDAO retrieves a password reset application by ID
func GetPasswordResetApplicationByIDDAO(id uint) (model.PasswordResetApplication, error) {
	var app model.PasswordResetApplication
	err := database.DB.First(&app, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return app, model.ErrNotFound
	}
	return app, err
}

// SearchPasswordResetApplicationsDAO retrieves password reset applications by filter
func SearchPasswordResetApplicationsDAO(filter model.RegisterApplicationFilter) ([]model.PasswordResetApplication, error) {
	query := database.DB.Model(&model.PasswordResetApplication{})
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
	var apps []model.PasswordResetApplication
	if err := query.Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

// CountPasswordResetApplicationsDAO returns total count for reset applications by filter
func CountPasswordResetApplicationsDAO(filter model.RegisterApplicationFilter) (int64, error) {
	query := database.DB.Model(&model.PasswordResetApplication{})
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
	result := database.DB.Model(&model.PasswordResetApplication{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return nil
}
