package service

import "fmt"

// SyncEvent sends a notification for sync results
func SyncEvent(entity string, monitorID uint, hostID uint, result SyncResult) {
	severity := syncSeverity(result)
	status := 1
	if result.Failed > 0 {
		status = 2
	}
	message := fmt.Sprintf("sync %s: total=%d added=%d updated=%d failed=%d", entity, result.Total, result.Added, result.Updated, result.Failed)
	ExecuteTriggersForEvent(AlertEvent{
		Severity:  severity,
		Status:    status,
		Message:   message,
		MonitorID: monitorID,
		HostID:    hostID,
		Entity:    entity,
		Added:     result.Added,
		Updated:   result.Updated,
		Failed:    result.Failed,
		Total:     result.Total,
	})
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
