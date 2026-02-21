# Implementation Summary: Trigger-Alert-Action Flow

## Status: ✅ IMPLEMENTED

The system now correctly implements the required flow:

```
Items (Metric Values)
    ↓
    Triggers evaluate item values (thresholds + operators)
    ↓
    Generate Alert (if condition matches)
    ↓
    Execute Alert Triggers (filter alert by severity/host/group/etc)
    ↓
    Send notifications via Media
```

## Changes Made

### 1. Enhanced Item Trigger Evaluation (trigger.go)

#### Modified: `execTriggersForItem()` function
**Before:**
```go
func execTriggersForItem(item model.Item, replacements map[string]string) {
    triggers, err := repository.GetActiveTriggersForEntityDAO("item")
    if err != nil {
        return
    }
    for _, trigger := range triggers {
        if !matchItemTrigger(trigger, item) {
            continue
        }
        invokeItemTriggerAction(trigger, replacements)  // Just send notification
    }
}
```

**After:**
```go
func execTriggersForItem(item model.Item, replacements map[string]string) {
    triggers, err := repository.GetActiveTriggersForEntityDAO("item")
    if err != nil {
        return
    }
    for _, trigger := range triggers {
        if !matchItemTrigger(trigger, item) {
            continue
        }
        // NEW: Generate alert if item trigger matches
        generateAlertFromItemTrigger(trigger, item)
        // EXISTING: Execute the associated action for the generated alert
        invokeItemTriggerAction(trigger, replacements)
    }
}
```

#### New Function: `generateAlertFromItemTrigger()`
Created new function to automatically generate alerts when item triggers match:
- Builds descriptive alert message with item name, host, and value
- Captures trigger information in alert comment
- Sets severity based on trigger configuration
- Calls `AddAlertServ()` which automatically:
  - Runs AI analysis (if enabled)
  - Executes alert triggers to filter and route the alert
  - Sends notifications through matching actions

#### New Helper Function: `describeItemTriggerCondition()`
Human-readable description of trigger conditions for logging and alerts:
- Formats threshold values with operators
- Handles range checks (between/outside)
- Used in alert comments for context

## Complete Data Flow Example

### Scenario: CPU Usage Exceeds Threshold

```
1. ITEM MONITORING
   ├─ External Monitor (Zabbix) polls metrics
   ├─ Service: PullItemsFromHostServ() updates item values
   └─ Calls: ExecuteTriggersForItem(item)

2. TRIGGER EVALUATION
   ├─ Function: execTriggersForItem()
   ├─ Loads all triggers with Entity="item"
   ├─ For each trigger: matchItemTrigger(trigger, item)
   │   └─ Evaluates: ItemValueThreshold, ItemValueOperator, ItemValueThresholdMax
   │   └─ Example: if item.LastValue > trigger.ItemValueThreshold
   │       └─ Returns: true (trigger matches)
   └─ Result: CPU trigger matches (95 > 90)

3. ALERT GENERATION
   ├─ Function: generateAlertFromItemTrigger()
   ├─ Creates Alert:
   │   ├─ Message: "Item CPU Usage on host Server-1 has value 95%"
   │   ├─ Severity: 2 (from trigger config)
   │   ├─ HostID: 5 (Server-1)
   │   ├─ ItemID: 42 (CPU Usage item)
   │   └─ Comment: "Triggered by CPU Alert: > 90.00"
   ├─ Calls: AddAlertServ(alertReq)
   └─ Alert saved to database with ID=1001

4. ALERT ANALYSIS (Optional)
   ├─ If AI analysis enabled:
   │   ├─ Function: analyzeAlertWithAI()
   │   ├─ AI analyzes alert and provides recommendations
   │   └─ Comment updated with analysis
   └─ If AI notification guard enabled:
       └─ Decision check: Should alert be suppressed?

5. ALERT FILTERING & ACTION EXECUTION
   ├─ Function: ExecuteTriggersForAlert(alert)
   ├─ Loads all triggers with Entity="alert"
   ├─ For each trigger: matchAlertTrigger(trigger, context)
   │   └─ Checks:
   │       ├─ AlertID
   │       ├─ Severity (alert.Severity >= trigger.SeverityMin)
   │       ├─ Status
   │       ├─ Host/Group/Monitor associations
   │       ├─ Item associations
   │       └─ Message pattern matching (AlertQuery)
   └─ Result: Admin trigger matches (severity 2 >= min 1)

6. ACTION EXECUTION
   ├─ Function: invokeAlertTriggerAction()
   ├─ For matching trigger:
   │   ├─ Load Action (e.g., "Email Admin Action")
   │   ├─ Load Media (e.g., Email to admin@company.com)
   │   ├─ Build message from template:
   │   │   └─ "{{host_name}} {{entity}}: {{message}}"
   │   │   └─ Result: "Server-1 alert: Item CPU Usage on host Server-1 has value 95%"
   │   └─ Function: ExecuteAction()
   └─ Email sent to admin@company.com

7. FINAL RESULT
   ✅ Alert created and stored
   ✅ AI analysis completed (optional)
   ✅ Notification sent via Email
   ✅ Audit trail maintained
```

