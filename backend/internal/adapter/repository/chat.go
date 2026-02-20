package repository

import (
	"errors"

	"nagare/internal/core/domain"
	"nagare/internal/database"

	"gorm.io/gorm"
)

// GetAllChatsDAO retrieves all chat messages from the database
func GetAllChatsDAO() ([]domain.Chat, error) {
	return GetChatsWithLimitDAO(10, 0)
}

// GetChatsWithLimitDAO retrieves chat messages with pagination
func GetChatsWithLimitDAO(limit, offset int) ([]domain.Chat, error) {
	var chats []domain.Chat
	if err := database.DB.Order("id DESC").Limit(limit).Offset(offset).Find(&chats).Error; err != nil {
		return []domain.Chat{}, nil
	}
	if chats == nil {
		chats = []domain.Chat{}
	}
	return chats, nil
}

// SearchChatsDAO retrieves chat messages by filter with pagination
func SearchChatsDAO(filter domain.ChatFilter) ([]domain.Chat, error) {
	query := database.DB.Model(&domain.Chat{})
	if filter.Query != "" {
		query = query.Where("content LIKE ?", "%"+filter.Query+"%")
	}
	if filter.Role != nil {
		query = query.Where("role = ?", *filter.Role)
	}
	if filter.ProviderID != nil {
		query = query.Where("provider_id = ?", *filter.ProviderID)
	}
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.Model != nil {
		query = query.Where("model = ?", *filter.Model)
	}
	limit := filter.Limit
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	var chats []domain.Chat
	if err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&chats).Error; err != nil {
		return []domain.Chat{}, nil
	}
	if chats == nil {
		chats = []domain.Chat{}
	}
	return chats, nil
}

// GetChatByIDDAO retrieves a chat by ID
func GetChatByIDDAO(id int) (domain.Chat, error) {
	var chat domain.Chat
	err := database.DB.First(&chat, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return chat, domain.ErrNotFound
	}
	return chat, err
}

// AddChatDAO creates a new chat message
func AddChatDAO(c domain.Chat) error {
	return database.DB.Create(&c).Error
}

// DeleteChatByIDDAO deletes a chat by ID
func DeleteChatByIDDAO(id int) error {
	return database.DB.Delete(&domain.Chat{}, id).Error
}

// UpdateChatDAO updates a chat by ID
func UpdateChatDAO(id int, c domain.Chat) error {
	return database.DB.Model(&domain.Chat{}).Where("id = ?", id).Updates(map[string]interface{}{
		"user_id":     c.UserID,
		"provider_id": c.ProviderID,
		"role":        c.Role,
		"content":     c.Content,
		"model":       c.LLMModel,
	}).Error
}
