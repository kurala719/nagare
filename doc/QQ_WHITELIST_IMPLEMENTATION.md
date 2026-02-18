# QQ Whitelist Access Control Implementation

## Overview
This document describes the complete implementation of QQ whitelist access control for the Nagare project. The system allows administrators to restrict which QQ users and groups can execute commands and receive alert notifications.

## Architecture

### Database Schema
New table: `qq_whitelists` (auto-created via GORM AutoMigrate)

```sql
CREATE TABLE qq_whitelists (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  qq_identifier VARCHAR(255) NOT NULL,
  type INT NOT NULL,
  nickname VARCHAR(255),
  can_command INT DEFAULT 1,
  can_receive INT DEFAULT 1,
  enabled INT DEFAULT 1,
  comment TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP,
  UNIQUE KEY idx_qq_type (qq_identifier, type),
  INDEX idx_qq_type_enabled (type, enabled)
);
```

**Fields:**
- `qq_identifier`: QQ user ID or group ID (string)
- `type`: 0 = QQ user, 1 = QQ group
- `nickname`: Display name for the user/group
- `can_command`: 1 = allowed to execute commands, 0 = blocked
- `can_receive`: 1 = allowed to receive alerts, 0 = blocked
- `enabled`: 1 = whitelist entry active, 0 = disabled
- `comment`: Admin notes

### Authorization Model
**Deny-by-default policy**: Unknown users/groups are rejected unless explicitly added to whitelist.

```
User seeks to:
  ├── Execute command → CheckQQWhitelistForCommand(qqID, isGroup)
  │   ├── Check if (qqID, 0) in whitelist with can_command=1 and enabled=1
  │   └── Check if (qqID, 1) in whitelist with can_command=1 and enabled=1
  │   └── If found and enabled → Allow; Otherwise → Reject
  │
  └── Receive alert → CheckQQWhitelistForAlert(qqID, isGroup)
      ├── Check if (qqID, 0) in whitelist with can_receive=1 and enabled=1
      └── Check if (qqID, 1) in whitelist with can_receive=1 and enabled=1
      └── If found and enabled → Allow; Otherwise → Reject
```

## Implementation Details

### 1. Data Model Layer
**File:** `backend/internal/model/entities.go`

```go
type QQWhitelist struct {
	gorm.Model
	QQIdentifier string `gorm:"index;uniqueIndex:idx_qq_type"`
	Type         int    `gorm:"uniqueIndex:idx_qq_type"`
	Nickname     string
	CanCommand   int `gorm:"default:1"`
	CanReceive   int `gorm:"default:1"`
	Enabled      int `gorm:"default:1"`
	Comment      string
}
```

### 2. Repository Layer (DAO)
**File:** `backend/internal/repository/qq_whitelist.go`

Implements CRUD operations:

```go
// Add new whitelist entry
AddQQWhitelistDAO(whitelist model.QQWhitelist) (model.QQWhitelist, error)

// Retrieve entry by ID and type
GetQQWhitelistDAO(qqID string, whitelistType int) (model.QQWhitelist, error)

// Update existing entry
UpdateQQWhitelistDAO(id uint, whitelist model.QQWhitelist) error

// Delete entry
DeleteQQWhitelistDAO(id uint) error

// List with optional filters
ListQQWhitelistDAO(whitelistType *int, enabled *int, limit int, offset int) ([]model.QQWhitelist, error)

// Count entries
CountQQWhitelistDAO(whitelistType *int, enabled *int) (int64, error)
```

### 3. Service Layer (Business Logic)
**File:** `backend/internal/service/qq_whitelist.go`

Implements DTOs and service functions:

