package main

import (
	"fmt"
	"nagare/internal/database"
	"nagare/internal/model"
	"nagare/internal/service"
	"os"
)

func main() {
	configPath := "configs/nagare_config.json"
	if err := service.InitConfigServ(configPath); err != nil {
		fmt.Printf("InitConfig failed: %v\n", err)
		os.Exit(1)
	}

	if err := database.InitDBFromConfig(); err != nil {
		fmt.Printf("InitDB failed: %v\n", err)
		os.Exit(1)
	}

	var items []model.Item
	database.DB.Where("hid = ?", 1).Find(&items)
	fmt.Printf("Found %d items for Host ID 1 (Zabbix server)\n", len(items))
	for _, item := range items {
		fmt.Printf("ID: %d, Name: %s, Value: %s\n", item.ID, item.Name, item.LastValue)
	}
}
