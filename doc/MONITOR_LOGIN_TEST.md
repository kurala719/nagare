# Monitor Login Testing Guide

## Overview
This document describes how to test the monitor login functionality that has been implemented.

## Changes Made

### Backend Changes

1. **Application Layer** ([monitor.go](nagare-v0.21/internal/web_server/application/monitor.go))
   - Modified `LoginMonitorServ()` to return `MonitorResp` with auth_token instead of just error
   - Modified `AddMonitorServ()` to return `MonitorResp` with auth_token and auto-login if credentials provided
   - Added context-based authentication flow

2. **Presentation Layer** ([monitor.go](nagare-v0.21/internal/web_server/presentation/monitor.go))
   - Updated `LoginMonitorCtrl()` to return monitor data with auth_token in response
   - Updated `AddMonitorCtrl()` to return created monitor data with auth_token

3. **Router** ([router.go](nagare-v0.21/cmd/web_server/router/router.go))
   - Changed `/monitor/:id/login` from GET to POST (more appropriate for authentication)

### Frontend Changes

1. **API Layer** ([monitors.js](nagare_web/src/api/monitors.js))
   - Added `loginMonitor(id)` function to call POST `/monitor/:id/login`

2. **Monitor View** ([Monitor.vue](nagare_web/src/views/Monitor.vue))
   - Added login button for each monitor card
   - Button shows "Login" if no auth_token, "Re-login" if auth_token exists
   - Added loading state for login operations
   - Auto-refresh every 30 seconds to update monitor status
   - Added manual refresh button in toolbar
   - Status indicator now shows green (authenticated) or red (not authenticated) based on auth_token
   - Auto-login when creating new monitor with credentials

3. **Internationalization** ([index.js](nagare_web/src/i18n/index.js))
   - Added translations for "login" and "reLogin" in English and Chinese

## Testing Steps

### 1. Test Manual Login (Existing Monitor)

```bash
# Using curl (now POST instead of GET)
curl -X POST http://localhost:8080/monitor/1/login \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json"
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "Zabbix Monitor",
    "url": "http://zabbix-server.com",
    "username": "admin",
    "password": "password",
    "auth_token": "SESSION_TOKEN_HERE",
    "type": "zabbix",
    "description": "Main Zabbix server"
  }
}
```

### 2. Test Creating Monitor with Auto-Login

```bash
curl -X POST http://localhost:8080/monitor/ \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Zabbix",
    "url": "http://your-zabbix-server/zabbix",
    "username": "Admin",
    "password": "zabbix",
    "type": "zabbix",
    "description": "Test monitor"
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "id": 2,
    "name": "Test Zabbix",
    "url": "http://your-zabbix-server/zabbix",
    "username": "Admin",
    "password": "zabbix",
    "auth_token": "AUTO_GENERATED_TOKEN",
    "type": "zabbix",
    "description": "Test monitor"
  }
}
```

**Note:** If credentials are invalid or server is unreachable, the monitor will still be created but `auth_token` will be empty.

### 3. Test Frontend Login

1. Navigate to the Monitors page in the web interface
2. You should see a "Login" or "Re-login" button for each monitor
3. Click the login button
4. Check for success/error message
5. Monitor status should change from red (offline) to green (online) if login succeeds

### 4. Test Auto-Refresh

1. Keep the Monitors page open
2. Every 30 seconds, the page will automatically refresh monitor data
3. You can also click the refresh button in the toolbar for manual refresh
4. Monitor statuses should update automatically

### 5. Test Creating Monitor with Auto-Login (Frontend)

1. Click "Create a new Monitor" button
2. Fill in all fields including:
   - Name
   - URL (e.g., `http://your-zabbix-server/zabbix`)
   - Username
   - Password
   - Type (e.g., `zabbix`)
   - Description
3. Click "Create"
4. The monitor should be created and automatically attempt to login
5. If successful, you'll see "Monitor created and logged in successfully!"
6. If login fails, you'll see "Monitor created successfully!" (can login manually later)

## Features Implemented

### ✅ Backend
- [x] Return auth_token in login response
- [x] Auto-login when creating monitor with credentials
- [x] Changed login endpoint to POST method
- [x] Proper error handling and token validation

### ✅ Frontend
- [x] Login button for each monitor
- [x] Visual status indicator (green = authenticated, red = not authenticated)
- [x] Loading state during login
- [x] Auto-refresh every 30 seconds
- [x] Manual refresh button
- [x] Auto-login on monitor creation
- [x] Internationalization support

## Troubleshooting

### Auth Token Not Received
- Check monitor credentials (username/password)
- Verify monitor URL is accessible
- Check monitor type is correctly set (e.g., "zabbix")
- Review backend logs for authentication errors

### Frontend Not Showing Status
- Check browser console for errors
- Verify API responses include `auth_token` field
- Clear browser cache and reload

### Auto-Refresh Not Working
- Check browser console for errors
- Verify component is properly mounted/unmounted
- Check that `refreshTimer` is being set

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/monitor/` | Get all monitors |
| GET | `/monitor/:id` | Get monitor by ID |
| POST | `/monitor/` | Create new monitor (auto-login if credentials provided) |
| PUT | `/monitor/:id` | Update monitor |
| DELETE | `/monitor/:id` | Delete monitor |
| POST | `/monitor/:id/login` | Authenticate with monitor and retrieve auth_token |

## Monitor Status Logic

A monitor is considered "online" (authenticated) if it has a non-empty `auth_token` field. The frontend checks:

```javascript
// Green checkmark if authenticated
monitor.auth_token ? true : false
```

This provides a simple visual indicator of which monitors are successfully authenticated and ready to sync data.
