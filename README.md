# Nagare (流) - Unified AIOps & Infrastructure Orchestration

Nagare is a high-performance, AI-native operations platform that bridges the gap between traditional monitoring (Zabbix/Prometheus) and autonomous system diagnosis. Built on a Go 1.24+ backend and a Vue 3 + Vite frontend, Nagare leverages Large Language Models (LLM) and a custom Retrieval-Augmented Generation (RAG) engine to provide context-aware troubleshooting.

## 🚀 Vision: Autonomous SRE
Nagare is not just a dashboard; it is an intelligent agent designed to reduce MTTR (Mean Time To Repair) by automating the "Detection -> Context Retrieval -> AI Analysis -> Action" pipeline.

---

## 📂 Technical Documentation Index

For exhaustive technical details, refer to the specialized manuals in the `doc/` directory:

| Document | Scope |
| :--- | :--- |
| [**Core Architecture**](./doc/ARCHITECTURE.md) | High-concurrency design, `jsoniter` optimizations, and data flow. |
| [**AI & RAG Engine**](./doc/AI_RAG_ENGINE.md) | Scoring algorithms, tokenization, and prompt engineering. |
| [**Automated Reporting**](./doc/REPORTING_SYSTEM.md) | PDF generation, Go-chart rendering, and cron orchestration. |
| [**Automation & Chaos**](./doc/AUTOMATION_CHAOS.md) | Ansible integration and Chaos Engineering (Alert Storm) logic. |
| [**WebSSH & Security**](./doc/WEBSSH_SECURITY.md) | PTY proxying, XSS prevention, and WebSocket safety. |
| [**Frontend Engineering**](./doc/FRONTEND_GUIDE.md) | Skeleton screens, code splitting, and Dev Tunnel interoperability. |
| [**API Technical Reference**](./doc/API_REFERENCE.md) | Comprehensive endpoint list, RBAC, and MCP support. |

---

## 🛠️ Performance & Engineering Benchmarks
- **Backend Optimization**: Uses `jsoniter` for high-speed serialization (20-30% CPU overhead reduction).
- **Frontend Perception**: Employs `el-skeleton` for smooth asynchronous loading.
- **Build Strategy**: Vite manual chunking for long-term browser caching of heavy libs (ECharts, xterm).
- **Network Resilience**: Automatic bypass of Microsoft Dev Tunnel phishing pages via header injection.

## 📄 License
Apache License 2.0
