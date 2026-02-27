package main

import (
	"fmt"
	"nagare/internal/database"
	"nagare/internal/model"
	"nagare/internal/service"
)

func main() {
	configPath := "configs/nagare_config.json"
	_ = service.InitConfigServ(configPath)
	_ = database.InitDBFromConfig()

	// 0. List all hosts
	var allHosts []model.Host
	database.DB.Find(&allHosts)
	fmt.Printf("ALL HOSTS:\n")
	for _, h := range allHosts {
		fmt.Printf("ID=%d, Name=%s, IP=%s\n", h.ID, h.Name, h.IPAddr)
	}

	// 1. Get Host ID for Zabbix server
	var hosts []model.Host
	database.DB.Where("name = ?", "Zabbix server").Find(&hosts)
	for _, h := range hosts {
		fmt.Printf("\nSEARCH HOST: ID=%d, Name=%s, IP=%s\n", h.ID, h.Name, h.IPAddr)
	}

	// 2. RUN EXACT QUERY FROM report.go
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
		Limit(10).
		Scan(&cpuResults)

	fmt.Printf("\nQUERY RESULTS (First 10):\n")
	for i, r := range cpuResults {
		fmt.Printf("[%d] ItemID=%d, ItemHIDFromStruct=%d, Name=%s, Value=%s, Host=%s\n", i, r.ID, r.HID, r.Name, r.LastValue, r.HostName)
	}
}
