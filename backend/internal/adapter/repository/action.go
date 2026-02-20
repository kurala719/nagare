package repository

import (
	"errors"

	"nagare/internal/core/domain"
	"nagare/internal/database"

	"gorm.io/gorm"
)

// GetAllActionsDAO retrieves all actions
func GetAllActionsDAO() ([]domain.Action, error) {
	var actions []domain.Action
	if err := database.DB.Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

// SearchActionsDAO retrieves actions by filter
func SearchActionsDAO(filter domain.ActionFilter) ([]domain.Action, error) {
	query := database.DB.Model(&domain.Action{})
	if filter.Query != "" {
		query = query.Where("name LIKE ? OR description LIKE ? OR template LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%", "%"+filter.Query+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	query = applySort(query, filter.SortBy, filter.SortOrder, map[string]string{
		"name":       "name",
		"status":     "status",
		"enabled":    "enabled",
		"media_id":   "media_id",
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
	var actions []domain.Action
	if err := query.Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

// CountActionsDAO returns total count for actions by filter
func CountActionsDAO(filter domain.ActionFilter) (int64, error) {
	query := database.DB.Model(&domain.Action{})
	if filter.Query != "" {
		query = query.Where("name LIKE ? OR description LIKE ? OR template LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%", "%"+filter.Query+"%")
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

// GetActionByIDDAO retrieves action by ID
func GetActionByIDDAO(id uint) (domain.Action, error) {
	var action domain.Action
	err := database.DB.First(&action, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return action, domain.ErrNotFound
	}
	return action, err
}

// GetActionsByMediaIDDAO retrieves actions by media ID
func GetActionsByMediaIDDAO(mediaID uint) ([]domain.Action, error) {
	var actions []domain.Action
	if err := database.DB.Where("media_id = ?", mediaID).Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

// AddActionDAO creates a new action
func AddActionDAO(action domain.Action) error {
	return database.DB.Create(&action).Error
}

// UpdateActionDAO updates action by ID
func UpdateActionDAO(id uint, action domain.Action) error {
	return database.DB.Model(&domain.Action{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":        action.Name,
		"media_id":    action.MediaID,
		"template":    action.Template,
		"enabled":     action.Enabled,
		"status":      action.Status,
		"description": action.Description,
	}).Error
}

// DeleteActionByIDDAO deletes action by ID
func DeleteActionByIDDAO(id uint) error {
	return database.DB.Delete(&domain.Action{}, id).Error
}

// UpdateActionStatusDAO updates only status for action
func UpdateActionStatusDAO(id uint, status int) error {
	return database.DB.Model(&domain.Action{}).Where("id = ?", id).Update("status", status).Error
}
