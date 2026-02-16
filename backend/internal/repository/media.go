package repository

import (
	"errors"

	"gorm.io/gorm"
	"nagare/internal/database"
	"nagare/internal/model"
)

// GetAllMediaDAO retrieves all media
func GetAllMediaDAO() ([]model.Media, error) {
	var media []model.Media
	if err := database.DB.Find(&media).Error; err != nil {
		return nil, err
	}
	return media, nil
}

// SearchMediaDAO retrieves media by filter
func SearchMediaDAO(filter model.MediaFilter) ([]model.Media, error) {
	query := database.DB.Model(&model.Media{})
	if filter.Query != "" {
		query = query.Where("name LIKE ? OR type LIKE ? OR target LIKE ? OR description LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%", "%"+filter.Query+"%", "%"+filter.Query+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.TypeID != nil {
		query = query.Where("media_type_id = ?", *filter.TypeID)
	}
	query = applySort(query, filter.SortBy, filter.SortOrder, map[string]string{
		"name":       "name",
		"status":     "status",
		"created_at": "created_at",
		"updated_at": "updated_at",
		"id":         "id",
	}, "id desc")
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}
	var media []model.Media
	if err := query.Find(&media).Error; err != nil {
		return nil, err
	}
	return media, nil
}

// CountMediaDAO returns total count for media by filter
func CountMediaDAO(filter model.MediaFilter) (int64, error) {
	query := database.DB.Model(&model.Media{})
	if filter.Query != "" {
		query = query.Where("name LIKE ? OR type LIKE ? OR target LIKE ? OR description LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%", "%"+filter.Query+"%", "%"+filter.Query+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.TypeID != nil {
		query = query.Where("media_type_id = ?", *filter.TypeID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// GetMediaByIDDAO retrieves media by ID
func GetMediaByIDDAO(id uint) (model.Media, error) {
	var media model.Media
	err := database.DB.First(&media, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return media, model.ErrNotFound
	}
	return media, err
}

// GetMediaByTypeIDDAO retrieves media by media type ID
func GetMediaByTypeIDDAO(typeID uint) ([]model.Media, error) {
	var media []model.Media
	if err := database.DB.Where("media_type_id = ?", typeID).Find(&media).Error; err != nil {
		return nil, err
	}
	return media, nil
}

// AddMediaDAO creates a new media
func AddMediaDAO(media model.Media) error {
	return database.DB.Create(&media).Error
}

// UpdateMediaDAO updates media by ID
func UpdateMediaDAO(id uint, media model.Media) error {
	return database.DB.Model(&model.Media{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":          media.Name,
		"type":          media.Type,
		"media_type_id": media.MediaTypeID,
		"target":        media.Target,
		"params":        media.Params,
		"enabled":       media.Enabled,
		"status":        media.Status,
		"description":   media.Description,
	}).Error
}

// DeleteMediaByIDDAO deletes media by ID
func DeleteMediaByIDDAO(id uint) error {
	return database.DB.Delete(&model.Media{}, id).Error
}

// UpdateMediaStatusDAO updates only status for media
func UpdateMediaStatusDAO(id uint, status int) error {
	return database.DB.Model(&model.Media{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateMediaTypeKeyForMediaDAO updates cached type key for media by media type ID
func UpdateMediaTypeKeyForMediaDAO(typeID uint, key string) error {
	return database.DB.Model(&model.Media{}).Where("media_type_id = ?", typeID).Update("type", key).Error
}

// UpdateMediaParamsDAO updates media params by ID
func UpdateMediaParamsDAO(id uint, params map[string]string) error {
	return database.DB.Model(&model.Media{}).Where("id = ?", id).Update("params", params).Error
}

// UpdateMediaParamsAndTargetDAO updates media params and target by ID
func UpdateMediaParamsAndTargetDAO(id uint, params map[string]string, target string) error {
	return database.DB.Model(&model.Media{}).Where("id = ?", id).Updates(map[string]interface{}{
		"params": params,
		"target": target,
	}).Error
}
