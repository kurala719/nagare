package repository

import (
	"errors"

	"gorm.io/gorm"
	"nagare/internal/database"
	"nagare/internal/model"
)

// GetAllSitesDAO retrieves all sites
func GetAllSitesDAO() ([]model.Site, error) {
	var sites []model.Site
	if err := database.DB.Find(&sites).Error; err != nil {
		return nil, err
	}
	return sites, nil
}

// SearchSitesDAO retrieves sites by filter
func SearchSitesDAO(filter model.SiteFilter) ([]model.Site, error) {
	query := database.DB.Model(&model.Site{})
	if filter.Query != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%")
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
	var sites []model.Site
	if err := query.Find(&sites).Error; err != nil {
		return nil, err
	}
	return sites, nil
}

// CountSitesDAO returns total count for sites by filter
func CountSitesDAO(filter model.SiteFilter) (int64, error) {
	query := database.DB.Model(&model.Site{})
	if filter.Query != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%")
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

// GetSiteByIDDAO retrieves a site by ID
func GetSiteByIDDAO(id uint) (model.Site, error) {
	var site model.Site
	err := database.DB.First(&site, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return site, model.ErrNotFound
	}
	return site, err
}

// AddSiteDAO creates a new site
func AddSiteDAO(site model.Site) error {
	return database.DB.Create(&site).Error
}

// UpdateSiteDAO updates a site by ID
func UpdateSiteDAO(id uint, site model.Site) error {
	return database.DB.Model(&model.Site{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":        site.Name,
		"description": site.Description,
		"enabled":     site.Enabled,
		"status":      site.Status,
	}).Error
}

// UpdateSiteStatusDAO updates only the status for a site
func UpdateSiteStatusDAO(id uint, status int) error {
	return database.DB.Model(&model.Site{}).Where("id = ?", id).Update("status", status).Error
}

// DeleteSiteByIDDAO deletes a site by ID
func DeleteSiteByIDDAO(id uint) error {
	return database.DB.Delete(&model.Site{}, id).Error
}
