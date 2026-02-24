# System Architecture Diagram

## High-Level Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                       NAGARE MONITORING SYSTEM                      │
└─────────────────────────────────────────────────────────────────────┘

┌──────────────────┐
│ External Monitor │  (Zabbix, Prometheus, SNMP, etc.)
│  (Metrics Data)  │
└────────┬─────────┘
         │
         │ Fetch metrics
         ↓
    ┌────────────────────────────────────────────────────┐
    │           ITEM SYNCHRONIZATION                     │
    │  (PullItemsFromMonitorServ, PullItemsFromHostServ) │
    └────────────┬─────────────────────────────────────┘
                 │
                 │ Update item.LastValue
                 ↓
    ┌────────────────────────────────────────────────────┐
    │        ITEM MONITORING DATABASE                    │
    │  ┌──────────────────────────────────────────────┐  │
    │  │ Item {id, name, hid, lastvalue, units, ...} │  │
    │  └──────────────────────────────────────────────┘  │
    └────────────┬─────────────────────────────────────┘
                 │
                 │ Call: ExecuteTriggersForItem()
                 ↓
    ┌────────────────────────────────────────────────────┐
    │       ITEM TRIGGER EVALUATION                      │
    │  execTriggersForItem()                             │
    │                                                    │
    │  For each item trigger:                           │
    │    1. matchItemTrigger()   ← Check threshold      │
    │    2. if match:                                   │
    │       a. generateAlertFromItemTrigger() [NEW]    │
    │       b. invokeItemTriggerAction()               │
    └────────┬───────────────────────────────────────┘
             │
             ├─────────────────────────┬──────────────────────┐
             │                         │                      │
      Match Condition?        Create Alert        Execute Action
             │                    ↓                    ↓
      NOT MATCHED?          Alert Message        Action Notification
             │              "Item X on host Y       (optional legacy)
             │               has value Z"
             │
             └──────────────────────────────┐
                                            ↓
    ┌────────────────────────────────────────────────────┐
    │        ALERT DATABASE & PROCESSING                │
    │  ┌──────────────────────────────────────────────┐  │
    │  │ Alert {id, message, severity, host_id,      │  │
    │  │        item_id, comment, ...}                │  │
    │  └──────────────────────────────────────────────┘  │
    │                                                    │
    │  1. AddAlertServ()      ← Save to DB              │
    │  2. analyzeAndNotifyAlert()                       │
    │     a. Optional: AI Analysis                      │
    │     b. Optional: AI Notification Guard            │
    │     c. Call: ExecuteTriggersForAlert()   [START]  │
    └────────────┬─────────────────────────────────────┘
                 │
                 │ Alert created, trigger alert processing
                 ↓
    ┌────────────────────────────────────────────────────┐
    │       ALERT TRIGGER EVALUATION                     │
    │  execTriggersForAlert()                            │
    │                                                    │
    │  For each alert trigger:                          │
    │    1. matchAlertTrigger()  ← Check filters:       │
    │       - Severity >= min?                          │
    │       - Status matches?                           │
    │       - Host matches?                             │
    │       - Group matches?                            │
    │       - Item matches?                             │
    │       - Message pattern matches?                  │
    │    2. if match:                                   │
    │       invokeAlertTriggerAction()                  │
    └────────┬───────────────────────────────────────┘
             │
             │ Alert passes all filters?
             ├─── NO ──→ Skip (alert not routed)
             │
             └─── YES ──→ Execute Action
                         ↓
    ┌────────────────────────────────────────────────────┐
    │            ACTION EXECUTION                        │
    │  invokeAlertTriggerAction()                        │
    │                                                    │
    │  1. Load Action {template, media_id}              │
    │  2. Load Media {type, target}                      │
    │  3. Format message:                               │
    │     - Replace {{host_name}}, {{value}}, etc.     │
    │  4. Call: ExecuteAction()                         │
    │     └─→ Send via media type                       │
    └────────┬───────────────────────────────────────┘
             │
             ↓
    ┌────────────────────────────────────────────────────┐
    │           NOTIFICATION DELIVERY                    │
    │  ┌──────────────────────────────────────────────┐  │
    │  │ Email    → SMTP Server → admin@company.com  │  │
    │  │ Webhook  → HTTP POST  → Slack/Teams/etc     │  │
    │  │ QQ       → QQ API     → Mobile notification │  │
    │  │ SMS      → SMS API    → Phone notification  │  │
    │  │ Custom   → Custom handler → Custom target   │  │
    │  └──────────────────────────────────────────────┘  │
    └────────┬───────────────────────────────────────┘
             │
             ↓
    ┌────────────────────────────────────────────────────┐
    │             AUDIT TRAIL SAVED                      │
    │  - Alert created with context                      │
    │  - Triggers evaluated (logged)                     │
    │  - Actions executed (logged)                       │
    │  - Notifications sent (logged)                     │
    │  - Full historical record maintained              │
    └────────────────────────────────────────────────────┘
