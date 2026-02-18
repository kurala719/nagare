package repository

import (
	"nagare/internal/database"
	"nagare/internal/model"
)

// AddLogDAO creates a new log entry
func AddLogDAO(entry model.LogEntry) error {
	return database.DB.Create(&entry).Error
}

// SearchLogsDAO retrieves logs by filter
func SearchLogsDAO(filter model.LogFilter) ([]model.LogEntry, error) {
	query := database.DB.Model(&model.LogEntry{})
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Severity != nil {
		query = query.Where("level = ?", *filter.Severity)
	}
	if filter.Query != "" {
		query = query.Where("message LIKE ? OR context LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%")
	}
	query = applySort(query, filter.SortBy, filter.SortOrder, map[string]string{
		"name":       "message",
		"message":    "message",
		"severity":   "level",
		"ip":         "ip",
		"context":    "context",
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
	var logs []model.LogEntry
	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// CountLogsDAO returns total count for logs by filter
func CountLogsDAO(filter model.LogFilter) (int64, error) {
	query := database.DB.Model(&model.LogEntry{})
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Severity != nil {
		query = query.Where("level = ?", *filter.Severity)
	}
	if filter.Query != "" {
		query = query.Where("message LIKE ? OR context LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
