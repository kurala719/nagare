package repository

import (
	"fmt"

	"nagare/internal/core/domain"
	"nagare/internal/database"
)

// AddQQWhitelistDAO adds a new QQ user or group to the whitelist
func AddQQWhitelistDAO(whitelist domain.QQWhitelist) (domain.QQWhitelist, error) {
	if err := database.DB.Create(&whitelist).Error; err != nil {
		return domain.QQWhitelist{}, fmt.Errorf("failed to add QQ whitelist: %w", err)
	}
	return whitelist, nil
}

// GetQQWhitelistDAO retrieves a whitelist entry by QQ ID and type
func GetQQWhitelistDAO(qqID string, whitelistType int) (domain.QQWhitelist, error) {
	var whitelist domain.QQWhitelist
	if err := database.DB.Where("qq_identifier = ? AND type = ?", qqID, whitelistType).First(&whitelist).Error; err != nil {
		return domain.QQWhitelist{}, err
	}
	return whitelist, nil
}

// UpdateQQWhitelistDAO updates a whitelist entry
func UpdateQQWhitelistDAO(id uint, whitelist domain.QQWhitelist) error {
	whitelist.ID = id
	return database.DB.Model(&whitelist).Updates(whitelist).Error
}

// DeleteQQWhitelistDAO deletes a whitelist entry
func DeleteQQWhitelistDAO(id uint) error {
	return database.DB.Delete(&domain.QQWhitelist{}, id).Error
}

// ListQQWhitelistDAO lists all QQ whitelist entries with optional filters
func ListQQWhitelistDAO(whitelistType *int, enabled *int, limit int, offset int) ([]domain.QQWhitelist, error) {
	var whitelist []domain.QQWhitelist
	query := database.DB

	if whitelistType != nil {
		query = query.Where("type = ?", *whitelistType)
	}
	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}

	if err := query.Limit(limit).Offset(offset).Find(&whitelist).Error; err != nil {
		return nil, err
	}
	return whitelist, nil
}

// CountQQWhitelistDAO counts QQ whitelist entries with optional filters
func CountQQWhitelistDAO(whitelistType *int, enabled *int) (int64, error) {
	var count int64
	query := database.DB

	if whitelistType != nil {
		query = query.Where("type = ?", *whitelistType)
	}
	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}

	if err := query.Model(&domain.QQWhitelist{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
