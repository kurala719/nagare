# Status Validation Implementation Summary

## Overview
Implemented a hierarchical health-check-first architecture for data synchronization. Before pulling or pushing monitor data, the system now validates that both the monitor and host are active. If either is inactive, the sync operation is skipped and appropriate error reasons are recorded in the new `status_description` field.

---

## Changes Made

### 1. Domain Entities (entities.go)
**Added `StatusDescription` field to three core entities:**

- **Monitor struct**: 
  - Added `StatusDescription string` - tracks reason for error status (e.g., "connection timeout", "authentication failed")

- **Host struct**: 
  - Added `StatusDescription string` - tracks reason for error status (e.g., "monitor is down", "connection failed")

- **Item struct**: 
  - Added `StatusDescription string` - tracks reason for error status (e.g., "host is down", "pull failed")

#### Purpose:
Separates error reasoning from the `comment` field, which is now reserved exclusively for human/AI operational notes. Each entity can now record why it entered an error state (status=2).

---

### 2. Infrastructure DAOs

#### host.go
- **Updated `UpdateHostDAO()`**: Now includes `status_description` in the update map
- **Added `UpdateHostStatusAndDescriptionDAO(id, status, statusDesc)`**: New DAO function to atomically update both status and status_description for hosts

#### item.go
- **Updated `UpdateItemDAO()`**: Now includes `status_description` in the update map
- **Added `UpdateItemStatusAndDescriptionDAO(id, status, statusDesc)`**: New DAO function to atomically update both status and status_description for items

#### monitor.go
- **Updated `UpdateMonitorDAO()`**: Now includes `status_description` in the update map
- **Added `UpdateMonitorStatusAndDescriptionDAO(id, status, statusDesc)`**: New DAO function to atomically update both status and status_description for monitors

---

### 3. Application Layer Status Helpers (status.go)

**Updated status management functions:**

- **`setHostStatusErrorWithReason(hid, reason)`**: Changed from using comment to using new `UpdateHostStatusAndDescriptionDAO()` - now stores error reasons in `status_description` field

- **`setItemStatusErrorWithReason(id, reason)`**: Changed from using comment to using new `UpdateItemStatusAndDescriptionDAO()` - now stores error reasons in `status_description` field

- **Added `setMonitorStatusErrorWithReason(mid, reason)`**: New function to store monitor error reasons in `status_description` field

---

### 4. Host Management Application Logic (host.go)

