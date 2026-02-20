package service

import (
	"nagare/internal/adapter/repository"
	"nagare/internal/core/domain"
)

// GetRetentionPoliciesServ retrieves all retention policies
func GetRetentionPoliciesServ() ([]domain.RetentionPolicy, error) {
	return repository.GetRetentionPoliciesDAO()
}

// UpdateRetentionPolicyServ updates or creates a retention policy
func UpdateRetentionPolicyServ(policy domain.RetentionPolicy) error {
	return repository.UpdateRetentionPolicyDAO(policy)
}

// PerformDataRetentionCleanupServ runs the cleanup for all enabled retention policies
func PerformDataRetentionCleanupServ() (map[string]int64, error) {
	policies, err := repository.GetRetentionPoliciesDAO()
	if err != nil {
		return nil, err
	}

	results := make(map[string]int64)
	for _, policy := range policies {
		if policy.Enabled != nil && *policy.Enabled == 1 && policy.RetentionDays > 0 {
			count, err := repository.CleanOldDataDAO(policy.DataType, policy.RetentionDays)
			if err != nil {
				LogService("error", "data retention cleanup failed for "+policy.DataType, map[string]interface{}{
					"error": err.Error(),
					"type":  policy.DataType,
					"days":  policy.RetentionDays,
				}, nil, "")
				continue
			}
			if count > 0 {
				results[policy.DataType] = count
			}
		}
	}

	if len(results) > 0 {
		LogService("info", "data retention cleanup performed", map[string]interface{}{
			"summary": results,
		}, nil, "")
	}

	return results, nil
}
