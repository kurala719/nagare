package repository

import (
	"nagare/internal/database"
	"nagare/internal/model"
)

// AddSiteMessageDAO creates a new site message
func AddSiteMessageDAO(msg model.SiteMessage) error {
	return database.DB.Create(&msg).Error
}

// GetSiteMessagesDAO retrieves site messages with filters
func GetSiteMessagesDAO(userID *uint, unreadOnly bool, limit, offset int) ([]model.SiteMessage, error) {
	var messages []model.SiteMessage
	query := database.DB.Order("created_at desc")
	if userID != nil {
		query = query.Where("user_id = ? OR user_id IS NULL", *userID)
	} else {
		query = query.Where("user_id IS NULL")
	}

	if unreadOnly {
		query = query.Where("is_read = 0")
	}
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Find(&messages).Error
	return messages, err
}

// CountSiteMessagesDAO returns the total number of messages for a user
func CountSiteMessagesDAO(userID *uint) (int64, error) {
	var count int64
	query := database.DB.Model(&model.SiteMessage{})
	if userID != nil {
		query = query.Where("user_id = ? OR user_id IS NULL", *userID)
	} else {
		query = query.Where("user_id IS NULL")
	}
	err := query.Count(&count).Error
	return count, err
}

// CountUnreadSiteMessagesDAO returns the number of unread messages for a user
func CountUnreadSiteMessagesDAO(userID *uint) (int64, error) {
	var count int64
	query := database.DB.Model(&model.SiteMessage{}).Where("is_read = 0")
	if userID != nil {
		query = query.Where("user_id = ? OR user_id IS NULL", *userID)
	} else {
		query = query.Where("user_id IS NULL")
	}
	
	err := query.Count(&count).Error
	return count, err
}

// MarkSiteMessageAsReadDAO marks a message as read
func MarkSiteMessageAsReadDAO(id uint) error {
	return database.DB.Model(&model.SiteMessage{}).Where("id = ?", id).Update("is_read", 1).Error
}

// MarkAllSiteMessagesAsReadDAO marks all messages for a user as read
func MarkAllSiteMessagesAsReadDAO(userID *uint) error {
	query := database.DB.Model(&model.SiteMessage{}).Where("is_read = 0")
	if userID != nil {
		query = query.Where("user_id = ? OR user_id IS NULL", *userID)
	} else {
		query = query.Where("user_id IS NULL")
	}
	return query.Update("is_read", 1).Error
}

// DeleteSiteMessageDAO deletes a site message
func DeleteSiteMessageDAO(id uint) error {
	return database.DB.Delete(&model.SiteMessage{}, id).Error
}
