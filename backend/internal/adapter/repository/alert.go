package repository

import (
	"errors"

	"nagare/internal/core/domain"
	"nagare/internal/database"

	"gorm.io/gorm"
)

// GetAllAlertsDAO retrieves all alerts from the database
func GetAllAlertsDAO() ([]domain.Alert, error) {
	var alerts []domain.Alert
	if err := database.DB.Find(&alerts).Error; err != nil {
		return nil, err
	}
	return alerts, nil
}

// SearchAlertsDAO retrieves alerts by filter
func SearchAlertsDAO(filter domain.AlertFilter) ([]domain.Alert, error) {
	query := database.DB.Model(&domain.Alert{})
	if filter.Query != "" {
		query = query.Where("message LIKE ?", "%"+filter.Query+"%")
	}
	if filter.Severity != nil {
		query = query.Where("severity = ?", *filter.Severity)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.AlarmID != nil {
		query = query.Where("alarm_id = ?", *filter.AlarmID)
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
	var alerts []domain.Alert
	if err := query.Find(&alerts).Error; err != nil {
		return nil, err
	}
	return alerts, nil
}

// CountAlertsDAO returns total count for alerts by filter
func CountAlertsDAO(filter domain.AlertFilter) (int64, error) {
	query := database.DB.Model(&domain.Alert{})
	if filter.Query != "" {
		query = query.Where("message LIKE ?", "%"+filter.Query+"%")
	}
	if filter.Severity != nil {
		query = query.Where("severity = ?", *filter.Severity)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.AlarmID != nil {
		query = query.Where("alarm_id = ?", *filter.AlarmID)
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
func GetAlertByIDDAO(id int) (domain.Alert, error) {
	var alert domain.Alert
	err := database.DB.First(&alert, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return alert, domain.ErrNotFound
	}
	return alert, err
}

// AddAlertDAO creates a new alert
func AddAlertDAO(alert *domain.Alert) error {
	return database.DB.Create(alert).Error
}

// UpdateAlertCommentDAO updates only the comment for an alert
func UpdateAlertCommentDAO(id int, comment string) error {
	return database.DB.Model(&domain.Alert{}).Where("id = ?", id).Update("comment", comment).Error
}

// DeleteAlertByIDDAO deletes an alert by ID
func DeleteAlertByIDDAO(id int) error {
	return database.DB.Delete(&domain.Alert{}, id).Error
}

// UpdateAlertDAO updates an existing alert
func UpdateAlertDAO(id int, alert domain.Alert) error {
	return database.DB.Model(&domain.Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"message":  alert.Message,
		"severity": alert.Severity,
		"status":   alert.Status,
		"alarm_id": alert.AlarmID,
		"host_id":  alert.HostID,
		"item_id":  alert.ItemID,
		"comment":  alert.Comment,
	}).Error
}
