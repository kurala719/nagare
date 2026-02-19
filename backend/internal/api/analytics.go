package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"nagare/internal/database"
	"nagare/internal/model"
)

// WordCount represents a word and its frequency
type WordCount struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// DailyCount represents a date and the number of alerts on that day
type DailyCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// GetAlertAnalyticsCtrl returns data for word cloud and heatmap
func GetAlertAnalyticsCtrl(c *gin.Context) {
	// 1. Get Word Cloud data (simple implementation: top words from last 1000 alerts)
	var alerts []model.Alert
	database.DB.Order("id desc").Limit(1000).Find(&alerts)

	wordMap := make(map[string]int)
	// Common stop words to ignore
	stopWords := map[string]bool{
		"a": true, "an": true, "the": true, "is": true, "are": true, "was": true, "were": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "with": true, "of": true,
		"and": true, "or": true, "not": true, "has": true, "been": true,
	}

	for _, alert := range alerts {
		words := strings.Fields(strings.ToLower(alert.Message))
		for _, w := range words {
			w = strings.Trim(w, `.,!?:;"'()`)
			if len(w) > 2 && !stopWords[w] {
				wordMap[w]++
			}
		}
	}

	wordCloud := make([]WordCount, 0)
	for name, value := range wordMap {
		if value > 1 { // Only include words appearing more than once
			wordCloud = append(wordCloud, WordCount{Name: name, Value: value})
		}
	}

	// 2. Get Heatmap data (counts per day for last 90 days)
	type result struct {
		Date  string
		Count int
	}
	var results []result
	
	// SQLite/MySQL specific query for grouping by date
	// Using a more portable GORM query
	ninetyDaysAgo := time.Now().AddDate(0, 0, -90)
	database.DB.Model(&model.Alert{}).
		Select("DATE(created_at) as date, count(*) as count").
		Where("created_at >= ?", ninetyDaysAgo).
		Group("DATE(created_at)").
		Order("date").
		Scan(&results)

	heatmap := make([][2]interface{}, 0)
	for _, r := range results {
		heatmap = append(heatmap, [2]interface{}{r.Date, r.Count})
	}

	respondSuccess(c, http.StatusOK, gin.H{
		"wordCloud": wordCloud,
		"heatmap":   heatmap,
	})
}
