# Dashboard Display Issues Fix Summary

## Issues Reported
- Topology chart showing nothing
- Health trend chart showing nothing  
- Host page showing nothing

## Root Cause Analysis

The primary issue is **authentication requirement**. All API endpoints require valid JWT authentication:

- `/api/v1/groups` - requires privilege level 1
- `/api/v1/hosts` - requires privilege level 1
- `/api/v1/monitors` - requires privilege level 1
- `/api/v1/system/health/history` - requires authentication

When these endpoints are called without authentication, they return 401/403 errors, causing the pages to display empty or error states.

## Fixes Applied

### 1. Topology Chart ([TopologyChart.vue](../frontend/src/views/dashboard/components/TopologyChart.vue))

**Changes:**
- Improved data extraction logic with `extractArray()` helper to handle multiple response formats
- Added comprehensive console logging for debugging:
  - Raw API responses
  - Extracted data arrays
  - Node/link counts
- Better error handling and logging

**Key improvements:**
```javascript
// More flexible data extraction
const extractArray = (res, ...paths) => {
  if (Array.isArray(res)) return res
  for (const path of paths) {
    const val = path.split('.').reduce((obj, key) => obj?.[key], res)
    if (Array.isArray(val)) return val
  }
  return []
}

const groups = extractArray(groupRes, 'data', 'data.groups', 'groups')
const hosts = extractArray(hostRes, 'data', 'data.hosts', 'data.items', 'hosts', 'items')
const monitors = extractArray(monitorRes, 'data', 'data.monitors', 'monitors')
```

### 2. Health Trend Chart ([HealthTrendChart.vue](../frontend/src/views/dashboard/components/HealthTrendChart.vue))

**Changes:**
- Added console logging for debugging:
  - Request parameters (from/to timestamps)
  - Raw response data
  - Extracted rows count
  - Series data points
- Improved error messages

### 3. Host Page ([Host.vue](../frontend/src/views/Host.vue))

**Changes:**
- Added console logging throughout data loading pipeline:
  - Monitor loading with response data
  - Group loading with response data
  - Host loading with request parameters and response
  - Extracted payload counts
- Helps identify exact point of failure in data pipeline

### 4. Health Score Backend ([health.go](../backend/internal/service/health.go))

**No changes needed** - Already updated with group metrics support:
- `GroupTotal`, `GroupActive`, `GroupImpacted` fields
- Proper weighting calculation (0.3 monitor, 0.2 group, 0.3 host, 0.2 item)

### 5. Network Status History ([history.go](../backend/internal/service/history.go))

**Already updated** with group fields in:
- `NetworkStatusHistoryResp` struct
- `GetNetworkStatusHistoryServ()` response mapping
- `recordNetworkStatusSnapshot()` snapshot creation

### 6. Dashboard Health Stats ([HealthStats.vue](../frontend/src/views/dashboard/components/HealthStats.vue))

**Already updated** with:
- 5 metric cards (score, monitors, groups, hosts, items)
- Responsive layout adjustments
- Group impacted count badge
- Purple gradient styling for groups icon

### 7. I18n Labels ([index.js](../frontend/src/i18n/index.js))

**Already added:**
- `activeGroups` / `活跃分组`
- `impactedGroups` / `受影响分组`
- `impactedLabel` / `受影响`

## Testing the Fixes

### Prerequisites

1. **Backend must be running:**
   ```powershell
   cd d:\Nagare_Project\nagare\backend
   .\bin\nagare-web-server
   ```

2. **You must be logged in** with valid credentials to access the dashboard

### Test Steps

1. **Start the frontend dev server:**
   ```powershell
   cd d:\Nagare_Project\nagare\frontend
   npm run dev
   ```

2. **Navigate to the application** (typically `http://localhost:5173`)

3. **Log in** with your credentials

4. **Open browser DevTools** (F12) → Console tab

