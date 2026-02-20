package repository

import (
	"nagare/internal/core/domain"
	"nagare/internal/database"
	"time"
)

// GetRetentionPoliciesDAO retrieves all retention policies
func GetRetentionPoliciesDAO() ([]domain.RetentionPolicy, error) {
	var policies []domain.RetentionPolicy
	if err := database.DB.Find(&policies).Order("data_type asc").Error; err != nil {
		return nil, err
	}
	return policies, nil
}

// GetRetentionPolicyByDataTypeDAO retrieves a retention policy by data type
func GetRetentionPolicyByDataTypeDAO(dataType string) (domain.RetentionPolicy, error) {
	var policy domain.RetentionPolicy
	if err := database.DB.Where("data_type = ?", dataType).First(&policy).Error; err != nil {
		return policy, err
	}
	return policy, nil
}

// UpdateRetentionPolicyDAO updates or creates a retention policy
func UpdateRetentionPolicyDAO(policy domain.RetentionPolicy) error {
	var existing domain.RetentionPolicy
	if err := database.DB.Where("data_type = ?", policy.DataType).First(&existing).Error; err == nil {
		// Update existing
		return database.DB.Model(&existing).Updates(map[string]interface{}{
			"retention_days": policy.RetentionDays,
			"enabled":        policy.Enabled,
			"description":    policy.Description,
		}).Error
	}
	// Create new
	return database.DB.Create(&policy).Error
}

// DeleteRetentionPolicyDAO deletes a retention policy
func DeleteRetentionPolicyDAO(id uint) error {
	return database.DB.Delete(&domain.RetentionPolicy{}, id).Error
}

// CleanOldDataDAO deletes data older than the specified retention period for a given type
func CleanOldDataDAO(dataType string, days int) (int64, error) {
	if days <= 0 {
		return 0, nil
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	var result int64

	switch dataType {
	case "logs":
		res := database.DB.Unscoped().Where("created_at < ?", cutoff).Delete(&domain.LogEntry{})
		result = res.RowsAffected
	case "alerts":
		res := database.DB.Unscoped().Where("created_at < ?", cutoff).Delete(&domain.Alert{})
		result = res.RowsAffected
	case "audit_logs":
		res := database.DB.Unscoped().Where("created_at < ?", cutoff).Delete(&domain.AuditLog{})
		result = res.RowsAffected
	case "item_history":
		res := database.DB.Unscoped().Where("sampled_at < ?", cutoff).Delete(&domain.ItemHistory{})
		result = res.RowsAffected
	case "host_history":
		res := database.DB.Unscoped().Where("sampled_at < ?", cutoff).Delete(&domain.HostHistory{})
		result = res.RowsAffected
	case "network_history":
		res := database.DB.Unscoped().Where("sampled_at < ?", cutoff).Delete(&domain.NetworkStatusHistory{})
		result = res.RowsAffected
	case "chat":
		res := database.DB.Unscoped().Where("created_at < ?", cutoff).Delete(&domain.Chat{})
		result = res.RowsAffected
	case "ansible_jobs":
		res := database.DB.Unscoped().Where("created_at < ?", cutoff).Delete(&domain.AnsibleJob{})
		result = res.RowsAffected
	case "reports":
		res := database.DB.Unscoped().Where("created_at < ?", cutoff).Delete(&domain.Report{})
		result = res.RowsAffected
	case "site_messages":
		res := database.DB.Unscoped().Where("created_at < ?", cutoff).Delete(&domain.SiteMessage{})
		result = res.RowsAffected
	}

	return result, nil
}
