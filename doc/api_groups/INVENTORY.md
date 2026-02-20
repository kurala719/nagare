# API Group: Infrastructure & Inventory

The Infrastructure API group manages the logical organization of your servers into Hosts and Groups.

---

## 🏢 1. Group Management

Groups are logical collections of hosts, often imported from Zabbix or Prometheus.

### **GET** `/api/v1/groups`
Searches and lists groups.
- **Parameters**: `q` (Search name), `status` (0=Inactive, 1=Active), `monitor_id` (Filter by source monitor).

### **GET** `/api/v1/groups/:id/detail`
Provides a detailed breakdown of a group, including the number of active, warning, and error hosts it contains.

### **POST** `/api/v1/groups`
Adds a new logical group.

### **PUT** `/api/v1/groups/:id`
Updates group metadata (e.g., name, description).

### **DELETE** `/api/v1/groups/:id`
Removes a group and its associations.

---

## 🖥️ 2. Host Management

Hosts are the individual servers being monitored.

### **GET** `/api/v1/hosts`
Searches and lists hosts.
- **Parameters**: `q` (Search name), `ip_addr` (Filter by IP), `m_id` (Source Monitor), `group_id` (Internal Group ID).

### **GET** `/api/v1/hosts/:id/history`
Retrieves the timeline of events (status changes, alerts, sync events) for this host.
- **Parameters**: `from`, `to` (Unix timestamps).

### **GET** `/api/v1/hosts/:id/ssh`
**WebSocket Endpoint**. Bridges a secure SSH session to the host directly in the browser terminal.
- **Required**: Host must have `ssh_user`, `ssh_password` (or key), and `ssh_port` configured.

### **POST** `/api/v1/hosts/:id/consult`
Sends all metrics and metadata for this host to the AI for a "General Health Audit" (RAG-enabled).

### **POST** `/api/v1/hosts`
Adds a new host.
- **Body**: `{ "name": "Host-A", "ip_addr": "192.168.1.1", "ssh_user": "...", "m_id": 1 }`

### **DELETE** `/api/v1/hosts/:id`
Removes a host and all its associated metrics (items) and history.

---

## 📡 3. Terminal (WebSSH)

### **GET** `/api/v1/terminal/ssh`
A generic WebSocket endpoint for direct SSH connections without pre-defined Host entities.
- **Query Params**: `ip`, `port`, `user`, `password`.
- **Logic**: This acts as a "Pass-through" proxy for the browser to communicate with any server.