5. **Navigate to Dashboard** and check console logs:

   **Expected logs for Topology Chart:**
   ```
   TopologyChart raw responses: {groupRes: {...}, hostRes: {...}, monitorRes: {...}}
   TopologyChart extracted data: {groups: Array(N), hosts: Array(M), monitors: Array(K)}
   TopologyChart graph: {nodeCount: X, linkCount: Y}
   ```

   **Expected logs for Health Trend:**
   ```
   HealthTrendChart fetching data: {from: ..., to: ...}
   HealthTrendChart raw response: {...}
   HealthTrendChart extracted rows: N
   HealthTrendChart series: {points: N, prevPoints: M}
   ```

6. **Navigate to Host page** and check console logs:

   **Expected logs:**
   ```
   Host page loading monitors...
   Host page monitors response: {...}
   Host page monitors loaded: N
   Host page loading groups...
   Host page groups response: {...}
   Host page groups loaded: M
   Host page loading hosts with params: {...}
   Host page hosts response: {...}
   Host page extracted payload: {count: K, total: K}
   Host page hosts loaded: K
   ```

### Debugging Authentication Issues

If you see authentication errors:

1. **Check JWT token:**
   - Open DevTools → Application → Local Storage
   - Look for auth token (key depends on your auth implementation)
   - Check if token exists and is not expired

2. **Check network requests:**
   - DevTools → Network tab
   - Look for 401/403 responses
   - Check if `Authorization: Bearer <token>` header is present

3. **Backend logs:**
   - Check backend console for auth-related errors
   - Look for JWT parsing errors or privilege issues

## API Response Formats

After authentication, the APIs should return data in these formats:

### Groups API: `/api/v1/groups`
```json
{
  "success": true,
  "data": [
    {
      "ID": 1,
      "Name": "Group Name",
      "Status": 1,
      "Enabled": 1,
      ...
    }
  ]
}
```

### Hosts API: `/api/v1/hosts`
```json
{
  "success": true,
  "data": {
    "items": [...],
    "total": 10
  }
}
```

### Monitors API: `/api/v1/monitors`
```json
{
  "success": true,
  "data": [...]
}
```

### Health API: `/api/v1/system/health`
```json
{
  "success": true,
  "data": {
    "score": 93,
    "monitor_total": 1,
    "monitor_active": 1,
    "group_total": 1,
    "group_active": 1,
    "group_impacted": 1,
    "host_total": 10,
    "host_active": 8,
    "item_total": 1868,
    "item_active": 1780
  }
}
```

### History API: `/api/v1/system/health/history`
```json
{
  "success": true,
  "data": [
    {
      "score": 93,
      "monitor_total": 1,
      "monitor_active": 1,
      "group_total": 1,
      "group_active": 1,
      "group_impacted": 1,
      "host_total": 10,
      "host_active": 8,
      "item_total": 1868,
      "item_active": 1780,
      "sampled_at": "2026-02-17T19:20:00Z"
    }
  ]
}
```

## Next Steps

1. **Ensure you have a valid user account** - if not, create one or check registration process
2. **Log in to the application**
3. **Test each page** with browser console open
4. **Review console logs** to verify data is being loaded correctly
5. **Report any specific errors** seen in console for further investigation

## Files Modified

### Frontend
- [frontend/src/views/dashboard/components/TopologyChart.vue](../frontend/src/views/dashboard/components/TopologyChart.vue)
- [frontend/src/views/dashboard/components/HealthTrendChart.vue](../frontend/src/views/dashboard/components/HealthTrendChart.vue)
- [frontend/src/views/Host.vue](../frontend/src/views/Host.vue)

### Backend
- Already updated in previous work:
  - [backend/internal/service/health.go](../backend/internal/service/health.go)
  - [backend/internal/service/history.go](../backend/internal/service/history.go)
  - [backend/internal/model/entities.go](../backend/internal/model/entities.go)

### Build Status
- ✅ Backend compiled successfully
- ✅ Frontend built successfully (dist folder created)