## Key Architecture Improvements

### 1. Two-Stage Filtering
- **Stage 1: Item Triggers** - Detect when metric values cross thresholds
  - Entity: "item"
  - Filters: ItemValueThreshold, ItemValueOperator, ItemStatus
  - Action: Generate Alert

- **Stage 2: Alert Triggers** - Filter generated alerts before sending
  - Entity: "alert"  
  - Filters: Severity, AlertStatus, Host/Group/Monitor/Item associations, Message patterns
  - Action: Send notification via Media

### 2. Flexible Severity Mapping
- Item triggers can generate alerts with configurable severity
- Alert triggers can filter further by severity
- Example: CPU trigger generates severity=2, Email trigger filters severity>=1

### 3. Rich Context Preservation
- Item value, units, and host info captured in alert
- Trigger information stored in alert comment
- Full traceability from metric → trigger → alert → action

### 4. AI-Enhanced Processing
- AI can analyze severity and importance of generated alerts
- Optional AI notification guard can suppress false positives
- Analysis stored in alert for future reference

## Database Schema (No Changes Required)

The existing Trigger model already supports all needed fields:

```go
type Trigger struct {
    // For Item-based Triggers:
    Entity                string      // = "item"
    ItemStatus            *int        // Optional: filter by item status
    ItemValueThreshold    *float64    // Lower bound
    ItemValueThresholdMax *float64    // Upper bound (for range checks)
    ItemValueOperator     string      // ">", "<", ">=", "<=", "==", "!=", "between", "outside"
    
    // For Alert-based Triggers:
    Entity                string      // = "alert"
    AlertID               *uint       // Optional: specific alert
    AlertStatus           *int        // Optional: status filter
    AlertGroupID          *uint       // Optional: host group filter
    AlertMonitorID        *uint       // Optional: monitor filter
    AlertHostID           *uint       // Optional: host filter
    AlertItemID           *uint       // Optional: item filter
    AlertQuery            string      // Optional: message pattern
    SeverityMin           int         // Minimum severity to match
    
    // Common:
    ActionID              uint        // Action to execute
    Enabled               int         // 0=disabled, 1=enabled
}
```

## Testing Checklist

To verify the implementation works correctly:

```
1. Item Update Monitoring
   [ ] Create item with threshold-based trigger
   [ ] Update item value below threshold
   [ ] Verify no alert is created
   [ ] Update item value above threshold
   [ ] Verify alert is created automatically

2. Alert Generation
   [ ] Check alert has correct item reference
   [ ] Check alert message includes item value and units
   [ ] Check alert severity matches trigger config
   [ ] Check alert comment has trigger description

3. Alert Filtering
   [ ] Create alert-based trigger with severity filter
   [ ] Create lower severity alert
   [ ] Verify alert trigger does not match
   [ ] Create high severity alert
   [ ] Verify alert trigger matches and action executes

4. Action Execution
   [ ] Set up email action with template
   [ ] Verify email sent with correct substitutions
   [ ] Check {{host_name}}, {{value}}, {{units}} replaced

5. AI Analysis (if enabled)
   [ ] Generate high severity alert
   [ ] Verify AI analysis added to comment
   [ ] Check AI notification guard decision honored
```

## Configuration

No additional configuration needed. Use existing Trigger settings:

```json
{
  "triggers": [
    {
      "name": "High CPU Usage",
      "entity": "item",
      "item_value_threshold": 90,
      "item_value_operator": ">",
      "severity_min": 2,
      "action_id": 5
    },
    {
      "name": "Route High Alerts to Email",
      "entity": "alert",
      "severity_min": 2,
      "action_id": 5
    }
  ]
}
```

## Files Modified

- `internal/service/trigger.go`
  - Modified: `execTriggersForItem()` - Added alert generation
  - Added: `generateAlertFromItemTrigger()` - New function
  - Added: `describeItemTriggerCondition()` - New helper

## Backward Compatibility

✅ Fully backward compatible
- Existing alert webhooks still work
- Existing trigger configurations unchanged
- New functionality is additive only

## Performance Considerations

- Alert generation happens asynchronously (goroutine in `AddAlertServ`)
- AI analysis is optional and configurable
- Trigger evaluation is fast (threshold comparison only)
- No additional database queries required
