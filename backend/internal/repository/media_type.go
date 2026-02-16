package repository

import (
	"errors"

	"gorm.io/gorm"
	"nagare/internal/database"
	"nagare/internal/model"
)

// GetAllMediaTypesDAO retrieves all media types
func GetAllMediaTypesDAO() ([]model.MediaType, error) {
	var types []model.MediaType
	if err := database.DB.Find(&types).Error; err != nil {
		return nil, err
	}
	return types, nil
}

// SearchMediaTypesDAO retrieves media types by filter
func SearchMediaTypesDAO(filter model.MediaTypeFilter) ([]model.MediaType, error) {
	query := database.DB.Model(&model.MediaType{})
	if filter.Query != "" {
		query = query.Where("name LIKE ? OR `key` LIKE ? OR description LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%", "%"+filter.Query+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
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
	var types []model.MediaType
	if err := query.Find(&types).Error; err != nil {
		return nil, err
	}
	return types, nil
}

// CountMediaTypesDAO returns total count for media types by filter
func CountMediaTypesDAO(filter model.MediaTypeFilter) (int64, error) {
	query := database.DB.Model(&model.MediaType{})
	if filter.Query != "" {
		query = query.Where("name LIKE ? OR `key` LIKE ? OR description LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%", "%"+filter.Query+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// GetMediaTypeByIDDAO retrieves media type by ID
func GetMediaTypeByIDDAO(id uint) (model.MediaType, error) {
	var mediaType model.MediaType
	err := database.DB.First(&mediaType, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return mediaType, model.ErrNotFound
	}
	return mediaType, err
}

// AddMediaTypeDAO creates a new media type
func AddMediaTypeDAO(mediaType model.MediaType) error {
	return database.DB.Create(&mediaType).Error
}

// UpdateMediaTypeDAO updates a media type by ID
func UpdateMediaTypeDAO(id uint, mediaType model.MediaType) error {
	return database.DB.Model(&model.MediaType{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":        mediaType.Name,
		"key":         mediaType.Key,
		"enabled":     mediaType.Enabled,
		"status":      mediaType.Status,
		"description": mediaType.Description,
		"template":    mediaType.Template,
		"fields":      mediaType.Fields,
	}).Error
}

// DeleteMediaTypeByIDDAO deletes a media type by ID
func DeleteMediaTypeByIDDAO(id uint) error {
	return database.DB.Delete(&model.MediaType{}, id).Error
}

// UpdateMediaTypeStatusDAO updates only status for media type
func UpdateMediaTypeStatusDAO(id uint, status int) error {
	return database.DB.Model(&model.MediaType{}).Where("id = ?", id).Update("status", status).Error
}
