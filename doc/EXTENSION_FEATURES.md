# Nagare Project Extension Features

This document describes the three major extension modules implemented to enhance the platform's immediate response, long-term memory, and intelligent analysis capabilities.

---

## 1. Interactive WebSSH Terminal
The WebSSH module allows operators to directly access remote host shells from the browser, significantly reducing MTTR (Mean Time To Repair).

### Key Components
- **Frontend**: Integrated `xterm.js` with `@xterm/addon-fit` for terminal emulation.
- **Backend**: 
    - WebSocket handler in `internal/api/webssh.go`.
    - PTY (Pseudo-Terminal) management using `golang.org/x/crypto/ssh`.
- **Security**: 
    - Credentials are encrypted using AES-GCM before storage.
    - WebSocket handshake is authenticated via JWT tokens passed in the query string.

### Data Flow
1. User clicks "Terminal" button.
2. Frontend opens WebSocket connection to `/api/v1/hosts/:id/ssh?token=...`.
3. Backend upgrades HTTP to WebSocket, validates token, and retrieves encrypted SSH credentials.
4. Backend establishes SSH connection to target host and requests a PTY.
5. Bi-directional data piping:
    - User Keystrokes -> WebSocket -> SSH Stdin.
    - SSH Stdout/Stderr -> WebSocket -> xterm.js Rendering.

---

## 2. Automated Report Generation System
A professional reporting engine that transforms monitoring data into actionable executive insights.

### Key Components
- **PDF Engine**: Uses `Maroto v2` for grid-based PDF layout.
- **Data Visualization**: Uses `go-chart/v2` to render PNG charts (Pie, Line, Bar) server-side.
- **Scheduling**: `robfig/cron` manages automated weekly and monthly generation tasks.

### Features
- **Visual Analytics**: Includes status distribution pie charts and alert trend line charts.
- **Deep Insights**: Statistical analysis of host stability (longest downtime and highest failure frequency).
- **Automation**: Configurable generation times and days via the web UI.

---

## 3. Lightweight RAG Knowledge Base
Enhances AI alert analysis by providing local context through Retrieval-Augmented Generation (RAG).

### Implementation Approach
Instead of a heavy vector database, this module uses a high-performance SQL-based keyword matching strategy, ideal for devops scenarios containing specific error codes and IPs.

### Workflow
1. **Knowledge Ingestion**: Operators record solutions for specific alerts in the Knowledge Base.
2. **Retrieval**: When a critical alert occurs, the system tokenizes the alert message and performs a `LIKE` search against topics and keywords.
3. **Augmentation**: The Top 3 matching entries are retrieved and injected into the Prompt.
4. **Generation**: The LLM (Gemini) receives the alert details *plus* the local context to provide a precise, business-aware recommendation.

### Example
- **KB Entry**: "Daily backup at 2 AM causes high CPU usage. Expected behavior."
- **Alert**: "Host-01 high CPU usage at 02:05 AM."
- **AI Output**: "This alert is likely caused by the scheduled daily backup task as noted in the local knowledge base. No immediate action is required."

---

## 4. Site-wide Notification System
A real-time event distribution system that keeps users informed of critical activities without requiring page refreshes.

### Key Components
- **WebSocket Hub**: A global thread-safe hub in the backend (`service/hub.go`) that manages all active browser connections.
- **Notification Center**: A frontend header component (`SiteMessageCenter.vue`) that handles real-time message reception and unread counts.
- **Message History**: Persistent storage of notifications allowing users to review past events.

### Automated Triggers
- **Sync Completion**: Notifies users when background synchronization of hosts/items finishes.
- **New Alerts**: Immediate pop-up when a new monitoring alert is ingested.
- **Report Ready**: Notifies users when an automated or manual report has finished generating.

---

## 5. Batch Operations & Enhanced Data Tracking
Improves operational efficiency when managing large numbers of resources.

### Features
- **Multi-resource Action**: Batch delete and update status/enabled state for Hosts, Groups, Items, and Reports.
- **Synchronization Tracking**: 
    - `LastSyncAt`: Tracks the exact time each resource was last updated from an external monitor.
    - `ExternalSource`: Identifies which monitoring system (e.g., Zabbix-Prod, Prom-Cluster-A) is the source of truth for the data.
