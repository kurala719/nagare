package repository

import (
	"errors"

	"nagare/internal/core/domain"
	"nagare/internal/database"

	"gorm.io/gorm"
)

// GetAllAlarmsDAO retrieves all alarms from the database
func GetAllAlarmsDAO() ([]domain.Alarm, error) {
	var alarms []domain.Alarm
	if err := database.DB.Find(&alarms).Error; err != nil {
		return nil, err
	}
	return alarms, nil
}

// SearchAlarmsDAO retrieves alarms by filter
func SearchAlarmsDAO(filter domain.AlarmFilter) ([]domain.Alarm, error) {
	query := database.DB.Model(&domain.Alarm{})
	if filter.Query != "" {
		query = query.Where("name LIKE ? OR url LIKE ? OR description LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%", "%"+filter.Query+"%")
	}
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	query = applySort(query, filter.SortBy, filter.SortOrder, map[string]string{
		"name":       "name",
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
	var alarms []domain.Alarm
	if err := query.Find(&alarms).Error; err != nil {
		return nil, err
	}
	return alarms, nil
}

// CountAlarmsDAO returns total count for alarms by filter
func CountAlarmsDAO(filter domain.AlarmFilter) (int64, error) {
	query := database.DB.Model(&domain.Alarm{})
	if filter.Query != "" {
		query = query.Where("name LIKE ? OR url LIKE ? OR description LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%", "%"+filter.Query+"%")
	}
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
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

// GetAlarmByIDDAO retrieves an alarm by ID
func GetAlarmByIDDAO(id uint) (domain.Alarm, error) {
	var alarm domain.Alarm
	err := database.DB.First(&alarm, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return alarm, domain.ErrNotFound
	}
	return alarm, err
}

// GetAlarmByEventTokenDAO retrieves an alarm by event token
func GetAlarmByEventTokenDAO(eventToken string) (domain.Alarm, error) {
	var alarm domain.Alarm
	err := database.DB.Where("event_token = ?", eventToken).First(&alarm).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return alarm, domain.ErrNotFound
	}
	return alarm, err
}

// AddAlarmDAO creates a new alarm
func AddAlarmDAO(a domain.Alarm) error {
	return database.DB.Create(&a).Error
}

// DeleteAlarmByIDDAO deletes an alarm by ID
func DeleteAlarmByIDDAO(id int) error {
	return database.DB.Delete(&domain.Alarm{}, id).Error
}

// UpdateAlarmDAO updates an alarm by ID
func UpdateAlarmDAO(id int, a domain.Alarm) error {
	return database.DB.Model(&domain.Alarm{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":               a.Name,
		"url":                a.URL,
		"username":           a.Username,
		"password":           a.Password,
		"auth_token":         a.AuthToken,
		"event_token":        a.EventToken,
		"description":        a.Description,
		"type":               a.Type,
		"enabled":            a.Enabled,
		"status":             a.Status,
		"status_description": a.StatusDescription,
	}).Error
}

// UpdateAlarmStatusDAO updates only the status for an alarm
func UpdateAlarmStatusDAO(id uint, status int) error {
	return database.DB.Model(&domain.Alarm{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateAlarmStatusAndDescriptionDAO updates status and status_description for an alarm
func UpdateAlarmStatusAndDescriptionDAO(id uint, status int, statusDesc string) error {
	return database.DB.Model(&domain.Alarm{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":             status,
		"status_description": statusDesc,
	}).Error
}

// UpdateAlarmAuthTokenDAO updates only the auth token for an alarm
func UpdateAlarmAuthTokenDAO(id uint, authToken string) error {
	return database.DB.Model(&domain.Alarm{}).Where("id = ?", id).Update("auth_token", authToken).Error
}

// UpdateAlarmEventTokenDAO updates only the event token for an alarm
func UpdateAlarmEventTokenDAO(id uint, eventToken string) error {
	return database.DB.Model(&domain.Alarm{}).Where("id = ?", id).Update("event_token", eventToken).Error
}
