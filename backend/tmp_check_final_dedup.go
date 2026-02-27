package main

import (
	"fmt"
	"nagare/internal/database"
	_ "nagare/internal/model" // Ensure models are registered if needed
	"nagare/internal/service"
)

func main() {
	configPath := "configs/nagare_config.json"
	_ = service.InitConfigServ(configPath)
	_ = database.InitDBFromConfig()

	type cpuResult struct {
		ID         uint   `gorm:"column:id"`
		HID        uint   `gorm:"column:hid"`
		Name       string `gorm:"column:name"`
		LastValue  string `gorm:"column:last_value"`
		Units      string `gorm:"column:units"`
		HostName   string `gorm:"column:host_name"`
		HostIP     string `gorm:"column:host_ip"`
		HostStatus int    `gorm:"column:host_status"`
	}
	var cpuResults []cpuResult
	database.DB.Table("items").
		Select("items.id, items.hid, items.name, items.last_value, items.units, hosts.name as host_name, hosts.ip_addr as host_ip, hosts.status as host_status").
		Joins("left join hosts on hosts.id = items.hid").
		Where("(items.name LIKE ? OR items.name LIKE ? OR items.name LIKE ?) AND items.last_value != ''", "%CPU%", "%cpu%", "%处理器%").
		Order("(items.last_value + 0) desc").
		Limit(20).
		Scan(&cpuResults)

	fmt.Printf("Deduplication Test:\n")
	seenAssets := make(map[string]bool)
	addedCount := 0
	for _, res := range cpuResults {
		assetKey := res.HostName + "|" + res.HostIP
		if assetKey == "|" || seenAssets[assetKey] {
			fmt.Printf("SKIPPED (Duplicate/Empty): Asset=%s, IP=%s, ItemID=%d\n", res.HostName, res.HostIP, res.ID)
			continue
		}
		seenAssets[assetKey] = true
		addedCount++
		fmt.Printf("ADDED: Asset=%s, IP=%s, Value=%s, ItemID=%d\n", res.HostName, res.HostIP, res.LastValue, res.ID)

		if addedCount >= 5 {
			break
		}
	}
}
