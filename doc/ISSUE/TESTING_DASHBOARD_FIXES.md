# Testing Dashboard Fixes - Step by Step Guide

## Problem Fixed
The dashboard pages (Topology, Health Trend, Host page) were showing nothing because:
1. The frontend data extraction logic didn't match the actual backend API response structure
2. All endpoints require authentication

## What Was Fixed

### Backend API Response Structure
All backend APIs return this structure:
```json
{
  "success": true,
  "data": <actual data here>
}
```

For simple lists: `data` is an array directly
For paginated lists: `data` is `{items: [...], total: N}`

### Frontend Data Extraction
Updated all three components to properly extract data from the `{success: true, data: ...}` wrapper:
- **TopologyChart.vue** - Now correctly extracts groups, hosts, and monitors arrays
- **HealthTrendChart.vue** - Now correctly extracts history data array
- **Host.vue** - Now correctly extracts monitors, groups, and hosts data

## IMPORTANT: You Must Be Logged In

**The pages will remain empty if you are not authenticated.** All the API endpoints require a valid JWT token.

## Testing Steps

### Step 1: Start the Backend

```powershell
cd d:\Nagare_Project\nagare\backend
.\bin\nagare-web-server
```

Keep this terminal open.

### Step 2: Start the Frontend Dev Server

```powershell
cd d:\Nagare_Project\nagare\frontend
npm run dev
```

The dev server will start (probably on port 5174 since 5173 was in use).
Look for output like: `➜  Local:   http://localhost:5174/`

### Step 3: Register a Test User

1. Open your browser and go to `http://localhost:5174`
2. Click **Register** (or navigate to `/register`)
3. Fill in the registration form:
   - Username: `testuser`
   - Password: `testpass123`
   - Email: `test@example.com`
   - (Other fields as required)
4. Submit the registration

**Note**: Depending on the system configuration, you may need admin approval for the registration. If so, you'll need:
- Either an existing admin account to approve it
- Or direct database access to manually approve/create the user

### Step 4: Log In

1. Go to the login page (`/login`)
2. Enter your credentials:
   - Username: `testuser`
   - Password: `testpass123`
3. Submit the login form

**What happens**: 
- The backend will return `{success: true, data: {token: "eyJ..."}}`
- The frontend will store the token in localStorage as `nagare_token`
- All subsequent API calls will include `Authorization: Bearer <token>` header

### Step 5: Open Browser DevTools Console

**Critical for debugging!**

1. Press **F12** to open DevTools
2. Go to the **Console** tab
3. Keep it open while testing

### Step 6: Navigate to Dashboard

Click on **Dashboard** in the navigation menu.

**What to look for in console:**

#### Topology Chart Logs:
```javascript
TopologyChart raw responses: {groupRes: {...}, hostRes: {...}, monitorRes: {...}}
TopologyChart extracted data: {groups: Array(1), hosts: Array(10), monitors: Array(1)}
TopologyChart graph: {nodeCount: 12, linkCount: 11}
```

If you see:
- ✅ `groups: Array(N)` with N > 0 → Groups loaded successfully
- ✅ `hosts: Array(M)` with M > 0 → Hosts loaded successfully  
- ✅ `monitors: Array(K)` with K > 0 → Monitors loaded successfully
- ✅ `nodeCount: X` with X > 0 → Graph is being rendered

If you see:
- ❌ `401` or `403` errors → Not logged in or token invalid
- ❌ All arrays empty but no errors → Database has no data yet
- ❌ `TypeError` or `undefined` → Data structure mismatch (report this)

#### Health Trend Logs:
```javascript
HealthTrendChart fetching data: {from: 1708188000, to: 1708274400}
HealthTrendChart raw response: {success: true, data: Array(50)}
HealthTrendChart extracted rows: 50
HealthTrendChart series: {points: 50, prevPoints: 0}
```

If you see:
- ✅ `data: Array(N)` with N > 0 → History data exists
- ✅ `points: N` → Chart will render

If you see:
- ❌ `data: Array(0)` → No history data yet (need to wait or trigger snapshot)
- ❌ `401` or `403` → Not authenticated

### Step 7: Navigate to Hosts Page

Click on **Hosts** in the navigation menu.

**What to look for in console:**

