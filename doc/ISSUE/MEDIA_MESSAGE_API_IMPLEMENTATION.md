# Media Message API Implementation

## Overview
This document describes the new Media Message API that allows message platforms (like QQ via OneBot 11) to send commands to the server and receive responses.

## Architecture

### Components Updated

#### 1. **Service Layer** (`backend/internal/service/im_command.go`)
- Enhanced `HandleIMCommand()` to process command messages
- Added support for three command types:
  - `/status` - Get system health status
  - `/get_alert` - Retrieve the latest active alerts
  - `/chat <message>` - Send a message to an LLM provider for processing
- Added `IMCommandContext` struct for tracking media source information
- Added helper functions:
  - `handleGetAlerts()` - Retrieves active alerts and formats them
  - `handleChatCommand()` - Sends chat messages to configured LLM providers

#### 2. **API Layer** (`backend/internal/api/media.go`)
- Added `HandleQQMessageCtrl()` - Main handler for incoming QQ messages from OneBot 11
- Added data structures for OneBot 11 message event format:
  - `OneBotMessageEvent` - Represents incoming message from QQ
  - `OneBotSender` - Sender information
  - `OneBotMessageResponse` - Response format for OneBot 11 compliance
  - `OneBotMessageResponseData` - Response data structure

#### 3. **Router Configuration** (`backend/cmd/server/router/router.go`)
- Added new route: `POST /api/v1/media/qq/message`
- Route accepts OneBot 11 HTTP event pushes without authentication
- Automatically processes commands and sends responses back through QQ media provider

## API Endpoint

### POST `/api/v1/media/qq/message`
Receives incoming messages from QQ (via OneBot 11/NapCat) and processes them as commands.

#### Request Format (OneBot 11 Event)
```json
{
  "post_type": "message",
  "message_type": "private|group",
  "user_id": 123456789,
  "group_id": 987654321,        // Only for group messages
  "message": "/status",          // The actual command message
  "raw_message": "/status",
  "message_id": 1,
  "time": 1234567890,
  "sender": {
    "user_id": 123456789,
    "nickname": "User",
    "role": "member"
  }
}
```

#### Response Format (OneBot 11 Compliant)
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

## Supported Commands

### 1. `/status`
Retrieves the current system health score and resource status.

**Response Example:**
```
Health Score: 85 (monitors 5/5, hosts 3/3, items 12/12)
```

### 2. `/get_alert`
Retrieves up to 10 active alerts from the system.

**Response Example:**
```
Active Alerts (3):
[1] Database connection timeout (Severity: 2)
[2] High CPU usage (Severity: 1)
[3] Disk space low (Severity: 2)
```

### 3. `/chat <message>`
Sends a message to the configured LLM provider and returns the AI response.

**Example:**
```
/chat What is the current network status?
```

**Response Example:**
```
Based on the available monitoring data, all systems are performing optimally with 99.2% uptime...
```

## Configuration Steps for QQ Integration

### 1. Set up NapCat (OneBot 11)
- Install NapCat on a machine with QQ access
- Configure it to connect to your Nagare server's HTTP event push endpoint

### 2. Configure NapCat HTTP Event Push
In NapCat configuration, set:
```
http_event_endpoint = http://your-server:8080/api/v1/media/qq/message
```

### 3. Add QQ Media in Nagare
- Go to Settings → Media
- Add new media with type "QQ"
- Configure the NapCat server URL (e.g., `http://127.0.0.1:3000`)

### 4. Configure LLM Provider (for /chat command)
- Go to Settings → Providers
- Add an LLM provider (OpenAI, Claude, etc.)
- Set it as default to use with QQ commands

## Implementation Details

### Message Flow
1. QQ user sends message starting with `/` to QQ bot
2. NapCat captures the message and sends HTTP POST to `/api/v1/media/qq/message`
3. Server receives OneBot 11 event
4. Message is parsed and passed to `HandleIMCommand()`
5. Command is executed (status, alert, or chat query)
6. Response is formatted
7. Server sends response back to QQ user via `SendIMReply()`
8. Response is returned to NapCat with status "ok"

### Command Processing Flow
```
OneBot 11 Event →
  ├─ Parse message and extract target (user/group)
  ├─ Pass message to HandleIMCommand()
  │  ├─ Check for /status → Get health score
  │  ├─ Check for /get_alert → Query active alerts
  │  └─ Check for /chat → Send to LLM provider
  ├─ Format response
  ├─ Send reply via QQ media provider
  └─ Return 200 OK to OneBot 11
```

## Security Considerations

⚠️ **Important**: The `/api/v1/media/qq/message` endpoint is intentionally left unauthenticated because:
- It receives webhooks from external systems (OneBot 11/NapCat)
- External systems cannot provide OAuth/JWT tokens in real-time event pushes
- The endpoint is specific and only accepts OneBot 11 event format

**Security measures implemented:**
- Event format validation (checks for valid OneBot 11 message event structure)
- Message content validation
- Target validation (requires valid user_id or group_id)

**Recommended additional security:**
- Configure a firewall rule to allow only the NapCat server IP
- Use HTTPS for the event push endpoint
- Consider adding a shared secret in the event payload

## Error Handling

All errors during command processing are:
1. Logged to the system log
2. Converted to user-friendly messages
3. Sent back to the QQ user
4. Still returns 200 OK to OneBot 11 (webhook convention)

## Testing

### Test Commands in QQ
```
# Get system status
/status

# Get active alerts
/get_alert

# Chat with AI
/chat What are the current system alerts?
/chat Explain the network topology
```

### curl Command Example
```bash
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
```

## Future Enhancements

Potential commands to add:
- `/hosts` - List all monitored hosts
- `/items` - List items by host
- `/alerts [status]` - Filter alerts by status
- `/help` - Show available commands
- `/subscribe_alerts` - Subscribe to alert notifications
- `/config` - View system configuration
- `/logs [filter]` - Search system logs
- `/action <name>` - Execute actions

## Files Modified

1. [backend/internal/service/im_command.go](backend/internal/service/im_command.go) - Enhanced command handling
2. [backend/internal/api/media.go](backend/internal/api/media.go) - New controller for QQ messages
3. [backend/cmd/server/router/router.go](backend/cmd/server/router/router.go) - New route configuration

## References

- [OneBot 11 Documentation](https://onebot.dev/)
- [NapCat (OneBot Implementation for QQ)](https://napneko.github.io/)
- [OneBot Network Communication](https://napneko.github.io/onebot/network)
- [OneBot API Reference](https://napneko.github.io/api/4.16.0)
