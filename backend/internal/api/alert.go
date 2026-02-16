package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"nagare/internal/service"
	"nagare/internal/model"
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
	eventToken := strings.TrimSpace(c.GetHeader("X-Monitor-Token"))
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
	if err := service.ValidateMonitorEventTokenServ(eventToken); err != nil {
		respondError(c, err)
		return
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	message := payloadString(payload, "message", "msg", "title", "subject", "alert")
	if strings.TrimSpace(message) == "" {
		respondBadRequest(c, "missing message")
		return
	}
	severity := payloadInt(payload, "severity", "level")
	hostID := payloadUint(payload, "host_id", "hostid")
	itemID := payloadUint(payload, "item_id", "itemid")
	comment := payloadString(payload, "comment", "detail", "details")
	if severity == 0 {
		severity = payloadSeverity(payload, "severity", "level", "priority")
	}
	if strings.TrimSpace(comment) == "" {
		comment = buildAlertContext(payload)
	}

	req := service.AlertReq{
		Message:  message,
		Severity: severity,
		HostID:   hostID,
		ItemID:   itemID,
		Comment:  comment,
	}

	if err := service.AddAlertServ(req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusAccepted, "alert accepted")
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
	case "disaster", "critical", "high":
		return 3
	case "average", "warning", "warn", "medium":
		return 2
	case "information", "info", "low", "normal":
		return 1
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
	filter := model.AlertFilter{
		Query:     c.Query("q"),
		Severity:  severity,
		Status:    status,
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
