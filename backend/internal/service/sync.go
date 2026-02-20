package service

import (
	"time"

	"github.com/spf13/viper"
	"nagare/internal/repository"
)

const (
	defaultSyncConcurrency = 2
)

// StartAutoSync starts periodic pulling of hosts and items from all monitors.
func StartAutoSync() {
	enabled := viper.GetBool("sync.enabled")
	if !viper.IsSet("sync.enabled") {
		enabled = true
	}
	if !enabled {
		LogSystem("info", "auto sync disabled via configuration", nil, nil, "")
		return
	}

	intervalSeconds := viper.GetInt("sync.interval_seconds")
	if intervalSeconds <= 0 {
		intervalSeconds = viper.GetInt("status_check.interval_seconds")
	}
	if intervalSeconds <= 0 {
		intervalSeconds = defaultStatusCheckIntervalSeconds
	}

	intervalSource := "sync"
	if viper.GetInt("sync.interval_seconds") <= 0 {
		intervalSource = "status_check"
	}
	LogSystem("info", "auto sync enabled", map[string]interface{}{"interval_seconds": intervalSeconds, "interval_source": intervalSource}, nil, "")

	go func() {
		ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
		defer ticker.Stop()

		pullAllMonitors()
		for range ticker.C {
			pullAllMonitors()
		}
	}()
}

func pullAllMonitors() {
	monitors, err := repository.GetAllMonitorsDAO()
	if err != nil {
		LogSystem("error", "auto sync failed to load monitors", map[string]interface{}{"error": err.Error()}, nil, "")
		return
	}

	limit := configuredLimit("sync.concurrency", defaultSyncConcurrency)
	runWithLimit(len(monitors), limit, func(i int) {
		monitor := monitors[i]
		if monitor.Enabled == 0 {
			return
		}
		if _, err := pullHostsFromMonitorServ(monitor.ID, false); err != nil {
			LogSystem("error", "auto sync hosts failed", map[string]interface{}{"monitor_id": monitor.ID, "error": err.Error()}, nil, "")
		}
		if _, err := pullItemsFromMonitorServ(monitor.ID, false); err != nil {
			LogSystem("error", "auto sync items failed", map[string]interface{}{"monitor_id": monitor.ID, "error": err.Error()}, nil, "")
		}
	})
}
