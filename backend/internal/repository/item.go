package repository

import (
	"errors"

	"gorm.io/gorm"
	"nagare/internal/database"
	"nagare/internal/model"
)

// GetAllItemsDAO retrieves all items from the database
func GetAllItemsDAO() ([]model.Item, error) {
	var items []model.Item
	if err := database.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// SearchItemsDAO retrieves items by filter
func SearchItemsDAO(filter model.ItemFilter) ([]model.Item, error) {
	query := database.DB.Model(&model.Item{})
	if filter.Query != "" {
		query = query.Where("name LIKE ? OR itemid LIKE ? OR hostid LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%", "%"+filter.Query+"%")
	}
	if filter.HID != nil {
		query = query.Where("hid = ?", *filter.HID)
	}
	if filter.ValueType != nil {
		query = query.Where("value_type = ?", *filter.ValueType)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.HostID != nil {
		query = query.Where("hostid = ?", *filter.HostID)
	}
	if filter.ItemID != nil {
		query = query.Where("itemid = ?", *filter.ItemID)
	}
	query = applySort(query, filter.SortBy, filter.SortOrder, map[string]string{
		"name":       "name",
		"status":     "status",
		"enabled":    "enabled",
		"id":         "id",
		"value":      "last_value",
		"created_at": "created_at",
		"updated_at": "updated_at",
	}, "id desc")
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}
	var items []model.Item
	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// CountItemsDAO returns total count for items by filter
func CountItemsDAO(filter model.ItemFilter) (int64, error) {
	query := database.DB.Model(&model.Item{})
	if filter.Query != "" {
		query = query.Where("name LIKE ? OR itemid LIKE ? OR hostid LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%", "%"+filter.Query+"%")
	}
	if filter.HID != nil {
		query = query.Where("hid = ?", *filter.HID)
	}
	if filter.ValueType != nil {
		query = query.Where("value_type = ?", *filter.ValueType)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.HostID != nil {
		query = query.Where("hostid = ?", *filter.HostID)
	}
	if filter.ItemID != nil {
		query = query.Where("itemid = ?", *filter.ItemID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// GetItems is an alias for GetAllItemsDAO (deprecated, use GetAllItemsDAO)
func GetItems() ([]model.Item, error) {
	return GetAllItemsDAO()
}

// GetItemByIDDAO retrieves an item by ID
func GetItemByIDDAO(id uint) (model.Item, error) {
	var item model.Item
	err := database.DB.First(&item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return item, model.ErrNotFound
	}
	return item, err
}

// GetItemsByHIDDAO retrieves all items for a specific host
func GetItemsByHIDDAO(hid uint) ([]model.Item, error) {
	var items []model.Item
	if err := database.DB.Where("hid = ?", hid).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// GetItemByHIDAndItemIDDAO retrieves an item by host ID and external item ID
func GetItemByHIDAndItemIDDAO(hid uint, itemID string) (model.Item, error) {
	var item model.Item
	err := database.DB.Where("hid = ? AND itemid = ?", hid, itemID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return item, model.ErrNotFound
	}
	return item, err
}

// AddItemDAO creates a new item
func AddItemDAO(item model.Item) error {
	return database.DB.Create(&item).Error
}

// DeleteItemByIDDAO deletes an item by ID
func DeleteItemByIDDAO(id uint) error {
	return database.DB.Delete(&model.Item{}, id).Error
}

// UpdateItemDAO updates an item by ID
func UpdateItemDAO(id uint, item model.Item) error {
	return database.DB.Model(&model.Item{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":               item.Name,
		"hid":                item.HID,
		"itemid":             item.ItemID,
		"hostid":             item.ExternalHostID,
		"value_type":         item.ValueType,
		"last_value":         item.LastValue,
		"units":              item.Units,
		"enabled":            item.Enabled,
		"status":             item.Status,
		"status_description": item.StatusDescription,
		"comment":            item.Comment,
		"last_sync_at":       item.LastSyncAt,
		"external_source":    item.ExternalSource,
	}).Error
}

// UpdateItemStatusDAO updates only the status for an item
func UpdateItemStatusDAO(id uint, status int) error {
	return database.DB.Model(&model.Item{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateItemStatusAndCommentDAO updates status and comment for an item
func UpdateItemStatusAndCommentDAO(id uint, status int, comment string) error {
	return database.DB.Model(&model.Item{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":  status,
		"comment": comment,
	}).Error
}

// UpdateItemStatusAndDescriptionDAO updates status and status_description for an item
func UpdateItemStatusAndDescriptionDAO(id uint, status int, statusDesc string) error {
	return database.DB.Model(&model.Item{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":             status,
		"status_description": statusDesc,
	}).Error
}
