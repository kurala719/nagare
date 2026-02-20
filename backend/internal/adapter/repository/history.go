package repository

import (
	"time"

	"nagare/internal/core/domain"
	"nagare/internal/database"
)

const defaultHistoryLimit = 500

func normalizeHistoryLimit(limit int) int {
	if limit <= 0 || limit > 2000 {
		return defaultHistoryLimit
	}
	return limit
}

// AddItemHistoryDAO stores a history snapshot for an item.
func AddItemHistoryDAO(history domain.ItemHistory) error {
	return database.DB.Create(&history).Error
}

// ListItemHistoryDAO returns item history entries ordered by sampled_at desc.
func ListItemHistoryDAO(itemID uint, from, to *time.Time, limit int) ([]domain.ItemHistory, error) {
	query := database.DB.Model(&domain.ItemHistory{}).Where("item_id = ?", itemID)
	if from != nil {
		query = query.Where("sampled_at >= ?", *from)
	}
	if to != nil {
		query = query.Where("sampled_at <= ?", *to)
	}
	limit = normalizeHistoryLimit(limit)
	var rows []domain.ItemHistory
	if err := query.Order("sampled_at desc").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// AddHostHistoryDAO stores a history snapshot for a host.
func AddHostHistoryDAO(history domain.HostHistory) error {
	return database.DB.Create(&history).Error
}

// ListHostHistoryDAO returns host history entries ordered by sampled_at desc.
func ListHostHistoryDAO(hostID uint, from, to *time.Time, limit int) ([]domain.HostHistory, error) {
	query := database.DB.Model(&domain.HostHistory{}).Where("host_id = ?", hostID)
	if from != nil {
		query = query.Where("sampled_at >= ?", *from)
	}
	if to != nil {
		query = query.Where("sampled_at <= ?", *to)
	}
	limit = normalizeHistoryLimit(limit)
	var rows []domain.HostHistory
	if err := query.Order("sampled_at desc").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// AddNetworkStatusHistoryDAO stores a network status snapshot.
func AddNetworkStatusHistoryDAO(history domain.NetworkStatusHistory) error {
	return database.DB.Create(&history).Error
}

// ListNetworkStatusHistoryDAO returns network status history entries ordered by sampled_at desc.
func ListNetworkStatusHistoryDAO(from, to *time.Time, limit int) ([]domain.NetworkStatusHistory, error) {
	query := database.DB.Model(&domain.NetworkStatusHistory{})
	if from != nil {
		query = query.Where("sampled_at >= ?", *from)
	}
	if to != nil {
		query = query.Where("sampled_at <= ?", *to)
	}
	limit = normalizeHistoryLimit(limit)
	var rows []domain.NetworkStatusHistory
	if err := query.Order("sampled_at desc").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
