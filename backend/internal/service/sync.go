package service

import (
	"fmt"
	"time"

	"nagare/internal/repository"

	"github.com/spf13/viper"
)

const (
	defaultSyncConcurrency = 2
)

// StartAutoSync starts periodic pulling of hosts and items from all monitors.
func StartAutoSync() {
	fmt.Println(">>> Starting AutoSync Service...")
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
	fmt.Printf(">>> AutoSync will run every %d seconds (source: %s)\n", intervalSeconds, intervalSource)
	LogSystem("info", "auto sync enabled", map[string]interface{}{"interval_seconds": intervalSeconds, "interval_source": intervalSource}, nil, "")

	LogSystem("info", "starting auto sync background task", nil, nil, "")
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
	fmt.Println(">>> AutoSync: Pulling data from monitors...")
	LogSystem("info", "auto sync started: pulling data from all monitors", nil, nil, "")
	monitors, err := repository.GetAllMonitorsDAO()
	if err != nil {
		LogSystem("error", "auto sync failed to load monitors", map[string]interface{}{"error": err.Error()}, nil, "")
		return
	}

	limit := configuredLimit("sync.concurrency", defaultSyncConcurrency)
	LogSystem("info", "auto sync processing monitors", map[string]interface{}{"count": len(monitors), "concurrency": limit}, nil, "")

	runWithLimit(len(monitors), limit, func(i int) {
		monitor := monitors[i]
		if monitor.Enabled == 0 {
			LogSystem("info", "auto sync skipping disabled monitor", map[string]interface{}{"monitor_id": monitor.ID, "name": monitor.Name}, nil, "")
			return
		}

		LogSystem("info", "auto sync syncing monitor", map[string]interface{}{"monitor_id": monitor.ID, "name": monitor.Name}, nil, "")
		if _, err := PullGroupsFromMonitorAutoSyncServ(monitor.ID); err != nil {
			LogSystem("error", "auto sync groups failed", map[string]interface{}{"monitor_id": monitor.ID, "name": monitor.Name, "error": err.Error()}, nil, "")
		}

		if _, err := pullHostsFromMonitorServ(monitor.ID, false); err != nil {
			LogSystem("error", "auto sync hosts failed", map[string]interface{}{"monitor_id": monitor.ID, "name": monitor.Name, "error": err.Error()}, nil, "")
		}
		if _, err := pullItemsFromMonitorServ(monitor.ID, false); err != nil {
			LogSystem("error", "auto sync items failed", map[string]interface{}{"monitor_id": monitor.ID, "name": monitor.Name, "error": err.Error()}, nil, "")
		}
	})
	LogSystem("info", "auto sync finished: all monitors processed", nil, nil, "")
}
