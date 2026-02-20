package service

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"nagare/internal/database"
	"nagare/internal/model"
	"nagare/internal/repository"
)

const (
	LogTypeSystem  = "system"
	LogTypeService = "service"
)

// LogResp represents a log response
type LogResp struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	Severity  int    `json:"severity"`
	Message   string `json:"message"`
	Context   string `json:"context"`
	UserID    *uint  `json:"user_id"`
	IP        string `json:"ip"`
	CreatedAt string `json:"created_at"`
}

func LogSystem(severity, message string, context map[string]interface{}, userID *uint, ip string) {
	logEntry(LogTypeSystem, logSeverityFromString(severity), message, context, userID, ip)
}

func LogService(severity, message string, context map[string]interface{}, userID *uint, ip string) {
	logEntry(LogTypeService, logSeverityFromString(severity), message, context, userID, ip)
}

func SearchLogsServ(filter model.LogFilter) ([]LogResp, error) {
	entries, err := repository.SearchLogsDAO(filter)
	if err != nil {
		return nil, err
	}
	result := make([]LogResp, 0, len(entries))
	for _, e := range entries {
		result = append(result, LogResp{
			ID:        int(e.ID),
			Type:      e.Type,
			Severity:  e.Severity,
			Message:   e.Message,
			Context:   e.Context,
			UserID:    e.UserID,
			IP:        e.IP,
			CreatedAt: e.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return result, nil
}

// CountLogsServ returns total count for logs by filter
func CountLogsServ(filter model.LogFilter) (int64, error) {
	return repository.CountLogsDAO(filter)
}

func ClearLogsServ(logType string) (int64, error) {
	return repository.ClearLogsDAO(logType)
}

func logEntry(logType string, severity int, message string, context map[string]interface{}, userID *uint, ip string) {
	ctx := ""
	if len(context) > 0 {
		if b, err := json.Marshal(context); err == nil {
			ctx = string(b)
		}
	}
	entry := model.LogEntry{
		Type:     logType,
		Severity: severity,
		Message:  message,
		Context:  ctx,
		UserID:   userID,
		IP:       ip,
	}
	if database.DB == nil {
		log.Printf("log write skipped (db not ready): %s", message)
		return
	}
	sqlDB, err := database.DB.DB()
	if err != nil {
		log.Printf("log write skipped (db not ready): %s", message)
		return
	}
	if err := sqlDB.Ping(); err != nil {
		log.Printf("log write skipped (db not ready): %s", message)
		return
	}
	if err := repository.AddLogDAO(entry); err != nil {
		log.Printf("log write failed: %v", err)
		return
	}
	if !shouldSkipLogTrigger(context) {
		ExecuteTriggersForLog(entry)
	}
}

func shouldSkipLogTrigger(context map[string]interface{}) bool {
	if context == nil {
		return false
	}
	if v, ok := context["skip_trigger"]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func logSeverityFromString(value string) int {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0
	}
	if n, err := strconv.Atoi(trimmed); err == nil {
		return n
	}
	switch strings.ToLower(trimmed) {
	case "warn", "warning":
		return 1
	case "error", "err":
		return 2
	case "info":
		return 0
	default:
		return 0
	}
}
