# Integration Guide: Trigger-Alert-Action Flow

## Overview
The Nagare system now implements the complete flow: **Items → Triggers → Alerts → Actions → Media**

## How to Use

### Step 1: Create an Item Trigger (Metric-based)

An Item Trigger evaluates when a metric value crosses a threshold.

**API: POST /trigger**

```json
{
  "name": "High CPU Usage Alert",
  "entity": "item",
  "item_value_threshold": 85,
  "item_value_operator": ">",
  "severity_min": 2,
  "action_id": 5,
  "enabled": 1
}
```

**Field Explanations:**
- `entity`: "item" - Triggers on metric value updates
- `item_value_threshold`: 85 - Lower bound for comparison
- `item_value_operator`: ">" - Trigger when value exceeds threshold
- `severity_min`: 2 - Generated alert will be severity 2
- `action_id`: 5 - Action to execute when trigger matches

**Supported Operators:**
- `>` - Greater than
- `<` - Less than
- `>=` - Greater than or equal to
- `<=` - Less than or equal to
- `==` or `=` - Equal to
- `!=` - Not equal to
- `between` - Value between min and max (requires `item_value_threshold_max`)
- `outside` - Value outside range (requires `item_value_threshold_max`)

**Example: Range Check**
```json
{
  "name": "Temperature Out of Range",
  "entity": "item",
  "item_value_threshold": 18,
  "item_value_threshold_max": 25,
  "item_value_operator": "outside",
  "severity_min": 1,
  "action_id": 5
}
```

### Step 2: Create an Alert Trigger (Alert Filtering)

An Alert Trigger filters generated alerts and decides if an action should execute.

**API: POST /trigger**

```json
{
  "name": "Route High Priority Alerts to Email",
  "entity": "alert",
  "severity_min": 2,
  "alert_host_id": 5,
  "action_id": 6,
  "enabled": 1
}
```

**Field Explanations:**
- `entity`: "alert" - Triggers on alert creation
- `severity_min`: 2 - Only match alerts with severity >= 2
- `alert_host_id`: 5 - Optional: Only match alerts from specific host
- `action_id`: 6 - Action to execute (send email)

**Optional Alert Filters:**
- `alert_id` - Specific alert ID
- `alert_status` - Alert status (0=active, 1=acknowledged, 2=resolved)
- `alert_host_id` - Host ID
- `alert_monitor_id` - Monitor system ID
- `alert_group_id` - Host group ID
- `alert_item_id` - Item/metric ID
- `alert_query` - Message text pattern (substring match, case-insensitive)

**Example: Filter by Item and Message**
```json
{
  "name": "Memory Alerts to Admin",
  "entity": "alert",
  "alert_item_id": 42,
  "alert_query": "memory",
  "severity_min": 1,
  "action_id": 7,
  "enabled": 1
}
```

### Step 3: Create an Action

An Action determines what happens when a trigger matches.

**API: POST /action**

```json
{
  "name": "Email Admin",
  "media_id": 3,
  "template": "ALERT: {{host_name}} - {{message}} (Severity: {{severity_label}})",
  "enabled": 1
}
```

**Available Message Placeholders:**

**For Item-Generated Alerts:**
- `{{host_name}}` - Host name
- `{{item_id}}` - Item/metric ID
- `{{name}}` - Item name
- `{{value}}` - Current metric value
- `{{units}}` - Measurement units
- `{{message}}` - Alert message
- `{{severity}}` - Numeric severity (0-5)
- `{{severity_label}}` - Human-readable severity (Normal, Warning, Critical)
- `{{status}}` - Alert status code
- `{{created_at}}` - ISO 8601 timestamp

### Step 4: Create Media

Media specifies how and where to send notifications.

**API: POST /media**

**Email Example:**
```json
{
  "name": "Admin Email",
  "type": "email",
  "target": "admin@company.com",
  "enabled": 1
}
```

**Other (Webhook) Example:**
```json
{
  "name": "Slack Webhook",
  "type": "other",
  "target": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
  "enabled": 1
}
```

**QQ (WeChat-like) Example:**
```json
{
  "name": "QQ Notification",
  "type": "qq",
  "target": "1234567890",
  "enabled": 1
}
```

