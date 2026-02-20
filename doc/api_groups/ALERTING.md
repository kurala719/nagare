# API Group: Alerting & Webhook Ingest

This group handles the lifecycle of alerts, from the moment they are received via webhook to their analysis and resolution.

---

## 🔔 1. Alert Lifecycle

Alerts are the core "Events" in Nagare.

### **POST** `/api/v1/alerts/webhook`
**The Primary Ingest Point**. Receives JSON payloads from Zabbix, Prometheus, or custom scripts.
- **Authentication**: Uses `X-Alarm-Token` header, `Bearer` token, or `event_token` in query/body.
- **Mapping**: Automatically extracts `message`, `severity`, `host_id`, and `item_id` from the payload.
- **Auto-Processing**: Triggers AI analysis and matching automation rules (Triggers).

### **GET** `/api/v1/alerts`
Searches and lists alerts.
- **Parameters**: `q` (search), `severity`, `status` (0=Active, 1=Ack, 2=Resolved), `host_id`, `alarm_id`.

### **POST** `/api/v1/alerts/:id/consult`
Triggers an on-demand AI diagnostic for a specific alert.
- **Output**: Detailed RAG-enabled summary, likely causes, and recommended actions.

### **POST** `/api/v1/alerts/generate-test`
Generates a batch of simulated alerts for testing dashboard performance and notification channels.

---

## 🚨 2. Alarm (Source) Configuration

Alarms represent the *sources* of alerts (e.g., a specific Zabbix server).

### **GET** `/api/v1/alarms`
Lists all alert sources.

### **POST** `/api/v1/alarms/:id/setup-media`
**One-Click Integration**. Automatically configures the required "Media Types" in the remote Zabbix server to point to Nagare's webhook.
- **Response**: Provides the exact Webhook URL and test commands.

### **POST** `/api/v1/alarms/:id/event-token/refresh`
Regenerates the security token for the alarm source. Useful if a token is leaked.
