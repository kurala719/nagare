package repository

import (
	"nagare/internal/database"
	"nagare/internal/model"
	"time"
)

// GetRetentionPoliciesDAO retrieves all retention policies
func GetRetentionPoliciesDAO() ([]model.RetentionPolicy, error) {
	var policies []model.RetentionPolicy
	if err := database.DB.Find(&policies).Order("data_type asc").Error; err != nil {
		return nil, err
	}
	return policies, nil
}

// GetRetentionPolicyByDataTypeDAO retrieves a retention policy by data type
func GetRetentionPolicyByDataTypeDAO(dataType string) (model.RetentionPolicy, error) {
	var policy model.RetentionPolicy
	if err := database.DB.Where("data_type = ?", dataType).First(&policy).Error; err != nil {
		return policy, err
	}
	return policy, nil
}

// UpdateRetentionPolicyDAO updates or creates a retention policy
func UpdateRetentionPolicyDAO(policy model.RetentionPolicy) error {
	var existing model.RetentionPolicy
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
	return database.DB.Delete(&model.RetentionPolicy{}, id).Error
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
		res := database.DB.Unscoped().Where("created_at < ?", cutoff).Delete(&model.LogEntry{})
		result = res.RowsAffected
	case "alerts":
		res := database.DB.Unscoped().Where("created_at < ?", cutoff).Delete(&model.Alert{})
		result = res.RowsAffected
	case "audit_logs":
		res := database.DB.Unscoped().Where("created_at < ?", cutoff).Delete(&model.AuditLog{})
		result = res.RowsAffected
	case "item_history":
		res := database.DB.Unscoped().Where("sampled_at < ?", cutoff).Delete(&model.ItemHistory{})
		result = res.RowsAffected
	case "host_history":
		res := database.DB.Unscoped().Where("sampled_at < ?", cutoff).Delete(&model.HostHistory{})
		result = res.RowsAffected
	case "network_history":
		res := database.DB.Unscoped().Where("sampled_at < ?", cutoff).Delete(&model.NetworkStatusHistory{})
		result = res.RowsAffected
	case "chat":
		res := database.DB.Unscoped().Where("created_at < ?", cutoff).Delete(&model.Chat{})
		result = res.RowsAffected
	case "ansible_jobs":
		res := database.DB.Unscoped().Where("created_at < ?", cutoff).Delete(&model.AnsibleJob{})
		result = res.RowsAffected
	case "reports":
		res := database.DB.Unscoped().Where("created_at < ?", cutoff).Delete(&model.Report{})
		result = res.RowsAffected
	case "site_messages":
		res := database.DB.Unscoped().Where("created_at < ?", cutoff).Delete(&model.SiteMessage{})
		result = res.RowsAffected
	}

	return result, nil
}
