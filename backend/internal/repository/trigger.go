package repository

import (
	"errors"

	"nagare/internal/database"
	"nagare/internal/model"

	"gorm.io/gorm"
)

// GetAllTriggersDAO retrieves all triggers
func GetAllTriggersDAO() ([]model.Trigger, error) {
	var triggers []model.Trigger
	if err := database.DB.Find(&triggers).Error; err != nil {
		return nil, err
	}
	return triggers, nil
}

// SearchTriggersDAO retrieves triggers by filter
func SearchTriggersDAO(filter model.TriggerFilter) ([]model.Trigger, error) {
	query := database.DB.Model(&model.Trigger{})
	if filter.Query != "" {
		query = query.Where("name LIKE ?", "%"+filter.Query+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Severity != nil {
		query = query.Where("severity = ?", *filter.Severity)
	}

	if filter.AlertID != nil {
		query = query.Where("alert_id = ?", *filter.AlertID)
	}

	if filter.ItemID != nil {
		query = query.Where("item_id = ?", *filter.ItemID)
	}

	query = applySort(query, filter.SortBy, filter.SortOrder, map[string]string{
		"name":       "name",
		"status":     "status",
		"enabled":    "enabled",
		"entity":     "entity",
		"severity":   "severity",
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
	var triggers []model.Trigger
	if err := query.Find(&triggers).Error; err != nil {
		return nil, err
	}
	return triggers, nil
}

// CountTriggersDAO returns total count for triggers by filter
func CountTriggersDAO(filter model.TriggerFilter) (int64, error) {
	query := database.DB.Model(&model.Trigger{})
	if filter.Query != "" {
		query = query.Where("name LIKE ?", "%"+filter.Query+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Severity != nil {
		query = query.Where("severity = ?", *filter.Severity)
	}

	if filter.AlertID != nil {
		query = query.Where("alert_id = ?", *filter.AlertID)
	}

	if filter.ItemID != nil {
		query = query.Where("item_id = ?", *filter.ItemID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// GetTriggerByIDDAO retrieves trigger by ID
func GetTriggerByIDDAO(id uint) (model.Trigger, error) {
	var trigger model.Trigger
	err := database.DB.First(&trigger, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return trigger, model.ErrNotFound
	}
	return trigger, err
}

// AddTriggerDAO creates a new trigger
func AddTriggerDAO(trigger model.Trigger) error {
	return database.DB.Create(&trigger).Error
}

// UpdateTriggerDAO updates trigger by ID
func UpdateTriggerDAO(id uint, trigger model.Trigger) error {
	return database.DB.Model(&model.Trigger{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":                     trigger.Name,
		"severity":                 trigger.Severity,
		"alert_id":                 trigger.AlertID,
		"item_status":              trigger.ItemStatus,
		"item_value_threshold":     trigger.ItemValueThreshold,
		"item_value_threshold_max": trigger.ItemValueThresholdMax,
		"item_value_operator":      trigger.ItemValueOperator,
		"enabled":                  trigger.Enabled,
		"status":                   trigger.Status,
	}).Error
}

// DeleteTriggerByIDDAO deletes trigger by ID
func DeleteTriggerByIDDAO(id uint) error {
	return database.DB.Delete(&model.Trigger{}, id).Error
}

// UpdateTriggerStatusDAO updates only status for trigger
func UpdateTriggerStatusDAO(id uint, status int) error {
	return database.DB.Model(&model.Trigger{}).Where("id = ?", id).Update("status", status).Error
}

// GetActiveTriggersForSeverityDAO retrieves triggers matching severity
func GetActiveTriggersForSeverityDAO(severity int) ([]model.Trigger, error) {
	var triggers []model.Trigger
	if err := database.DB.Where("enabled = ? AND severity <= ?", 1, severity).Find(&triggers).Error; err != nil {
		return nil, err
	}
	return triggers, nil
}

// GetActiveTriggersDAO retrieves all active triggers
func GetActiveTriggersDAO() ([]model.Trigger, error) {
	var triggers []model.Trigger
	if err := database.DB.Where("enabled = ?", 1).Find(&triggers).Error; err != nil {
		return nil, err
	}
	return triggers, nil
}
