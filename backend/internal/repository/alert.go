package repository

import (
	"nagare/internal/database"
	"nagare/internal/model"
	"strings"

	"gorm.io/gorm"
)

type AlertWithContext struct {
	model.Alert
	HostID      *uint  `gorm:"column:host_id"`
	HostName    string `gorm:"column:host_name"`
	GroupID     *uint  `gorm:"column:group_id"`
	GroupName   string `gorm:"column:group_name"`
	MonitorID   *uint  `gorm:"column:monitor_id"`
	MonitorName string `gorm:"column:monitor_name"`
	ItemName    string `gorm:"column:item_name"`
	AlarmName   string `gorm:"column:alarm_name"`
}

func alertWithContextQuery() *gorm.DB {
	return database.DB.Model(&model.Alert{}).
		Select("alerts.*, hosts.id as host_id, hosts.name as host_name, `groups`.id as group_id, `groups`.name as group_name, items.name as item_name, COALESCE(alarms.name, monitors.name, 'System') as alarm_name, monitors.id as monitor_id, monitors.name as monitor_name").
		Joins("LEFT JOIN items ON items.id = alerts.item_id OR (alerts.item_id > 0 AND items.external_id = alerts.item_id)").
		Joins("LEFT JOIN hosts ON hosts.id = items.host_id").
		Joins("LEFT JOIN `groups` ON `groups`.id = hosts.group_id").
		Joins("LEFT JOIN alarms ON alarms.id = alerts.alarm_id").
		Joins("LEFT JOIN monitors ON monitors.id = alerts.alarm_id AND alarms.id IS NULL")
}

// GetAllAlertsDAO retrieves all alerts from the database
func GetAllAlertsDAO() ([]model.Alert, error) {
	var alerts []model.Alert
	if err := database.DB.Model(&model.Alert{}).Order("id desc").Find(&alerts).Error; err != nil {
		return nil, err
	}
	return alerts, nil
}

func GetAllAlertsWithContextDAO() ([]AlertWithContext, error) {
	var alerts []AlertWithContext
	if err := alertWithContextQuery().Order("alerts.id desc").Scan(&alerts).Error; err != nil {
		return nil, err
	}
	return alerts, nil
}

// SearchAlertsDAO retrieves alerts by filter
func SearchAlertsDAO(filter model.AlertFilter) ([]model.Alert, error) {
	query := database.DB.Model(&model.Alert{})

	if filter.HostID != nil {
		query = query.Joins("LEFT JOIN items ON items.id = alerts.item_id OR (alerts.item_id > 0 AND items.external_id = alerts.item_id)")
	}

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
		query = query.Where("items.host_id = ?", *filter.HostID)
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
	if err := query.Find(&alerts).Error; err != nil {
		return nil, err
	}
	return alerts, nil
}

func SearchAlertsWithContextDAO(filter model.AlertFilter) ([]AlertWithContext, error) {
	query := alertWithContextQuery()

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
		query = query.Where("items.host_id = ?", *filter.HostID)
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

	var alerts []AlertWithContext
	if err := query.Scan(&alerts).Error; err != nil {
		return nil, err
	}
	return alerts, nil
}

// CountAlertsDAO returns total count for alerts by filter
func CountAlertsDAO(filter model.AlertFilter) (int64, error) {
	query := database.DB.Model(&model.Alert{})

	if filter.HostID != nil {
		query = query.Joins("LEFT JOIN items ON items.id = alerts.item_id OR (alerts.item_id > 0 AND items.external_id = alerts.item_id)")
	}

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
		query = query.Where("items.host_id = ?", *filter.HostID)
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
	var alerts []model.Alert
	err := database.DB.Model(&model.Alert{}).Where("id = ?", id).Limit(1).Find(&alerts).Error
	if err != nil {
		return model.Alert{}, err
	}
	if len(alerts) == 0 {
		return model.Alert{}, gorm.ErrRecordNotFound
	}
	return alerts[0], nil
}

