package repository

import (
	"errors"

	"nagare/internal/core/domain"
	"nagare/internal/database"

	"gorm.io/gorm"
)

// GetAllGroupsDAO retrieves all groups
func GetAllGroupsDAO() ([]domain.Group, error) {
	var groups []domain.Group
	if err := database.DB.Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

// SearchGroupsDAO retrieves groups by filter
func SearchGroupsDAO(filter domain.GroupFilter) ([]domain.Group, error) {
	query := database.DB.Model(&domain.Group{})
	if filter.Query != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.MonitorID != nil {
		query = query.Where("m_id = ?", *filter.MonitorID)
	}
	query = applySort(query, filter.SortBy, filter.SortOrder, map[string]string{
		"name":       "name",
		"status":     "status",
		"enabled":    "enabled",
		"monitor_id": "m_id",
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
	var groups []domain.Group
	if err := query.Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

// CountGroupsDAO returns total count for groups by filter
func CountGroupsDAO(filter domain.GroupFilter) (int64, error) {
	query := database.DB.Model(&domain.Group{})
	if filter.Query != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.MonitorID != nil {
		query = query.Where("m_id = ?", *filter.MonitorID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// GetGroupByIDDAO retrieves a group by ID
func GetGroupByIDDAO(id uint) (domain.Group, error) {
	var group domain.Group
	err := database.DB.First(&group, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return group, domain.ErrNotFound
	}
	return group, err
}

// AddGroupDAO creates a new group
func AddGroupDAO(group domain.Group) error {
	return database.DB.Create(&group).Error
}

// GetGroupByExternalIDDAO retrieves a group by its external ID and monitor ID
func GetGroupByExternalIDDAO(externalID string, monitorID uint) (domain.Group, error) {
	var group domain.Group
	err := database.DB.Where("external_id = ? AND m_id = ?", externalID, monitorID).First(&group).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return group, domain.ErrNotFound
	}
	return group, err
}

// UpdateGroupDAO updates a group by ID
func UpdateGroupDAO(id uint, group domain.Group) error {
	return database.DB.Model(&domain.Group{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":            group.Name,
		"description":     group.Description,
		"enabled":         group.Enabled,
		"status":          group.Status,
		"m_id":            group.MonitorID,
		"external_id":     group.ExternalID,
		"last_sync_at":    group.LastSyncAt,
		"external_source": group.ExternalSource,
		"health_score":    group.HealthScore,
	}).Error
}

// UpdateGroupHealthScoreDAO updates only the health score for a group
func UpdateGroupHealthScoreDAO(id uint, score int) error {
	return database.DB.Model(&domain.Group{}).Where("id = ?", id).Update("health_score", score).Error
}

// UpdateGroupStatusDAO updates only the status for a group
func UpdateGroupStatusDAO(id uint, status int) error {
	return database.DB.Model(&domain.Group{}).Where("id = ?", id).Update("status", status).Error
}

// DeleteGroupByIDDAO deletes a group by ID
func DeleteGroupByIDDAO(id uint) error {
	return database.DB.Delete(&domain.Group{}, id).Error
}