#### PullHostsFromMonitorServ(mid)
**Added pre-flight monitor health check:**
- Retrieves monitor and checks if `status == 1` (active)
- If monitor is not active:
  - Calls `setMonitorStatusErrorWithReason()` with reason (either generic "monitor is not active" or monitor's statusDescription if set)
  - Logs warning event with status details
  - Returns early without attempting to pull hosts
- If monitor is active, proceeds with normal pull flow

#### PushHostsFromMonitorServ(mid)
**Added pre-flight monitor health check:**
- Retrieves monitor and checks if `status == 1` (active)
- If monitor is not active:
  - Calls `setMonitorStatusErrorWithReason()` with reason
  - Logs warning event with status details
  - Returns early without attempting to push hosts
- If monitor is active, proceeds with normal push flow

---

### 5. Item Management Application Logic (item.go)

#### PullItemsFromMonitorServ(mid)
**Added pre-flight monitor health check:**
- Retrieves monitor and validates `status == 1` (active)
- If monitor is not active:
  - Calls `setMonitorStatusErrorWithReason()` with status reason
  - Marks all associated hosts as error with "monitor is down" reason (cascading failure)
  - Logs warning event
  - Returns early without processing hosts
- If monitor is active, proceeds to pull items from all hosts

#### PullItemsFromHostServ(mid, hid)
**Added pre-flight host health check:**
- After validating monitor consistency, checks if `host.status == 1` (active)
- If host is not active:
  - Calls `setHostStatusErrorWithReason()` with status reason
  - Logs warning event
  - Returns early without fetching items from monitor
- If host is active, proceeds with full item pull flow

#### PushItemsFromMonitorServ(mid)
**Added pre-flight monitor health check:**
- Retrieves monitor and validates `status == 1` (active)
- If monitor is not active:
  - Calls `setMonitorStatusErrorWithReason()` with status reason
  - Logs warning event
  - Returns early without processing hosts
- If monitor is active, proceeds to push items from all hosts

#### PushItemsFromHostServ(mid, hid)
**Added pre-flight host health check:**
- Retrieves host and validates `status == 1` (active)
- If host is not active:
  - Calls `setHostStatusErrorWithReason()` with status reason
  - Logs warning event
  - Returns early without pushing items to monitor
- If host is active, proceeds with normal push flow

---

## Synchronization Flow Hierarchy

### Pull Operations
```
PullItemsFromMonitorServ(mid)
  ├─ Check: monitor.status == 1?
  │  └─ If NO → Set monitor error with reason, mark all hosts as "monitor is down" → RETURN
  ├─ For each host in monitor:
  │  └─ PullItemsFromHostServ(mid, hid)
  │     ├─ Check: host.status == 1?
  │     │  └─ If NO → Set host error with reason → RETURN
  │     └─ Pull items from monitor for this host
```

### Push Operations
```
PushItemsFromMonitorServ(mid)
  ├─ Check: monitor.status == 1?
  │  └─ If NO → Set monitor error with reason → RETURN
  ├─ For each host in monitor:
  │  └─ PushItemsFromHostServ(mid, hid)
  │     ├─ Check: host.status == 1?
  │     │  └─ If NO → Set host error with reason → RETURN
  │     └─ Push items from all items for this host
```

---

## Status Values Reference
- **0**: Inactive (disabled/not monitored)
- **1**: Active (healthy, ready for sync)
- **2**: Error (problem detected; reason in statusDescription)
- **3**: Syncing (operation in progress)

---

## Benefits

1. **Cascading Failure Prevention**: Skips expensive remote API calls when upstream components (monitor/host) are down
2. **Clear Error Tracking**: Error reasons stored in dedicated `status_description` field instead of polluting `comment` field
3. **Better Observability**: Logs include status and reason, enabling troubleshooting
4. **Atomicity**: Status and reason updated together, preventing data inconsistency
5. **Resource Efficiency**: Eliminates wasted sync attempts for unavailable monitors/hosts
6. **Clean Separation**: Comment field remains for human/AI operational notes only

---

## Database Migration Required

To support the new `status_description` column, a migration must be run:

```sql
-- Add status_description column to monitors table
ALTER TABLE monitors ADD COLUMN status_description VARCHAR(255) DEFAULT '';

-- Add status_description column to hosts table
ALTER TABLE hosts ADD COLUMN status_description VARCHAR(255) DEFAULT '';

-- Add status_description column to items table
ALTER TABLE items ADD COLUMN status_description VARCHAR(255) DEFAULT '';
```

---

## Testing Recommendations

1. **Monitor Down Scenario**: Set monitor.status=2 and verify:
   - PullItemsFromMonitorServ returns early with error
   - All hosts marked as error with "monitor is down" reason
   - statusDescription field populated correctly

2. **Host Down Scenario**: Set host.status=2 and verify:
   - PullItemsFromHostServ returns early with error
   - statusDescription field contains host status reason
   - Items for that host not pulled

3. **Monitor Recovery**: After fixing monitor issue and setting status=1:
   - Verify sync operations proceed normally
   - statusDescription can be cleared or inherited from host/item recovery

4. **Cascading Updates**: Verify monitor failure cascades to all hosts with proper error reasons

---

## Files Modified

1. `domain/entities.go` - Added StatusDescription to Monitor, Host, Item
2. `infrastructure/host.go` - Added UpdateHostStatusAndDescriptionDAO
3. `infrastructure/item.go` - Added UpdateItemStatusAndDescriptionDAO
4. `infrastructure/monitor.go` - Added UpdateMonitorStatusAndDescriptionDAO
5. `application/status.go` - Updated error reason handlers to use new DAO functions
6. `application/host.go` - Added health checks in Pull/Push operations
7. `application/item.go` - Added health checks in Pull/Push operations

---

## Backward Compatibility

- All changes are backward compatible
- Existing `status` field logic unchanged
- New `status_description` field is optional (defaults to empty string)
- Existing code paths continue to work; health checks are additive safeguards
