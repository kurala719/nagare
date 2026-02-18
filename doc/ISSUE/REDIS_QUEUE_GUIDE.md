# Redis Task Queue Implementation Guide

## Overview

This implementation provides a non-blocking asynchronous task queue system using Redis to solve synchronization blocking issues during host/item syncing operations.

## Features

- **Async Operations**: Queue long-running sync tasks instead of blocking HTTP requests
- **Worker Pool**: Multiple goroutine workers process queued tasks concurrently
- **Retry Logic**: Failed tasks are retried up to 3 times before being sent to dead letter queue
- **Mock Alert Generator**: Generate test alerts for development/testing
- **Task Status Monitoring**: Check queue statistics via API

## Architecture

```
┌─────────────────────────────────────────┐
│         HTTP Request (Sync API)         │
├─────────────────────────────────────────┤
│  POST /api/v1/monitors/:id/hosts/pull   │
│  201 Accepted { task_id: "..." }        │
└────────────────┬────────────────────────┘
                 │
                 ▼
         ┌───────────────┐
         │   Redis       │
         │   Queue       │
         └───────┬───────┘
                 │
         ┌───────┴────────┐
         │                │
    ┌────▼────┐  ┌────▼────┐
    │ Worker 1│  │ Worker 2│ ... (N workers)
    └────┬────┘  └────┬────┘
         │            │
         └────┬───────┘
              │
              ▼
       ┌──────────────┐
       │  Processing  │
       │ (Pull hosts) │
       │ (Pull items) │
       │ (Gen alerts) │
       └──────────────┘
```

## Configuration

Add Redis configuration to `nagare_config.json`:

```json
{
  "redis": {
    "addr": "localhost:6379"
  }
}
```

Or set environment variable:
```bash
export REDIS_ADDR="localhost:6379"
```

## API Endpoints

### Async Host Pull
```bash
POST /api/v1/monitors/:id/hosts/pull-async

Response (202 Accepted):
{
  "message": "Host pull task queued",
  "task_id": "pull_hosts:1707996000000000000",
  "monitor_id": 1
}
```

### Async Item Pull
```bash
POST /api/v1/monitors/:m_id/hosts/:h_id/items/pull-async

Response (202 Accepted):
{
  "message": "Item pull task queued",
  "task_id": "pull_items:1707996000000000001",
  "monitor_id": 1,
  "host_id": 5
}
```

### Generate Test Alerts
```bash
POST /api/v1/alerts/generate-test?count=5

Response (201 Created):
{
  "message": "Test alerts generated",
  "count": 5
}
```

### Get Alert Score
```bash
GET /api/v1/alerts/score

Response (200 OK):
{
  "score": 78
}
```

### Queue Statistics
```bash
GET /api/v1/queue/stats

Response (200 OK):
{
  "pull_hosts": 3,
  "pull_items": 5,
  "generate_alerts": 2,
  "pull_host": 0,
  "pull_item": 0,
  "push_host": 0,
  "push_item": 0
}
```

## Worker Configuration

Workers are started automatically during application startup. To adjust worker count, edit `backend/internal/service/worker.go`:

```go
func StartTaskWorkers() {
    workerCount := 4  // Change this value
    // ...
}
```

## Test Scenario

1. **Start Application**:
```bash
cd backend
go run cmd/server/main.go
```

2. **Queue Async Tasks**:
```bash
curl -X POST http://localhost:8080/api/v1/monitors/1/hosts/pull-async \
  -H "Authorization: Bearer <token>"

# Returns immediately with task_id
```

3. **Monitor Queue**:
```bash
curl -X GET http://localhost:8080/api/v1/queue/stats \
  -H "Authorization: Bearer <token>"

# Shows pending tasks: { "pull_hosts": 1, ... }
```

4. **Generate Test Alerts**:
```bash
curl -X POST "http://localhost:8080/api/v1/alerts/generate-test?count=10" \
  -H "Authorization: Bearer <token>"
```

5. **Check Alert Health**:
```bash
curl -X GET http://localhost:8080/api/v1/alerts/score \
  -H "Authorization: Bearer <token>"

# { "score": 65 }
```

## Task Types

| Type | Queue Name | Worker Handler |
|------|-----------|-----------------|
| `pull_hosts` | `nagare:queue:pull_hosts` | PullHostsFromMonitorServ |
| `pull_items` | `nagare:queue:pull_items` | PullItemsFromMonitorServ |
| `pull_host` | `nagare:queue:pull_host` | PullHostFromMonitorServ |
| `pull_item` | `nagare:queue:pull_item` | PullItemFromMonitorServ |
| `push_host` | `nagare:queue:push_host` | PushHostToMonitorServ |
| `push_item` | `nagare:queue:push_item` | PushItemToMonitorServ |
| `generate_alerts` | `nagare:queue:generate_alerts` | GenerateTestAlerts |

## Error Handling

- Failed tasks are retried up to 3 times
- Tasks exceeding max retries are moved to dead letter queue: `nagare:queue:dead`
- Worker logs failures with task ID for debugging

## Performance Benefits

- **Non-blocking**: HTTP requests return immediately with 202 Accepted
- **Parallel processing**: Multiple workers handle tasks concurrently
- **Scalable**: Can add more workers or Redis instances for higher throughput
- **Resilient**: Automatic retry and dead letter queue for fault tolerance

## Monitoring

Check task queue metrics:
```bash
# Get queue stats
curl http://localhost:8080/api/v1/queue/stats

# Tail application logs for worker activity
tail -f logs/system.log | grep worker
```

## Future Enhancements

- Task result callbacks via webhook
- Task progress tracking
- Distributed worker pool across multiple servers
- Redis Sentinel for high availability
- Dead letter queue analysis and recovery
