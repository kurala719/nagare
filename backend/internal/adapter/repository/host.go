package repository

import (
	"errors"
	"time"

	"nagare/internal/core/domain"
	"nagare/internal/database"

	"gorm.io/gorm"
)

// HostWithItems represents a host with its associated items
// (used by application layer for host+items responses)
type HostWithItems struct {
	Host  domain.Host
	Items []domain.Item
}

// GetAllHostsDAO retrieves all hosts from the database
func GetAllHostsDAO() ([]domain.Host, error) {
	var hosts []domain.Host
	if err := database.DB.Find(&hosts).Error; err != nil {
		return nil, err
	}
	return hosts, nil
}

// SearchHostsDAO retrieves hosts by filter
func SearchHostsDAO(filter domain.HostFilter) ([]domain.Host, error) {
	query := database.DB.Model(&domain.Host{})
	if filter.Query != "" {
		query = query.Where("name LIKE ? OR hostid LIKE ? OR ip_addr LIKE ? OR description LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%", "%"+filter.Query+"%", "%"+filter.Query+"%")
	}
	if filter.MID != nil {
		query = query.Where("m_id = ?", *filter.MID)
	}
	if filter.GroupID != nil {
		query = query.Where("group_id = ?", *filter.GroupID)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.IPAddr != nil {
		query = query.Where("ip_addr = ?", *filter.IPAddr)
	}
	query = applySort(query, filter.SortBy, filter.SortOrder, map[string]string{
		"name":       "name",
		"status":     "status",
		"enabled":    "enabled",
		"m_id":       "m_id",
		"group_id":   "group_id",
		"hostid":     "hostid",
		"ip_addr":    "ip_addr",
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
	var hosts []domain.Host
	if err := query.Find(&hosts).Error; err != nil {
		return nil, err
	}
	return hosts, nil
}

// CountHostsDAO returns total count for hosts by filter
func CountHostsDAO(filter domain.HostFilter) (int64, error) {
	query := database.DB.Model(&domain.Host{})
	if filter.Query != "" {
		query = query.Where("name LIKE ? OR hostid LIKE ? OR ip_addr LIKE ? OR description LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%", "%"+filter.Query+"%", "%"+filter.Query+"%")
	}
	if filter.MID != nil {
		query = query.Where("m_id = ?", *filter.MID)
	}
	if filter.GroupID != nil {
		query = query.Where("group_id = ?", *filter.GroupID)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.IPAddr != nil {
		query = query.Where("ip_addr = ?", *filter.IPAddr)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// GetHostByIDDAO retrieves a host by ID
func GetHostByIDDAO(id uint) (domain.Host, error) {
	var host domain.Host
	err := database.DB.First(&host, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return host, domain.ErrNotFound
	}
	return host, err
}

// GetHostByHostIDDAO retrieves a host by external host ID
func GetHostByHostIDDAO(hostid string) (domain.Host, error) {
	var host domain.Host
	err := database.DB.Where("hostid = ?", hostid).First(&host).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return host, domain.ErrNotFound
	}
	return host, err
}

// GetHostByMIDAndHostIDDAO retrieves a host by monitor ID and external host ID
func GetHostByMIDAndHostIDDAO(mid uint, hostid string) (domain.Host, error) {
	var host domain.Host
	err := database.DB.Where("hostid = ? AND m_id = ?", hostid, mid).First(&host).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return host, domain.ErrNotFound
	}
	return host, err
}

// AddHostDAO creates a new host in the database
func AddHostDAO(h domain.Host) error {
	return database.DB.Create(&h).Error
}

// DeleteHostByIDDAO deletes a host by ID
func DeleteHostByIDDAO(id uint) error {
	return database.DB.Delete(&domain.Host{}, id).Error
}

// DeleteHostByMIDDAO deletes all hosts associated with a monitor
func DeleteHostByMIDDAO(mid uint) error {
	return database.DB.Where("m_id = ?", mid).Delete(&domain.Host{}).Error
}

// UpdateHostDAO updates a host by ID
func UpdateHostDAO(id uint, h domain.Host) error {
	// Use individual Update calls to ensure all fields including zero values are updated
	// This bypasses GORM's zero-value skipping behavior
	db := database.DB.Model(&domain.Host{}).Where("id = ?", id).
		Update("name", h.Name).
		Update("hostid", h.Hostid).
		Update("m_id", h.MonitorID).
		Update("group_id", h.GroupID).
		Update("description", h.Description).
		Update("enabled", h.Enabled).
		Update("status", h.Status).
		Update("status_description", h.StatusDescription).
		Update("active_available", h.ActiveAvailable).
		Update("ip_addr", h.IPAddr).
		Update("comment", h.Comment).
		Update("ssh_user", h.SSHUser).
		Update("ssh_port", h.SSHPort).
		Update("last_sync_at", h.LastSyncAt).
		Update("external_source", h.ExternalSource).
		Update("health_score", h.HealthScore)

	if h.SSHPassword != "" {
		db = db.Update("ssh_password", h.SSHPassword)
	}

	return db.Error
}

// UpdateHostHealthScoreDAO updates only the health score for a host
func UpdateHostHealthScoreDAO(id uint, score int) error {
	return database.DB.Model(&domain.Host{}).Where("id = ?", id).Update("health_score", score).Error
}

// UpdateHostStatusDAO updates only the status for a host
func UpdateHostStatusDAO(id uint, status int) error {
	return database.DB.Model(&domain.Host{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateHostStatusAndCommentDAO updates status and comment for a host
func UpdateHostStatusAndCommentDAO(id uint, status int, comment string) error {
	return database.DB.Model(&domain.Host{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":  status,
		"comment": comment,
	}).Error
}

// UpdateHostStatusAndDescriptionDAO updates status and status_description for a host
func UpdateHostStatusAndDescriptionDAO(id uint, status int, statusDesc string) error {
	return database.DB.Model(&domain.Host{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":             status,
		"status_description": statusDesc,
	}).Error
}

// UpdateHostLastSyncAtDAO updates only the last_sync_at for a host
func UpdateHostLastSyncAtDAO(id uint, lastSyncAt *time.Time) error {
	return database.DB.Model(&domain.Host{}).Where("id = ?", id).Update("last_sync_at", lastSyncAt).Error
}

// GetHostWithItemsByIDDAO retrieves a specific host with its items
func GetHostWithItemsByIDDAO(id uint) (HostWithItems, error) {
	host, err := GetHostByIDDAO(id)
	if err != nil {
		return HostWithItems{}, err
	}

	items, err := GetItemsByHIDDAO(id)
	if err != nil {
		items = []domain.Item{}
	}

	return HostWithItems{
		Host:  host,
		Items: items,
	}, nil
}
