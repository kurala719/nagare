# Nagare API Technical Reference

All Nagare API endpoints are versioned and follow a standardized REST pattern with JWT authentication.

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

### RBAC Model
Nagare uses a numeric `Privilege` level system:
- **Level 1 (User)**: Analytical read-only access + AI Chat.
- **Level 2 (Manager)**: Monitor/Host configuration + Automated Reporting.
- **Level 3 (Admin)**: System configuration + Audit logs + RBAC management.

---

## 2. API Response Codes & Meanings
Beyond standard HTTP status codes, Nagare provides specific error strings in the `error` field of the JSON wrapper.

| Error Code | HTTP Status | Description |
| :--- | :--- | :--- |
| `invalid_auth_token` | 401 | The JWT token is missing, expired, or tampered with. |
| `insufficient_privileges` | 403 | You are trying to perform an action above your level. |
| `resource_not_found` | 404 | The requested host, group, or monitor does not exist. |
| `ai_provider_timeout` | 504 | The AI took too long to answer. Check your network or provider. |
| `llm_quota_exceeded` | 429 | You have hit your AI API rate limits (e.g. Gemini/OpenAI). |
| `database_connection_fail` | 500 | Nagare's Brain lost contact with its MySQL heart. |
| `ansible_runtime_error` | 500 | A playbook execution failed during the startup phase. |

---

## 3. Key Endpoint Categories

### 2.1 Intelligence
- `POST /api/v1/alerts/:id/consult`: Run RAG-augmented diagnostic.
- `POST /api/v1/hosts/:id/consult`: Summarize host health based on metrics.
- `GET /api/v1/mcp/sse`: Connect to MCP agent event stream (Server-Sent Events).
- `POST /api/v1/chats`: Send a message to the AI Chat with `mode: "roast"` for audit.

### 2.2 Monitoring
- `GET /api/v1/monitors/`: List all monitoring sources.
- `POST /api/v1/monitors/:id/login`: Authenticate source node.
- `POST /api/v1/alerts/webhook`: Universal unauthenticated ingest point.
- `POST /api/v1/monitors/:id/push`: Push local config changes back to Zabbix.

### 2.3 Operations
- `GET /api/v1/hosts/:id/ssh`: Bridge to WebSSH WebSocket.
- `POST /api/v1/ansible/playbooks/:id/run`: Trigger a robot script on a host or group.
- `GET /api/v1/reports/:id/download`: Retrieve generated PDF.

## 3. Remote Access Interop
All endpoints support the `X-Tunnel-Skip-AntiPhishing-Page: true` header for Microsoft Dev Tunnel bypass.
