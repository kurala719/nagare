# Code Structure Reference (Updated)

## Command Processing Pipeline

```go
// 1. HandleQQMessageCtrl() receives OneBot 11 event
// Location: internal/api/media.go

type OneBotMessageEvent struct {
    PostType    string  // "message"
    MessageType string  // "private" or "group"
    UserID      int64   // Recipient user ID
    GroupID     int64   // Recipient group ID (for groups)
    Message     string  // Command message (e.g., "/status")
}

// 2. Controller routes message to command handler
result, err := service.HandleIMCommand(message)

// 3. HandleIMCommand() processes the command
// Location: internal/service/im_command.go

func HandleIMCommand(message string) (IMCommandResult, error) {
    trimmed := strings.TrimSpace(message)
    lower := strings.ToLower(trimmed)
    
    // Check for /status command
    if strings.HasPrefix(lower, "/status") || lower == "status" {
        return handleStatus()
    }
    
    // Check for /get_alert command
    if strings.HasPrefix(lower, "/get_alert") {
        return handleGetAlerts(trimmed)
    }
    
    // Check for /chat command
    if strings.HasPrefix(lower, "/chat") {
        content := strings.TrimSpace(trimmed[5:])
        return handleChatCommand(content)
    }
    
    return IMCommandResult{Reply: "Unsupported command..."}, nil
}

// 4. Command handlers execute business logic

// Status Handler
func handleStatus() IMCommandResult {
    score, err := GetHealthScoreServ()  // Get health metrics
    reply := fmt.Sprintf("Health Score: %d ...", score.Score)
    return IMCommandResult{Reply: reply}, nil
}

// Get Alerts Handler
func handleGetAlerts() IMCommandResult {
    status := 0  // Active only
    limit := 10
    filter := model.AlertFilter{Status: &status, Limit: limit}
    alerts, err := SearchAlertsServ(filter)  // Query database
    
    var reply strings.Builder
    for _, alert := range alerts {
        reply.WriteString(fmt.Sprintf("[%d] %s\n", alert.ID, alert.Message))
    }
    return IMCommandResult{Reply: reply.String()}, nil
}

// Chat Handler
func handleChatCommand(content string) IMCommandResult {
    providers, err := GetAllProvidersServ()  // Get LLM providers
    chatReq := ChatReq{
        ProviderID: uint(providers[0].ID),
        Content:    content,
        Privileges: 1,
    }
    resp, err := SendChatServ(chatReq)  // Send to LLM
    return IMCommandResult{Reply: resp.Content}, nil
}

// 5. Send response back to user
err = service.SendIMReply("qq", replyTarget, result.Reply)
// Location: internal/repository/media/media.go (or similar in infrastructure)
// Uses media.Service to send through QQ provider

// 6. Return success to OneBot 11
resp := OneBotMessageResponse{
    Status:  "ok",
    Retcode: 0,
}
c.JSON(http.StatusOK, resp)
```

## Data Flow Diagram

