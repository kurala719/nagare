# Nagare (流) - Enterprise AIOps & Infrastructure Orchestration

Nagare is a high-performance, AI-native operations platform designed to bridge the gap between traditional monitoring (Zabbix/Prometheus) and autonomous system diagnosis. Built on a Go 1.24+ backend and a Vue 3 + Vite frontend, Nagare leverages Large Language Models (LLM) and a custom Retrieval-Augmented Generation (RAG) engine to provide context-aware troubleshooting and automated remediation.

## 📂 Technical Documentation Manuals

| Manual | Description |
| :--- | :--- |
| [**Architecture**](./doc/ARCHITECTURE.md) | System design, high-concurrency Go patterns, and `jsoniter` optimization. |
| [**AI & RAG Engine**](./doc/AI_RAG_ENGINE.md) | Precision scoring algorithms, tokenization, and prompt engineering. |
| [**WebSSH & Security**](./doc/WEBSSH_SECURITY.md) | WebSocket-PTY proxying, XSS prevention, and terminal protocols. |
| [**Reporting & Async Tasks**](./doc/REPORTING_SYSTEM.md) | Redis workers, Maroto PDF generation, and Go-chart rendering. |
| [**Automation & Chaos**](./doc/AUTOMATION_CHAOS.md) | Ansible dynamic inventory and Alert Storm simulation logic. |
| [**Frontend Engineering**](./doc/FRONTEND_GUIDE.md) | UX skeleton screens, manual chunking, and tunnel interoperability. |
| [**API Reference**](./doc/API_REFERENCE.md) | RESTful specifications, RBAC levels, and MCP agent support. |

---

## ⚡ Engineering Highlights

### 1. High-Performance Serialization
Nagare replaces the standard library with `jsoniter` across all high-load paths.
- **Backend Build**: `go build -tags=jsoniter -o server ./cmd/server`
- **Result**: 20-30% reduction in CPU overhead during Zabbix/Prometheus webhook ingestion.

### 2. Autonomous RAG Scoring
Context retrieval uses a precise keyword relevance algorithm:
$$S = \sum_{t \in T} \mathbb{1}(t \in KB_{entry}) \times 2$$
This ensures AI diagnostics are grounded in your specific infrastructure runbooks.

### 3. Dev-Tunnel Interoperability
Native bypass for Microsoft Dev Tunnels using automated header injection (`X-Tunnel-Skip-AntiPhishing-Page`), enabling seamless remote testing and webhook delivery.

## 📄 License
Apache License 2.0
