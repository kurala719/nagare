# API Group: Monitoring Systems & Metrics

This group manages the external monitoring sources (like Zabbix or Prometheus) and the specific metrics (Items) being tracked.

---

## 📡 1. Monitor Management

Monitors are the external "Eyes" of Nagare.

### **GET** `/api/v1/monitors`
Lists all configured monitoring systems.
- **Parameters**: `q` (search), `status` (0=Inactive, 1=Active).

### **POST** `/api/v1/monitors`
Adds a new monitoring source.
- **Body**: `{ "name": "Zabbix Prod", "url": "...", "username": "...", "password": "...", "type": 1 }`

### **POST** `/api/v1/monitors/:id/login`
Tests and saves the authentication token for the remote monitor.

### **POST** `/api/v1/monitors/:id/event-token`
Generates a unique `event_token` used for receiving webhooks from this specific monitor.

### **POST** `/api/v1/monitors/:id/check`
Performs an instant health check on the connection to the monitor.

---

## 📊 2. Item (Metric) Management

Items are specific metrics (e.g., "CPU Load", "Free Memory") associated with a host.

### **GET** `/api/v1/items`
Lists all monitored items across the system.
- **Parameters**: `hid` (Filter by internal Host ID), `q` (Search name).

### **GET** `/api/v1/items/:id/history`
Retrieves time-series data for a specific metric.
- **Parameters**: `from`, `to` (Unix timestamps), `limit` (Default: 500).

### **POST** `/api/v1/items/:id/consult`
Sends the current value and history of this metric to the AI for an expert assessment.

---

## 🔄 3. Synchronization (Pull/Push)

Nagare can bidirectional sync with external monitors.

### **POST** `/api/v1/monitors/:id/groups/pull`
Fetches all host groups from the remote Zabbix/Prometheus and creates them in Nagare.

### **POST** `/api/v1/monitors/:id/hosts/pull`
Fetches all hosts from the remote monitor.

### **POST** `/api/v1/monitors/:id/hosts/:hid/items/pull`
Fetches all specific metrics for a specific host from the remote monitor.

### **POST** `/api/v1/monitors/:id/hosts/:hid/items/push`
Pushes configuration changes (like updated thresholds or descriptions) back to the remote monitoring system.
