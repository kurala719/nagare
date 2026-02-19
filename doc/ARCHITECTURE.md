# Nagare Project Architecture

## Purpose
Nagare is a unified monitoring and automation platform composed of a Go-based backend API server and a Vue 3-based frontend web UI. It centralizes monitors, hosts, items, alerts, media providers, and actions, offering both a human-friendly interface and programmatic APIs for seamless operations and integrations.

## Architecture At A Glance
- **Backend**: Go (Gin) HTTP server located in the `backend` directory.
- **Frontend**: Vue 3 + Vite Single Page Application (SPA) located in the `frontend` directory.
- **API Base Path**: `/api/v1`.
- **Authentication**: JWT-based with a robust privilege level system enforced by middleware.
- **Integrations**: Includes QQ (OneBot 11) message webhook, MCP server endpoints, and various media providers.
- **Data Storage**: Uses GORM for database interactions, supporting various SQL dialects.
- **Task Queue**: Optional Redis integration for background task processing.

## Auth And Privilege Model
- **Privilege 1**: Basic authenticated access (read-only on most resources).
- **Privilege 2**: Operator/Admin access (create/update/delete on most resources).
- **Privilege 3**: Super Admin access (system configuration and user administration).
- **Public Endpoints**: Selected webhook and system health endpoints are intentionally unauthenticated.
- **MCP Endpoints**: Protected by a dedicated API key middleware.

## Backend API Reference (Endpoint-By-Endpoint)
All endpoints below are prefixed by `/api/v1`.

### Configuration (Privilege 3)
- `GET /config/` : Get main configuration
- `PUT /config/` : Update main configuration
- `POST /config/save` : Save configuration to disk
- `POST /config/reload` : Reload configuration from disk

### Monitors
**Read (Privilege 1):**
- `GET /monitors/` : Search monitors
- `GET /monitors/:id` : Get monitor by ID

**Write (Privilege 2):**
- `POST /monitors/` : Create monitor
- `PUT /monitors/:id` : Update monitor
- `DELETE /monitors/:id` : Delete monitor
- `POST /monitors/:id/login` : Authenticate monitor (obtain auth token)
- `POST /monitors/check` : Check all monitor status
- `POST /monitors/:id/check` : Check monitor status

### Groups
**Read (Privilege 1):**
- `GET /groups/` : Search groups
- `GET /groups/:id` : Get group by ID
- `GET /groups/:id/detail` : Get group detail

**Write (Privilege 2):**
- `POST /groups/` : Create group
- `PUT /groups/:id` : Update group
- `DELETE /groups/:id` : Delete group
- `POST /groups/:id/pull` : Pull group data from monitors
- `POST /groups/:id/push` : Push group data to monitors
- `POST /groups/check` : Check all group status
- `POST /groups/:id/check` : Check group status

### Hosts
**Read (Privilege 1):**
- `GET /hosts/` : Search hosts
- `GET /hosts/:id` : Get host by ID
- `GET /hosts/:id/history` : Get host history
- `POST /hosts/:id/consult` : Consult host
- `GET /hosts/:id/ssh` : WebSSH WebSocket endpoint (Host-specific)

**Generic Terminal (Privilege 1):**
- `GET /terminal/ssh` : WebSSH WebSocket endpoint (Ad-hoc connection via query params)

**Write (Privilege 2):**
- `POST /hosts/` : Create host
- `PUT /hosts/:id` : Update host
- `DELETE /hosts/:id` : Delete host

**Monitor Host Sync (Privilege 2):**
- `POST /monitors/:id/hosts/pull` : Pull all hosts from monitor
- `POST /monitors/:id/hosts/:hid/pull` : Pull single host from monitor
- `POST /monitors/:id/hosts/push` : Push all hosts to monitor
- `POST /monitors/:id/hosts/:hid/push` : Push single host to monitor

### Site Messages (Privilege 1)
- `GET /site-messages/` : Get notifications for current user
- `GET /site-messages/unread-count` : Get number of unread notifications
- `PUT /site-messages/:id/read` : Mark a message as read
- `PUT /site-messages/read-all` : Mark all messages as read
- `DELETE /site-messages/:id` : Delete a notification
- `GET /site-messages/ws` : WebSocket endpoint for real-time notifications

### Knowledge Base
**Read (Privilege 1):**
- `GET /knowledge-base` : Search/List knowledge base entries
- `GET /knowledge-base/:id` : Get specific entry