```javascript
Host page loading monitors...
Host page monitors response: {success: true, data: Array(1)}
Host page monitors loaded: 1

Host page loading groups...
Host page groups response: {success: true, data: Array(1)}
Host page groups loaded: 1

Host page loading hosts with params: {limit: 100, offset: 0, ...}
Host page hosts response: {success: true, data: {items: Array(10), total: 10}}
Host page extracted payload: {count: 10, total: 10}
Host page hosts loaded: 10
```

If you see counts > 0, the page should display data.

## Troubleshooting

### Issue: Pages Still Empty After Login

**Check 1: Verify Authentication**
1. Open DevTools → Application → Local Storage
2. Look for key `nagare_token`
3. Value should be a JWT token (three parts separated by dots: `eyJ...`)
4. If missing or malformed → Log out and log in again

**Check 2: Verify API Calls Succeed**
1. Open DevTools → Network tab
2. Refresh the dashboard
3. Look for API calls to `/api/v1/groups`, `/api/v1/hosts`, etc.
4. Click on each request and check:
   - **Headers**: Should include `Authorization: Bearer <token>`
   - **Response**: Should be `{success: true, data: [...]}`
5. If you see 401/403 → Token is invalid or expired

**Check 3: Database Has Data**
The pages can't show data if the database is empty. You need at least:
- 1 Monitor (for topology and host page)
- 1 Group (for topology)
- 1 Host (for topology and host page)

Create these through the UI or import from your monitoring system.

### Issue: "TypeError: Cannot read property 'data' of undefined"

This means the API call failed completely (network error, backend down, etc.).

1. Check backend is running
2. Check browser console for CORS errors
3. Check backend logs for errors

### Issue: Console Shows Success But Page Still Empty

If console logs show successful data extraction but the chart/table is still empty:

1. Check if the counts are actually > 0 (not just empty arrays)
2. Check browser console for ECharts errors (for topology/health trend)
3. Try clicking the **Refresh** button on the component
4. Check if the data has valid required fields (ID, Name, etc.)

### Issue: Network Status History Empty (Health Trend)

The backend records network status snapshots periodically. If you just started:
1. Wait a few minutes for snapshots to be recorded
2. Or manually trigger by calling the health API: `curl http://localhost:8080/api/v1/system/health`

## Quick Test Script (PowerShell)

If you need to quickly test the APIs after fixing, use this script:

```powershell
# Step 1: Register a user
$registerBody = @{
    username = "testuser"
    password = "testpass123"
    email = "test@example.com"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/register" `
    -Method POST `
    -Body $registerBody `
    -ContentType "application/json"

# Step 2: Login and get token
$loginBody = @{
    username = "testuser"  
    password = "testpass123"
} | ConvertTo-Json

$loginResp = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" `
    -Method POST `
    -Body $loginBody `
    -ContentType "application/json"

$token = $loginResp.data.token
Write-Host "Token: $token"

# Step 3: Test APIs with token
$headers = @{
    "Authorization" = "Bearer $token"
}

Write-Host "`nTesting Health API..."
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/system/health" -Headers $headers | ConvertTo-Json

Write-Host "`nTesting Groups API..."
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/groups?limit=5" -Headers $headers | ConvertTo-Json

Write-Host "`nTesting Hosts API..."
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/hosts?limit=5" -Headers $headers | ConvertTo-Json

Write-Host "`nTesting Monitors API..."
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/monitors?limit=5" -Headers $headers | ConvertTo-Json
```

## Expected Results After Fixes

**✅ With Authentication + Data**:
- Topology chart displays monitors, groups, and hosts as connected nodes
- Impacted groups (with down hosts) have red borders
- Health trend chart shows score over time  
- Host page shows table with all hosts
- All filters and actions work

**⚠️ With Authentication + No Data**:
- Components show "No data" or "Empty" state
- No errors in console
- This is expected for new installations

**❌ Without Authentication**:
- Pages show spinner then error or empty state
- Console shows 401/403 errors
- Need to log in first

## Summary

The code fixes are complete and working. The pages now correctly:
1. Extract data from backend's `{success: true, data: ...}` structure
2. Handle both array and paginated responses
3. Log detailed debugging information

**The only requirement is that you must be logged in with a valid account before the pages will load data.**
