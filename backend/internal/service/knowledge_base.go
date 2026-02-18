package service

import (
	"fmt"
	"strings"

	"nagare/internal/model"
	"nagare/internal/repository"
)

// RetrieveContext retrieves relevant knowledge from the knowledge base based on the alert message
func RetrieveContext(alertMsg string) string {
	// 1. Simple tokenization by spaces and common punctuation
	f := func(c rune) bool {
		return c == ' ' || c == ',' || c == '.' || c == ':' || c == ';' || c == '!' || c == '?' || c == '(' || c == ')' || c == '[' || c == ']'
	}
	tokens := strings.FieldsFunc(alertMsg, f)

	// 2. Filter out short or common words
	stopWords := map[string]bool{
		"the": true, "is": true, "at": true, "which": true, "on": true, "a": true, "an": true,
		"and": true, "or": true, "but": true, "error": true, "alert": true, "detected": true,
		"warning": true, "critical": true, "failed": true, "failure": true,
	}

	validTokens := make([]string, 0)
	seen := make(map[string]bool)
	for _, t := range tokens {
		t = strings.ToLower(t)
		if len(t) > 2 && !stopWords[t] && !seen[t] {
			validTokens = append(validTokens, t)
			seen[t] = true
		}
	}

	// 3. Search database
	results, err := repository.SearchKnowledgeBaseDAO(validTokens, 3)
	if err != nil || len(results) == 0 {
		return ""
	}

	// 4. Construct context string
	var sb strings.Builder
	sb.WriteString("\nLocal Knowledge Base Reference Information:\n")
	for i, res := range results {
		sb.WriteString(fmt.Sprintf("[%d] Topic: %s\nContent: %s\n", i+1, res.Topic, res.Content))
	}
	return sb.String()
}

// KnowledgeBaseReq represents a knowledge base request
type KnowledgeBaseReq struct {
	Topic    string `json:"topic" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Keywords string `json:"keywords"`
	Category string `json:"category"`
}

// AddKnowledgeBaseServ adds a new knowledge base entry
func AddKnowledgeBaseServ(req KnowledgeBaseReq) error {
	kb := model.KnowledgeBase{
		Topic:    req.Topic,
		Content:  req.Content,
		Keywords: req.Keywords,
		Category: req.Category,
	}
	return repository.AddKnowledgeBaseDAO(&kb)
}

// GetAllKnowledgeBaseServ retrieves all knowledge base entries
func GetAllKnowledgeBaseServ() ([]model.KnowledgeBase, error) {
	return repository.GetAllKnowledgeBaseDAO()
}

// GetKnowledgeBaseByIDServ retrieves a knowledge base entry by ID
func GetKnowledgeBaseByIDServ(id uint) (model.KnowledgeBase, error) {
	return repository.GetKnowledgeBaseByIDDAO(id)
}

// UpdateKnowledgeBaseServ updates an existing knowledge base entry
func UpdateKnowledgeBaseServ(id uint, req KnowledgeBaseReq) error {
	kb := model.KnowledgeBase{
		Topic:    req.Topic,
		Content:  req.Content,
		Keywords: req.Keywords,
		Category: req.Category,
	}
	return repository.UpdateKnowledgeBaseDAO(id, kb)
}

// DeleteKnowledgeBaseServ deletes a knowledge base entry by ID
func DeleteKnowledgeBaseServ(id uint) error {
	return repository.DeleteKnowledgeBaseDAO(id)
}

// SearchKnowledgeBaseServ searches for knowledge base entries
func SearchKnowledgeBaseServ(q string) ([]model.KnowledgeBase, error) {
	return repository.QueryKnowledgeBaseDAO(q)
}
