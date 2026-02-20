package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TaskType defines the type of task
type TaskType string

const (
	TaskTypePullHostsFromMonitor  TaskType = "pull_hosts"
	TaskTypePullGroupsFromMonitor TaskType = "pull_groups"
	TaskTypePullItemsFromMonitor  TaskType = "pull_items"
	TaskTypePullHostFromMonitor   TaskType = "pull_host"
	TaskTypePullItemFromMonitor   TaskType = "pull_item"
	TaskTypePushHostToMonitor     TaskType = "push_host"
	TaskTypePushItemToMonitor     TaskType = "push_item"
	TaskTypeGenerateAlerts        TaskType = "generate_alerts"
)

// Task represents a task to be queued
type Task struct {
	ID        string                 `json:"id"`
	Type      TaskType               `json:"type"`
	Params    map[string]interface{} `json:"params"`
	CreatedAt int64                  `json:"created_at"`
	Retries   int                    `json:"retries"`
	MaxRetry  int                    `json:"max_retry"`
}

// TaskQueue manages task queueing with Redis
type TaskQueue struct {
	client   *redis.Client
	queues   map[TaskType]string
	deadLett string
}

// New creates a new TaskQueue instance
func New(redisAddr string) (*TaskQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	tq := &TaskQueue{
		client:   client,
		queues:   make(map[TaskType]string),
		deadLett: "nagare:queue:dead",
	}

	// Register queue names
	tq.queues[TaskTypePullHostsFromMonitor] = "nagare:queue:pull_hosts"
	tq.queues[TaskTypePullGroupsFromMonitor] = "nagare:queue:pull_groups"
	tq.queues[TaskTypePullItemsFromMonitor] = "nagare:queue:pull_items"
	tq.queues[TaskTypePullHostFromMonitor] = "nagare:queue:pull_host"
	tq.queues[TaskTypePullItemFromMonitor] = "nagare:queue:pull_item"
	tq.queues[TaskTypePushHostToMonitor] = "nagare:queue:push_host"
	tq.queues[TaskTypePushItemToMonitor] = "nagare:queue:push_item"
	tq.queues[TaskTypeGenerateAlerts] = "nagare:queue:generate_alerts"

	return tq, nil
}

// Enqueue adds a task to the queue
func (tq *TaskQueue) Enqueue(ctx context.Context, tt TaskType, params map[string]interface{}) (string, error) {
	task := &Task{
		ID:        fmt.Sprintf("%s:%d", tt, time.Now().UnixNano()),
		Type:      tt,
		Params:    params,
		CreatedAt: time.Now().Unix(),
		MaxRetry:  3,
	}

	taskJSON, err := json.Marshal(task)
	if err != nil {
		return "", err
	}

	queueName := tq.queues[tt]
	if queueName == "" {
		return "", fmt.Errorf("unknown task type: %s", tt)
	}

	if err := tq.client.LPush(ctx, queueName, taskJSON).Err(); err != nil {
		return "", err
	}

	return task.ID, nil
}

// Dequeue retrieves a task from the queue
func (tq *TaskQueue) Dequeue(ctx context.Context, tt TaskType, timeout time.Duration) (*Task, error) {
	queueName := tq.queues[tt]
	if queueName == "" {
		return nil, fmt.Errorf("unknown task type: %s", tt)
	}

	result, err := tq.client.BRPop(ctx, timeout, queueName).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No task available
		}
		return nil, err
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid response from Redis")
	}

	var task Task
	if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
		return nil, err
	}

	return &task, nil
}

// DequeueAny retrieves a task from any of the specified task types
func (tq *TaskQueue) DequeueAny(ctx context.Context, types []TaskType, timeout time.Duration) (*Task, error) {
	var queueNames []string
	for _, tt := range types {
		if qName, ok := tq.queues[tt]; ok {
			queueNames = append(queueNames, qName)
		}
	}

	if len(queueNames) == 0 {
		return nil, fmt.Errorf("no valid task types provided")
	}

	// BRPop returns [key, value]
	result, err := tq.client.BRPop(ctx, timeout, queueNames...).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No task available (timeout)
		}
		return nil, err
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid response from Redis")
	}

	// result[0] is the queue name, result[1] is the task JSON
	taskJSON := result[1]

	var task Task
	if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
		return nil, err
	}

	return &task, nil
}

// Requeue adds an existing task back to the queue (e.g. for retry)
func (tq *TaskQueue) Requeue(ctx context.Context, task *Task) error {
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return err
	}

	queueName := tq.queues[task.Type]
	if queueName == "" {
		return fmt.Errorf("unknown task type: %s", task.Type)
	}

	if err := tq.client.LPush(ctx, queueName, taskJSON).Err(); err != nil {
		return err
	}

	return nil
}

// SendToDeadLetter moves a task to dead letter queue after max retries
func (tq *TaskQueue) SendToDeadLetter(ctx context.Context, task *Task, reason string) error {
	data := map[string]interface{}{
		"task":      task,
		"reason":    reason,
		"timestamp": time.Now().Unix(),
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return tq.client.LPush(ctx, tq.deadLett, dataJSON).Err()
}

// GetQueueLength returns the length of a queue
func (tq *TaskQueue) GetQueueLength(ctx context.Context, tt TaskType) (int64, error) {
	queueName := tq.queues[tt]
	if queueName == "" {
		return 0, fmt.Errorf("unknown task type: %s", tt)
	}

	return tq.client.LLen(ctx, queueName).Result()
}

// GetAllQueueStats returns stats for all queues
func (tq *TaskQueue) GetAllQueueStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)
	for tt, queueName := range tq.queues {
		len, err := tq.client.LLen(ctx, queueName).Result()
		if err != nil {
			return nil, err
		}
		stats[string(tt)] = len
	}
	return stats, nil
}

// Close closes the Redis connection
func (tq *TaskQueue) Close() error {
	return tq.client.Close()
}

// FlushAll clears all queues (for testing)
func (tq *TaskQueue) FlushAll(ctx context.Context) error {
	for _, queueName := range tq.queues {
		if err := tq.client.Del(ctx, queueName).Err(); err != nil {
			return err
		}
	}
	return tq.client.Del(ctx, tq.deadLett).Err()
}
