package main

import (
	"fmt"
	"log"

	"nagare/internal/repository"
	"nagare/internal/service"
)

func main() {
	err := service.InitConfigServ("nagare_config.json")
	if err != nil {
		log.Fatalf("InitConfigServ error: %v", err)
	}

	minLog := repository.GetConfigInt("site_message.min_log_severity")
	minAlert := repository.GetConfigInt("site_message.min_alert_severity")

	fmt.Printf("Parsed Min Log Severity: %d\n", minLog)
	fmt.Printf("Parsed Min Alert Severity: %d\n", minAlert)

	// Simulate a Log Message check
	checkLog(1)
	checkLog(2)
	checkLog(4)

	// Simulate an Alert check
}

func checkLog(sev int) {
	minLog := repository.GetConfigInt("site_message.min_log_severity")
	if sev < minLog {
		fmt.Printf("Log Sev %d: SKIPPED (sev < %d)\n", sev, minLog)
	} else {
		fmt.Printf("Log Sev %d: ALLOWED (sev >= %d)\n", sev, minLog)
	}
}