func GetAlertByIDWithContextDAO(id int) (AlertWithContext, error) {
	var alerts []AlertWithContext
	err := alertWithContextQuery().Where("alerts.id = ?", id).Limit(1).Find(&alerts).Error
	if err != nil {
		return AlertWithContext{}, err
	}
	if len(alerts) == 0 {
		return AlertWithContext{}, gorm.ErrRecordNotFound
	}
	return alerts[0], nil
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
		"message":     alert.Message,
		"external_id": alert.ExternalID,
		"severity":    alert.Severity,
		"status":      alert.Status,
		"alarm_id":    alert.AlarmID,
		"item_id":     alert.ItemID,
		"comment":     alert.Comment,
	}).Error
}

// FindLatestUnresolvedAlertByAlarmAndItemDAO finds the newest unresolved alert for the same alarm+item.
func FindLatestUnresolvedAlertByAlarmAndItemDAO(alarmID uint, itemID uint) (model.Alert, error) {
	var alerts []model.Alert
	if alarmID == 0 || itemID == 0 {
		return model.Alert{}, nil
	}
	err := database.DB.Model(&model.Alert{}).
		Where("alarm_id = ? AND item_id = ? AND status <> 2", alarmID, itemID).
		Order("id desc").
		Limit(1).
		Find(&alerts).Error
	if err != nil {
		return model.Alert{}, err
	}
	if len(alerts) == 0 {
		return model.Alert{}, nil
	}
	return alerts[0], nil
}

// UpdateAlertStatusAndCommentDAO updates status and comment for an alert by ID.
func UpdateAlertStatusAndCommentDAO(id uint, status int, comment string) error {
	return database.DB.Model(&model.Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":  status,
		"comment": comment,
	}).Error
}

// FindLatestUnresolvedAlertByEventDAO finds the newest unresolved alert matching alarm and event id marker in comment.
func FindLatestUnresolvedAlertByEventDAO(alarmID uint, eventID string) (model.Alert, error) {
	var alerts []model.Alert
	if alarmID == 0 || strings.TrimSpace(eventID) == "" {
		return model.Alert{}, nil
	}
	eid := strings.TrimSpace(eventID)
	err := database.DB.Model(&model.Alert{}).
		Where("alarm_id = ? AND status <> 2 AND external_id = ?", alarmID, eid).
		Order("id desc").
		Limit(1).
		Find(&alerts).Error
	if err == nil && len(alerts) > 0 {
		return alerts[0], nil
	}
	if err != nil {
		return model.Alert{}, err
	}

	// Try searching by comment marker
	err = database.DB.Model(&model.Alert{}).
		Where("alarm_id = ? AND status <> 2 AND comment LIKE ?", alarmID, "%event_id="+eid+"%").
		Order("id desc").
		Limit(1).
		Find(&alerts).Error
	if err != nil {
		return model.Alert{}, err
	}
	if len(alerts) == 0 {
		return model.Alert{}, nil
	}
	return alerts[0], nil
}

// FindLatestUnresolvedAlertByExternalIDDAO finds newest unresolved alert by external_id.
func FindLatestUnresolvedAlertByExternalIDDAO(externalID string) (model.Alert, error) {
	var alerts []model.Alert
	eid := strings.TrimSpace(externalID)
	if eid == "" {
		return model.Alert{}, nil
	}
	err := database.DB.Model(&model.Alert{}).
		Where("external_id = ? AND status <> 2", eid).
		Order("id desc").
		Limit(1).
		Find(&alerts).Error
	if err != nil {
		return model.Alert{}, err
	}
	if len(alerts) == 0 {
		return model.Alert{}, nil
	}
	return alerts[0], nil
}

// FindLatestUnresolvedAlertByFingerprintDAO finds the newest unresolved alert using a conservative fingerprint.
func FindLatestUnresolvedAlertByFingerprintDAO(alarmID uint, itemID uint, message string) (model.Alert, error) {
	var alerts []model.Alert
	query := database.DB.Model(&model.Alert{}).Where("status <> 2")
	if alarmID > 0 {
		query = query.Where("alarm_id = ?", alarmID)
	}
	if itemID > 0 {
		query = query.Where("item_id = ?", itemID)
	}
	if strings.TrimSpace(message) != "" {
		query = query.Where("message = ?", strings.TrimSpace(message))
	}
	err := query.Order("id desc").Limit(1).Find(&alerts).Error
	if err != nil {
		return model.Alert{}, err
	}
	if len(alerts) == 0 {
		return model.Alert{}, nil
	}
	return alerts[0], nil
}