```

## Entity Relationships

```
┌──────────────────────────────────────────────────────────────┐
│                      ENTITY MODEL                            │
└──────────────────────────────────────────────────────────────┘

Monitor (Zabbix, Prometheus)
  │
  ├── contains → Host (Server-1, Server-2, ...)
  │               │
  │               ├── contains → Item (CPU, Memory, Disk, ...)
  │               │               │
  │               │               ├── has value → LastValue (95%, 4.2GB, ...)
  │               │               │
  │               │               └── updates trigger → Item Trigger
  │               │                                      │
  │               │                                      └─→ generates → Alert
  │               │
  │               └── referenced in → Alert (alert.host_id)
  │
  └── referenced in → Alert (alert.monitor_id)

Group (Database Servers, Web Servers, ...)
  │
  └── contains hosts → Host → referenced in → Alert (alert.group_id)

Trigger (Item-based)
  │
  ├── evaluates → Item values
  ├── uses filter → ItemValueThreshold, ItemValueOperator
  ├── references → Action
  │
  └─ when matched ──→ generates → Alert

Trigger (Alert-based)
  │
  ├── filters → Alert
  ├── uses filters → Severity, Status, Host, Item, Group, Message
  ├── references → Action
  │
  └─ when matched ──→ invokes → Action

Action
  │
  ├── uses template → Message with {{placeholders}}
  ├── references → Media
  │
  └─ when invoked ──→ sends via → Media

Media
  │
  ├── type → email, webhook, qq, sms, custom
  ├── target → address/endpoint/number
  │
  └─ when used ──→ sends → Notification

Alert
  │
  ├── references → Item (alert.item_id)
  ├── references → Host (alert.host_id)
  ├── references → Monitor (implicit via host)
  ├── references → Group (implicit via host)
  ├── triggered by → Item Trigger
  ├── filtered by → Alert Trigger(s)
  │
  └─ invoices → Action(s) → Media
```

## Processing Sequence

```
Time ──────────────────────────────────────────────────────────→

T0: Metric Poll
    │
    Monitor fetches CPU metric: 95%
    │
    └─→ Item.LastValue = "95"

T1: Item Update
    │
    ExecuteTriggersForItem(item)
    │
    └─→ matchItemTrigger(trigger, item)
        - Is CPU (95) > Threshold (85)?
        - YES ✓

T2: Alert Generation [NEW FLOW]
    │
    generateAlertFromItemTrigger(trigger, item)
    │
    ├─→ Build message: "Item CPU on host Server-1 has value 95%"
    ├─→ Create Alert object
    │
    └─→ AddAlertServ(alertReq)
        │
        └─→ Alert saved to DB (id=1001)

T3: Alert Analysis (Optional)
    │
    ├─ analyzeAndNotifyAlert()
    │
    ├─→ AI Analysis (if enabled)
    │   └─ Generates recommendations
    │
    ├─→ AI Notification Guard (if enabled)
    │   └─ Decision: Should notify?
    │
    └─→ ExecuteTriggersForAlert(alert)

T4: Alert Trigger Evaluation
    │
    matchAlertTrigger(alert_trigger, alert)
    - Is severity (2) >= min (2)?
    - Is host matched?
    - Is item matched?
    - All YES ✓

T5: Action Execution
    │
    invokeAlertTriggerAction(trigger)
    │
    ├─→ Load Action (Email Alert)
    ├─→ Load Media (admin@company.com)
    │
    └─→ ExecuteAction(action, media)
        │
        └─→ Format template
            - {{host_name}} → "Server-1"
            - {{message}} → "Item CPU..."
            - {{value}} → "95"
            - {{units}} → "%"

T6: Media Dispatch
    │
    Send via SMTP
    │
    └─→ Email received at admin@company.com ✓

T7: Audit Logged
    │
    - Alert #1001 created
    - Trigger #5 evaluated (matched)
    - Action #3 executed (succeeded)
    - Email sent (success logged)
