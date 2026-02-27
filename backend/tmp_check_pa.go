package main

import (
	"fmt"
	"nagare/internal/database"
	"nagare/internal/model"
	"os"
)

func main() {
	os.Setenv("NAGARE_CONFIG_PATH", "../configs/nagare_config.json")
	if err := database.InitDBFromConfig(); err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer database.CloseDB(database.DB)

	var pas []model.PacketAnalysis
	database.DB.Order("id desc").Limit(5).Find(&pas)

	for _, pa := range pas {
		fmt.Printf("ID: %d\nFile: %s\nStatus: %d\nRisk: %s\nAnalysis (first 100): %.100s\nRawContent (first 100): %.100s\n---\n",
			pa.ID, pa.FilePath, pa.Status, pa.RiskLevel, pa.Analysis, pa.RawContent)
	}
}
