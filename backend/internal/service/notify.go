package service

import (
	"fmt"
	"time"
)

// SyncEvent sends a notification for sync results
func SyncEvent(entity string, monitorID uint, hostID uint, result SyncResult) {
	go func() {
		message := fmt.Sprintf("sync %s: total=%d added=%d updated=%d failed=%d", entity, result.Total, result.Added, result.Updated, result.Failed)
		
		// Add Site Message
		title := fmt.Sprintf("Sync %s Finished", entity)
		msgSeverity := 1 // success
		if result.Failed > 0 {
			msgSeverity = 3 // error
		}
		_ = CreateSiteMessageServ(title, message, "sync", msgSeverity, nil)

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