```
┌──────────────────────────────────────────────────────────────────┐
│                    QQ User sends message                         │
│                     (e.g., "/status")                            │
└──────────────────┬───────────────────────────────────────────────┘
                   │
                   ▼
┌──────────────────────────────────────────────────────────────────┐
│                  NapCat (OneBot 11)                              │
│          Captures message from QQ                                │
└──────────────────┬───────────────────────────────────────────────┘
                   │
                   │ HTTP POST
                   │ OneBot 11 Event JSON
                   ▼
┌──────────────────────────────────────────────────────────────────┐
│              Nagare Server                                        │
│  POST /api/v1/media/qq/message                                    │
│                                                                   │
│  1. HandleQQMessageCtrl() (internal/api)                         │
│     ├─ Parse OneBotMessageEvent JSON                             │
│     ├─ Extract message & target (user/group)                     │
│     └─ Call HandleIMCommand()                                    │
│                                                                   │
│  2. HandleIMCommand() (internal/service)                         │
│     ├─ Check command type (/status, /get_alert, /chat)          │
│     ├─ Route to appropriate handler                              │
│     └─ Execute business logic                                    │
│                                                                   │
│  3. Handler Functions                                            │
│     ├─ handleStatus() → GetHealthScoreServ()                     │
│     ├─ handleGetAlerts() → SearchAlertsServ()                    │
│     └─ handleChatCommand() → SendChatServ()                      │
│                                                                   │
│  4. SendIMReply()                                                │
│     └─ Media.Service.SendMessage() → QQ Provider                 │
│                                                                   │
│  5. Return 200 OK to OneBot 11                                   │
└──────────────────┬───────────────────────────────────────────────┘
                   │
                   │ HTTP Response
                   │ {"status":"ok","retcode":0}
                   ▼
┌──────────────────────────────────────────────────────────────────┐
│                  NapCat (OneBot 11)                              │
│          Acknowledges receipt                                    │
└──────────────────┬───────────────────────────────────────────────┘
                   │
                   │ At same time, response sent to QQ via
                   │ NapCat API: /send_msg
                   │
                   ▼
┌──────────────────────────────────────────────────────────────────┐
│                    QQ User receives response                      │
│              (e.g., "Health Score: 85 ...")                      │
└──────────────────────────────────────────────────────────────────┘
```

## File Locations (Updated)

```
nagare-v0.21/ (Backend)
├── cmd/
│   └── server/
│       └── router/
│           └── router.go                    [MODIFIED]
│               └── setupMediaRoutes()
│                   └── media.POST("/qq/message")
│
└── internal/
    ├── api/ (was presentation)
    │   └── media.go                       [MODIFIED]
    │       ├── HandleQQMessageCtrl()
    │       ├── OneBotMessageEvent
    │       ├── OneBotSender
    │       ├── OneBotMessageResponse
    │       └── (existing controllers)
    │
    ├── service/ (was application)
    │   ├── im_command.go                  [MODIFIED]
    │   │   ├── HandleIMCommand()
    │   │   ├── handleGetAlerts()
    │   │   ├── handleChatCommand()
    │   │   └── HandleIMCommandWithContext()
    │   │
    │   └── (other services)
    │       ├── GetHealthScoreServ()
    │       ├── SearchAlertsServ()
    │       ├── SendChatServ()
    │       └── GetAllProvidersServ()
    │
    ├── repository/ (was infrastructure)
    │   └── (DAOs and external adapters)
    │
    └── model/ (was domain)
        └── (Entities)
```

## Command Handler Pattern

```go
// All command handlers follow this pattern:

// 1. Check for command
if strings.HasPrefix(lower, "/command_name") {
    return handleCommandName(trimmed)
}

// 2. Implement handler
func handleCommandName(input string) (IMCommandResult, error) {
    // Parse input parameters if needed
    // Call service layer functions
    // Format response
    // Return result
    return IMCommandResult{Reply: formattedResponse}, nil
}
```

## Route Configuration

```go
// In cmd/server/router/router.go setupMediaRoutes()

func setupMediaRoutes(rg RouteGroup) {
    media := rg.Group("/media")
    
    // Existing routes
    media.GET("/", api.SearchMediaCtrl).Use(api.PrivilegesMiddleware(1))
    media.GET("/:id", api.GetMediaByIDCtrl).Use(api.PrivilegesMiddleware(1))
    media.POST("/", api.AddMediaCtrl).Use(api.PrivilegesMiddleware(2))
    media.PUT("/:id", api.UpdateMediaCtrl).Use(api.PrivilegesMiddleware(2))
    media.DELETE("/:id", api.DeleteMediaByIDCtrl).Use(api.PrivilegesMiddleware(2))
    
    // New webhook route (NO AUTHENTICATION)
    media.POST("/qq/message", api.HandleQQMessageCtrl)
}
```