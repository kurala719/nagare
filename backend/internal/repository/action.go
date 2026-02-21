package repository

import (
	"errors"

	"gorm.io/gorm"
	"nagare/internal/database"
	"nagare/internal/model"
)

// GetAllActionsDAO retrieves all actions
func GetAllActionsDAO() ([]model.Action, error) {
	var actions []model.Action
	if err := database.DB.Preload("Users").Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

// SearchActionsDAO retrieves actions by filter
func SearchActionsDAO(filter model.ActionFilter) ([]model.Action, error) {
	query := database.DB.Model(&model.Action{}).Preload("Users")
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
	var actions []model.Action
	if err := query.Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

// CountActionsDAO returns total count for actions by filter
func CountActionsDAO(filter model.ActionFilter) (int64, error) {
	query := database.DB.Model(&model.Action{})
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
func GetActionByIDDAO(id uint) (model.Action, error) {
	var action model.Action
	err := database.DB.Preload("Users").First(&action, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return action, model.ErrNotFound
	}
	return action, err
}

// GetActionsByMediaIDDAO retrieves actions by media ID
func GetActionsByMediaIDDAO(mediaID uint) ([]model.Action, error) {
	var actions []model.Action
	if err := database.DB.Preload("Users").Where("media_id = ?", mediaID).Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

// AddActionDAO creates a new action
func AddActionDAO(action model.Action) error {
	return database.DB.Create(&action).Error
}

// UpdateActionDAO updates action by ID
func UpdateActionDAO(id uint, action model.Action) error {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&model.Action{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":         action.Name,
		"media_id":     action.MediaID,
		"template":     action.Template,
		"enabled":      action.Enabled,
		"status":       action.Status,
		"description":  action.Description,
		"severity_min": action.SeverityMin,
		"trigger_id":   action.TriggerID,
		"host_id":      action.HostID,
		"group_id":     action.GroupID,
		"alert_status": action.AlertStatus,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update many-to-many relationship
	if err := tx.Model(&model.Action{Model: gorm.Model{ID: id}}).Association("Users").Replace(action.Users); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// DeleteActionByIDDAO deletes action by ID
func DeleteActionByIDDAO(id uint) error {
	return database.DB.Delete(&model.Action{}, id).Error
}

// UpdateActionStatusDAO updates only status for action
func UpdateActionStatusDAO(id uint, status int) error {
	return database.DB.Model(&model.Action{}).Where("id = ?", id).Update("status", status).Error
}
