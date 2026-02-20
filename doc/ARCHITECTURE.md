# Nagare System Architecture & Service Functions

Nagare is built as a robust, decoupled, clean-architecture application. It acts as an intelligent "Brain" overlay for traditional monitoring tools (like Zabbix and Prometheus).

---

## 1. High-Level Architecture

### Backend: Go (Golang)
- **Framework**: `gin-gonic` for HTTP routing, middleware (CORS, JWT auth, logging, audit).
- **ORM**: `gorm` handling database operations and schema migrations.
- **Task Queue**: Redis-backed asynchronous job processing (`queue.go`).
- **Communication**: WebSocket hubs (`hub.go`, `qq_ws.go`, `webssh.go`) for real-time bidirectional messaging.
- **Integrations**: Standardized Provider interfaces for AI (`ai.go`), Email (`gmail.go`), and Webhooks (`webhook.go`).

### Frontend: Vue 3
- **Build Tool**: Vite.
- **UI Framework**: Element Plus.
- **Routing**: `vue-router` with JWT-based route guarding.
- **State**: Vue Composition API and reactive stores.

---

## 2. Core Service Functions (`internal/service`)

The `service` package contains all the business logic, decoupled from HTTP transport.

### 2.1 Intelligence & RAG Engine (`ai.go`, `knowledge_base.go`)
- **`ConsultAlertServ(alertID uint)`**: Orchestrates the RAG process. Retrieves an alert, its associated host, and items. It then searches the `KnowledgeBase` and historical `Chat` logs to provide context to the configured LLM (Gemini/OpenAI) for root cause analysis.
- **`SendChatServ(req ChatReq)`**: A generalized conversational endpoint for the user. Supports a `UseTools` mode allowing the AI to query the database directly.
- **`AddKnowledgeBaseServ(req KnowledgeBaseReq)`**: Ingests new network documentation or Markdown into the system's local memory for RAG.

### 2.2 Monitoring & Synchronization (`sync.go`, `monitor.go`)
- **`SyncHostsServ(monitorID uint)`**: Connects to the external inventory source (Zabbix/Prometheus), pulls the latest endpoints, and writes them to the local `hosts` table.
- **`SyncItemsServ(hostID uint)`**: Iterates over a specific host, pulling the latest metrics from the external source into the `items` and `item_histories` tables.
- **`StartAutoSync()`**: A background goroutine that triggers `SyncHosts` and `SyncItems` on an interval defined in `configs/nagare_config.json`.

### 2.3 Operations & Automation (`ansible.go`, `webssh.go`, `chaos.go`)
- **`RunPlaybookServ(jobID uint)`**: Executes a YAML Ansible playbook against a specified host group. Handles environment preparation, sub-process execution, and output capture.
- **`HandleWebSSH(c *gin.Context)`**: Upgrades an HTTP connection to a WebSocket, establishes an SSH session using `golang.org/x/crypto/ssh`, and bridges the IO streams (PTY) to the browser terminal.
- **`TriggerAlertStormServ(count int)`**: Generates synthetic, high-severity alerts to test the resilience of triggers, actions, and media delivery (Chaos Engineering).

### 2.4 Status & Health (`status.go`, `health.go`)
- **`RecomputeAllStatuses()`**: Evaluates the health of all monitors, hosts, groups, and items based on recent sync results.
- **`RecomputeMonitorRelated(mid uint)`**: specifically triggers a cascading update of dependent entities if a monitoring source goes offline.
- **`CalculateHealthScore(host Host)`**: Generates a 0-100 score based on active vs inactive critical items. Averages host scores into a global network health metric.

### 2.5 Alert Logic & Delivery (`alert.go`, `trigger.go`, `media.go`, `notify.go`)
- **`ProcessIncomingAlert(raw JSON)`**: Normalizes webhooks from Zabbix/Prometheus into the `Alert` structure.
- **`EvaluateTriggers(alert Alert)`**: Checks if a new alert meets the criteria of any defined `Trigger` (e.g., Severity > 2 AND HostGroup = 'WebServers').
- **`ExecuteAction(action Action, alert Alert)`**: Formats the alert data into the action's template.
- **`GlobalQQWSManager.HandleReverseWS()`**: Maintains a persistent OneBot 11 Reverse WebSocket connection for instant IM bot communication and command parsing (`/status`, `/alerts`).
- **`SendGmailServ()`**: Delivers formatted templates to administrator email addresses using the Google Workspace API.

---

## 3. Data Flow Example: Incident Resolution

1. **Detection**: Zabbix detects high CPU and pushes to Nagare (`POST /alerts/webhook`).
2. **Ingestion**: `ProcessIncomingAlert` saves to `alerts` table.
3. **Trigger**: `EvaluateTriggers` finds a matching rule (`Severity > 1`).
4. **Action**: `ExecuteAction` formats the message ("High CPU on Web01").
5. **Delivery**: System dispatches a message via Gmail and NapCat (QQ WebSocket).
6. **Interaction**: Admin types `/consult Web01` in QQ.
7. **Intelligence**: `GlobalQQWSManager` parses the command, invokes `ConsultAlertServ`, queries the Gemini API with history/metrics, and replies with a diagnosis.
8. **Remediation**: Admin navigates to the dashboard, opens the `Terminal.vue` WebSSH component, and restarts the service.