## Complete Example Flow

### Setup

**1. Create Item Trigger (CPU):**
```
POST /trigger
{
  "name": "CPU Alert Trigger",
  "entity": "item",
  "item_value_threshold": 85,
  "item_value_operator": ">",
  "severity_min": 2,
  "action_id": 100,  // Dummy, will be filtered by alert trigger
  "enabled": 1
}
→ Trigger ID: 1
```

**2. Create Alert Trigger (Route to Email):**
```
POST /trigger
{
  "name": "High Alerts to Email",
  "entity": "alert",
  "severity_min": 2,
  "action_id": 5,  // Email action
  "enabled": 1
}
→ Trigger ID: 2
```

**3. Create Email Action:**
```
POST /action
{
  "name": "Send Email Alert",
  "media_id": 1,  // Email media
  "template": "⚠️ ALERT: {{host_name}} - {{message}}",
  "enabled": 1
}
→ Action ID: 5
```

**4. Create Email Media:**
```
POST /media
{
  "name": "Admin Email",
  "type": "email",
  "target": "admin@company.com",
  "enabled": 1
}
→ Media ID: 1
```

### Execution Flow

**Step 1: Metric Update**
```
Monitor fetches CPU metric for Server-1
Value: 90% (> 85% threshold)
System calls: ExecuteTriggersForItem()
```

**Step 2: Item Trigger Evaluation**
```
Trigger #1 (CPU Alert Trigger) matches:
  - Entity: "item" ✓
  - Value 90 > Threshold 85 ✓
  
Action: generateAlertFromItemTrigger()
Creates Alert:
  - ID: 1001
  - Message: "Item CPU on host Server-1 has value 90%"
  - Severity: 2
  - HostID: 5
  - ItemID: 42
  - Comment: "Triggered by CPU Alert Trigger: > 85.00"
```

**Step 3: Alert Trigger Evaluation**
```
Alert #1001 created
System calls: ExecuteTriggersForAlert()

Trigger #2 (High Alerts to Email) matches:
  - Entity: "alert" ✓
  - Severity 2 >= 2 ✓
  
Action: invokeAlertTriggerAction()
```

**Step 4: Action Execution**
```
Execute Action #5:
  - Template: "⚠️ ALERT: {{host_name}} - {{message}}"
  - Substitutions:
    - {{host_name}} = "Server-1"
    - {{message}} = "Item CPU on host Server-1 has value 90%"
  - Result: "⚠️ ALERT: Server-1 - Item CPU on host Server-1 has value 90%"

Send via Media #1 (Email):
  - Type: email
  - Target: admin@company.com
  - Body: "⚠️ ALERT: Server-1 - Item CPU on host Server-1 has value 90%"

Email received by: admin@company.com ✓
```

## Advanced Filtering Scenarios

### Scenario 1: Route by Severity and Host

```json
// Trigger: Send high severity alerts to ops team
{
  "name": "Critical to OPS",
  "entity": "alert",
  "severity_min": 3,
  "alert_host_id": 10,
  "action_id": 8  // OPS Email Action
}
```

### Scenario 2: Route by Item Type

```json
// Trigger: Send memory alerts to special handler
{
  "name": "Memory Issues to DevOps",
  "entity": "alert",
  "alert_item_id": 42,  // Memory item
  "action_id": 9        // DevOps Webhook Action
}
```

### Scenario 3: Route by Message Pattern

```json
// Trigger: Security alerts to security team
{
  "name": "Security to SecOps",
  "entity": "alert",
  "alert_query": "security authentication unauthorized",
  "severity_min": 2,
  "action_id": 10  // Security Email
}
```

### Scenario 4: Multiple Actions per Trigger

Create multiple alert triggers with the same filters but different actions:

```json
// Alert Trigger 1: Email
{
  "name": "Route to Email",
  "entity": "alert",
  "severity_min": 2,
  "action_id": 5  // Email action
}

// Alert Trigger 2: Slack
{
  "name": "Route to Slack",
  "entity": "alert",
  "severity_min": 2,
  "action_id": 11  // Slack webhook action
}

// Alert Trigger 3: PagerDuty
{
  "name": "Route to PagerDuty",
  "entity": "alert",
  "severity_min": 3,  // Only critical
  "action_id": 12  // PagerDuty webhook action
}
```

