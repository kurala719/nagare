# Nagare (流) - The AI-Powered Brain for Your IT Infrastructure

**Nagare** (Japanese for "Flow") is a smart platform that watches over your servers and applications. Unlike traditional systems that just "beep" when something breaks, Nagare uses Artificial Intelligence to understand **why** it broke and tells you **how** to fix it.

## 🌟 What makes Nagare special? (The Layman's View)
- **It Listens**: It connects to your existing tools (Zabbix) to hear every "heartbeat" of your system.
- **It Remembers**: It has a "Knowledge Base" (RAG). When a problem occurs, it looks up your past notes and manuals to find a solution.
- **It Thinks**: It uses advanced AI (Google Gemini) to analyze errors. It acts like a senior engineer who is awake 24/7.
- **It Acts**: You can fix problems directly from your browser using a built-in "Command Center" (WebSSH) or automated "Robot Scripts" (Ansible).

---

## 📖 The "Human-Friendly" Manual
New to Nagare? Check out our **[Nagare User Manual (A Guide for Everyone)](./NAGARE_USER_MANUAL.md)**. 
It explains every part of the system in plain English.

---

## 📂 Technical Navigation
If you are a developer or an engineer, explore our deep-dive manuals:

| Manual | Concept | Technical Focus |
| :--- | :--- | :--- |
| [**Architecture**](./doc/ARCHITECTURE.md) | The Nervous System | Go 1.24, Gin, High-Concurrency. |
| [**Database Schema**](./doc/DATABASE_SCHEMA.md) | The Storage Engine | MySQL/GORM, ERD, History Tracking. |
| [**Deployment Guide**](./doc/DEPLOYMENT_GUIDE.md) | Production & Staging | Nginx, systemd, JWT Secrets, HTTPS. |
| [**Developer Guide**](./doc/DEVELOPER_GUIDE.md) | Code Standards | DDD Layering, Vue 3 Composition API. |
| [**Integrations**](./doc/INTEGRATIONS.md) | Connecting Monitors | Zabbix Webhooks, Custom Integrations. |
| [**AI Configuration**](./doc/AI_CONFIGURATION.md) | The Brain Setup | Gemini, OpenAI, RAG Tuning. |
| [**Playbook Authoring**](./doc/PLAYBOOK_AUTHORING.md) | Robot Scripts | Ansible YAML, Dynamic Inventory. |
| [**Security & RBAC**](./doc/RBAC_SECURITY_MODEL.md) | Access Control | Privilege Levels, JWT, WebSSH Security. |
| [**Troubleshooting**](./doc/TROUBLESHOOTING.md) | Fixing Issues | Common Errors, Performance Tuning. |
| [**WebSSH & Security**](./doc/WEBSSH_SECURITY.md) | The Command Center | WebSocket Proxy, xterm.js, XSS Defense. |
| [**Reporting System**](./doc/REPORTING_SYSTEM.md) | The Weekly Checkup | PDF Rendering, Go-Charts, Cron Tasks. |
| [**Frontend Guide**](./doc/FRONTEND_GUIDE.md) | The Interface | Vue 3, Vite, Perceived Speed Optimization. |
| [**Communication**](./doc/COMMUNICATION_NOTIFICATIONS.md) | Notifications | WebSockets, QQ Bot, Whitelist Security. |
| [**API Reference**](./doc/API_REFERENCE.md) | The Language | RESTful Endpoints, RBAC, MCP Protocol. |

---

## ⚡ Engineering Benchmarks
- **High Speed**: Optimized JSON processing (`jsoniter`) makes it 30% faster than standard tools.
- **Remote Ready**: Works perfectly with Microsoft Dev Tunnels (no "Anti-Phishing" blocks).
- **Future Proof**: Supports the **MCP Protocol**, allowing other AI agents to talk to Nagare.

## 📄 License
Apache License 2.0
