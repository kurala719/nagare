package main

import (
	"fmt"
	"os"

	"nagare/internal/database"
	"nagare/internal/repository"
	"nagare/internal/service"
)

func main() {
	os.Setenv("NAGARE_CONFIG_PATH", "../configs/nagare_config.json")
	if err := database.InitDBFromConfig(); err != nil {
		fmt.Println("DB Init Error:", err)
		return
	}

	hosts, err := repository.GetAllHostsDAO()
	if err != nil {
		fmt.Println("Error fetching hosts:", err)
		return
	}

	if len(hosts) == 0 {
		fmt.Println("No hosts exist in the DB to test.")
		return
	}

	target := hosts[len(hosts)-1] // Just look at the last one
	fmt.Printf("Analyzing Host ID: %d, Name: %s\n", target.ID, target.Name)
	fmt.Printf("  MonitorID: %d\n", target.MonitorID)
	fmt.Printf("  Hostid (external Zabbix ID): '%s'\n", target.Hostid)

	if target.MonitorID > 0 && target.Hostid != "" {
		fmt.Println("Condition to trigger Zabbix Delete is MET. Calling DeleteHostFromMonitorServ directly...")
		err := service.DeleteHostFromMonitorServ(target.MonitorID, target.Hostid)
		if err != nil {
			fmt.Printf("  -> FAILED: %v\n", err)
		} else {
			fmt.Println("  -> SUCCESS")
		}
	} else {
		fmt.Println("Condition to trigger Zabbix Delete is NOT MET. Would skip.")
	}
}
