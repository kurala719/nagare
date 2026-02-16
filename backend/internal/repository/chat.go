package repository

import (
	"errors"

	"gorm.io/gorm"
	"nagare/internal/database"
	"nagare/internal/model"
)

// GetAllChatsDAO retrieves all chat messages from the database
func GetAllChatsDAO() ([]model.Chat, error) {
	return GetChatsWithLimitDAO(10, 0)
}

// GetChatsWithLimitDAO retrieves chat messages with pagination
func GetChatsWithLimitDAO(limit, offset int) ([]model.Chat, error) {
	var chats []model.Chat
	if err := database.DB.Order("id DESC").Limit(limit).Offset(offset).Find(&chats).Error; err != nil {
		return []model.Chat{}, nil
	}
	if chats == nil {
		chats = []model.Chat{}
	}
	return chats, nil
}

// SearchChatsDAO retrieves chat messages by filter with pagination
func SearchChatsDAO(filter model.ChatFilter) ([]model.Chat, error) {
	query := database.DB.Model(&model.Chat{})
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
	var chats []model.Chat
	if err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&chats).Error; err != nil {
		return []model.Chat{}, nil
	}
	if chats == nil {
		chats = []model.Chat{}
	}
	return chats, nil
}

// GetChatByIDDAO retrieves a chat by ID
func GetChatByIDDAO(id int) (model.Chat, error) {
	var chat model.Chat
	err := database.DB.First(&chat, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return chat, model.ErrNotFound
	}
	return chat, err
}

// AddChatDAO creates a new chat message
func AddChatDAO(c model.Chat) error {
	return database.DB.Create(&c).Error
}

// DeleteChatByIDDAO deletes a chat by ID
func DeleteChatByIDDAO(id int) error {
	return database.DB.Delete(&model.Chat{}, id).Error
}

// UpdateChatDAO updates a chat by ID
func UpdateChatDAO(id int, c model.Chat) error {
	return database.DB.Model(&model.Chat{}).Where("id = ?", id).Updates(map[string]interface{}{
		"user_id":     c.UserID,
		"provider_id": c.ProviderID,
		"role":        c.Role,
		"content":     c.Content,
		"model":       c.LLMModel,
	}).Error
}
