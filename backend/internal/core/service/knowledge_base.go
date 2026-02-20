package service

import (
	"fmt"
	"strings"

	"nagare/internal/adapter/repository"
	"nagare/internal/core/domain"
)

// RetrieveContext retrieves relevant knowledge from the knowledge base based on the alert message
func RetrieveContext(alertMsg string) string {
	// 1. Smarter tokenization by spaces and common punctuation, including IP patterns
	f := func(c rune) bool {
		return c == ' ' || c == ',' || c == '.' || c == ':' || c == ';' || c == '!' || c == '?' || c == '(' || c == ')' || c == '[' || c == ']'
	}
	tokens := strings.FieldsFunc(alertMsg, f)

	// 2. Filter out short or common words, but keep specific entities like IP segments
	stopWords := map[string]bool{
		"the": true, "is": true, "at": true, "which": true, "on": true, "a": true, "an": true,
		"and": true, "or": true, "but": true, "error": true, "alert": true, "detected": true,
		"warning": true, "critical": true, "failed": true, "failure": true, "nagare": true,
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

	// 3. Search database and perform keyword-based re-ranking
	results, err := repository.SearchKnowledgeBaseDAO(validTokens, 10) // Fetch more for re-ranking
	if err != nil || len(results) == 0 {
		return ""
	}

	// Re-ranking by keyword hit count
	type scoredResult struct {
		kb    domain.KnowledgeBase
		score int
	}
	scored := make([]scoredResult, 0, len(results))
	for _, res := range results {
		score := 0
		content := strings.ToLower(res.Content + " " + res.Topic + " " + res.Keywords)
		for _, token := range validTokens {
			if strings.Contains(content, token) {
				score += 2 // Boost score for direct keyword matches
			}
		}
		scored = append(scored, scoredResult{res, score})
	}

	// Simple bubble sort for demonstration (could use sort package for production)
	for i := 0; i < len(scored)-1; i++ {
		for j := 0; j < len(scored)-i-1; j++ {
			if scored[j].score < scored[j+1].score {
				scored[j], scored[j+1] = scored[j+1], scored[j]
			}
		}
	}

	// 4. Construct context string (top 3)
	var sb strings.Builder
	sb.WriteString("\n--- Relevant Operations Knowledge Base (RAG) ---\n")
	limit := 3
	if len(scored) < limit {
		limit = len(scored)
	}
	for i := 0; i < limit; i++ {
		res := scored[i].kb
		sb.WriteString(fmt.Sprintf("[%d] Topic: %s (Relevance: %d)\nContext: %s\n", i+1, res.Topic, scored[i].score, res.Content))
	}
	sb.WriteString("--------------------------------------------------\n")
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
	kb := domain.KnowledgeBase{
		Topic:    req.Topic,
		Content:  req.Content,
		Keywords: req.Keywords,
		Category: req.Category,
	}
	return repository.AddKnowledgeBaseDAO(&kb)
}

// GetAllKnowledgeBaseServ retrieves all knowledge base entries
func GetAllKnowledgeBaseServ() ([]domain.KnowledgeBase, error) {
	return repository.GetAllKnowledgeBaseDAO()
}

// GetKnowledgeBaseByIDServ retrieves a knowledge base entry by ID
func GetKnowledgeBaseByIDServ(id uint) (domain.KnowledgeBase, error) {
	return repository.GetKnowledgeBaseByIDDAO(id)
}

// UpdateKnowledgeBaseServ updates an existing knowledge base entry
func UpdateKnowledgeBaseServ(id uint, req KnowledgeBaseReq) error {
	kb := domain.KnowledgeBase{
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
func SearchKnowledgeBaseServ(q string) ([]domain.KnowledgeBase, error) {
	return repository.QueryKnowledgeBaseDAO(q)
}