```go
type QQWhitelistReq struct {
	QQIdentifier string
	Type         int
	Nickname     string
	CanCommand   int
	CanReceive   int
	Enabled      int
	Comment      string
}

type QQWhitelistResp struct {
	ID           uint
	QQIdentifier string
	Type         string // "user" or "group"
	Nickname     string
	CanCommand   bool
	CanReceive   bool
	Enabled      bool
	Comment      string
}

// Service functions (CRUD)
AddQQWhitelistServ(req QQWhitelistReq) (QQWhitelistResp, error)
GetQQWhitelistServ(id uint) (QQWhitelistResp, error)
UpdateQQWhitelistServ(id uint, req QQWhitelistReq) error
DeleteQQWhitelistServ(id uint) error
ListQQWhitelistServ(whitelistType *int, enabled *int, limit int, offset int) ([]QQWhitelistResp, error)
CountQQWhitelistServ(whitelistType *int, enabled *int) (int64, error)

// Authorization checks
CheckQQWhitelistForCommand(qqID string, isGroup bool) bool
CheckQQWhitelistForAlert(qqID string, isGroup bool) bool
```

### 4. API Layer (REST Controllers)
**File:** `backend/internal/api/qq_whitelist.go`

Endpoints (requires privilege level 2):

```
GET    /api/v1/qq-whitelist?type=0&enabled=1&limit=100&offset=0
POST   /api/v1/qq-whitelist
PUT    /api/v1/qq-whitelist/:id
DELETE /api/v1/qq-whitelist/:id
```

**Query Parameters:**
- `type`: (optional) Filter by 0=user or 1=group
- `enabled`: (optional) Filter by 0=disabled or 1=enabled
- `limit`: (optional) Results per page, default 100
- `offset`: (optional) Pagination offset, default 0

### 5. Integration Points

#### A. Command Execution
**File:** `backend/internal/api/media.go` - HandleQQMessageCtrl()

```go
// Extract QQ ID and determine if group
qqID, isGroup := parseQQTarget(event.GroupID, event.UserID)

// Check whitelist before processing command
if !CheckQQWhitelistForCommand(qqID, isGroup) {
    return 401-like response with "unauthorized" message
}

// Process command if authorized
```

**Target Format:**
- Single user: `user_123456789`
- Group: `group_123456789`

#### B. Alert Delivery
**File:** `backend/internal/service/action.go` - sendMediaMessage()

```go
// Check if media type is QQ
if lowerType == "qq" || lowerType == "qrobot" {
    qqID, isGroup := parseQQTarget(media.Target)
    
    // Check alert whitelist
    if !CheckQQWhitelistForAlert(qqID, isGroup) {
        LogService("info", "send message skipped (QQ alert whitelist)", ...)
        return nil  // Skip alert delivery
    }
}

// Send alert if authorized
```

### 6. Database Migration
**File:** `backend/cmd/server/main.go` - initDBTables()

Added `&model.QQWhitelist{}` to `database.DB.AutoMigrate()` call to auto-create table on startup.

## API Usage Examples

### Add QQ user to whitelist (allow commands and alerts)
```bash
curl -X POST http://localhost:8080/api/v1/qq-whitelist \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "qq_identifier": "123456789",
    "type": 0,
    "nickname": "System Admin",
    "can_command": 1,
    "can_receive": 1,
    "enabled": 1,
    "comment": "Main administrator"
  }'
```

### Add QQ group to whitelist (alerts only)
```bash
curl -X POST http://localhost:8080/api/v1/qq-whitelist \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "qq_identifier": "987654321",
    "type": 1,
    "nickname": "Operations Team",
    "can_command": 0,
    "can_receive": 1,
    "enabled": 1,
    "comment": "Alert recipients only"
  }'
```

### List all active user whitelists
```bash
curl -X GET "http://localhost:8080/api/v1/qq-whitelist?type=0&enabled=1" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json"
```

### Disable a whitelist entry (without deleting)
```bash
curl -X PUT http://localhost:8080/api/v1/qq-whitelist/1 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": 0
  }'
```

