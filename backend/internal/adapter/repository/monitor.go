package repository

import (
	"errors"

	"nagare/internal/core/domain"
	"nagare/internal/database"

	"gorm.io/gorm"
)

// GetAllMonitorsDAO retrieves all monitors from the database
func GetAllMonitorsDAO() ([]domain.Monitor, error) {
	var monitors []domain.Monitor
	if err := database.DB.Find(&monitors).Error; err != nil {
		return nil, err
	}
	return monitors, nil
}

// SearchMonitorsDAO retrieves monitors by filter
func SearchMonitorsDAO(filter domain.MonitorFilter) ([]domain.Monitor, error) {
	query := database.DB.Model(&domain.Monitor{})
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
	var monitors []domain.Monitor
	if err := query.Find(&monitors).Error; err != nil {
		return nil, err
	}
	return monitors, nil
}

// CountMonitorsDAO returns total count for monitors by filter
func CountMonitorsDAO(filter domain.MonitorFilter) (int64, error) {
	query := database.DB.Model(&domain.Monitor{})
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

// GetMonitorByIDDAO retrieves a monitor by ID
func GetMonitorByIDDAO(id uint) (domain.Monitor, error) {
	var monitor domain.Monitor
	err := database.DB.First(&monitor, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return monitor, domain.ErrNotFound
	}
	return monitor, err
}

// GetMonitorByEventTokenDAO retrieves a monitor by event token
func GetMonitorByEventTokenDAO(eventToken string) (domain.Monitor, error) {
	var monitor domain.Monitor
	err := database.DB.Where("event_token = ?", eventToken).First(&monitor).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return monitor, domain.ErrNotFound
	}
	return monitor, err
}

// AddMonitorDAO creates a new monitor
func AddMonitorDAO(m domain.Monitor) error {
	return database.DB.Create(&m).Error
}

// DeleteMonitorByIDDAO deletes a monitor by ID
func DeleteMonitorByIDDAO(id int) error {
	return database.DB.Delete(&domain.Monitor{}, id).Error
}

// UpdateMonitorDAO updates a monitor by ID
func UpdateMonitorDAO(id int, m domain.Monitor) error {
	return database.DB.Model(&domain.Monitor{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":               m.Name,
		"url":                m.URL,
		"username":           m.Username,
		"password":           m.Password,
		"auth_token":         m.AuthToken,
		"event_token":        m.EventToken,
		"description":        m.Description,
		"type":               m.Type,
		"enabled":            m.Enabled,
		"status":             m.Status,
		"status_description": m.StatusDescription,
		"health_score":       m.HealthScore,
	}).Error
}

// UpdateMonitorHealthScoreDAO updates only the health score for a monitor
func UpdateMonitorHealthScoreDAO(id uint, score int) error {
	return database.DB.Model(&domain.Monitor{}).Where("id = ?", id).Update("health_score", score).Error
}

// UpdateMonitorStatusDAO updates only the status for a monitor
func UpdateMonitorStatusDAO(id uint, status int) error {
	return database.DB.Model(&domain.Monitor{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateMonitorStatusAndDescriptionDAO updates status and status_description for a monitor
func UpdateMonitorStatusAndDescriptionDAO(id uint, status int, statusDesc string) error {
	return database.DB.Model(&domain.Monitor{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":             status,
		"status_description": statusDesc,
	}).Error
}

// UpdateMonitorAuthTokenDAO updates only the auth token for a monitor
func UpdateMonitorAuthTokenDAO(id uint, authToken string) error {
	return database.DB.Model(&domain.Monitor{}).Where("id = ?", id).Update("auth_token", authToken).Error
}

// UpdateMonitorEventTokenDAO updates only the event token for a monitor
func UpdateMonitorEventTokenDAO(id uint, eventToken string) error {
	return database.DB.Model(&domain.Monitor{}).Where("id = ?", id).Update("event_token", eventToken).Error
}
