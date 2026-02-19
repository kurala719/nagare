# Nagare Code Structure Reference

A guide to the monorepo organization for the Nagare project.

## 1. Backend Organization (`/backend`)

### `cmd/server/`
The entry point of the Go application.
-   `main.go`: Initializes database, Redis, cron tasks, and the router.
-   `router/router.go`: Defines the Gin router, middleware (Audit log, JWT, Privilege), and route groups.

### `internal/`
Core logic, shielded from external packages.
-   `api/`: HTTP handlers. (e.g., `webssh.go` for terminal, `alert.go` for webhooks).
-   `mcp/`: Implementation of the Model Context Protocol.
-   `model/`: GORM entities and filter structs.
-   `repository/`: Data Access Objects (DAO). Includes `llm/` and `monitors/`.
-   `service/`: The "Brain". Orchestrates AI diagnostics (`ai.go`), report generation (`report.go`), and RAG logic (`knowledge_base.go`).
-   `service/utils/`: Charting engines (`charts.go`) and cryptography utilities.

### `pkg/queue/`
A reusable Redis-backed task queue library for Nagare.

### `configs/`
`nagare_config.json`: Centralized configuration for AI providers, databases, and system ports.

## 2. Frontend Organization (`/frontend`)

### `src/`
-   `api/`: Centralized Axios/authFetch calls for each backend service.
-   `components/`: Reusable UI elements (Charts, Status Tags).
-   `layout/`: `MainLayout.vue` with dynamic navigation.
-   `utils/`: 
    -   `request.js`: Axios instance with JWT and Dev Tunnel headers.
    -   `authFetch.js`: Standard Fetch wrapper with the same optimizations.
-   `views/`: Feature-specific pages.
    -   `dashboard/`: AI dashboard with skeleton screen optimization.
    -   `Terminal.vue`: xterm.js terminal integration.
    -   `Item.vue`: Data-intensive monitoring view with server-side pagination.

### `public/`
Static assets (favicon, etc.) served directly by the web server.

## 3. Data Flow

1.  **Monitor Ingestion**: `Monitor -> Webhook -> Gin -> Service -> GORM`.
2.  **AI Diagnosis**: `Alert -> RAG Engine -> Gemini/OpenAI -> Result -> Audit Log`.
3.  **Real-time UI**: `Backend -> WebSocket -> xterm.js / Reactive Refs`.
4.  **Async Reporting**: `Cron -> Redis Queue -> Maroto PDF Engine -> Site Message`.
