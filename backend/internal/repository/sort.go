package repository

import (
	"strings"

	"gorm.io/gorm"
)

func applySort(query *gorm.DB, sortBy string, sortOrder string, allowed map[string]string, fallback string) *gorm.DB {
	key := strings.TrimSpace(strings.ToLower(sortBy))
	column, ok := allowed[key]
	if !ok || column == "" {
		if fallback == "" {
			return query
		}
		return query.Order(fallback)
	}

	order := strings.TrimSpace(strings.ToLower(sortOrder))
	if order != "desc" {
		order = "asc"
	}
	return query.Order(column + " " + order)
}
