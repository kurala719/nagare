package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"nagare/internal/model"
	"nagare/internal/service"

	"github.com/gin-gonic/gin"
)

// GetAllAlertsCtrl handles GET /alerts
func GetAllAlertsCtrl(c *gin.Context) {
	alerts, err := service.GetAllAlertsServ()
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, alerts)
}

// AlertWebhookCtrl handles POST /alerts/webhook
func AlertWebhookCtrl(c *gin.Context) {
	// High-level entry log
	service.LogService("info", "webhook entry", map[string]interface{}{"method": c.Request.Method, "url": c.Request.URL.String()}, nil, "")

	// Ensure we always respond to prevent Zabbix timeout
	defer func() {
		if r := recover(); r != nil {
			service.LogService("error", "webhook panic recovered", map[string]interface{}{"panic": r}, nil, "")
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "internal server error"})
		}
	}()

	service.LogService("debug", "webhook request received", map[string]interface{}{
		"remote_addr": c.ClientIP(),
		"user_agent":  c.GetHeader("User-Agent"),
	}, nil, "")

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		service.LogService("error", "webhook invalid JSON", map[string]interface{}{"error": err.Error()}, nil, "")
		respondBadRequest(c, err.Error())
		return
	}

	service.LogService("debug", "webhook payload received", map[string]interface{}{
		"payload_keys": getMapKeys(payload),
	}, nil, "")

	eventToken := strings.TrimSpace(c.GetHeader("X-Alarm-Token"))
	if eventToken == "" {
		eventToken = strings.TrimSpace(c.GetHeader("X-Monitor-Token"))
	}
	if eventToken == "" {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if strings.HasPrefix(authHeader, "Bearer ") {
			eventToken = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		}
	}
	if eventToken == "" {
		eventToken = strings.TrimSpace(c.Query("token"))
	}
	if eventToken == "" {
		eventToken = strings.TrimSpace(c.Query("event_token"))
	}

	// Check payload for token if not found in headers/query
	if eventToken == "" {
		if t := payloadString(payload, "token", "event_token", "auth_token"); t != "" {
			eventToken = t
		}
	}

	var alarmID uint
	if eventToken == "" {
		service.LogService("warn", "webhook missing token", map[string]interface{}{
			"headers": c.Request.Header,
		}, nil, "")
		respondError(c, model.ErrUnauthorized)
		return
	}

	service.LogService("debug", "webhook token found", map[string]interface{}{
		"token_prefix": eventToken[:min(8, len(eventToken))],
	}, nil, "")

	if alarm, err := service.GetAlarmByEventTokenServ(eventToken); err == nil {
		alarmID = alarm.ID
		service.LogService("debug", "webhook alarm identified", map[string]interface{}{"alarm_id": alarmID}, nil, "")
	} else if errors.Is(err, model.ErrNotFound) || errors.Is(err, model.ErrUnauthorized) {
		// Try to find if it's a monitor token
		if monitor, mErr := service.GetMonitorByEventTokenServ(eventToken); mErr == nil {
			service.LogService("debug", "webhook monitor identified from token", map[string]interface{}{"monitor_id": monitor.ID}, nil, "")
			// Try to find a matching alarm for this monitor by name
			if alarms, aErr := service.SearchAlarmsServ(model.AlarmFilter{Query: monitor.Name}); aErr == nil && len(alarms) > 0 {
				for _, a := range alarms {
					if strings.EqualFold(a.Name, monitor.Name) {
						alarmID = uint(a.ID)
						break
					}
				}
			}
			// If still 0, use the monitor ID. The repository join supports COALESCE(alarms.name, monitors.name).
			if alarmID == 0 {
				alarmID = monitor.ID
			}
		} else {
			if err := service.ValidateMonitorEventTokenServ(eventToken); err != nil {
				service.LogService("warn", "webhook token validation failed", map[string]interface{}{"error": err.Error()}, nil, "")
				respondError(c, err)
				return
			}
			service.LogService("debug", "webhook monitor token validated", nil, nil, "")
		}
	} else {
		service.LogService("error", "webhook token lookup failed", map[string]interface{}{"error": err.Error()}, nil, "")
		respondError(c, err)
		return
	}

	message := payloadString(payload, "message", "msg", "title", "subject", "alert", "alert_message")
	// If message looks like unexpanded Zabbix macros or is empty, try fallback fields
	if strings.TrimSpace(message) == "" || strings.Contains(message, "{ALERT.") || strings.Contains(message, "{EVENT.") {
		if name := payloadString(payload, "name", "problem", "trigger_name", "trigger", "event_name"); name != "" {
			message = name
		}
	}

	if strings.TrimSpace(message) == "" {
		service.LogService("warn", "webhook missing message", map[string]interface{}{"payload_keys": getMapKeys(payload)}, nil, "")
		respondBadRequest(c, "missing message")
		return
	}

	severity := payloadInt(payload, "severity", "level", "event_nseverity", "trigger_severity")
	hostID := payloadUint(payload, "host_id", "hostid")
	itemID := payloadUint(payload, "item_id", "itemid")
	triggerID := payloadUint(payload, "trigger_id", "triggerid", "event_objectid")
	hostName := payloadString(payload, "host", "hostname", "host_name")
	itemName := payloadString(payload, "item", "itemname", "item_name")
	comment := payloadString(payload, "comment", "detail", "details")
	if alarmID == 0 {
		alarmID = payloadUint(payload, "alarm_id", "alarmid")
	}

	// Normalize severity if it's from Zabbix
	if alarmID > 0 {
		if alarm, err := service.GetAlarmByIDServ(alarmID); err == nil {
			if alarm.Type == 1 { // Zabbix
				switch severity {
				case 5:
					severity = 4 // Disaster -> Critical
				case 4:
					severity = 3 // High -> High
				case 3:
					severity = 2 // Average -> Medium
				case 2:
					severity = 2 // Warning -> Medium
				case 1:
					severity = 1 // Information -> Low
				case 0:
					severity = 0 // Not classified -> Info
				}
			}
		}
	}

	if severity == 0 {
		severity = payloadSeverity(payload, "severity", "level", "priority", "event_severity", "trigger_severity")
	}
	if strings.TrimSpace(comment) == "" {
		comment = buildAlertContext(payload)
	}

	statusStr := payloadString(payload, "status", "state", "event_value")
	status := 0 // Default to open
	if statusStr != "" {
		switch strings.ToLower(statusStr) {
		case "resolved", "ok", "0":
			status = 2 // Resolved
		case "acknowledged", "ack":
			status = 1
		case "problem", "open", "1":
			status = 0
		}
	}

	service.LogService("info", "webhook creating alert", map[string]interface{}{
		"alarm_id": alarmID,
		"severity": severity,
		"host_id":  hostID,
		"host_name": hostName,
		"status":   status,
		"message":  message[:min(100, len(message))],
	}, nil, "")

	req := service.AlertReq{
		Message:   message,
		Severity:  severity,
		Status:    status,
		AlarmID:   alarmID,
		HostID:    hostID,
		ItemID:    itemID,
		HostName:  hostName,
		ItemName:  itemName,
		Comment:   comment,
	}
	if triggerID > 0 {
		req.TriggerID = &triggerID
	}

	if err := service.AddAlertServ(req); err != nil {
		service.LogService("error", "webhook failed to create alert", map[string]interface{}{"error": err.Error()}, nil, "")
		respondError(c, err)
		return
	}

	service.LogService("info", "webhook alert created successfully", map[string]interface{}{"alarm_id": alarmID}, nil, "")
	respondSuccessMessage(c, http.StatusAccepted, "alert accepted")
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func payloadString(payload map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := payload[key]; ok {
			switch v := value.(type) {
			case string:
				return v
			case []byte:
				return string(v)
			}
		}
	}
	return ""
}

