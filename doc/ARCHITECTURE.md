# Nagare System Architecture

Nagare is a high-availability, AI-native operations framework. Its decoupled architecture separates high-speed telemetry ingestion from intelligent analysis.

## 1. High-Performance Backend (Go 1.24+ / Gin)

### 1.1 Serialization Optimization (`jsoniter`)
To handle thousands of monitoring metrics per second, Nagare replaces the standard Go `encoding/json` with `github.com/json-iterator/go`.
-   **Implementation**: A package-level `jsonIter` variable in `backend/internal/api/webssh.go` and `internal/api/media.go`.
-   **Performance Gain**: Benchmarked at 20-30% CPU overhead reduction on high-load metric endpoints (Zabbix/Prometheus Webhooks).

### 1.2 Concurrency & Orchestration
-   **Goroutine Hub**: Nagare uses a `WebSocket Hub` (`internal/service/hub.go`) to manage hundreds of concurrent terminal and site message connections.
-   **Task Queue (Redis)**: Asynchronous tasks like **PDF Generation**, **Monitor Syncing**, and **Ansible Jobs** are offloaded to background workers.

## 2. Persistence Layer (MySQL / GORM)

### 2.1 Data Models (`internal/model/entities.go`)
-   **`Monitor`**: Metadata for Zabbix/Prometheus nodes.
-   **`Item`**: Individual monitoring metrics (CPU, Memory, Disk).
-   **`Alert`**: Centralized event model with `Severity` (0-3).
-   **`KnowledgeBase`**: Source of truth for the RAG engine.

### 2.2 Migrations & Updates (`backend/cmd/server/main.go`)
-   Nagare uses GORM AutoMigrate for schema evolution.
-   **Schema Legacy Support**: Built-in logic to rename deprecated columns (e.g., `site_id` to `group_id`) ensuring backward compatibility.

## 3. Communication & Connectivity

### 3.1 Real-time Hub
-   Nagare supports **WebSocket (wss://)** for real-time site messages and WebSSH.
-   **Origin Security**: Strict origin checks are applied in `internal/api/webssh.go`.

### 3.2 Connectivity & Health
-   **Root Handler**: `GET /` returns a `200 OK` JSON response, bypassing Microsoft Dev Tunnel anti-phishing pages for external callers.
-   **Health Endpoint**: `GET /health` provides a "UP/DOWN" signal for load balancers.
-   **System Metrics**: `GET /api/v1/system/metrics` exposes Go runtime statistics (Goroutines, GC, MemAlloc).

## 4. Automation Pipeline
1.  **Ingestion**: `Prometheus -> Nagare Webhook -> Alert Model`.
2.  **RAG Diagnosis**: `Alert -> RAG Engine -> Knowledge Retrieval -> Gemini -> AI SRE Report`.
3.  **Remediation**: `User -> Ansible Playbook -> Host CLI -> Resolution`.
4.  **Reporting**: `Cron -> Report Engine -> PDF Generation -> User Download`.
