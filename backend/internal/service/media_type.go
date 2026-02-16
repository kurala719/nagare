package service

import (
	"fmt"

	"nagare/internal/model"
	"nagare/internal/repository"
)

// MediaTypeReq represents a media type request
type MediaTypeReq struct {
	Name        string                   `json:"name" binding:"required"`
	Key         string                   `json:"key" binding:"required"`
	Enabled     int                      `json:"enabled"`
	Description string                   `json:"description"`
	Template    string                   `json:"template"`
	Fields      []model.MediaParamField `json:"fields"`
}

// MediaTypeResp represents a media type response
type MediaTypeResp struct {
	ID          int                      `json:"id"`
	Name        string                   `json:"name"`
	Key         string                   `json:"key"`
	Enabled     int                      `json:"enabled"`
	Status      int                      `json:"status"`
	Description string                   `json:"description"`
	Template    string                   `json:"template"`
	Fields      []model.MediaParamField `json:"fields"`
}

func GetAllMediaTypesServ() ([]MediaTypeResp, error) {
	types, err := repository.GetAllMediaTypesDAO()
	if err != nil {
		return nil, fmt.Errorf("failed to get media types: %w", err)
	}
	result := make([]MediaTypeResp, 0, len(types))
	for _, t := range types {
		result = append(result, mediaTypeToResp(t))
	}
	return result, nil
}

func SearchMediaTypesServ(filter model.MediaTypeFilter) ([]MediaTypeResp, error) {
	types, err := repository.SearchMediaTypesDAO(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search media types: %w", err)
	}
	result := make([]MediaTypeResp, 0, len(types))
	for _, t := range types {
		result = append(result, mediaTypeToResp(t))
	}
	return result, nil
}

// CountMediaTypesServ returns total count for media types by filter
func CountMediaTypesServ(filter model.MediaTypeFilter) (int64, error) {
	return repository.CountMediaTypesDAO(filter)
}

func GetMediaTypeByIDServ(id uint) (MediaTypeResp, error) {
	mediaType, err := repository.GetMediaTypeByIDDAO(id)
	if err != nil {
		return MediaTypeResp{}, fmt.Errorf("failed to get media type: %w", err)
	}
	return mediaTypeToResp(mediaType), nil
}

func AddMediaTypeServ(req MediaTypeReq) (MediaTypeResp, error) {
	mediaType := model.MediaType{
		Name:        req.Name,
		Key:         req.Key,
		Enabled:     req.Enabled,
		Status:      determineMediaTypeStatus(model.MediaType{Enabled: req.Enabled, Name: req.Name, Key: req.Key}),
		Description: req.Description,
		Template:    req.Template,
		Fields:      req.Fields,
	}
	if err := repository.AddMediaTypeDAO(mediaType); err != nil {
		return MediaTypeResp{}, fmt.Errorf("failed to add media type: %w", err)
	}
	return mediaTypeToResp(mediaType), nil
}

func UpdateMediaTypeServ(id uint, req MediaTypeReq) error {
	updated := model.MediaType{
		Name:        req.Name,
		Key:         req.Key,
		Enabled:     req.Enabled,
		Status:      determineMediaTypeStatus(model.MediaType{Enabled: req.Enabled, Name: req.Name, Key: req.Key}),
		Description: req.Description,
		Template:    req.Template,
		Fields:      req.Fields,
	}
	if err := repository.UpdateMediaTypeDAO(id, updated); err != nil {
		return err
	}
	_ = repository.UpdateMediaTypeKeyForMediaDAO(id, updated.Key)
	_, _ = recomputeMediaTypeStatus(id)
	mediaList, _ := repository.GetMediaByTypeIDDAO(id)
	for _, media := range mediaList {
		_, _ = recomputeMediaStatus(media.ID)
		actions, _ := repository.GetActionsByMediaIDDAO(media.ID)
		for _, action := range actions {
			_, _ = recomputeActionStatus(action.ID)
		}
	}
	return nil
}

func DeleteMediaTypeByIDServ(id uint) error {
	return repository.DeleteMediaTypeByIDDAO(id)
}

func mediaTypeToResp(mediaType model.MediaType) MediaTypeResp {
	return MediaTypeResp{
		ID:          int(mediaType.ID),
		Name:        mediaType.Name,
		Key:         mediaType.Key,
		Enabled:     mediaType.Enabled,
		Status:      mediaType.Status,
		Description: mediaType.Description,
		Template:    mediaType.Template,
		Fields:      mediaType.Fields,
	}
}
