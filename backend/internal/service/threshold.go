package service

import (
	"fmt"
	"strconv"
	"time"

	"nagare/internal/model"
	"nagare/internal/repository"
)

// CheckItemThresholds evaluates all metrics for a host against predefined safe thresholds
func CheckItemThresholds(hid uint) {
	items, err := repository.GetItemsByHIDDAO(hid)
	if err != nil {
		return
	}

	host, err := repository.GetHostByIDDAO(hid)
	if err != nil {
		return
	}

	for _, item := range items {
		evaluateThreshold(host, item)
	}
}

func evaluateThreshold(host model.Host, item model.Item) {
	if item.Enabled == 0 || item.LastValue == "" || item.LastValue == "N/A" {
		return
	}

	val, err := strconv.ParseFloat(item.LastValue, 64)
	if err != nil {
		return
	}

	var alertMsg string
	var severity int = 0 // 0:Info, 1:Warning, 2:Critical

	// Huawei CPU Thresholds
	if item.Name == "hwCpuDevDuty" || item.Name == "hwEntityCpuUsage" {
		if val >= 90 {
			alertMsg = fmt.Sprintf("Critical CPU Usage: %.2f%% on %s", val, host.Name)
			severity = 2
		} else if val >= 75 {
			alertMsg = fmt.Sprintf("High CPU Usage: %.2f%% on %s", val, host.Name)
			severity = 1
		}
	}

	// Huawei Memory Thresholds
	if item.Name == "hwEntityMemUsage" || item.Name == "mem_usage_pct" {
		if val >= 95 {
			alertMsg = fmt.Sprintf("Critical Memory Usage: %.2f%% on %s", val, host.Name)
			severity = 2
		} else if val >= 85 {
			alertMsg = fmt.Sprintf("High Memory Usage: %.2f%% on %s", val, host.Name)
			severity = 1
		}
	}

	// Huawei Temperature Thresholds
	if item.Name == "hwEntityTemperature" {
		// Smart-Fix: If temperature is > 200, it's likely scaled by 10 (e.g. 350 = 35.0C)
		if val > 200 {
			val = val / 10.0
		}
		
		if val >= 75 {
			alertMsg = fmt.Sprintf("Critical Temperature: %.2f°C on %s", val, host.Name)
			severity = 2
		} else if val >= 60 {
			alertMsg = fmt.Sprintf("High Temperature: %.2f°C on %s", val, host.Name)
			severity = 1
		}
	}

	if alertMsg != "" {
		triggerAlert(host, item, alertMsg, severity)
	}
}

func triggerAlert(host model.Host, item model.Item, message string, severity int) {
	// Simple suppression: don't create same alert for same item if one exists in last 1 hour
	recentAlerts, err := repository.SearchAlertsDAO(model.AlertFilter{
		HostID: uintPtrToIntPtr(&host.ID),
		ItemID: uintPtrToIntPtr(&item.ID),
		Status: pointerToInt(0), // Active
	})

	if err == nil && len(recentAlerts) > 0 {
		for _, a := range recentAlerts {
			if time.Since(a.CreatedAt) < 1*time.Hour {
				return // Suppress redundant alert
			}
		}
	}

	alert := model.Alert{
		Message:  message,
		Severity: severity,
		Status:   0,
		HostID:   host.ID,
		ItemID:   item.ID,
		Comment:  fmt.Sprintf("Automatically detected by Nagare Threshold Engine at %s", time.Now().Format(time.RFC1123)),
	}

	_ = repository.AddAlertDAO(&alert)
	
	// Site Message for Real-time notification
	_ = CreateSiteMessageServ("Threshold Alert", message, "alert", severity, nil)

	// Async AI Analysis and Trigger execution
	go analyzeAndNotifyAlert(alert)
}

func pointerToInt(i int) *int {
	return &i
}

func uintPtrToIntPtr(u *uint) *int {
	if u == nil {
		return nil
	}
	i := int(*u)
	return &i
}
