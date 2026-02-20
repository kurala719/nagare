package service

import (
	"nagare/internal/adapter/repository"
	"nagare/internal/core/domain"
)

type QQWhitelistReq struct {
	QQIdentifier string `json:"qq_identifier"`
	Type         int    `json:"type"` // 0 = user, 1 = group
	Nickname     string `json:"nickname"`
	CanCommand   int    `json:"can_command"`
	CanReceive   int    `json:"can_receive"`
	Enabled      int    `json:"enabled"`
	Comment      string `json:"comment"`
}

type QQWhitelistResp struct {
	ID           uint   `json:"id"`
	QQIdentifier string `json:"qq_identifier"`
	Type         int    `json:"type"`
	Nickname     string `json:"nickname"`
	CanCommand   int    `json:"can_command"`
	CanReceive   int    `json:"can_receive"`
	Enabled      int    `json:"enabled"`
	Comment      string `json:"comment"`
	CreatedAt    string `json:"created_at"`
}

// AddQQWhitelistServ adds a new QQ user or group to the whitelist
func AddQQWhitelistServ(req QQWhitelistReq) (QQWhitelistResp, error) {
	whitelist := domain.QQWhitelist{
		QQIdentifier: req.QQIdentifier,
		Type:         req.Type,
		Nickname:     req.Nickname,
		CanCommand:   req.CanCommand,
		CanReceive:   req.CanReceive,
		Enabled:      req.Enabled,
		Comment:      req.Comment,
	}

	result, err := repository.AddQQWhitelistDAO(whitelist)
	if err != nil {
		return QQWhitelistResp{}, err
	}

	return qqWhitelistToResp(result), nil
}

// GetQQWhitelistServ retrieves a whitelist entry
func GetQQWhitelistServ(qqID string, whitelistType int) (QQWhitelistResp, error) {
	whitelist, err := repository.GetQQWhitelistDAO(qqID, whitelistType)
	if err != nil {
		return QQWhitelistResp{}, err
	}

	return qqWhitelistToResp(whitelist), nil
}

// UpdateQQWhitelistServ updates a whitelist entry
func UpdateQQWhitelistServ(id uint, req QQWhitelistReq) error {
	whitelist := domain.QQWhitelist{
		QQIdentifier: req.QQIdentifier,
		Type:         req.Type,
		Nickname:     req.Nickname,
		CanCommand:   req.CanCommand,
		CanReceive:   req.CanReceive,
		Enabled:      req.Enabled,
		Comment:      req.Comment,
	}

	return repository.UpdateQQWhitelistDAO(id, whitelist)
}

// DeleteQQWhitelistServ deletes a whitelist entry
func DeleteQQWhitelistServ(id uint) error {
	return repository.DeleteQQWhitelistDAO(id)
}

// ListQQWhitelistServ lists all QQ whitelist entries
func ListQQWhitelistServ(whitelistType *int, enabled *int, limit int, offset int) ([]QQWhitelistResp, error) {
	whitelist, err := repository.ListQQWhitelistDAO(whitelistType, enabled, limit, offset)
	if err != nil {
		return nil, err
	}

	result := make([]QQWhitelistResp, 0, len(whitelist))
	for _, w := range whitelist {
		result = append(result, qqWhitelistToResp(w))
	}

	return result, nil
}

// CountQQWhitelistServ counts QQ whitelist entries
func CountQQWhitelistServ(whitelistType *int, enabled *int) (int64, error) {
	return repository.CountQQWhitelistDAO(whitelistType, enabled)
}

func qqWhitelistToResp(w domain.QQWhitelist) QQWhitelistResp {
	return QQWhitelistResp{
		ID:           w.ID,
		QQIdentifier: w.QQIdentifier,
		Type:         w.Type,
		Nickname:     w.Nickname,
		CanCommand:   w.CanCommand,
		CanReceive:   w.CanReceive,
		Enabled:      w.Enabled,
		Comment:      w.Comment,
		CreatedAt:    w.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