### Delete a whitelist entry
```bash
curl -X DELETE http://localhost:8080/api/v1/qq-whitelist/1 \
  -H "Authorization: Bearer <token>"
```

## Testing

### Test Command Authorization
1. Create a QQ user whitelist entry with `can_command=1`
2. Send a command via QQ: `/cmd help`
3. Verify command is executed
4. Disable the whitelist entry
5. Try command again
6. Verify 401 authorization error

### Test Alert Delivery
1. Create a QQ user whitelist entry with `can_receive=1`
2. Trigger an alert that would send to QQ
3. Verify alert is delivered
4. Set `can_receive=0`
5. Trigger another alert
6. Verify alert is skipped silently (logged as "send message skipped")

### Test Group Authorization
1. Create a QQ group whitelist entry
2. Send group command: Verify works if `can_command=1`, blocked if `can_command=0`
3. Send group alert: Verify delivered if `can_receive=1`, skipped if `can_receive=0`

## Security Considerations

1. **Deny-by-default**: Unknown QQ users/groups cannot execute commands or receive alerts
2. **Atomic permissions**: Each whitelist entry has independent `can_command` and `can_receive` flags
3. **Enable/disable**: Admins can temporarily disable access without deleting entries
4. **Audit trail**: Comments field for admin notes about who/why entries were added
5. **User vs Group**: Separate whitelist types allow different policies for individuals vs teams

## Future Enhancements

1. **Time-based access**: Add `allowed_from` and `allowed_until` timestamps
2. **Command restrictions**: Only allow specific commands for certain users
3. **Quota limits**: Limit alerts per user/group per time period
4. **Audit logging**: Track all authorization decisions
5. **Frontend UI**: Add whitelist management panel to Nagare dashboard
6. **Multi-provider**: Support whitelist for other IM platforms (Slack, DingTalk, etc.)

## Files Modified

| File | Changes |
|------|---------|
| `backend/internal/model/entities.go` | Added QQWhitelist struct |
| `backend/internal/repository/qq_whitelist.go` | NEW - DAO layer for whitelist CRUD |
| `backend/internal/service/qq_whitelist.go` | NEW - Service layer with business logic |
| `backend/internal/service/im_command.go` | Added CheckQQWhitelistForCommand, CheckQQWhitelistForAlert, getQQWhitelist |
| `backend/internal/service/action.go` | Added whitelist check in sendMediaMessage for alert delivery |
| `backend/internal/api/media.go` | Added whitelist check in HandleQQMessageCtrl for commands |
| `backend/internal/api/qq_whitelist.go` | NEW - REST controllers for whitelist management |
| `backend/cmd/server/router/router.go` | Added setupQQWhitelistRoutes function and route registration |
| `backend/cmd/server/main.go` | Added QQWhitelist to AutoMigrate in initDBTables |

## Deployment Steps

1. **Update code**: Pull the latest changes
2. **Rebuild**: `cd backend && go build -o nagare-web-server ./cmd/server`
3. **Restart service**: Restart the Nagare backend service
   - Database migration runs automatically on startup
   - `qq_whitelists` table is created if it doesn't exist
4. **Configure whitelist**: Add authorized users/groups via API or database
5. **Verify**: Test command and alert delivery with whitelisted users

## Troubleshooting

### Commands rejected immediately
- Check if user/group is in whitelist: `GET /api/v1/qq-whitelist?type=0`
- Verify `enabled=1` for the entry
- Verify `can_command=1` for the entry
- Check logs for "send message skipped (QQ alert whitelist)" messages

### Alerts not delivered
- Verify whitelist entry exists with `can_receive=1`
- Check if `enabled=0` (disabled entry won't accept alerts)
- Review logs for "send message skipped" entries
- Ensure media target format is `user_QQID` or `group_QQID`

### Database table creation failed
- Check database permissions
- Verify database is running
- Check `database.DB` is initialized in `gormdb.go`
- Review GORM migration logs in server startup output
