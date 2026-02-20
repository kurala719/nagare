# Nagare API Technical Reference

All Nagare API endpoints are versioned and follow a standardized REST pattern with JWT authentication. The system relies on Go (`gin-gonic`) and follows a clean MVC-style architecture.

---

## 1. Global Standards

### Response Wrapper
Every response from Nagare is wrapped in a consistent JSON structure:
```json
{
  "success": true,
  "data": { ... },
  "message": "Operation completed successfully.",
  "error": "Error code string (if success is false)"
}
```

### RBAC Privilege Model
Nagare uses a numeric `Privilege` level system handled by `PrivilegesMiddleware`:
- **Level 1 (User)**: Analytical read-only access, site messages, and AI Chat.
- **Level 2 (Manager)**: Monitor, Host configuration, automated reporting, playbook execution, triggering chaos storms.
- **Level 3 (Admin)**: System configuration, audit logs, and full user management.

---

## 2. API Endpoints & Function Routes

Below are the primary API endpoint groups defined in `router.go` and their underlying descriptions. All routes are prefixed with `/api/v1` except public system routes.

### 2.1 Public Authentication (`/auth`)
- `POST /login`: Generate JWT token for users.
- `POST /register`: Request system access.
- `POST /send-code`: Send email verification for registration.
- `POST /reset-request`: Submit a password reset application.
- `POST /reset` *(Privilege 1)*: Reset password via authenticated request.

### 2.2 Intelligence & AI (`/chats`, `/mcp`, `/providers`)
- `POST /chats`: Send a message to the active AI (Gemini/OpenAI) for RAG or general queries.
- `GET /mcp/sse`: Connect to the Model Context Protocol (MCP) server event stream.
- `GET /providers`: List all registered AI providers.
- `POST /providers`: Add a new AI model provider (e.g. Gemini, OpenAI API compatible).
- `POST /alerts/:id/consult`: Request AI analysis for a specific alert incident.

### 2.3 Monitoring Configuration (`/monitors`, `/alarms`)
- `GET /monitors`: List external inventory sources (Zabbix/Prometheus).
- `POST /monitors/:id/login`: Authenticate source node.
- `POST /monitors/:id/sync/hosts`: Pull all hosts and groups from the external monitoring tool.
- `GET /alarms`: List external alerting sources.

### 2.4 Infrastructure & Endpoints (`/hosts`, `/groups`, `/items`)
- `GET /hosts`: Search endpoint assets.
- `POST /hosts`: Add a new server endpoint.
- `GET /hosts/:id/ssh`: Open a WebSSH terminal connection over WebSocket.
- `GET /items`: List all metrics, logs, and checks.
- `POST /hosts/:id/sync`: Force item state synchronization for a specific host.

### 2.5 Alerts, Actions & Media (`/alerts`, `/media`, `/actions`, `/triggers`)
- `POST /alerts/webhook`: Universal unauthenticated ingest point for external monitoring tool pushes.
- `GET /media`: List notification targets (Gmail, Webhook, QQ).
- `POST /media/:id/test`: Trigger a dummy notification to verify target configuration.
- `GET /media/qq/ws`: Connect NapCat OneBot 11 Reverse WebSocket for IM bot integration.
- `POST /actions`: Create notification templates bound to specific media.
- `POST /triggers`: Define conditions under which alerts execute actions.

### 2.6 Operations & Automation (`/ansible`, `/reports`, `/chaos`)
- `GET /ansible/playbooks`: Retrieve YAML automation scripts.
- `POST /ansible/playbooks/:id/run`: Trigger a playbook execution on a specific host or group.
- `POST /chaos/alert-storm`: Simulate a massive incoming alert event to test resilience.
- `POST /reports/generate/weekly`: Manually trigger weekly PDF generation.
- `GET /reports/:id/download`: Retrieve generated PDF document.

### 2.7 Identity & Audit (`/users`, `/audit-logs`, `/register-applications`)
- `GET /users` *(Privilege 2)*: Search active directory users.
- `PUT /register-applications/:id/approve` *(Privilege 3)*: Grant a new user access.
- `GET /audit-logs` *(Privilege 3)*: View internal API mutations, logins, and settings changes.

### 2.8 System Configuration (`/config`, `/retention`)
- `GET /config`: Load current application runtime variables.
- `PUT /config`: Update settings and optionally persist to disk (`configs/nagare_config.json`).
- `GET /retention`: List active data pruning policies (e.g. clean logs older than 30 days).

---

## 3. Remote Access Interop
All endpoints support the `X-Tunnel-Skip-AntiPhishing-Page: true` header for Microsoft Dev Tunnel bypass.
