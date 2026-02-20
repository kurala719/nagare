package repository

import (
	"nagare/internal/core/domain"
	"nagare/internal/database"
)

// AddKnowledgeBaseDAO adds a new knowledge base entry
func AddKnowledgeBaseDAO(kb *domain.KnowledgeBase) error {
	return database.DB.Create(kb).Error
}

// GetAllKnowledgeBaseDAO retrieves all knowledge base entries
func GetAllKnowledgeBaseDAO() ([]domain.KnowledgeBase, error) {
	var kbs []domain.KnowledgeBase
	err := database.DB.Find(&kbs).Error
	return kbs, err
}

// GetKnowledgeBaseByIDDAO retrieves a knowledge base entry by ID
func GetKnowledgeBaseByIDDAO(id uint) (domain.KnowledgeBase, error) {
	var kb domain.KnowledgeBase
	err := database.DB.First(&kb, id).Error
	return kb, err
}

// UpdateKnowledgeBaseDAO updates an existing knowledge base entry
func UpdateKnowledgeBaseDAO(id uint, kb domain.KnowledgeBase) error {
	return database.DB.Model(&domain.KnowledgeBase{}).Where("id = ?", id).Updates(kb).Error
}

// DeleteKnowledgeBaseDAO deletes a knowledge base entry by ID
func DeleteKnowledgeBaseDAO(id uint) error {
	return database.DB.Delete(&domain.KnowledgeBase{}, id).Error
}

// SearchKnowledgeBaseDAO searches for knowledge base entries by keywords
func SearchKnowledgeBaseDAO(tokens []string, limit int) ([]domain.KnowledgeBase, error) {
	var kbs []domain.KnowledgeBase
	query := database.DB.Model(&domain.KnowledgeBase{})

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
func QueryKnowledgeBaseDAO(q string) ([]domain.KnowledgeBase, error) {
	var kbs []domain.KnowledgeBase
	err := database.DB.Where("topic LIKE ? OR content LIKE ? OR keywords LIKE ?", "%"+q+"%", "%"+q+"%", "%"+q+"%").Find(&kbs).Error
	return kbs, err
}
