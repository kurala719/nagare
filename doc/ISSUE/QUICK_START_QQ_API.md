# QQ Media Message API - Quick Start Guide

## What Was Implemented

A new API endpoint that allows QQ (via OneBot 11/NapCat) to send commands to Nagare server and receive responses.

## Endpoint

```
POST /api/v1/media/qq/message
```

## Supported Commands

| Command | Description | Example |
|---------|-------------|---------|
| `/status` | Get system health score | `/status` |
| `/get_alert` | Get active alerts (max 10) | `/get_alert` |
| `/chat` | Chat with LLM provider | `/chat What is the network status?` |

## How to Test

### Method 1: Using curl (Direct Testing)

```bash
# Test /status command
curl -X POST http://localhost:8080/api/v1/media/qq/message \
  -H "Content-Type: application/json" \
  -d '{
    "post_type": "message",
    "message_type": "private",
    "user_id": 123456789,
    "message": "/status",
    "message_id": 1,
    "time": 1234567890
  }'

# Test /get_alert command
curl -X POST http://localhost:8080/api/v1/media/qq/message \
  -H "Content-Type: application/json" \
  -d '{
    "post_type": "message",
    "message_type": "private",
    "user_id": 123456789,
    "message": "/get_alert",
    "message_id": 2,
    "time": 1234567890
  }'

# Test /chat command
curl -X POST http://localhost:8080/api/v1/media/qq/message \
  -H "Content-Type: application/json" \
  -d '{
    "post_type": "message",
    "message_type": "private",
    "user_id": 123456789,
    "message": "/chat What systems are currently experiencing issues?",
    "message_id": 3,
    "time": 1234567890
  }'
```

### Method 2: Using PowerShell

```powershell
# Test /status command
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

### Method 3: Real QQ Integration with NapCat

1. **Install NapCat** on a machine with QQ:
   ```bash
   # Install NapCat (see https://napneko.github.io for details)
   ```

2. **Configure NapCat** to send events to Nagare:
   ```
   Set the event push URL to: http://your-nagare-server:8080/api/v1/media/qq/message
   ```

3. **Add QQ Media in Nagare Dashboard**:
   - Go to Settings → Media
   - Add new media type: QQ
   - Set NapCat server URL

4. **Send Commands in QQ**:
   - Add the QQ bot as friend
   - Send messages: `/status`, `/get_alert`, `/chat ...`

## Response Format

Successful response (200 OK):
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

The actual command response is automatically sent back to the QQ user.

## Example Command Responses

### `/status` Response
```
Health Score: 85 (monitors 5/5, hosts 3/3, items 12/12)
```

### `/get_alert` Response
```
Active Alerts (2):
[1] Database connection timeout (Severity: 2)
[2] High CPU usage (Severity: 1)
```

### `/chat` Response
```
Based on your network status, all monitored systems are operating normally. 
The database has 2GB free space, and network latency is within normal ranges.
```

## Troubleshooting

### Issue: "No LLM provider configured"
- **Solution**: Add an LLM provider in Settings → Providers

### Issue: "No active alerts"
- **Cause**: This is normal if there are no active alerts
- **Response**: "No active alerts."

### Issue: Command not recognized
- **Check**: Ensure message starts with `/` and use correct command name
- **Example**: `/status`, `/get_alert`, `/chat message`

### Issue: Message not sent back to QQ
- **Check**: Verify NapCat server URL is configured correctly in Nagare
- **Check**: Ensure firewall allows communication between servers
- **Check**: Review system logs for SendIMReply errors

## Code Changes Summary

### 1. `backend/internal/service/im_command.go`
```go
// Enhanced HandleIMCommand to support:
// - /status (existing)
// - /get_alert (new)
// - /chat (new)

// New helper functions:
// - handleGetAlerts()
// - handleChatCommand()
// - HandleIMCommandWithContext()
```

### 2. `backend/internal/api/media.go`
```go
// New controller:
// - HandleQQMessageCtrl()

// New data structures:
// - OneBotMessageEvent
// - OneBotSender
// - OneBotMessageResponse
// - OneBotMessageResponseData
```

### 3. `backend/cmd/server/router/router.go`
```go
// New route added to setupMediaRoutes():
// media.POST("/qq/message", api.HandleQQMessageCtrl)
```

## Key Features

✅ OneBot 11 protocol compliant  
✅ Automatic command parsing from message content  
✅ Support for private and group chats  
✅ Automatic response sending back to sender  
✅ Error handling and logging  
✅ No authentication required (webhook pattern)  
✅ Extensible command system  

## Next Steps

1. **For Development**: Test with curl commands above
2. **For Production**: 
   - Set up NapCat server
   - Configure firewall rules
   - Add security IP whitelist
   - Test end-to-end with real QQ account

3. **For Enhancement**: Add more commands like:
   - `/hosts` - List all hosts
   - `/logs` - Search logs
   - `/subscribe` - Subscribe to alerts
   - `/help` - List available commands
