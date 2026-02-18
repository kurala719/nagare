# Code Structure Reference

## Command Processing Pipeline

```go
// 1. HandleQQMessageCtrl() receives OneBot 11 event
// Location: backend/internal/api/media.go

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
// Location: backend/internal/service/im_command.go

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
// Location: backend/internal/repository/media/media.go (or similar in infrastructure)
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

## Backend File Structure

```
backend/
├── cmd/
│   └── server/
│       └── router/
│           └── router.go                    [Route Definitions]
│               └── setupMediaRoutes()
│                   └── media.POST("/qq/message")
│
└── internal/
    ├── api/ (Presentation Layer)
    │   ├── media.go                       [Controllers]
    │   ├── webssh.go                      [WebSSH WebSocket Handler]
    │   ├── knowledge_base.go              [Knowledge Base Controllers]
    │   ├── report.go                      [Report Management Controllers]
    │   ├── report_config.go               [Report Config Controllers]
    │   └── (existing controllers)
    │
    ├── service/ (Business Logic Layer)
    │   ├── im_command.go                  [IM Logic]
    │   ├── knowledge_base.go              [RAG and KB Logic]
    │   ├── report.go                      [PDF Generation Logic]
    │   ├── cron.go                        [Task Scheduling]
    │   └── (other services)
    │       ├── GetHealthScoreServ()
    │       ├── SearchAlertsServ()
    │       ├── SendChatServ()
    │       └── GetAllProvidersServ()
    │
    ├── repository/ (Data Access Layer)
    │   ├── knowledge_base.go              [KB DAOs]
    │   ├── report.go                      [Report DAOs]
    │   └── (DAOs and external adapters)
    │
    ├── model/ (Domain Model)
    │   └── entities.go                    [Core Entities including KB and Report]
    │
    └── utils/
        ├── charts.go                      [Chart Generation Utility]
        └── crypto.go                      [AES Encryption Utility]
```

## Frontend File Structure

```
frontend/
├── src/
│   ├── api/          [API Client Modules]
│   ├── components/   [Reusable Vue Components]
│   ├── views/        [Page Components (mapped to routes)]
│   ├── router/       [Vue Router Configuration]
│   ├── utils/        [Helper Functions]
│   └── App.vue       [Root Component]
├── package.json      [Dependencies and Scripts]
└── vite.config.js    [Vite Configuration]
```
