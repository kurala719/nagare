package repository

import (
	"nagare/internal/core/domain"
	"nagare/internal/database"
)

// AddAuditLogDAO creates a new audit log entry
func AddAuditLogDAO(entry domain.AuditLog) error {
	return database.DB.Create(&entry).Error
}

// SearchAuditLogsDAO retrieves audit logs by filter
func SearchAuditLogsDAO(limit, offset int, query string) ([]domain.AuditLog, int64, error) {
	var logs []domain.AuditLog
	var total int64

	db := database.DB.Model(&domain.AuditLog{})
	if query != "" {
		db = db.Where("username LIKE ? OR action LIKE ? OR path LIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("id desc").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
