# Quick Reference: Trigger-Alert-Action System

## System Architecture

```
Metrics (Items) → Triggers (Threshold) → Alerts (Notification) → Actions (Media)
```

## Three Entity Types

### 1. **Items** (Metrics)
- Monitor external metrics (CPU, Memory, etc.)
- Have values that change over time
- Tracked in database with history

### 2. **Triggers** (Rules)
- Two types based on `entity` field:
  - `"item"` - Evaluate metric values against thresholds
  - `"alert"` - Filter generated alerts before sending

### 3. **Actions** (Delivery)
- Template message with placeholders
- Reference media (email, webhook, QQ, etc.)
- Executed when triggers match

## API Quick Start

```bash
# Create Item Trigger (Threshold-based)
POST /trigger
{
  "name": "High CPU",
  "entity": "item",
  "item_value_threshold": 85,
  "item_value_operator": ">",
  "severity_min": 2,
  "action_id": 5
}

# Create Alert Trigger (Filter-based)
POST /trigger
{
  "name": "Route to Email",
  "entity": "alert",
  "severity_min": 2,
  "action_id": 5
}

# Create Action
POST /action
{
  "name": "Email Alert",
  "media_id": 1,
  "template": "ALERT: {{host_name}} - {{message}}"
}

# Create Media
POST /media
{
  "name": "Admin Email",
  "type": "email",
  "target": "admin@company.com"
}
```

## Default Severity Mapping

| Level | Label | Usage |
|-------|-------|-------|
| 0 | Info | Informational events |
| 1 | Warning | Non-critical issues |
| 2 | Critical | Important alerts |
| 3+ | Severe | Emergency situations |

## Item Trigger Operators

| Operator | Meaning | Example |
|----------|---------|---------|
| `>` | Greater than | `CPU > 85` |
| `<` | Less than | `Memory < 10` |
| `>=` | Greater or equal | `Response >= 2000ms` |
| `<=` | Less or equal | `Connections <= 100` |
| `==` or `=` | Equal | `Status = 1` |
| `!=` | Not equal | `State != "normal"` |
| `between` | Range (inclusive) | `Temp 18-25` |
| `outside` | Outside range | `Humidity outside 40-60` |

## Message Placeholders

### Item-Generated Alerts

| Placeholder | Description | Example |
|------------|-------------|---------|
| `{{host_name}}` | Host name | "Server-1" |
| `{{item_id}}` | Item/metric ID | "42" |
| `{{name}}` | Item name | "CPU Usage" |
| `{{value}}` | Current value | "95" |
| `{{units}}` | Measurement unit | "%" |
| `{{message}}` | Alert message | "Item CPU Usage on..." |
| `{{severity}}` | Numeric severity | "2" |
| `{{severity_label}}` | Text severity | "Critical" |
| `{{status}}` | Alert status | "0" (active) |

## Threshold Check Examples

```json
// CPU > 85%
{
  "entity": "item",
  "item_value_threshold": 85,
  "item_value_operator": ">",
  "severity_min": 2
}

// Memory < 10% free
{
  "entity": "item",
  "item_value_threshold": 10,
  "item_value_operator": "<",
  "severity_min": 2
}

// Temperature 18-25°C
{
  "entity": "item",
  "item_value_threshold": 18,
  "item_value_threshold_max": 25,
  "item_value_operator": "between",
  "severity_min": 1
}

// Disk free outside 20-80%
{
  "entity": "item",
  "item_value_threshold": 20,
  "item_value_threshold_max": 80,
  "item_value_operator": "outside",
  "severity_min": 2
}
```

## Alert Filter Examples

```json
// All high severity
{
  "entity": "alert",
  "severity_min": 2
}

// Specific host
{
  "entity": "alert",
  "alert_host_id": 5,
  "severity_min": 1
}

// Specific item
{
  "entity": "alert",
  "alert_item_id": 42,
  "severity_min": 1
}

// Message pattern
{
  "entity": "alert",
  "alert_query": "memory",
  "severity_min": 1
}

// Combination
{
  "entity": "alert",
  "alert_host_id": 5,
  "alert_query": "cpu",
  "severity_min": 2
}
```

