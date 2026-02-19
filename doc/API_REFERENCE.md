# Nagare API Technical Reference

All Nagare API endpoints follow a standardized RESTful pattern with JWT authentication.

## 1. Global Endpoints

### Root & Health Checks
-   `GET /`: Returns a 200 OK JSON confirmation (Bypass anti-phishing/tunnel checks).
-   `GET /health`: Returns system status (`UP/DOWN`).
-   `GET /api/v1/system/metrics`: Real-time Go runtime stats (Goroutines, Memory).

## 2. Intelligence & AI

### Diagnostic Consultation
-   `POST /api/v1/alerts/:id/consult`: Trigger an AI-powered RAG diagnosis for a specific alert.
-   `POST /api/v1/chats/`: Normal LLM interaction with "Persona" support.
-   `POST /api/v1/hosts/:id/consult`: Analyze a host's entire resource trend using LLM.

## 3. Monitoring & Data Ingestion

### Monitor Management
-   `GET /api/v1/monitors/`: Search and list monitors.
-   `POST /api/v1/monitors/:id/login`: Authenticate and obtain tokens from Zabbix/Prometheus.
-   `POST /api/v1/monitors/:id/hosts/pull`: Sync all hosts from a specific monitoring node.

### Alert Webhooks (Unauthenticated)
-   `POST /api/v1/alerts/webhook`: Generic webhook for Prometheus Alertmanager.
-   `POST /api/v1/media/qq/message`: QQ OneBot 11 message handler.

## 4. Operational Assets

### WebSSH
-   `GET /api/v1/hosts/:id/ssh`: Establish a WebSocket bridge to a specific host's terminal.
-   `GET /api/v1/terminal/ssh`: Establish an ad-hoc terminal connection (query params required).

### Reports
-   `GET /api/v1/reports/`: List available reports.
-   `GET /api/v1/reports/:id/download`: Download a generated PDF report.

## 5. Security & Privileges

Endpoints are protected by `PrivilegesMiddleware`:
-   **Level 1 (User)**: Read access, personal settings, and chat.
-   **Level 2 (Manager)**: Manage monitors, hosts, and reports.
-   **Level 3 (Admin)**: System configuration, audit logs, and user management.
