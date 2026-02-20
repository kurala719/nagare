package service

import (
	"fmt"
	"strings"

	"nagare/internal/adapter/repository"
	"nagare/internal/core/domain"
)

// MediaReq represents a media request
type MediaReq struct {
	Name        string                 `json:"name" binding:"required"`
	Type        string                 `json:"type" binding:"required"`
	Target      string                 `json:"target" binding:"required"`
	Params      map[string]interface{} `json:"params"`
	Enabled     int                    `json:"enabled"`
	Description string                 `json:"description"`
}

// MediaResp represents a media response
type MediaResp struct {
	ID          int               `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Target      string            `json:"target"`
	Params      map[string]string `json:"params"`
	Enabled     int               `json:"enabled"`
	Status      int               `json:"status"`
	Description string            `json:"description"`
}

func GetAllMediaServ() ([]MediaResp, error) {
	media, err := repository.GetAllMediaDAO()
	if err != nil {
		return nil, fmt.Errorf("failed to get media: %w", err)
	}
	result := make([]MediaResp, 0, len(media))
	for _, m := range media {
		result = append(result, mediaToResp(m))
	}
	return result, nil
}

func SearchMediaServ(filter domain.MediaFilter) ([]MediaResp, error) {
	media, err := repository.SearchMediaDAO(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search media: %w", err)
	}
	result := make([]MediaResp, 0, len(media))
	for _, m := range media {
		result = append(result, mediaToResp(m))
	}
	return result, nil
}

// CountMediaServ returns total count for media by filter
func CountMediaServ(filter domain.MediaFilter) (int64, error) {
	return repository.CountMediaDAO(filter)
}

func GetMediaByIDServ(id uint) (MediaResp, error) {
	media, err := repository.GetMediaByIDDAO(id)
	if err != nil {
		return MediaResp{}, fmt.Errorf("failed to get media: %w", err)
	}
	return mediaToResp(media), nil
}

func TestMediaServ(id uint) error {
	media, err := repository.GetMediaByIDDAO(id)
	if err != nil {
		return err
	}
	if media.Enabled == 0 {
		return fmt.Errorf("media is disabled")
	}

	testMessage := "Nagare Media Test Connection: Your notification system is working correctly."
	return SendIMReply(media.Type, media.Target, testMessage)
}

func AddMediaServ(req MediaReq) (MediaResp, error) {
	params := coerceMediaParams(req.Params)
	media := domain.Media{
		Name:        req.Name,
		Type:        req.Type,
		Target:      req.Target,
		Params:      params,
		Enabled:     req.Enabled,
		Status:      determineMediaStatus(domain.Media{Enabled: req.Enabled, Type: req.Type, Target: req.Target}),
		Description: req.Description,
	}
	if err := repository.AddMediaDAO(media); err != nil {
		return MediaResp{}, fmt.Errorf("failed to add media: %w", err)
	}
	return mediaToResp(media), nil
}

func UpdateMediaServ(id uint, req MediaReq) error {
	existing, err := repository.GetMediaByIDDAO(id)
	if err != nil {
		return err
	}
	params := coerceMediaParams(req.Params)
	updated := domain.Media{
		Name:        req.Name,
		Type:        req.Type,
		Target:      req.Target,
		Params:      params,
		Enabled:     req.Enabled,
		Status:      existing.Status,
		Description: req.Description,
	}
	// Preserve status unless enabled state, type or target changed
	if req.Enabled != existing.Enabled || req.Type != existing.Type || req.Target != existing.Target {
		updated.Status = determineMediaStatus(domain.Media{Enabled: req.Enabled, Type: req.Type, Target: req.Target})
	}
	if err := repository.UpdateMediaDAO(id, updated); err != nil {
		return err
	}
	_, _ = recomputeMediaStatus(id)
	return nil
}

func DeleteMediaByIDServ(id uint) error {
	return repository.DeleteMediaByIDDAO(id)
}

func mediaToResp(media domain.Media) MediaResp {
	return MediaResp{
		ID:          int(media.ID),
		Name:        media.Name,
		Type:        media.Type,
		Target:      media.Target,
		Params:      media.Params,
		Enabled:     media.Enabled,
		Status:      media.Status,
		Description: media.Description,
	}
}

func coerceMediaParams(incoming map[string]interface{}) map[string]string {
	params := map[string]string{}
	for key, value := range incoming {
		if value == nil {
			params[key] = ""
			continue
		}
		params[key] = strings.TrimSpace(fmt.Sprint(value))
	}
	return params
}

// BackfillMediaParamsAndTargetsServ is now a no-op as MediaType is removed
func BackfillMediaParamsAndTargetsServ() (int, int, error) {
	return 0, 0, nil
}
