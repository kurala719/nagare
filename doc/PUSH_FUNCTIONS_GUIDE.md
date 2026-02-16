# Host and Item Push Functions - Implementation Guide

## Overview
The push functions sync data from the local Nagare database to remote monitoring systems (Zabbix, Prometheus, etc.). This is the inverse of the pull functions which sync data from remote monitors to local database.

## Architecture

### Host Push Functions

#### 1. `PushHostToMonitorServ(mid uint, id uint) error`
**Purpose**: Push a single host from local database to a remote monitor

**Location**: [service/host.go](../backend/internal/service/host.go)

**Flow**:
1. Retrieve host from local database by ID
2. Validate that host belongs to the specified monitor
3. Create monitor client and authenticate
4. Log the push operation (ready for API implementation)
5. Return any errors

**Endpoint**: `GET /host/push/:mid/:id`

**Example**:
```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/host/push/1/5
```

#### 2. `PushHostsFromMonitorServ(mid uint) (SyncResult, error)`
**Purpose**: Push all hosts for a monitor from local database to remote

**Location**: [service/host.go](../backend/internal/service/host.go)

**Flow**:
1. Get all hosts for the specified monitor
2. Iterate through each host
3. Call `PushHostToMonitorServ()` for each
4. Aggregate results (added, failed, total)
5. Return sync result

**Endpoint**: `GET /host/push/:mid`

**Example**:
```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/host/push/1
```

**Response**:
```json
{
  "success": true,
  "data": {
    "added": 5,
    "updated": 0,
    "failed": 0,
    "total": 5
  }
}
```

---

### Item Push Functions

#### 1. `PushItemToMonitorServ(mid, hid, id uint) error`
**Purpose**: Push a single item from local database to a remote monitor

**Location**: [service/item.go](../backend/internal/service/item.go)

**Flow**:
1. Retrieve item, host, and monitor from local database
2. Validate relationships (item belongs to host, host belongs to monitor)
3. Create monitor client and authenticate
4. Log the push operation (ready for API implementation)
5. Return any errors

**Endpoint**: `GET /item/push/:mid/:hid/:id`

**Example**:
```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/item/push/1/3/12
```

#### 2. `PushItemsFromHostServ(mid, hid uint) (SyncResult, error)`
**Purpose**: Push all items for a specific host to remote monitor

**Location**: [service/item.go](../backend/internal/service/item.go)

**Flow**:
1. Get all items for the specified host
2. Iterate through each item
3. Call `PushItemToMonitorServ()` for each
4. Aggregate results
5. Return sync result

**Endpoint**: `GET /item/push/:mid/:hid`

**Example**:
```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/item/push/1/3
```

#### 3. `PushItemsFromMonitorServ(mid uint) (SyncResult, error)`
**Purpose**: Push all items from all hosts for a monitor to remote

**Location**: [service/item.go](../backend/internal/service/item.go)

**Flow**:
1. Get all hosts for the specified monitor
2. For each host, call `PushItemsFromHostServ()`
3. Aggregate all results
4. Return combined sync result

**Endpoint**: `GET /item/push/:mid`

**Example**:
```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/item/push/1
```

---

## API Reference

### Host Push Endpoints

| Method | Endpoint | Description | Auth Level |
|--------|----------|-------------|-----------|
| GET | `/host/push/:mid` | Push all hosts for monitor | admin (≥2) |
| GET | `/host/push/:mid/:id` | Push single host | admin (≥2) |

### Item Push Endpoints

| Method | Endpoint | Description | Auth Level |
|--------|----------|-------------|-----------|
| GET | `/item/push/:mid` | Push all items for monitor | admin (≥2) |
| GET | `/item/push/:mid/:hid` | Push all items for host | admin (≥2) |
| GET | `/item/push/:mid/:hid/:id` | Push single item | admin (≥2) |

---

## Response Format

### Success Response
```json
{
  "success": true,
  "data": {
    "added": 10,
    "updated": 2,
    "failed": 0,
    "total": 12
  }
}
```

### Error Response
```json
{
  "success": false,
  "error": "failed to create monitor client: connection refused"
}
```

---

## SyncResult Structure

```go
type SyncResult struct {
    Added   int // Number of successfully pushed items
    Updated int // Number of updated items
    Failed  int // Number of failed push operations
    Total   int // Total number of items processed
}
```

---

## Current Implementation Status

✅ **Completed**:
- Application layer functions fully implemented
- Presentation controller functions created
- Router endpoints configured
- Error handling and validation
- Logging support

🔄 **Ready for Extension**:
- Add actual API calls to monitoring systems
- Implement batch push operations
- Add transaction support for consistency
- Support for conditional push (only changed items)
- Webhook notifications after push

---

## Future Enhancements

### 1. Actual Monitor API Integration
Replace the placeholder log statements with actual API calls:

```go
// Example for Zabbix
resp, err := client.CreateHost(context.Background(), &monitors.Host{
    Name:        host.Name,
    Description: host.Description,
    // ... other fields
})
if err != nil {
    return fmt.Errorf("failed to push host to Zabbix: %w", err)
}
```

### 2. Batch Operations
Optimize performance for pushing multiple items at once

### 3. Conflict Resolution
Handle cases where items differ between local and remote:
- Always push local (current behavior ready)
- Always pull remote (merge behavior)
- Manual resolution

### 4. Differential Sync
Only push items that have changed since last sync using timestamps

### 5. Undo/Rollback
Support rolling back pushed changes

---

## Testing

### Unit Tests
Test each function individually with mocked infrastructure layer

### Integration Tests
Test against real monitoring system instances

### Example Test
```go
func TestPushHostToMonitor(t *testing.T) {
    // Setup
    monitor := &domain.Monitor{ID: 1, Name: "Test Monitor", Type: "zabbix"}
    host := &domain.Host{ID: 1, Name: "test-host", MonitorID: 1}
    
    // Execute
    err := application.PushHostToMonitorServ(1, 1)
    
    // Assert
    assert.NoError(t, err)
}
```

---

## Troubleshooting

### "host does not belong to the specified monitor"
Ensure the host's `MonitorID` matches the monitor ID in the URL

### "failed to authenticate with monitor"
Verify:
- Monitor credentials are correct
- Monitor is online and accessible
- Network connectivity

### Empty results
- Check that items/hosts exist in the local database
- Verify the monitor ID and host ID parameters
