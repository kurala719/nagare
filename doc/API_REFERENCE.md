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

## 3. Automation & Chaos

### Ansible Automation
-   `GET /api/v1/ansible/inventory`: Serves an Ansible-compatible dynamic JSON inventory.
-   `POST /api/v1/ansible/playbooks/`: Create or update an Ansible playbook.
-   `POST /api/v1/ansible/playbooks/:id/run`: Execute a playbook on targeted hosts.
-   `POST /api/v1/ansible/playbooks/recommend`: Get AI-recommended playbooks for a specific alert.

### Chaos Engineering
-   `POST /api/v1/chaos/alert-storm`: Simulate a high-intensity failure event across the system.

## 4. Operational Assets

### WebSSH
-   `GET /api/v1/hosts/:id/ssh`: Establish a WebSocket bridge to a specific host's terminal.
-   `GET /api/v1/terminal/ssh`: Establish an ad-hoc terminal connection (query params required).

### Reports
-   `GET /api/v1/reports/`: List available reports.
-   `GET /api/v1/reports/:id/download`: Download a generated PDF report.

## 5. Intelligence & Agent Support (MCP)

### Model Context Protocol (MCP)
Nagare provides a standard MCP interface for external AI agents:
-   `GET /api/v1/mcp/sse`: Connect to the MCP Server-Sent Events stream.
-   `POST /api/v1/mcp/message`: Send and receive MCP-compatible tool call messages.

## 6. Security & Privileges

Endpoints are protected by `PrivilegesMiddleware`:
-   **Level 1 (User)**: Read access, personal settings, and chat.
-   **Level 2 (Manager)**: Manage monitors, hosts, reports, and Ansible playbooks.
-   **Level 3 (Admin)**: System configuration, audit logs, and user management.
