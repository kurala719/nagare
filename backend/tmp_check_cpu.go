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
	// Try the original query
	_ = database.DB.Model(&model.Item{}).
		Where("name LIKE ? OR name LIKE ?", "%CPU%", "%cpu%").
		Order("CAST(last_value AS DECIMAL) desc").
		Limit(5).
		Find(&items).Error

	// Test the JOIN query
	type cpuResult struct {
		model.Item
		HostName   string
		HostIP     string
		HostStatus int
	}
	var cpuResults []cpuResult
	err := database.DB.Table("items").
		Select("items.*, hosts.name as host_name, hosts.ip_addr as host_ip, hosts.status as host_status").
		Joins("left join hosts on hosts.id = items.hid").
		Where("(items.name LIKE ? OR items.name LIKE ? OR items.name LIKE ?) AND items.last_value != ''", "%CPU%", "%cpu%", "%处理器%").
		Order("(items.last_value + 0) desc").
		Limit(5).
		Scan(&cpuResults).Error

	fmt.Printf("\nJOIN Query: Found %d items, Error: %v\n", len(cpuResults), err)
	for _, res := range cpuResults {
		fmt.Printf("Item ID: %d, Name: %s, Value: %s, HostName: %s, HostIP: %s\n", res.ID, res.Name, res.LastValue, res.HostName, res.HostIP)
	}
}
