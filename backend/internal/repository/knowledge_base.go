package repository

import (
	"nagare/internal/database"
	"nagare/internal/model"
)

// AddKnowledgeBaseDAO adds a new knowledge base entry
func AddKnowledgeBaseDAO(kb *model.KnowledgeBase) error {
	return database.DB.Create(kb).Error
}

// GetAllKnowledgeBaseDAO retrieves all knowledge base entries
func GetAllKnowledgeBaseDAO() ([]model.KnowledgeBase, error) {
	var kbs []model.KnowledgeBase
	err := database.DB.Find(&kbs).Error
	return kbs, err
}

// GetKnowledgeBaseByIDDAO retrieves a knowledge base entry by ID
func GetKnowledgeBaseByIDDAO(id uint) (model.KnowledgeBase, error) {
	var kb model.KnowledgeBase
	err := database.DB.First(&kb, id).Error
	return kb, err
}

// UpdateKnowledgeBaseDAO updates an existing knowledge base entry
func UpdateKnowledgeBaseDAO(id uint, kb model.KnowledgeBase) error {
	return database.DB.Model(&model.KnowledgeBase{}).Where("id = ?", id).Updates(kb).Error
}

// DeleteKnowledgeBaseDAO deletes a knowledge base entry by ID
func DeleteKnowledgeBaseDAO(id uint) error {
	return database.DB.Delete(&model.KnowledgeBase{}, id).Error
}

// SearchKnowledgeBaseDAO searches for knowledge base entries by keywords
func SearchKnowledgeBaseDAO(tokens []string, limit int) ([]model.KnowledgeBase, error) {
	var kbs []model.KnowledgeBase
	query := database.DB.Model(&model.KnowledgeBase{})

	if len(tokens) > 0 {
		for _, token := range tokens {
			query = query.Or("keywords LIKE ?", "%"+token+"%")
			query = query.Or("topic LIKE ?", "%"+token+"%")
		}
	}

	err := query.Limit(limit).Find(&kbs).Error
	return kbs, err
}

// QueryKnowledgeBaseDAO performs a general search on the knowledge base
func QueryKnowledgeBaseDAO(q string) ([]model.KnowledgeBase, error) {
	var kbs []model.KnowledgeBase
	err := database.DB.Where("topic LIKE ? OR content LIKE ? OR keywords LIKE ?", "%"+q+"%", "%"+q+"%", "%"+q+"%").Find(&kbs).Error
	return kbs, err
}