**Write (Privilege 2):**
- `POST /knowledge-base` : Create knowledge entry
- `PUT /knowledge-base/:id` : Update knowledge entry
- `DELETE /knowledge-base/:id` : Delete knowledge entry

### Reports (Privilege 2)
- `GET /reports` : List generated reports
- `GET /reports/:id` : Get report metadata
- `GET /reports/config` : Get reporting configuration
- `PUT /reports/config` : Update reporting configuration
- `POST /reports/generate/weekly` : Trigger weekly report generation
- `POST /reports/generate/monthly` : Trigger monthly report generation
- `DELETE /reports/:id` : Delete report
- `GET /reports/:id/download` : Download PDF report file

### Alerts
**Public Webhook:**
- `POST /alerts/webhook` : Inbound webhook for alerts (no auth)

**Read (Privilege 1):**
- `GET /alerts/` : Search alerts
- `GET /alerts/:id` : Get alert by ID
- `POST /alerts/:id/consult` : Consult alert

**Write (Privilege 2):**
- `POST /alerts/` : Create alert
- `PUT /alerts/:id` : Update alert
- `DELETE /alerts/:id` : Delete alert

### System (Public)
- `GET /system/health` : Get health score
- `GET /system/health/history` : Get network status history
- `GET /system/metrics` : Get network metrics

### IM (Public)
- `POST /im/command` : Send IM command

### Media Types
**Read (Privilege 1):**
- `GET /media-types/` : Search media types
- `GET /media-types/:id` : Get media type by ID

**Write (Privilege 2):**
- `POST /media-types/` : Create media type
- `PUT /media-types/:id` : Update media type
- `DELETE /media-types/:id` : Delete media type

### Media
**Public Webhook:**
- `POST /media/qq/message` : QQ OneBot 11 message webhook (no auth)

**Read (Privilege 1):**
- `GET /media/` : Search media
- `GET /media/:id` : Get media by ID

**Write (Privilege 2):**
- `POST /media/` : Create media
- `PUT /media/:id` : Update media
- `DELETE /media/:id` : Delete media

### Actions
**Read (Privilege 1):**
- `GET /actions/` : Search actions
- `GET /actions/:id` : Get action by ID

**Write (Privilege 2):**
- `POST /actions/` : Create action
- `PUT /actions/:id` : Update action
- `DELETE /actions/:id` : Delete action

### Triggers
**Read (Privilege 1):**
- `GET /triggers/` : Search triggers
- `GET /triggers/:id` : Get trigger by ID

**Write (Privilege 2):**
- `POST /triggers/` : Create trigger
- `PUT /triggers/:id` : Update trigger
- `DELETE /triggers/:id` : Delete trigger

### Logs (Privilege 2)
- `GET /logs/system` : Get system logs
- `GET /logs/service` : Get service logs

### Items
**Read (Privilege 1):**
- `GET /items/` : Search items
- `GET /items/:id` : Get item by ID
- `GET /items/:id/history` : Get item history
- `POST /items/:id/consult` : Consult item

**Write (Privilege 2):**
- `POST /items/` : Create item
- `PUT /items/:id` : Update item
- `DELETE /items/:id` : Delete item
- `POST /items/hosts/:hid/import` : Import items by host ID from monitor

**Monitor Item Sync (Privilege 2):**
- `POST /monitors/:id/items/pull` : Pull all items from monitor
- `POST /monitors/:id/items/push` : Push all items to monitor

**Monitor Host Item Sync (Privilege 2):**
- `POST /monitors/:id/hosts/:hid/items/pull` : Pull all items for a host
- `POST /monitors/:id/hosts/:hid/items/:item_id/pull` : Pull a single item for a host
- `POST /monitors/:id/hosts/:hid/items/push` : Push all items for a host
- `POST /monitors/:id/hosts/:hid/items/:item_id/push` : Push a single item for a host

### Chats (Privilege 1)
- `GET /chats/` : Search chats
- `POST /chats/` : Send chat

### Providers
**Read (Privilege 1):**
- `GET /providers/` : Search providers
- `GET /providers/:id` : Get provider by ID

**Write (Privilege 2):**
- `POST /providers/` : Create provider
- `PUT /providers/:id` : Update provider
- `DELETE /providers/:id` : Delete provider
- `POST /providers/check` : Check all provider status
- `POST /providers/:id/check` : Check provider status

