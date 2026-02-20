# Nagare Developer Guide: Code Standards & Architecture

This guide is for engineers who want to extend Nagare, add new APIs, or modify the frontend. 

---

## 🏗️ Backend (Go) Standards

Nagare follows a **Domain-Driven Design (DDD)** lite approach, organized into four main layers:

### 1. The Directory Structure
- `cmd/server/`: Entry point and HTTP routing.
- `internal/api/`: **Handlers (Controllers)**. They parse input, call services, and return JSON.
- `internal/service/`: **Business Logic**. This is where the core logic (e.g., AI scoring, PDF generation) lives.
- `internal/repository/`: **Data Access (DAO)**. All database queries (GORM) happen here.
- `internal/model/`: **Entities**. Shared data structures used across all layers.

### 2. Adding a New API Endpoint
1. **Define Model**: Add a struct in `internal/model/entities.go`.
2. **Create DAO**: Add a function in `internal/repository/` to handle DB operations.
3. **Write Service**: Add business logic in `internal/service/`. Use the DAO.
4. **Create Controller**: Add a handler in `internal/api/`. Call the service.
5. **Register Route**: Add the endpoint to `cmd/server/router/router.go`.

### 3. Error Handling
Always return the standardized `APIResponse` from controllers:
```go
c.JSON(http.StatusOK, APIResponse{
    Success: true,
    Data: result,
})
```

### 4. Concurrency & Queues
Use the `pkg/queue` for long-running tasks. Never block an HTTP request for more than 500ms. If a task takes longer (like syncing 1000 hosts), run it in a Goroutine and update the status in the DB.

---

## 🎨 Frontend (Vue 3) Standards

### 1. Composition API
We use the `<script setup>` syntax for all new components. It is cleaner and more efficient.

### 2. State Management
- For global UI state (like notifications or user profile), use a reactive store in `src/utils/store.js` (or Pinia if the project grows).
- For local component state, use `ref()` and `reactive()`.

### 3. API Communication
Always use the centralized Axios client in `src/api/`. 
Example:
```javascript
import request from '@/utils/request'

export function getHosts(params) {
  return request({
    url: '/api/v1/hosts',
    method: 'get',
    params
  })
}
```

### 4. Internationalization (i18n)
Never hardcode text. Add the translation to `src/i18n/` and use `$t('key')` in templates.

---

## 🧪 Testing
- **Unit Tests**: Place `_test.go` files alongside the code.
- **API Tests**: Use the scripts in the root `tests/` directory (PowerShell/Shell) to verify real API responses.
- **Frontend Tests**: Use Vitest for unit testing Vue components.

---

## 🛠️ Tooling
- **Formatting**: Run `go fmt` and `npm run lint`.
- **Docs**: Documentation is stored in Markdown in the `doc/` directory. If you change a feature, update the corresponding `.md` file.
