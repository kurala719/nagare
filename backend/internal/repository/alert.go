package repository

import (
	"nagare/internal/database"
	"nagare/internal/model"
)

// GetAllAlertsDAO retrieves all alerts from the database
func GetAllAlertsDAO() ([]model.Alert, error) {
	var alerts []model.Alert
	if err := database.DB.Model(&model.Alert{}).
		Select("alerts.*, hosts.name as host_name, items.name as item_name, COALESCE(alarms.name, monitors.name, 'System') as alarm_name").
		Joins("LEFT JOIN hosts ON hosts.id = alerts.host_id OR (alerts.host_id > 0 AND hosts.hostid = alerts.host_id)").
		Joins("LEFT JOIN items ON items.id = alerts.item_id OR (alerts.item_id > 0 AND items.itemid = alerts.item_id)").
		Joins("LEFT JOIN alarms ON alarms.id = alerts.alarm_id").
		Joins("LEFT JOIN monitors ON monitors.id = alerts.alarm_id AND alarms.id IS NULL").
		Order("alerts.id desc").
		Scan(&alerts).Error; err != nil {
		return nil, err
	}
	return alerts, nil
}

// SearchAlertsDAO retrieves alerts by filter
func SearchAlertsDAO(filter model.AlertFilter) ([]model.Alert, error) {
	query := database.DB.Model(&model.Alert{}).
		Select("alerts.*, hosts.name as host_name, items.name as item_name, COALESCE(alarms.name, monitors.name, 'System') as alarm_name").
		Joins("LEFT JOIN hosts ON hosts.id = alerts.host_id OR (alerts.host_id > 0 AND hosts.hostid = alerts.host_id)").
		Joins("LEFT JOIN items ON items.id = alerts.item_id OR (alerts.item_id > 0 AND items.itemid = alerts.item_id)").
		Joins("LEFT JOIN alarms ON alarms.id = alerts.alarm_id").
		Joins("LEFT JOIN monitors ON monitors.id = alerts.alarm_id AND alarms.id IS NULL")

	if filter.Query != "" {
		query = query.Where("alerts.message LIKE ?", "%"+filter.Query+"%")
	}
	if filter.Severity != nil {
		query = query.Where("alerts.severity = ?", *filter.Severity)
	}
	if filter.Status != nil {
		query = query.Where("alerts.status = ?", *filter.Status)
	}
	if filter.AlarmID != nil {
		query = query.Where("alerts.alarm_id = ?", *filter.AlarmID)
	}
	if filter.HostID != nil {
		query = query.Where("alerts.host_id = ?", *filter.HostID)
	}
	if filter.ItemID != nil {
		query = query.Where("alerts.item_id = ?", *filter.ItemID)
	}
	query = applySort(query, filter.SortBy, filter.SortOrder, map[string]string{
		"name":       "alerts.message",
		"message":    "alerts.message",
		"severity":   "alerts.severity",
		"status":     "alerts.status",
		"created_at": "alerts.created_at",
		"updated_at": "alerts.updated_at",
		"id":         "alerts.id",
	}, "alerts.id desc")
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}
	var alerts []model.Alert
	if err := query.Scan(&alerts).Error; err != nil {
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
func GetAlertByIDDAO(id int) (model.Alert, error) {
	var alert model.Alert
	err := database.DB.Model(&model.Alert{}).
		Select("alerts.*, hosts.name as host_name, items.name as item_name, COALESCE(alarms.name, monitors.name, 'System') as alarm_name").
		Joins("LEFT JOIN hosts ON hosts.id = alerts.host_id OR (alerts.host_id > 0 AND hosts.hostid = alerts.host_id)").
		Joins("LEFT JOIN items ON items.id = alerts.item_id OR (alerts.item_id > 0 AND items.itemid = alerts.item_id)").
		Joins("LEFT JOIN alarms ON alarms.id = alerts.alarm_id").
		Joins("LEFT JOIN monitors ON monitors.id = alerts.alarm_id AND alarms.id IS NULL").
		Where("alerts.id = ?", id).
		First(&alert).Error
	if err != nil {
		return alert, err
	}
	return alert, nil
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
		"alarm_id": alert.AlarmID,
		"trigger_id": alert.TriggerID,
		"host_id":  alert.HostID,
		"item_id":  alert.ItemID,
		"comment":  alert.Comment,
	}).Error
}
