# Nagare Frontend Architecture & Views Guide

The Nagare frontend is built with **Vue 3 (Composition API & Options API mix)**, styled with **Element Plus** component library, and packaged by **Vite**. State management and routing are handled by standard Vue ecosystem tools. The project supports internationalization via `vue-i18n`.

---

## 1. Directory Structure

```text
frontend/src/
├── api/          # Axios HTTP clients for all backend endpoints
├── assets/       # Static images, global CSS variables, and icons
├── components/   # Reusable UI widgets (e.g. Chat, Terminal)
├── i18n/         # Language dictionaries (en, zh-CN)
├── layout/       # Main application shell (Sidebar, Header, Content)
├── router/       # Vue Router configuration and route guards
├── styles/       # Global styling overrides
├── utils/        # Helpers (e.g. request.js interceptors)
└── views/        # Top-level page components corresponding to routes
```

---

## 2. Core Views & Functionality

The `views` directory contains the primary pages a user interacts with.

### 2.1 Databoard & System Overview
- **`Home.vue`**: Landing page containing quick navigation and system summary.
- **`dashboard/SystemStatus.vue`**: Primary operational dashboard visualizing network health, recent alerts, and infrastructure topology.
- **`dashboard/Analytics.vue`**: Charts and trends for long-term health scores.
- **`StatusPage.vue`**: Public-facing read-only status page (similar to Atlassian/GitHub status pages).

### 2.2 Infrastructure Management
- **`Monitor.vue`**: Manages external monitoring tool connections (Zabbix/Prometheus).
- **`Group.vue` & `GroupDetail.vue`**: Logical server clusters and their aggregated health.
- **`Host.vue` & `HostDetail.vue`**: Lists individual servers. The detail view shows specific items, history charts, and an AI diagnostic action.
- **`Item.vue` & `ItemDetail.vue`**: Individual metrics and checks (e.g., CPU load, HTTP ping).

### 2.3 Incident Response & Alerting
- **`Alarm.vue`**: Configures incoming webhook endpoints.
- **`Alert.vue`**: Displays active, acknowledged, or resolved incidents. Integrates the "Consult AI" button for RAG-assisted triage.
- **`Trigger.vue`**: Logic rules mapping specific alert conditions to execution actions.
- **`Action.vue`**: Notification templates triggered by rules.

### 2.4 Automation & Operations
- **`AnsiblePlaybook.vue`**: IDE-like view for authoring YAML automation scripts with AI-generation assistance.
- **`AnsibleJob.vue`**: Log output of historical playbook executions against hosts.
- **`Terminal.vue`**: Fully functional WebSSH client utilizing `xterm.js` to connect directly to endpoints.

### 2.5 Intelligence & AI
- **`Provider.vue`**: Configuration of AI models (Gemini, OpenAI) including API keys.
- **`KnowledgeBase.vue`**: Markdown/text upload center to augment the AI with local network documentation and playbooks.

### 2.6 Communications
- **`Media.vue`**: Notification targets (Gmail, Webhook, QQ). Includes a "Test" button to verify configurations.
- **`SiteMessage.vue`**: Real-time internal system notifications powered by WebSocket.
- **`Reports.vue`**: Automated or manual PDF report generation based on infrastructure data.

### 2.7 Security & Settings
- **`Configuration.vue`**: Centralized application settings (Database, MCP, AI, Sync timers).
- **`AuditLog.vue`**: Read-only log of all `POST/PUT/DELETE` API actions by users.
- **`Log.vue`**: Internal application and service logs.
- **`User.vue`**: Identity management and RBAC privilege assignment.
- **`RegisterApplication.vue`**: Administrator approval queue for new users.
- **`Retention.vue`**: Data lifecycle policies for automatic database pruning.
- **`Profile.vue`**: Current user's settings, avatar, and password.

---

## 3. Key Frontend Workflows

1. **Authentication Guard**: `router/index.js` uses a `beforeEach` hook to verify JWT presence. If a user is unauthenticated, they are redirected to `Login.vue`.
2. **I18n (Language)**: Controlled by a global toggle. Changes immediately update all localized strings defined in `i18n/index.js`.
3. **API Interceptors**: `utils/request.js` catches `401 Unauthorized` errors and automatically clears local storage, pushing the user to login.
4. **Real-time Updates**: `SiteMessage.vue` and `Terminal.vue` establish dedicated WebSocket connections for push notifications and bi-directional SSH streaming respectively.