Result: One alert triggers multiple notifications!

## Testing Instructions

### Test 1: Simple Item Threshold

1. Create item trigger: `CPU > 75%`
2. Create alert trigger: severity >= 1 → email
3. Create email action and media
4. Manually update item to 80%
5. Verify email received

**API to test:**
```bash
# Update item value
PUT /item/42
{ "value": "80" }

# Check alerts were created
GET /alerts?with_total=true
```

### Test 2: Range Check

1. Create item trigger: `Temperature outside 18-25°C`
2. Create alert trigger: severity >= 1 → email
3. Update item to 16°C (below range)
4. Verify alert created and email sent

### Test 3: Multiple Actions

1. Create item trigger: `Memory > 80%`
2. Create two alert triggers filtering same item
3. First trigger → Email action
4. Second trigger → Slack action
5. Update item to 85%
6. Verify both email and Slack message received

### Test 4: AI Analysis

1. Enable AI analysis
2. Follow Test 1
3. Check alert comment has AI analysis
4. Verify AI notification guard respected

## Troubleshooting

### Issue: Trigger not matching

**Check:**
1. Trigger enabled? (`enabled: 1`)
2. Item trigger operator correct? (>, <, >=, <=, =, !=)
3. Threshold value comparison? (90 > 85, not "90" > "85")
4. Action enabled? (action must exist and be enabled)
5. Media enabled? (media must exist and be enabled)

**Debug:**
```bash
# View all triggers
GET /trigger/search?limit=100

# Check specific trigger
GET /trigger/1

# Check action
GET /action/5

# Check media
GET /media/1
```

### Issue: Alert created but no notification

**Check:**
1. Alert created? (`GET /alerts`)
2. Alert severity matches filter? (alert.severity >= trigger.severity_min)
3. Alert status correct? (0 = active, not resolved)
4. Second trigger (alert trigger) enabled and matching?

**Debug:**
```bash
# View recent alerts
GET /alerts/search?limit=10&order=desc&sort=created_at

# View alert details
GET /alerts/1001

# Check if triggers execute
GET /trigger/search?entity=alert
```

### Issue: Template substitution not working

**Check:**
1. Placeholder spelling exact (`{{host_name}}` not `{{hostname}}`)
2. Field exists in context (item triggers provide item fields)
3. Action template syntax correct

**Valid Placeholders:**
- `{{host_name}}` ✓
- `{{hostname}}` ✗ (not valid)
- `{{host}}` ✗ (not valid)

## Performance Notes

- **Item trigger evaluation**: O(1) per item → very fast
- **Alert filtering**: O(n) triggers × O(1) per check → fast
- **Alert creation**: Async (goroutine), doesn't block metric update
- **AI analysis**: Optional, can be disabled if too slow

## Architecture Diagram

```
External Monitor (Zabbix)
    ↓
  PULL ITEMS
    ↓
  Item Value Updated
    ↓
  ExecuteTriggersForItem()
    ↓
    ├─→ Item Trigger Evaluation (threshold check)
    │     ↓
    │   Match? → generateAlertFromItemTrigger()
    │             ↓
    │           Create Alert (with item context)
    │             ↓
    │           AddAlertServ()
    │             ↓
    │           asyncAnalyzeAndNotifyAlert()
    │             ↓
    │           ExecuteTriggersForAlert()
    │             ↓
    │             └─→ Alert Trigger Evaluation (filter)
    │                   ↓
    │                 Match? → invokeAlertTriggerAction()
    │                           ↓
    │                         ExecuteAction()
    │                           ↓
    │                         Send via Media
    │                           ↓
    │                         Notification Sent! ✓
    │
    └─→ invokeItemTriggerAction() (legacy, for action connection)
          ↓
        (executes if matching)
```

## Summary

The system now implements true **metric-driven alerting**:
1. ✓ Items update with metric values
2. ✓ Triggers detect when values cross thresholds
3. ✓ Alerts are generated automatically
4. ✓ Alert triggers filter and route the alerts
5. ✓ Actions format and send notifications via media
