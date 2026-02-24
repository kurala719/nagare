# Nagare System Architecture: Trigger-Alert-Action Flow

## Current System Overview

The Nagare monitoring system follows a clean architecture pattern with the following flow:

```
Items (Metrics) -> Alerts -> Triggers -> Actions -> Media (Notification)
```

## Components

### 1. Items (Metrics)
- **Model**: `Item` - Represents monitoring metrics from external sources (Zabbix, SNMP, etc.)
- **Properties**: 
  - `ID`, `Name`, `HID` (Host ID), `ItemID` (External ID)
  - `LastValue` (current metric value)
  - `ValueType`, `Units`
  - `Status`, `Enabled`
  - `ItemHistory` - Historical values tracked over time

**Current Role**: Items are passive data holders. They track metrics but don't directly trigger anything.

### 2. Alerts
- **Model**: `Alert` - Represents critical events/conditions
- **Properties**:
  - `ID`, `Message`, `Severity` (0-5+)
  - `Status` (0=active, 1=acknowledged, 2=resolved)
  - `HostID`, `ItemID`, `AlarmID`
  - `Comment` (AI analysis results)

**Current Role**: Alerts come from external sources (webhook) or internal events. They trigger actions through triggers.

### 3. Triggers
- **Model**: `Trigger` - Rules that filter alerts/logs/items and execute actions
- **Properties**:
  - `Entity` - Type of event ("alert", "log", "item")
  - `SeverityMin` - Minimum severity threshold
  - `ActionID` - Associated action to execute
  - Filter conditions:
    - For alerts: `AlertID`, `AlertStatus`, `AlertGroupID`, `AlertMonitorID`, `AlertHostID`, `AlertItemID`, `AlertQuery`
    - For logs: `LogType`, `LogSeverity`, `LogQuery`
    - For items: `ItemStatus`, `ItemValueThreshold`, `ItemValueThresholdMax`, `ItemValueOperator`

**Current Role**: Triggers filter events and decide whether to execute an action. They currently react to alerts/logs/items but don't generate alerts based on item values.

### 4. Actions
- **Model**: `Action` - Execution rules for notifications
- **Properties**:
  - `ID`, `Name`, `MediaID`
  - `Template` - Message template with placeholders
  - `Status`, `Enabled`

**Current Role**: Actions format messages and determine which media to use.

### 5. Media
- **Model**: `Media` - Notification delivery channels
- **Properties**:
  - `Type` (email, other, QQ, etc.)
  - `Target` (email address, webhook/other URL, QQ account)
  - `Params` - Type-specific configuration
  - `Status`, `Enabled`

**Current Role**: Actual notification delivery channel.

## Recommended System Flow (REQUIRED)

```
Items (Metric Values) 
    ↓
    → Trigger Evaluation (based on thresholds/operators)
         ↓
         → Generate Alert (if condition met)
              ↓
              → Actions Filter (match trigger criteria)
                   ↓
                   → Send via Media
```

### The Three-Stage Process

**Stage 1: Trigger Detection** (Items → Alerts)
- Triggers evaluate item values using thresholds and operators
- Conditions: `ItemValueThreshold`, `ItemValueThresholdMax`, `ItemValueOperator` (>, <, >=, <=, ==, !=, between)
- When condition is met: Generate an Alert

**Stage 2: Alert Filtering** (Alerts → Matching Triggers)
- Additional triggers filter the generated alert by:
  - Alert severity, status
  - Host/Group/Monitor/Item associations
  - Alert message patterns
  - These are the action-triggering filters

**Stage 3: Action Execution** (Triggers → Media)
- Matching triggers execute their associated actions
- Actions format messages with templates
- Actions send notifications through media channels

## Required Implementation Changes

### 1. Add Item-Trigger Link (Model Enhancement)
```go
// In Trigger model
ItemIDs []uint // Multiple items can trigger this trigger
```

### 2. Enable Item Value Monitoring
Current trigger fields for items:
- `ItemStatus` - Filter by item status
- `ItemValueThreshold` - Lower bound
- `ItemValueThresholdMax` - Upper bound (for range checks)
- `ItemValueOperator` - Comparison operator

These already exist but need to be properly connected to the item evaluation flow.

### 3. Implement Item → Alert Generation
New service function: `EvaluateItemTriggersServ(item Item)`
- Fetch all triggers with `Entity == "item"`
- Evaluate each trigger's conditions against item value
- If match: Create Alert with references to item, host, monitor, group
- Execute action triggers for the generated alert

### 4. Enhance Action Filtering
Actions should:
1. Receive the generated alert
2. Check if any triggers match
3. Filter by severity, status, relationships
4. Execute matching trigger's action

## Data Flow Example

```
Scenario: CPU Usage Alert

1. Item Update:
   Host: "Server-1", Item: "CPU Usage", Value: 95%
   
2. Trigger Evaluation:
   Trigger: "High CPU"
   - Entity: "item"
   - ItemID: 42 (CPU Usage item)
   - ItemValueOperator: ">"
   - ItemValueThreshold: 90
   - ActionID: 5 (Email Admin)
   
3. Condition Met: 95 > 90
   → Create Alert:
      Message: "High CPU detected on Server-1: 95%"
      Severity: 2
      HostID, ItemID, MonitorID populated
   
4. Action Execution:
   After alert created, execute trigger 5
   → Action 5: Send via Media 3 (Admin Email)
   → Template: "{{host_name}} CPU is {{value}}{{units}}"
   → Result: Email sent to admin@company.com
```

## Current Implementation Status

### ✅ Already Implemented
- Item model with value tracking
- Trigger model with item filter fields
- Action and Media models
- Alert creation from webhooks
- Trigger execution for alerts (reactive)
- Message template substitution

### ❌ Needs Implementation
- Scheduled item evaluation function
- Item-based trigger evaluation logic
- Automatic alert generation from item thresholds
- Enhanced filtering for generated alerts

## Architecture Benefits

1. **Separation of Concerns**: Items measure, triggers decide, actions execute
2. **Flexible Filtering**: Multiple trigger levels for precise control
3. **Extensible**: Easily add new item types, trigger types, or media types
4. **Asynchronous**: Can handle high volume of events
5. **AI-Enhanced**: Supports AI analysis/suppression of alerts

## Files to Modify

### Core Services
- `internal/service/trigger.go` - Add item evaluation logic
- `internal/service/alert.go` - Enhance alert generation
- `internal/service/action.go` - Enhance action execution

### API Endpoints
- `internal/api/trigger.go` - May need new endpoints for item-trigger links
- `internal/api/item.go` - May need manual trigger evaluation endpoint

### Database/Repository
- `internal/repository/trigger.go` - Add item-based queries
- Database schema migration if needed

## Next Steps

1. ✓ Understand current architecture
2. Implement `EvaluateItemTriggersServ()` function
3. Add cron/scheduled evaluation for items
4. Enhance `ExecuteTriggersForItem()` to generate alerts
5. Test end-to-end: Item value → Alert → Action → Notification
6. Add API endpoint to manually trigger item evaluation
7. Add monitoring/logging for trigger evaluations
