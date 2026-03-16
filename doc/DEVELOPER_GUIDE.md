# Nagare Developer Guide: Code Standards And Architecture

This guide is for engineers who extend Nagare, add APIs, or modify the frontend.

## Backend (Go) Standards

Nagare follows a DDD-lite structure with clear layering.

### Directory Structure

- `cmd/server/`: App entry and route registration.
- `internal/api/`: HTTP handlers.
- `internal/service/`: Business logic.
- `internal/repository/`: Data access.
- `internal/model/`: Shared entities and DTOs.

### Adding A New API Endpoint

1. Define or update model structs in `internal/model/`.
2. Add repository functions in `internal/repository/`.
3. Add service logic in `internal/service/`.
4. Expose a handler in `internal/api/`.
5. Register route in `backend/cmd/server/router/`.

### Error Handling

Use unified API responses from handlers:

```go
c.JSON(http.StatusOK, APIResponse{
    Success: true,
    Data: result,
})
```

### Concurrency And Background Work

Use `pkg/queue` for long tasks. Avoid blocking request handlers with heavy operations.

## Frontend (Vue 3) Standards

### Component Style

Use `<script setup>` for new components where possible.

### State

- Keep local state in `ref` and `reactive`.
- Keep shared state in reusable utility or store modules under `src/utils`.

### API Access

Use centralized request clients under `src/api` and `src/utils/request.js`.

### I18n

Do not hardcode user-facing text. Add keys under `src/i18n`.

## Testing And Regression Policy

- Backend regression tests live beside code, for example:
  - `backend/cmd/server/router/routes_regression_test.go`
- Run backend checks with `go test ./...` from `backend/`.
- Frontend unit test assets were removed in current cleanup pass; frontend verification is currently build-based:
  - `npm run build` from `frontend/`

If frontend unit tests are reintroduced, restore both test files and related test dependencies together.

## Tooling

- Format Go with `gofmt`.
- Validate frontend with `npm run build`.
- Keep docs in `doc/` synchronized with implementation changes.