### Auth, Users, Registration
**Public Auth:**
- `POST /auth/login` : Login
- `POST /auth/register` : Register

**Authenticated (Privilege 1):**
- `POST /auth/reset` : Reset password

**Register Applications (Privilege 3):**
- `GET /register-applications/` : List pending register applications
- `PUT /register-applications/:id/approve` : Approve
- `PUT /register-applications/:id/reject` : Reject

**Legacy Register Applications (Privilege 3):**
- `GET /register-application/` : List pending register applications
- `PUT /register-application/:id/approve` : Approve
- `PUT /register-application/:id/reject` : Reject

**Users Read (Privilege 2):**
- `GET /users/` : Search users
- `GET /users/:id` : Get user by ID

**Users Write (Privilege 3):**
- `POST /users/` : Create user
- `PUT /users/:id` : Update user
- `DELETE /users/:id` : Delete user

### User Information
**Self (Privilege 1):**
- `GET /user-info/me` : Get current user info
- `POST /user-info/me` : Create current user info
- `PUT /user-info/me` : Update current user info
- `DELETE /user-info/me` : Delete current user info

**Admin (Privilege 3):**
- `GET /user-info/users/:user_id` : Get user info by user ID
- `PUT /user-info/users/:user_id` : Update user info by user ID

### MCP (API Key)
- `GET /mcp/sse` : Server-sent events stream
- `POST /mcp/message` : Send MCP message

## Frontend Pages And Routes
All routes are defined in the SPA router and map to Vue view components.

**Public:**
- `/login` -> Login
- `/register` -> Register

**Authenticated (Privilege 1):**
- `/reset-password` -> ResetPassword
- `/dashboard` -> Dashboard
- `/alert` -> Alert
- `/host` -> Host
- `/host/:id/detail` -> HostDetail
- `/host/:id/terminal` -> Terminal (WebSSH, Host-specific)
- `/terminal` -> Terminal (WebSSH, with host selector and direct connect)
- `/group` -> Group
- `/group/:id/detail` -> GroupDetail
- `/monitor` -> Monitor
- `/knowledge-base` -> KnowledgeBase
- `/item` -> Item
- `/item/:id/detail` -> ItemDetail
- `/host/:hostId/items` -> Redirect to /item with hostId filter
- `/profile` -> Profile
- `/site-messages` -> SiteMessages (Notification History)

**Privilege 2:**
- `/provider` -> Provider
- `/media` -> Media
- `/media-type` -> MediaType
- `/action` -> Action
- `/trigger` -> Trigger
- `/log` -> Log
- `/user` -> User
- `/reports` -> Reports

**Privilege 3:**
- `/register-application` -> RegisterApplication
- `/system` -> System

**Default:**
- `/` -> Redirects to `/dashboard`

## Integrations
### QQ OneBot 11
- **Webhook Endpoint**: `POST /api/v1/media/qq/message`
- **Supported Commands**: `/status`, `/get_alert`, `/chat <message>`
- **Typical Flow**: QQ user -> OneBot (NapCat) -> Nagare webhook -> command execution -> reply to QQ
- **Docs**: `QUICK_START_QQ_API.md`, `MEDIA_MESSAGE_API_IMPLEMENTATION.md`

### Media Providers
- Configured via the Media and Media Type modules.
- Used to send notifications and IM replies (including QQ).

### MCP
- API key protected endpoints for message and SSE streaming.

## Operations
### Frontend
- **Install dependencies**: `npm install`
- **Dev server**: `npm run dev`
- **Production build**: `npm run build`

### Backend
- **Main server entry**: `backend/cmd/server/main.go`
- **Router definition**: `backend/cmd/server/router/router.go`
- **Config file**: `backend/configs/nagare_config.json`

## Source Map
- **Backend Routes**: `backend/cmd/server/router/router.go`
- **Frontend Routes**: `frontend/src/router/index.js`
- **QQ Webhook Docs**: `QUICK_START_QQ_API.md`, `MEDIA_MESSAGE_API_IMPLEMENTATION.md`
- **Sync and Status Logic**: `STATUS_VALIDATION_IMPLEMENTATION.md`
- **Push Functions**: `PUSH_FUNCTIONS_GUIDE.md`
- **Monitor Login**: `MONITOR_LOGIN_TEST.md`
