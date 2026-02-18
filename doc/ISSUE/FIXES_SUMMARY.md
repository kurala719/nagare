# Fixes Summary - 2026-02-16

## 1. Infinite Retry Loop in Task Queue
**Issue:**
When a task failed and needed to be retried, the `handleTaskFailure` function called `TaskQueue.Enqueue`. The `Enqueue` method creates a *new* task with a fresh ID and resets `Retries` to 0 (default). This caused failed tasks to be retried infinitely, ignoring the `MaxRetry` limit.

**Fix:**
- Added `Requeue(ctx, task)` method to `TaskQueue` (in `pkg/queue/queue.go`). This method pushes the *existing* task struct (with incremented `Retries`) back to the queue.
- Updated `handleTaskFailure` (in `application/worker.go`) to use `Requeue` instead of `Enqueue`.

## 2. Worker Latency/Starvation
**Issue:**
The worker loop sequentially checked queues (`pull_hosts`, `pull_items`, `generate_alerts`) using blocking pops (`BRPop`) with a 5-second timeout.
If the `pull_hosts` queue was empty, the worker would block for 5 seconds before checking `pull_items`. This introduced significant latency for lower-priority tasks and potential starvation if the first queue was always empty but the worker was just waiting.

**Fix:**
- Added `DequeueAny(ctx, types, timeout)` to `TaskQueue` (in `pkg/queue/queue.go`). This uses Redis `BRPOP` with multiple keys to listen to all specified queues simultaneously.
- Refactored `startWorker` (in `application/worker.go`) to use `DequeueAny` with all 3 task types. Now the worker wakes up immediately if *any* queue has a task.

## 3. Hardcoded Secrets & Configuration
**Issue:**
`nagare_config.json` contained hardcoded secrets (database passwords, API keys).
While the file might be for local development, the application did not support overriding these values via environment variables, posing a security risk for deployment.

**Fix:**
- Updated `InitConfig` (in `infrastructure/config.go`) to enable Viper's environment variable support.
- Configured prefix `NAGARE_` and automatic key replacement (e.g., `NAGARE_DATABASE_PASSWORD` overrides `database.password`).

## Verification
- Added `pkg/queue/queue_test.go` to verify queue operations (skipped if Redis is unavailable).
- Backend compiles successfully.