## Execution Flow (Simplified)

```
1. Monitor sends metric → Item updated
2. Item trigger evaluates threshold
   └─ Match? → Create Alert
3. Alert trigger filters alert
   └─ Match? → Execute Action
4. Action formats message → Send via Media
```

## Common Mistakes to Avoid

❌ **Using text comparison for numbers**
```json
// WRONG: "90" > "85" fails (string comparison)
{
  "item_value_threshold": "85",
  "item_value_operator": ">"
}
```

✅ **Use numeric values**
```json
// CORRECT: 90 > 85 works (numeric comparison)
{
  "item_value_threshold": 85,
  "item_value_operator": ">"
}
```

---

❌ **Wrong placeholder names**
```json
// WRONG: {{hostname}} (not valid)
{
  "template": "Alert: {{hostname}}"
}
```

✅ **Use exact names**
```json
// CORRECT: {{host_name}} (valid)
{
  "template": "Alert: {{host_name}}"
}
```

---

❌ **Action without media**
```json
// WRONG: Action exists but media_id points to nothing
{
  "media_id": 999  // doesn't exist
}
```

✅ **Ensure dependencies exist**
```bash
# Create media first, then action
POST /media → returns id 1
POST /action with media_id: 1
```

## Debugging Commands

```bash
# List all triggers
curl http://localhost:8080/trigger/search?limit=100

# List triggers for alerts
curl http://localhost:8080/trigger/search?entity=alert&limit=100

# List triggers for items
curl http://localhost:8080/trigger/search?entity=item&limit=100

# View recent alerts
curl http://localhost:8080/alerts/search?limit=20&order=desc

# Check specific item
curl http://localhost:8080/items/:id

# View all actions
curl http://localhost:8080/action/search?limit=100

# View all media
curl http://localhost:8080/media/search?limit=100
```

## Performance Notes

- **Item trigger evaluation**: Fast (single operator check)
- **Alert creation**: Async (doesn't block metric updates)
- **Alert filtering**: Fast (linear, not nested loops)
- **Media sending**: Async (doesn't block system)

## Testing Flow

```
1. Create item with high threshold
2. Create item trigger for that threshold
3. Create action with email media
4. Create alert trigger to route emails
5. Update item value past threshold
6. Verify:
   - Alert created in database
   - Email sent to recipient
   - Alert has correct item reference
   - Message has substituted values
```

## Key Differences from Other Systems

| Feature | Nagare | Typical |
|---------|--------|---------|
| Threshold detection | Via triggers | Via rules |
| Alert generation | Automatic | Webhook-based |
| Filtering | Two-stage | Single |
| Placeholders | In actions | In alerts |
| AI support | Optional | No |

## When to Use Each Entity Type

### Item Triggers
- CPU > 85%
- Memory < 10%
- Disk free outside 20-80%
- Response time > 5000ms
- Database query timeout exceeded

### Alert Triggers
- Route all high severity to PagerDuty
- Route all security alerts to SecOps
- Route all critical to phone call
- Suppress low severity during maintenance
- Route specific hosts to specific teams

## Next Steps

1. Create item triggers for critical metrics
2. Create alert triggers for routing
3. Create actions with templates
4. Create media with targets
5. Test with manual metric updates
6. Monitor alerts for correctness
7. Tune thresholds based on false positive rate
8. Scale to all metrics

---

**Need more help?** See the full guides:
- `SYSTEM_ARCHITECTURE.md` - Detailed architecture
- `INTEGRATION_GUIDE.md` - Complete integration steps
- `IMPLEMENTATION_SUMMARY.md` - What changed
- `CODE_CHANGES.md` - Exact code modifications