func payloadInt(payload map[string]interface{}, keys ...string) int {
	for _, key := range keys {
		if value, ok := payload[key]; ok {
			switch v := value.(type) {
			case float64:
				return int(v)
			case int:
				return v
			case int64:
				return int(v)
			case string:
				// If it's a macro placeholder, return 0 instead of trying to parse
				if strings.Contains(v, "{") && strings.Contains(v, "}") {
					return 0
				}
				if parsed, err := strconv.Atoi(v); err == nil {
					return parsed
				}
			}
		}
	}
	return 0
}

func payloadUint(payload map[string]interface{}, keys ...string) uint {
	value := payloadInt(payload, keys...)
	if value < 0 {
		return 0
	}
	return uint(value)
}

func payloadSeverity(payload map[string]interface{}, keys ...string) int {
	for _, key := range keys {
		if value, ok := payload[key]; ok {
			switch v := value.(type) {
			case string:
				return parseSeverityLabel(v)
			case float64:
				return int(v)
			}
		}
	}
	return 0
}

func parseSeverityLabel(value string) int {
	lower := strings.ToLower(strings.TrimSpace(value))
	switch lower {
	case "disaster", "critical":
		return 4
	case "high":
		return 3
	case "average", "warning", "warn", "medium":
		return 2
	case "low":
		return 1
	case "information", "info", "normal":
		return 0
	default:
		return 0
	}
}

