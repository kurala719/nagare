package service

import (
	"context"
	"fmt"
	"time"

	"nagare/pkg/queue"
)

var TaskQueue *queue.TaskQueue

// SetTaskQueue sets the global task queue instance
func SetTaskQueue(tq *queue.TaskQueue) {
	TaskQueue = tq
}

// StartTaskWorkers starts worker goroutines for processing tasks
func StartTaskWorkers() {
	workerCount := 4

	for i := 0; i < workerCount; i++ {
		go startWorker(i)
	}

	// Start periodic scheduler
	go startScheduler()

	LogSystem("info", "task workers started", map[string]interface{}{"worker_count": workerCount}, nil, "")
}

func startScheduler() {
	// Periodic heavy tasks (Discovery/Sync) every 5 minutes
	discoveryTicker := time.NewTicker(5 * time.Minute)
	// Metric polling every 1 minute
	pollTicker := time.NewTicker(1 * time.Minute)
	
	defer discoveryTicker.Stop()
	defer pollTicker.Stop()

	// Initial run
	go schedulePeriodicTasks()
	go scheduleMetricPolling()

	for {
		select {
		case <-discoveryTicker.C:
			schedulePeriodicTasks()
		case <-pollTicker.C:
			scheduleMetricPolling()
		}
	}
}

func schedulePeriodicTasks() {
	ctx := context.Background()
	monitors, err := GetAllMonitorsServ()
	if err != nil {
		LogService("error", "failed to get monitors for scheduling", map[string]interface{}{"error": err.Error()}, nil, "scheduler")
		return
	}

	for _, monitor := range monitors {
		if monitor.Status != 1 && monitor.Status != 3 { // Only sync active or syncing monitors
			continue
		}

		// Schedule Group Sync
		if _, err := TaskQueue.Enqueue(ctx, queue.TaskTypePullGroupsFromMonitor, map[string]interface{}{
			"monitor_id": float64(monitor.ID),
		}); err != nil {
			LogService("error", "failed to enqueue pull groups task", map[string]interface{}{"monitor_id": monitor.ID, "error": err.Error()}, nil, "scheduler")
		}

		// Schedule Host Discovery/Sync
		if _, err := TaskQueue.Enqueue(ctx, queue.TaskTypePullHostsFromMonitor, map[string]interface{}{
			"monitor_id": float64(monitor.ID),
		}); err != nil {
			LogService("error", "failed to enqueue pull hosts task", map[string]interface{}{"monitor_id": monitor.ID, "error": err.Error()}, nil, "scheduler")
		}
	}
}

func scheduleMetricPolling() {
	ctx := context.Background()
	monitors, err := GetAllMonitorsServ()
	if err != nil {
		return
	}

	for _, monitor := range monitors {
		if monitor.Status != 1 && monitor.Status != 3 {
			continue
		}

		// Schedule Item Poll (The actual data collection)
		if _, err := TaskQueue.Enqueue(ctx, queue.TaskTypePullItemsFromMonitor, map[string]interface{}{
			"monitor_id": float64(monitor.ID),
		}); err != nil {
			LogService("error", "failed to enqueue pull items task", map[string]interface{}{"monitor_id": monitor.ID, "error": err.Error()}, nil, "scheduler")
		}
	}
}

func startWorker(id int) {
	workerName := fmt.Sprintf("worker-%d", id)
	ctx := context.Background()

	taskTypes := []queue.TaskType{
		queue.TaskTypePullHostsFromMonitor,
		queue.TaskTypePullGroupsFromMonitor,
		queue.TaskTypePullItemsFromMonitor,
		queue.TaskTypeGenerateAlerts,
	}

	for {
		// Use DequeueAny to listen to all queues simultaneously
		// This avoids the latency issue where we block on an empty high-priority queue
		// while lower-priority queues have work.
		task, err := TaskQueue.DequeueAny(ctx, taskTypes, 5*time.Second)
		if err != nil {
			// If error is just a timeout (nil task), we continue
			// If it's a real Redis error, log it
			if task == nil {
				// Timeout, just loop again
				continue
			}
			LogService("error", "failed to dequeue task", map[string]interface{}{"error": err.Error()}, nil, workerName)
			time.Sleep(1 * time.Second) // Backoff on error
			continue
		}

		if task == nil {
			continue
		}

		var processErr error
		switch task.Type {
		case queue.TaskTypePullHostsFromMonitor:
			processErr = processPullHostsTask(task)
		case queue.TaskTypePullGroupsFromMonitor:
			processErr = processPullGroupsTask(task)
		case queue.TaskTypePullItemsFromMonitor:
			processErr = processPullItemsTask(task)
		case queue.TaskTypeGenerateAlerts:
			processErr = processGenerateAlertsTask(task)
		default:
			processErr = fmt.Errorf("unknown task type: %s", task.Type)
		}

		if processErr != nil {
			LogService("error", fmt.Sprintf("failed to process %s task", task.Type), map[string]interface{}{"task_id": task.ID, "error": processErr.Error()}, nil, workerName)
			handleTaskFailure(ctx, task)
		}
	}
}

func processPullHostsTask(task *queue.Task) error {
	monitorID, ok := task.Params["monitor_id"].(float64)
	if !ok {
		return fmt.Errorf("invalid monitor_id")
	}

	_, err := PullHostsFromMonitorServ(uint(monitorID))
	return err
}

func processPullGroupsTask(task *queue.Task) error {
	monitorID, ok := task.Params["monitor_id"].(float64)
	if !ok {
		return fmt.Errorf("invalid monitor_id")
	}

	_, err := PullGroupsFromMonitorServ(uint(monitorID))
	return err
}

func processPullItemsTask(task *queue.Task) error {
	monitorID, ok := task.Params["monitor_id"].(float64)
	if !ok {
		return fmt.Errorf("invalid monitor_id")
	}

	_, err := PullItemsFromMonitorServ(uint(monitorID))
	return err
}

func processGenerateAlertsTask(task *queue.Task) error {
	count, ok := task.Params["count"].(float64)
	if !ok {
		count = 5
	}

	return GenerateTestAlerts(int(count))
}

func handleTaskFailure(ctx context.Context, task *queue.Task) {
	task.Retries++
	if task.Retries >= task.MaxRetry {
		_ = TaskQueue.SendToDeadLetter(ctx, task, "max retries exceeded")
		LogService("warn", "task sent to dead letter", map[string]interface{}{"task_id": task.ID, "retries": task.Retries}, nil, "")
	} else {
		// Re-queue the task
		err := TaskQueue.Requeue(ctx, task)
		if err != nil {
			LogService("error", "failed to re-queue task", map[string]interface{}{"task_id": task.ID, "error": err.Error()}, nil, "")
		}
	}
}
