package main

import (
	"fmt"
	"nagare/internal/database"
	"nagare/internal/model"
	"nagare/internal/service"
	"time"
)

func main() {
	if err := service.InitConfigServ("configs/nagare_config.json"); err != nil {
		fmt.Println("Init config error:", err)
		return
	}
	if err := database.InitDBFromConfig(); err != nil {
		fmt.Println("Init DB Error:", err)
		return
	}

	type result struct {
		Date  string `json:"date"`
		Count int    `json:"count"`
	}
	var results []result
	ninetyDaysAgo := time.Now().AddDate(0, 0, -90)
	err := database.DB.Model(&model.Alert{}).
		Select("DATE_FORMAT(created_at, '%Y-%m-%d') as date, count(*) as count").
		Where("created_at >= ?", ninetyDaysAgo).
		Group("DATE_FORMAT(created_at, '%Y-%m-%d')").
		Order("date").
		Scan(&results).Error
	if err != nil {
		fmt.Println("Error:", err)
	}
	heatmap := make([][2]interface{}, 0)
	for _, r := range results {
		heatmap = append(heatmap, [2]interface{}{r.Date, r.Count})
	}
	fmt.Printf("Heatmap Data: %+v\n", heatmap)
}