func buildAlertContext(payload map[string]interface{}) string {
	trigger := payloadString(payload, "trigger", "trigger_name", "name")
	host := payloadString(payload, "host", "hostname")
	eventID := payloadString(payload, "event_id", "eventid")
	eventTime := payloadString(payload, "event_time", "clock", "time")
	if trigger == "" && host == "" && eventID == "" && eventTime == "" {
		return ""
	}
	return fmt.Sprintf("trigger=%s host=%s event_id=%s event_time=%s", trigger, host, eventID, eventTime)
}

// SearchAlertsCtrl handles GET /alert/search
func SearchAlertsCtrl(c *gin.Context) {
	severity, err := parseOptionalInt(c, "severity")
	if err != nil {
		respondBadRequest(c, "invalid severity")
		return
	}
	status, err := parseOptionalInt(c, "status")
	if err != nil {
		respondBadRequest(c, "invalid status")
		return
	}
	withTotal, _ := parseOptionalBool(c, "with_total")
	limit := 100
	if l, err := parseOptionalInt(c, "limit"); err == nil && l != nil {
		limit = *l
	}
	offset := 0
	if o, err := parseOptionalInt(c, "offset"); err == nil && o != nil {
		offset = *o
	}
	hostID, err := parseOptionalInt(c, "host_id")
	if err != nil {
		respondBadRequest(c, "invalid host_id")
		return
	}
	itemID, err := parseOptionalInt(c, "item_id")
	if err != nil {
		respondBadRequest(c, "invalid item_id")
		return
	}
	alarmID, err := parseOptionalInt(c, "alarm_id")
	if err != nil {
		respondBadRequest(c, "invalid alarm_id")
		return
	}
	filter := model.AlertFilter{
		Query:     c.Query("q"),
		Severity:  severity,
		Status:    status,
		AlarmID:   alarmID,
		HostID:    hostID,
		ItemID:    itemID,
		Limit:     limit,
		Offset:    offset,
		SortBy:    c.Query("sort"),
		SortOrder: c.Query("order"),
	}
	alerts, err := service.SearchAlertsServ(filter)
	if err != nil {
		respondError(c, err)
		return
	}
	if withTotal != nil && *withTotal {
		total, err := service.CountAlertsServ(filter)
		if err != nil {
			respondError(c, err)
			return
		}
		respondSuccess(c, http.StatusOK, gin.H{"items": alerts, "total": total})
		return
	}
	respondSuccess(c, http.StatusOK, alerts)
}

// GetAlertByIDCtrl handles GET /alerts/:id
func GetAlertByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid alert ID")
		return
	}

	alert, err := service.GetAlertByIDServ(id)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, alert)
}

// AddAlertCtrl handles POST /alerts
func AddAlertCtrl(c *gin.Context) {
	var req service.AlertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	if err := service.AddAlertServ(req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusCreated, "alert created")
}

// DeleteAlertByIDCtrl handles DELETE /alerts/:id
func DeleteAlertByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid alert ID")
		return
	}

	if err := service.DeleteAlertServ(id); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "alert deleted")
}

// UpdateAlertCtrl handles PUT /alerts/:id
func UpdateAlertCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid alert ID")
		return
	}

	var req service.AlertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	if err := service.UpdateAlertServ(id, req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "alert updated")
}

// GenerateTestAlertsCtrl handles POST /api/v1/alerts/generate-test
func GenerateTestAlertsCtrl(c *gin.Context) {
	count := 5
	if c, ok := c.GetQuery("count"); ok {
		if parsedCount, err := strconv.Atoi(c); err == nil {
			count = parsedCount
		}
	}

	if count <= 0 || count > 100 {
		count = 5
	}

	if err := service.GenerateTestAlerts(count); err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, http.StatusCreated, gin.H{
		"message": "Test alerts generated",
		"count":   count,
	})
}

// GetAlertScoreCtrl handles GET /api/v1/alerts/score
func GetAlertScoreCtrl(c *gin.Context) {
	score, err := service.CalculateAlertScore()
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{
		"score": score,
	})
}
