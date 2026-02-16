package queue

import (
	"context"
	"testing"
	"time"
)

func TestQueueOperations(t *testing.T) {
	// Try to connect to local Redis
	tq, err := New("localhost:6379")
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer tq.Close()

	ctx := context.Background()
	// Flush for clean state
	if err := tq.FlushAll(ctx); err != nil {
		t.Fatalf("Failed to flush queue: %v", err)
	}

	// Test Enqueue
	params := map[string]interface{}{"foo": "bar"}
	taskID, err := tq.Enqueue(ctx, TaskTypePullHostsFromMonitor, params)
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	if taskID == "" {
		t.Fatal("Expected task ID, got empty")
	}

	// Test DequeueAny
	types := []TaskType{TaskTypePullHostsFromMonitor, TaskTypeGenerateAlerts}
	task, err := tq.DequeueAny(ctx, types, 2*time.Second)
	if err != nil {
		t.Fatalf("DequeueAny failed: %v", err)
	}
	if task == nil {
		t.Fatal("Expected task, got nil")
	}
	if task.ID != taskID {
		t.Errorf("Expected task ID %s, got %s", taskID, task.ID)
	}
	if task.Type != TaskTypePullHostsFromMonitor {
		t.Errorf("Expected task type %s, got %s", TaskTypePullHostsFromMonitor, task.Type)
	}

	// Test Requeue
	task.Retries = 1
	err = tq.Requeue(ctx, task)
	if err != nil {
		t.Fatalf("Requeue failed: %v", err)
	}

	// Dequeue again to verify it's back and retries preserved
	task2, err := tq.DequeueAny(ctx, types, 2*time.Second)
	if err != nil {
		t.Fatalf("DequeueAny after Requeue failed: %v", err)
	}
	if task2 == nil {
		t.Fatal("Expected task after Requeue, got nil")
	}
	if task2.ID != taskID {
		t.Errorf("Expected task ID %s, got %s", taskID, task2.ID)
	}
	if task2.Retries != 1 {
		t.Errorf("Expected retries 1, got %d", task2.Retries)
	}
}
