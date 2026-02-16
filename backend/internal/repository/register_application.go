package repository

import (
	"errors"

	"gorm.io/gorm"
	"nagare/internal/database"
	"nagare/internal/model"
)

// CreateRegisterApplicationDAO creates a new registration application
func CreateRegisterApplicationDAO(app model.RegisterApplication) error {
	return database.DB.Create(&app).Error
}

// GetRegisterApplicationByIDDAO retrieves a registration application by ID
func GetRegisterApplicationByIDDAO(id uint) (model.RegisterApplication, error) {
	var app model.RegisterApplication
	err := database.DB.First(&app, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return app, model.ErrNotFound
	}
	return app, err
}

// GetRegisterApplicationByUsernameDAO retrieves a registration application by username
func GetRegisterApplicationByUsernameDAO(username string) (model.RegisterApplication, error) {
	var app model.RegisterApplication
	err := database.DB.Where("username = ?", username).Order("id DESC").First(&app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return app, model.ErrNotFound
	}
	return app, err
}

// SearchRegisterApplicationsDAO retrieves registration applications by filter
func SearchRegisterApplicationsDAO(filter model.RegisterApplicationFilter) ([]model.RegisterApplication, error) {
	query := database.DB.Model(&model.RegisterApplication{})
	if filter.Query != "" {
		query = query.Where("username LIKE ?", "%"+filter.Query+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	query = applySort(query, filter.SortBy, filter.SortOrder, map[string]string{
		"name":       "username",
		"username":   "username",
		"status":     "status",
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
	var apps []model.RegisterApplication
	if err := query.Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

// CountRegisterApplicationsDAO returns total count for register applications by filter
func CountRegisterApplicationsDAO(filter model.RegisterApplicationFilter) (int64, error) {
	query := database.DB.Model(&model.RegisterApplication{})
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
	result := database.DB.Model(&model.RegisterApplication{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return nil
}
