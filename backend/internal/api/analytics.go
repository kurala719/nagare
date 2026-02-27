package api

import (
	"net/http"
	"strings"
	"time"

	"nagare/internal/database"
	"nagare/internal/model"
	"nagare/internal/service"

	"github.com/gin-gonic/gin"
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

// GetAlertAnalyticsCtrl returns data for word cloud, heatmap, and trends
func GetAlertAnalyticsCtrl(c *gin.Context) {
	// 1. Get Word Cloud data (top words from last 1000 alerts)
	var alerts []model.Alert
	database.DB.Order("id desc").Limit(1000).Find(&alerts)

	wordMap := make(map[string]int)
	stopWords := map[string]bool{
		"a": true, "an": true, "the": true, "is": true, "are": true, "was": true, "were": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "with": true, "of": true,
		"and": true, "or": true, "not": true, "has": true, "been": true, "failed": true, "problem": true,
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
		if value > 1 {
			wordCloud = append(wordCloud, WordCount{Name: name, Value: value})
		}
	}

	// 2. Get Heatmap data (counts per day for last 90 days)
	type result struct {
		Date  string `json:"date"`
		Count int    `json:"count"`
	}
	var results []result
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

	// 3. Get Severity Distribution
	type sevResult struct {
		Severity int `json:"severity"`
		Count    int `json:"count"`
	}
	var sevResults []sevResult
	database.DB.Model(&model.Alert{}).
		Select("severity, count(*) as count").
		Group("severity").
		Scan(&sevResults)

	// 4. Get Top Noisy Hosts
	type hostResult struct {
		HostID uint   `json:"host_id"`
		Count  int    `json:"count"`
		Name   string `json:"name"`
	}
	var hostResults []hostResult
	database.DB.Table("alerts").
		Select("host_id, count(*) as count, hosts.name").
		Joins("left join hosts on hosts.id = alerts.host_id").
		Where("alerts.host_id > 0").
		Group("host_id").
		Order("count desc").
		Limit(10).
		Scan(&hostResults)

	// 5. Get Alert Trend (last 14 days)
	var trendResults []result
	fourteenDaysAgo := time.Now().AddDate(0, 0, -14)
	database.DB.Model(&model.Alert{}).
		Select("DATE(created_at) as date, count(*) as count").
		Where("created_at >= ?", fourteenDaysAgo).
		Group("DATE(created_at)").
		Order("date").
		Scan(&trendResults)

	// 6. Summary Stats
	var totalAlerts int64
	database.DB.Model(&model.Alert{}).Count(&totalAlerts)

	healthScore, _ := service.GetHealthScoreServ()

	var activeHosts int64
	database.DB.Model(&model.Host{}).Where("status = ?", 1).Count(&activeHosts)

	respondSuccess(c, http.StatusOK, gin.H{
		"wordCloud":    wordCloud,
		"heatmap":      heatmap,
		"severityDist": sevResults,
		"topHosts":     hostResults,
		"trend":        trendResults,
		"summary": gin.H{
			"totalAlerts":  totalAlerts,
			"systemHealth": healthScore.Score,
			"activeHosts":  activeHosts,
		},
	})
}
