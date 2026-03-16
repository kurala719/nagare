# Nagare Frontend Architecture And Views Guide

Nagare frontend uses Vue 3 with Element Plus and Vite. Routing is handled by Vue Router, and localization uses vue-i18n.

## Directory Structure

```text
frontend/src/
├── api/         # API wrappers
├── assets/      # Static assets and global styles
├── components/  # Reusable components
├── i18n/        # Localization resources
├── layout/      # Main shell
├── router/      # Route table and guards
├── utils/       # Shared utilities
└── views/       # Page-level views
```

## Core Views

### Authentication And Entry

- `Login.vue`
- `Register.vue`
- `ResetPassword.vue`
- `StatusPage.vue`

### Monitoring And Inventory

- `Monitor.vue`, `MonitorDetail.vue`
- `Group.vue`, `GroupDetail.vue`
- `Host.vue`, `HostDetail.vue`
- `Item.vue`, `ItemDetail.vue`
- `PacketAnalysis.vue`

### Alerting And Delivery

- `Alarm.vue`
- `Alert.vue`
- `Trigger.vue`
- `Action.vue`
- `Media.vue`
- `SiteMessage.vue`

### AI And Knowledge

- `Provider.vue`
- `KnowledgeBase.vue`
- `Analytics.vue`

### System And Admin

- `SystemStatus.vue`
- `Configuration.vue`
- `AuditLog.vue`
- `Log.vue`
- `Reports.vue`
- `Retention.vue`
- `User.vue`
- `RegisterApplication.vue`
- `Profile.vue`
- `About.vue`

### Dashboard Container

- `dashboard/Dashboard.vue`

## Key Frontend Workflows

1. Route guard in `router/index.js` checks token and privilege.
2. HTTP interception in `utils/request.js` handles auth failures.
3. WebSocket views support real-time messaging and terminal sessions.
4. Route-level lazy loading keeps initial load responsive.

## Build And Verification

- Development: `npm run dev`
- Production build: `npm run build`
- Preview build: `npm run preview`

Current baseline validation is build-based. If unit testing is reintroduced later, document test entry points here.
