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

	// Check for hosts named "Zabbix server"
	var hosts []model.Host
	database.DB.Where("name = ?", "Zabbix server").Find(&hosts)
	fmt.Printf("Found %d hosts with name 'Zabbix server'\n", len(hosts))
	for _, h := range hosts {
		fmt.Printf("Host ID: %d, Name: %s, IP: %s\n", h.ID, h.Name, h.IPAddr)
	}

	// Run the CPU query and print raw details including HID
	type cpuResult struct {
		model.Item
		HostName   string
		HostIP     string
		HostStatus int
	}
	var cpuResults []cpuResult
	database.DB.Table("items").
		Select("items.*, hosts.name as host_name, hosts.ip_addr as host_ip, hosts.status as host_status").
		Joins("left join hosts on hosts.id = items.hid").
		Where("(items.name LIKE ? OR items.name LIKE ? OR items.name LIKE ?) AND items.last_value != ''", "%CPU%", "%cpu%", "%处理器%").
		Order("(items.last_value + 0) desc").
		Limit(20).
		Scan(&cpuResults)

	fmt.Printf("\nTop 20 CPU Items found:\n")
	seenHosts := make(map[uint]bool)
	for i, res := range cpuResults {
		duplicate := "New"
		if seenHosts[res.HID] {
			duplicate = "DUPLICATE"
		}
		seenHosts[res.HID] = true
		fmt.Printf("[%d] ItemID: %d, Name: %s, Value: %s, HID: %d, HostName: %s, Status: %s\n",
			i, res.ID, res.Name, res.LastValue, res.HID, res.HostName, duplicate)
	}
}
