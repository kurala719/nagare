package service

import (
	"fmt"
	"nagare/internal/adapter/repository"
	"nagare/internal/core/domain"
)

// SiteMessageResp represents a site message response
type SiteMessageResp struct {
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Type      string `json:"type"`
	Severity  int    `json:"severity"`
	IsRead    int    `json:"is_read"`
	CreatedAt string `json:"created_at"`
}

// CreateSiteMessageServ creates and broadcasts a new site message
func CreateSiteMessageServ(title, content, msgType string, severity int, userID *uint) error {
	msg := domain.SiteMessage{
		Title:    title,
		Content:  content,
		Type:     msgType,
		Severity: severity,
		UserID:   userID,
		IsRead:   0,
	}

	if err := repository.AddSiteMessageDAO(msg); err != nil {
		return fmt.Errorf("failed to save site message: %w", err)
	}

	// Broadcast real-time
	BroadcastMessage(map[string]interface{}{
		"event": "site_message",
		"data": SiteMessageResp{
			ID:        msg.ID,
			Title:     msg.Title,
			Content:   msg.Content,
			Type:      msg.Type,
			Severity:  msg.Severity,
			IsRead:    msg.IsRead,
			CreatedAt: msg.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	})

	return nil
}

// GetSiteMessagesServ retrieves site messages for a user
func GetSiteMessagesServ(userID *uint, unreadOnly bool, limit, offset int) ([]SiteMessageResp, error) {
	messages, err := repository.GetSiteMessagesDAO(userID, unreadOnly, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get site messages: %w", err)
	}

	result := make([]SiteMessageResp, 0, len(messages))
	for _, m := range messages {
		result = append(result, SiteMessageResp{
			ID:        m.ID,
			Title:     m.Title,
			Content:   m.Content,
			Type:      m.Type,
			Severity:  m.Severity,
			IsRead:    m.IsRead,
			CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return result, nil
}

// MarkSiteMessageAsReadServ marks a message as read
func MarkSiteMessageAsReadServ(id uint) error {
	return repository.MarkSiteMessageAsReadDAO(id)
}

// MarkAllSiteMessagesAsReadServ marks all messages as read
func MarkAllSiteMessagesAsReadServ(userID *uint) error {
	return repository.MarkAllSiteMessagesAsReadDAO(userID)
}

// GetUnreadCountServ returns the unread count
func GetUnreadCountServ(userID *uint) (int64, error) {
	return repository.CountUnreadSiteMessagesDAO(userID)
}

// GetTotalMessagesCountServ returns the total count
func GetTotalMessagesCountServ(userID *uint) (int64, error) {
	return repository.CountSiteMessagesDAO(userID)
}

// DeleteSiteMessageServ deletes a message
func DeleteSiteMessageServ(id uint) error {
	return repository.DeleteSiteMessageDAO(id)
}
