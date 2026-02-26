package service

import (
	"fmt"
	"time"
)

// SyncEvent sends a notification for sync results
func SyncEvent(entity string, monitorID uint, hostID uint, result SyncResult) {
	go func() {
		message := fmt.Sprintf("sync %s: total=%d added=%d updated=%d failed=%d", entity, result.Total, result.Added, result.Updated, result.Failed)
		
		// Log event - this will trigger a Site Message via central log Entry
		logSeverity := "info"
		if result.Failed > 0 {
			logSeverity = "error"
		}
		LogService(logSeverity, fmt.Sprintf("Sync %s Finished: %s", entity, message), map[string]interface{}{
			"entity":     entity,
			"monitor_id": monitorID,
			"host_id":    hostID,
			"result":     result,
		}, nil, "")

		// BROADCAST to Frontend via WebSocket to force UI refresh
		broadcastSyncUpdate(entity, monitorID, hostID)
	}()
}

func broadcastSyncUpdate(entity string, mid, hid uint) {
	msg := map[string]interface{}{
		"type":       "sync_complete",
		"entity":     entity,
		"monitor_id": mid,
		"host_id":    hid,
		"timestamp":  time.Now().Unix(),
	}
	BroadcastMessage(msg)
}

func syncSeverity(result SyncResult) int {
	if result.Failed > 0 {
		return 3
	}
	if result.Added > 0 || result.Updated > 0 {
		return 1
	}
	return 0
}
