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
		Limit(20).
		Scan(&cpuResults).Error

	if err != nil {
		fmt.Printf("Query failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Query returned %d items\n", len(cpuResults))

	seenHosts := make(map[uint]bool)
	finalHosts := 0
	for _, res := range cpuResults {
		if seenHosts[res.HID] {
			fmt.Printf("DUPLICATE REMOVED: HostID: %d, ItemName: %s\n", res.HID, res.Name)
			continue
		}
		seenHosts[res.HID] = true
		finalHosts++
		fmt.Printf("ADDED: HostID: %d, ItemName: %s, Value: %s\n", res.HID, res.Name, res.LastValue)

		if finalHosts >= 5 {
			break
		}
	}

	fmt.Printf("\nFinal count of unique hosts: %d\n", finalHosts)
}
