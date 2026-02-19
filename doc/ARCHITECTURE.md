# Nagare System Architecture

Nagare is designed as a decoupled, monorepo system that prioritizes ingestion throughput and intelligent response latency.

## 1. Backend: The Go Engine (Go 1.24+)

### 1.1 Serialization Strategy (`jsoniter`)
To maximize throughput for high-frequency monitoring data, Nagare replaces `encoding/json` with `github.com/json-iterator/go`.
- **Implementation**: Used explicitly in `webssh.go` and globally enabled via build tags.
- **Impact**: Significant performance gains in `POST /api/v1/alerts/webhook` which processes hundreds of concurrent JSON payloads from Zabbix/Prometheus.

### 1.2 Concurrency & Real-time Hub
- **Global Hub (`internal/service/hub.go`)**: Manages long-lived WebSocket connections for site messages. It uses a thread-safe registration/broadcast pattern to prevent race conditions.
- **Task Queue (Redis)**: Non-blocking execution of heavy workloads (PDF generation, bulk monitor syncs) via the `pkg/queue` library.

## 2. Persistence & Models (MySQL / GORM)

### 2.1 Unified Monitoring Model
Nagare standardizes heterogeneous monitoring data:
- **`Monitor`**: Metadata for the source (Zabbix/Prometheus).
- **`Item`**: Individual metrics (CPU, Mem, Latency).
- **`Alert`**: Event-based data with standardized severities (0-3).

### 2.2 Knowledge Graph Metadata
- **`KnowledgeBase`**: Optimized for RAG retrieval with dedicated fields for `topic`, `keywords`, and `content`.

## 3. Global Data Flow

1. **Ingestion Layer**: `Monitor Node -> HTTP Webhook -> Gin Router -> Ingestion Service`.
2. **Analysis Layer**: `Alert Event -> RAG Engine -> Knowledge Retrieval -> LLM Provider -> Diagnostic Report`.
3. **Execution Layer**: `User Action -> Ansible Playbook -> SSH Target -> Status Feedback`.
4. **Presentation Layer**: `Service Event -> WebSocket Hub -> Vue 3 Reactive View`.

## 4. Connectivity & Deployment
- **Root Health Check**: `GET /` and `GET /health` provide unauthenticated heartbeat signals for load balancers.
- **Tunnel Bypass**: Injects headers to bypass Microsoft Dev Tunnel anti-phishing pages, ensuring external webhooks reach the backend without manual intervention.
