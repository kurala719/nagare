# QQ Webhook 401 Error - Troubleshooting Guide

## Current Setup
✅ Route configured: `POST /api/v1/media/qq/message`  
✅ No authentication middleware applied  
✅ Follows same pattern as working webhooks (`/alerts/webhook`, `/im/command`)

## Root Cause Analysis

The route is configured correctly in the code. The 401 error is most likely caused by one of these:

### 1. **Server Not Recompiled** (Most Likely)
If you made changes to the router but the Go server is still running the old binary:

**Solution:**
```powershell
# Stop the running server (Ctrl+C or kill the process)

# Rebuild and run
cd d:\Nagare_Project\workd\nagare-v0.21
go build -o web_server.exe ./cmd/web_server
./web_server.exe

# OR use go run (slower but ensures fresh build)
go run ./cmd/web_server/main.go
```

### 2. **Testing Wrong URL**
Make sure you're testing the correct endpoint:
- ✅ Correct: `POST http://localhost:8080/api/v1/media/qq/message`
- ❌ Wrong: `POST http://localhost:8080/api/v1/qq/message`
- ❌ Wrong: `POST http://localhost:8080/media/qq/message`

### 3. **Port Mismatch**
Check your server's actual port in the config or logs. Default is 8080 but could be different.

## Testing Steps

### Step 1: Rebuild the Server
```powershell
cd d:\Nagare_Project\workd\nagare-v0.21

# Clean build
Remove-Item -Path web_server.exe -ErrorAction SilentlyContinue

# Build fresh
go build -o web_server.exe ./cmd/web_server

# Run
./web_server.exe
```

### Step 2: Test the Endpoint
Run the test script:
```powershell
cd d:\Nagare_Project\workd
.\test_qq_webhook.ps1
```

OR test manually:
```powershell
$body = @{
    post_type = "message"
    message_type = "private"
    user_id = 123456789
    message = "/status"
    message_id = 1
    time = [DateTimeOffset]::Now.ToUnixTimeSeconds()
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:8080/api/v1/media/qq/message" `
    -Method POST `
    -Headers @{"Content-Type"="application/json"} `
    -Body $body
```

### Step 3: Verify Route Registration
Check server startup logs. You should see routes being registered. If the logs show the server starting on a different port, use that port.

### Step 4: Compare with Working Endpoint
Test the IM command endpoint (which doesn't require auth):
```powershell
$body = @{
    message = "/status"
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:8080/api/v1/im/command" `
    -Method POST `
    -Headers @{"Content-Type"="application/json"} `
    -Body $body
```

If this works but `/media/qq/message` doesn't, then the server hasn't been rebuilt.

## Debugging

If still getting 401 after rebuild:

1. **Check server logs** for the exact error
2. **Add debug logging** to verify route is registered:
   ```go
   media.POST("/qq/message", func(c *gin.Context) {
       fmt.Println("QQ webhook endpoint hit!")
       presentation.HandleQQMessageCtrl(c)
   })
   ```

3. **List all routes** - Add this temporarily to InitRouter():
   ```go
   for _, route := range r.Routes() {
       fmt.Printf("Route: %s %s\n", route.Method, route.Path)
   }
   ```

## Expected Behavior

### Success Response:
```json
{
  "status": "ok",
  "retcode": 0,
  "message": "success",
  "data": {
    "message_id": null
  }
}
```

The command result will be sent back to QQ automatically via the media provider.

## Quick Fix Checklist
- [ ] Server rebuilt (`go build`)
- [ ] Server restarted  
- [ ] Correct URL tested (`/api/v1/media/qq/message`)
- [ ] Correct port (check server logs)
- [ ] Valid JSON body sent
- [ ] POST method used (not GET)

## Still Not Working?

If after rebuilding it still doesn't work, the issue might be more complex. Please provide:
1. Server startup logs
2. The exact curl/PowerShell command you're using
3. The full error response
4. Output of running the test script
