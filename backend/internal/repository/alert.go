package repository

import (
	"errors"

	"gorm.io/gorm"
	"nagare/internal/database"
	"nagare/internal/model"
)

// GetAllAlertsDAO retrieves all alerts from the database
func GetAllAlertsDAO() ([]model.Alert, error) {
	var alerts []model.Alert
	if err := database.DB.Find(&alerts).Error; err != nil {
		return nil, err
	}
	return alerts, nil
}

// SearchAlertsDAO retrieves alerts by filter
func SearchAlertsDAO(filter model.AlertFilter) ([]model.Alert, error) {
	query := database.DB.Model(&model.Alert{})
	if filter.Query != "" {
		query = query.Where("message LIKE ?", "%"+filter.Query+"%")
	}
	if filter.Severity != nil {
		query = query.Where("severity = ?", *filter.Severity)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.HostID != nil {
		query = query.Where("host_id = ?", *filter.HostID)
	}
	if filter.ItemID != nil {
		query = query.Where("item_id = ?", *filter.ItemID)
	}
	query = applySort(query, filter.SortBy, filter.SortOrder, map[string]string{
		"name":       "message",
		"message":    "message",
		"severity":   "severity",
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
	var alerts []model.Alert
	if err := query.Find(&alerts).Error; err != nil {
		return nil, err
	}
	return alerts, nil
}

// CountAlertsDAO returns total count for alerts by filter
func CountAlertsDAO(filter model.AlertFilter) (int64, error) {
	query := database.DB.Model(&model.Alert{})
	if filter.Query != "" {
		query = query.Where("message LIKE ?", "%"+filter.Query+"%")
	}
	if filter.Severity != nil {
		query = query.Where("severity = ?", *filter.Severity)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.HostID != nil {
		query = query.Where("host_id = ?", *filter.HostID)
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

// GetAlertByIDDAO retrieves an alert by ID
func GetAlertByIDDAO(id int) (model.Alert, error) {
	var alert model.Alert
	err := database.DB.First(&alert, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return alert, model.ErrNotFound
	}
	return alert, err
}

// AddAlertDAO creates a new alert
func AddAlertDAO(alert *model.Alert) error {
	return database.DB.Create(alert).Error
}

// UpdateAlertCommentDAO updates only the comment for an alert
func UpdateAlertCommentDAO(id int, comment string) error {
	return database.DB.Model(&model.Alert{}).Where("id = ?", id).Update("comment", comment).Error
}

// DeleteAlertByIDDAO deletes an alert by ID
func DeleteAlertByIDDAO(id int) error {
	return database.DB.Delete(&model.Alert{}, id).Error
}

// UpdateAlertDAO updates an existing alert
func UpdateAlertDAO(id int, alert model.Alert) error {
	return database.DB.Model(&model.Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"message":  alert.Message,
		"severity": alert.Severity,
		"status":   alert.Status,
		"host_id":  alert.HostID,
		"item_id":  alert.ItemID,
		"comment":  alert.Comment,
	}).Error
}
