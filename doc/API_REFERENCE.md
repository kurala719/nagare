# Nagare API Technical Reference

## 1. Request Standards

### Base URL
`/api/v1`

### Response Structure
```json
{
  "success": true,
  "data": { ... },
  "message": "Optional feedback",
  "error": "Optional error string"
}
```

## 2. Endpoint Catalog (Highlights)

### 2.1 Intelligence & Agents
- `POST /alerts/:id/consult`: Trigger RAG diagnostic.
- `GET /mcp/sse`: Connect to Model Context Protocol stream.
- `POST /chats/`: Direct LLM interface with persona support.

### 2.2 System & Metrics
- `GET /`: Root heartbeat (Bypass tunnel phishing).
- `GET /health`: Binary status.
- `GET /system/metrics`: Go runtime telemetry (MemAlloc, Goroutines).

### 2.3 Operational Resources
- `GET /hosts/`: Search/List with pagination.
- `GET /monitors/:id/hosts/pull`: Synchronize from source.
- `POST /ansible/playbooks/:id/run`: Trigger remediation.

## 3. RBAC & Privilege Levels

Middleware (`PrivilegesMiddleware`) enforces the following hierarchy:
- **Level 1 (User)**: Read metrics, view alerts, chat with AI.
- **Level 2 (Manager)**: Manage monitors, update host credentials, generate reports.
- **Level 3 (Admin)**: System configuration, Audit logs, User/RBAC management.

## 4. Connectivity Bypass
All endpoints support the `X-Tunnel-Skip-AntiPhishing-Page: true` header to ensure compatibility with Microsoft Dev Tunnels.
