package service

import (
	"context"
	"fmt"

	"nagare/pkg/queue"
)

// PullHostsFromMonitorAsyncServ queues a host pull task
func PullHostsFromMonitorAsyncServ(mid uint, recordHistory bool) (string, error) {
	if TaskQueue == nil {
		return "", fmt.Errorf("task queue not initialized")
	}

	params := map[string]interface{}{
		"monitor_id":     mid,
		"record_history": recordHistory,
	}

	taskID, err := TaskQueue.Enqueue(context.Background(), queue.TaskTypePullHostsFromMonitor, params)
	if err != nil {
		return "", err
	}

	LogService("info", "pull hosts task queued", map[string]interface{}{"monitor_id": mid, "task_id": taskID}, nil, "")
	return taskID, nil
}

// PullItemsFromMonitorAsyncServ queues an item pull task
func PullItemsFromMonitorAsyncServ(mid, hid uint, recordHistory bool) (string, error) {
	if TaskQueue == nil {
		return "", fmt.Errorf("task queue not initialized")
	}

	params := map[string]interface{}{
		"monitor_id":     mid,
		"host_id":        hid,
		"record_history": recordHistory,
	}

	taskID, err := TaskQueue.Enqueue(context.Background(), queue.TaskTypePullItemsFromMonitor, params)
	if err != nil {
		return "", err
	}

	LogService("info", "pull items task queued", map[string]interface{}{"monitor_id": mid, "host_id": hid, "task_id": taskID}, nil, "")
	return taskID, nil
}

// GetQueueStats returns stats for all task queues
func GetQueueStats() (map[string]int64, error) {
	if TaskQueue == nil {
		return nil, fmt.Errorf("task queue not initialized")
	}

	return TaskQueue.GetAllQueueStats(context.Background())
}