```

## State Machines

### Alert Trigger Evaluation Flow

```
┌─────────────────────────┐
│   Alert Created         │
│  (severity=2)           │
│  (host_id=5)            │
└────────────┬────────────┘
             │
             ↓
    ┌─────────────────────────────────────┐
    │ Load All Alert Triggers             │
    │ (entity="alert")                    │
    └────────┬────────────────────────────┘
             │
             ↓
    ┌─────────────────────────────────────┐
    │ For Each Trigger:                   │
    │ Check Filters                       │
    └────────┬────────────────────────────┘
             │
      ┌──────┴──────────┬──────────┬──────────┐
      │                 │          │          │
      ↓                 ↓          ↓          ↓
  Severity    Status      Host         Item
  >= min?     matches?     matches?     matches?
     │            │            │            │
   YES          YES          YES          YES
    │            │            │            │
    └─────┬──────┴────────────┴────────────┘
          │
          ↓
  ┌─────────────────┐
  │  All Pass?      │
  └────┬────────┬───┘
       │        │
      YES      NO
       │        │
       │        ↓
       │   Skip Alert
       │   (no action)
       │
       ↓
  ┌──────────────────────────────┐
  │ Execute Trigger Action       │
  │ - Get Action                 │
  │ - Get Media                  │
  │ - Format Message             │
  │ - Send Notification          │
  └──────────────────────────────┘
```

### Item Trigger Evaluation Flow

```
┌──────────────────────┐
│  Item Updated        │
│  LastValue = "95"    │
└──────┬───────────────┘
       │
       ↓
┌────────────────────────────────┐
│ ExecuteTriggersForItem()        │
│ Load Item Triggers             │
│ (entity="item")                │
└────┬───────────────────────────┘
     │
     ↓
┌─────────────────────┐
│ For Each Trigger:   │
│ Evaluate Condition  │
└────┬────────────────┘
     │
     ├──→ Parse item.LastValue (95)
     ├──→ Get threshold (85)
     ├──→ Get operator (>)
     │
     └──→ Compare: 95 > 85?
         │
         ├─ YES ✓
         │  │
         │  ├─→ generateAlertFromItemTrigger()  [NEW]
         │  │    └─ Create Alert with context
         │  │
         │  └─→ invokeItemTriggerAction()
         │       └─ Execute action (legacy)
         │
         └─ NO
            └─ Skip (continue)
```

## Database Schema Relationships

```
┌─────────────────────────────────┐
│          Trigger Table          │
├─────────────┬───────────────────┤
│ ID          │ 1, 2, 3, ...      │
│ Name        │ "High CPU", ...   │
│ Entity      │ "item" / "alert"  │
│ ActionID    │ FK → Action       │
│             │                   │
│ ITEM Fields │                   │
│ ├ ItemStatus│ *int              │
│ ├ ItemValue │ *float64          │
│ │  Threshold│                   │
│ ├ ItemValue │ *float64          │
│ │  ThresholdMax                 │
│ └ ItemValue │ string            │
│   Operator  │                   │
│             │                   │
│ ALERT Fields│                   │
│ ├ AlertID   │ *uint             │
│ ├ AlertStatus│ *int             │
│ ├ AlertHost │ *uint             │
│ │  ID        │                   │
│ ├ AlertItem │ *uint             │
│ │  ID        │                   │
│ ├ AlertGroup│ *uint             │
│ │  ID        │                   │
│ ├ AlertMonitor│ *uint           │
│ │  ID        │                   │
│ └ AlertQuery│ string            │
│             │                   │
│ SeverityMin │ 0-5               │
│ Enabled     │ 0/1               │
└─────────────┴───────────────────┘
        │
        └─→ FK: Action.ID
            │
            ├─→ Template (message)
            └─→ FK: Media.ID
                │
                ├─→ Type (email, webhook, ...)
                └─→ Target (address, endpoint, ...)

┌─────────────────────────────────┐
│          Alert Table            │
├─────────────┬───────────────────┤
│ ID          │ 1001, 1002, ...   │
│ Message     │ Alert text        │
│ Severity    │ 0-5               │
│ Status      │ 0=active, 1=ack   │
│ HostID      │ FK → Host         │
│ ItemID      │ FK → Item         │
│ AlarmID     │ FK → Alarm        │
│ Comment     │ AI analysis       │
│ CreatedAt   │ Timestamp         │
└─────────────┴───────────────────┘
```

## Summary

The Nagare system implements a complete monitoring pipeline:

1. **Items** - Metrics from external sources
2. **Item Triggers** - Threshold-based detection
3. **Alerts** - Persistent event records
4. **Alert Triggers** - Smart filtering and routing
5. **Actions** - Message formatting
6. **Media** - Notification delivery

This creates a robust, flexible, and maintainable alerting system that automatically converts metrics into actionable notifications.
