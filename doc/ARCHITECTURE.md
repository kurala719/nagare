# Nagare System Architecture

Nagare follows a decoupled, monorepo architecture designed for high availability and low-latency response.

## 1. High-Level Components

### Backend (Golang)
- **Gin Web Framework**: Handles routing and middleware. Optimized with `jsoniter` for high-frequency metric ingestion.
- **RAG Engine**: Implements a "Retrieve -> Re-rank -> Prompt" pipeline for alert diagnosis.
- **Task Queue (Redis)**: Asynchronous processing for report generation and monitor syncing.
- **WebSocket Hub**: Manages real-time communications for WebSSH and Site Messages.

### Frontend (Vue 3)
- **State Management**: Composition API with reactive refs.
- **Network Layer**: Axios and Fetch utilities with built-in Dev Tunnel bypass headers.
- **UI Architecture**: Component-based design with Skeleton Screens to optimize perceived loading performance.

## 2. API Design Patterns

### Response Format
All API responses follow a unified structure:
```json
{
  "success": true,
  "data": { ... },
  "message": "Operation successful",
  "error": ""
}
```

### Authentication & Authorization
- **JWT**: Stateless authentication via `Authorization: Bearer <token>`.
- **Privilege Levels**: 
  - `Level 1`: Read-only / Basic interaction.
  - `Level 2`: Management (Add/Edit monitors, Generate reports).
  - `Level 3`: Administrative (User management, System config).

## 3. Performance Optimizations
- **JSON Serialization**: Replaced standard library with `jsoniter` for significant CPU savings on high-load endpoints (Webhooks, Metrics).
- **Vite Manual Chunks**: Large libraries (Element-Plus, ECharts, xterm) are split into separate bundles for better caching.
- **Async Processing**: Long-running report tasks are offloaded to Redis workers.

## 4. Connectivity & Deployment
- **Microsoft Dev Tunnels**: Backend and Frontend automatically inject `X-Tunnel-Skip-AntiPhishing-Page` to ensure seamless remote development.
- **Health Checks**: Standard `/health` and `/` endpoints provided for load balancer and heartbeat monitoring.
