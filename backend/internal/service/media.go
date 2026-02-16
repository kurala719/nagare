package service

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"nagare/internal/model"
	"nagare/internal/repository"
)

// MediaReq represents a media request
type MediaReq struct {
	Name        string                 `json:"name" binding:"required"`
	Type        string                 `json:"type"`
	MediaTypeID uint                   `json:"media_type_id"`
	Target      string                 `json:"target"`
	Params      map[string]interface{} `json:"params"`
	Enabled     int                    `json:"enabled"`
	Description string                 `json:"description"`
}

// MediaResp represents a media response
type MediaResp struct {
	ID          int               `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	MediaTypeID uint              `json:"media_type_id"`
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

func SearchMediaServ(filter model.MediaFilter) ([]MediaResp, error) {
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
func CountMediaServ(filter model.MediaFilter) (int64, error) {
	return repository.CountMediaDAO(filter)
}

func GetMediaByIDServ(id uint) (MediaResp, error) {
	media, err := repository.GetMediaByIDDAO(id)
	if err != nil {
		return MediaResp{}, fmt.Errorf("failed to get media: %w", err)
	}
	return mediaToResp(media), nil
}

func AddMediaServ(req MediaReq) (MediaResp, error) {
	_, mediaTypeKey, mediaTypeStatus, params, target, err := resolveMediaTypeAndTarget(req)
	if err != nil {
		return MediaResp{}, err
	}
	media := model.Media{
		Name:        req.Name,
		Type:        mediaTypeKey,
		MediaTypeID: req.MediaTypeID,
		Target:      target,
		Params:      params,
		Enabled:     req.Enabled,
		Status:      determineMediaStatus(model.Media{Enabled: req.Enabled, Type: mediaTypeKey, Target: target}),
		Description: req.Description,
	}
	if mediaTypeStatus == 2 && req.Enabled != 0 {
		media.Status = 2
	}
	if err := repository.AddMediaDAO(media); err != nil {
		return MediaResp{}, fmt.Errorf("failed to add media: %w", err)
	}
	return mediaToResp(media), nil
}

func UpdateMediaServ(id uint, req MediaReq) error {
	_, mediaTypeKey, mediaTypeStatus, params, target, err := resolveMediaTypeAndTarget(req)
	if err != nil {
		return err
	}
	updated := model.Media{
		Name:        req.Name,
		Type:        mediaTypeKey,
		MediaTypeID: req.MediaTypeID,
		Target:      target,
		Params:      params,
		Enabled:     req.Enabled,
		Status:      determineMediaStatus(model.Media{Enabled: req.Enabled, Type: mediaTypeKey, Target: target}),
		Description: req.Description,
	}
	if mediaTypeStatus == 2 && req.Enabled != 0 {
		updated.Status = 2
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

func mediaToResp(media model.Media) MediaResp {
	return MediaResp{
		ID:          int(media.ID),
		Name:        media.Name,
		Type:        media.Type,
		MediaTypeID: media.MediaTypeID,
		Target:      media.Target,
		Params:      media.Params,
		Enabled:     media.Enabled,
		Status:      media.Status,
		Description: media.Description,
	}
}

func resolveMediaTypeAndTarget(req MediaReq) (model.MediaType, string, int, map[string]string, string, error) {
	if req.MediaTypeID == 0 {
		return model.MediaType{}, "", 2, nil, "", fmt.Errorf("media_type_id is required")
	}
	mediaType, err := repository.GetMediaTypeByIDDAO(req.MediaTypeID)
	if err != nil {
		return model.MediaType{}, "", 2, nil, "", fmt.Errorf("invalid media_type_id")
	}
	mediaTypeKey := mediaType.Key
	mediaTypeStatus := mediaType.Status
	if mediaType.Enabled == 0 {
		mediaTypeStatus = 2
	}
	params, err := normalizeMediaParams(mediaType, coerceMediaParams(req.Params))
	if err != nil {
		return mediaType, mediaTypeKey, mediaTypeStatus, nil, "", err
	}
	var target string
	if strings.TrimSpace(mediaType.Template) != "" {
		target, err = renderMediaTemplate(mediaType.Template, params)
		if err != nil {
			return mediaType, mediaTypeKey, mediaTypeStatus, nil, "", err
		}
	} else {
		target = strings.TrimSpace(req.Target)
		if target == "" {
			return mediaType, mediaTypeKey, mediaTypeStatus, nil, "", fmt.Errorf("target is required")
		}
	}
	return mediaType, mediaTypeKey, mediaTypeStatus, params, target, nil
}

func normalizeMediaParams(mediaType model.MediaType, incoming map[string]string) (map[string]string, error) {
	params := map[string]string{}
	for key, value := range incoming {
		params[key] = value
	}
	for _, field := range mediaType.Fields {
		key := strings.TrimSpace(field.Key)
		if key == "" {
			continue
		}
		value := strings.TrimSpace(params[key])
		if value == "" && field.Default != "" {
			value = field.Default
			params[key] = value
		}
		if field.Required && strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("missing required param: %s", key)
		}
		if field.Pattern != "" && value != "" {
			re, err := regexp.Compile(field.Pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid pattern for param %s", key)
			}
			if !re.MatchString(value) {
				return nil, fmt.Errorf("param %s does not match pattern", key)
			}
		}
	}
	return params, nil
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
func renderMediaTemplate(template string, params map[string]string) (string, error) {
	re := regexp.MustCompile(`{{\s*([a-zA-Z0-9_]+)\s*}}`)
	missing := map[string]struct{}{}
	result := re.ReplaceAllStringFunc(template, func(match string) string {
		key := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(match, "{{"), "}}"))
		value, ok := params[key]
		if !ok || strings.TrimSpace(value) == "" {
			missing[key] = struct{}{}
			return match
		}
		return value
	})
	if len(missing) > 0 {
		return "", fmt.Errorf("template has unresolved params")
	}
	return strings.TrimSpace(result), nil
}

func BackfillMediaParamsAndTargetsServ() (int, int, error) {
	mediaList, err := repository.GetAllMediaDAO()
	if err != nil {
		return 0, 0, err
	}
	mediaTypes, err := repository.GetAllMediaTypesDAO()
	if err != nil {
		return 0, 0, err
	}
	mediaTypeMap := map[uint]model.MediaType{}
	for _, mediaType := range mediaTypes {
		mediaTypeMap[mediaType.ID] = mediaType
	}
	updated := 0
	skipped := 0
	for _, media := range mediaList {
		mediaType, ok := mediaTypeMap[media.MediaTypeID]
		if !ok {
			skipped++
			continue
		}
		normalized, err := normalizeMediaParams(mediaType, media.Params)
		if err != nil {
			skipped++
			continue
		}
		paramsChanged := !reflect.DeepEqual(media.Params, normalized)
		target := media.Target
		targetChanged := false
		if strings.TrimSpace(mediaType.Template) != "" {
			rendered, err := renderMediaTemplate(mediaType.Template, normalized)
			if err == nil && rendered != "" && rendered != target {
				target = rendered
				targetChanged = true
			}
		}
		if paramsChanged || targetChanged {
			if targetChanged {
				err = repository.UpdateMediaParamsAndTargetDAO(media.ID, normalized, target)
			} else {
				err = repository.UpdateMediaParamsDAO(media.ID, normalized)
			}
			if err != nil {
				return updated, skipped, err
			}
			updated++
		}
	}
	return updated, skipped, nil
}
