package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"nagare/internal/model"
	"nagare/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// GetAllAlarmsCtrl handles GET /alarms
func GetAllAlarmsCtrl(c *gin.Context) {
	alarms, err := service.GetAllAlarmsServ()
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, alarms)
}

// SearchAlarmsCtrl handles GET /alarms
func SearchAlarmsCtrl(c *gin.Context) {
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
	typeValue, err := parseOptionalInt(c, "type")
	if err != nil {
		respondBadRequest(c, "invalid type")
		return
	}
	filter := model.AlarmFilter{
		Query:     c.Query("q"),
		Type:      typeValue,
		Status:    status,
		Limit:     limit,
		Offset:    offset,
		SortBy:    c.Query("sort"),
		SortOrder: c.Query("order"),
	}
	alarms, err := service.SearchAlarmsServ(filter)
	if err != nil {
		respondError(c, err)
		return
	}
	if withTotal != nil && *withTotal {
		total, err := service.CountAlarmsServ(filter)
		if err != nil {
			respondError(c, err)
			return
		}
		respondSuccess(c, http.StatusOK, gin.H{"items": alarms, "total": total})
		return
	}
	respondSuccess(c, http.StatusOK, alarms)
}

// GetAlarmByIDCtrl handles GET /alarms/:id
func GetAlarmByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid alarm ID")
		return
	}

	alarm, err := service.GetAlarmByIDServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, alarm)
}

// AddAlarmCtrl handles POST /alarms
func AddAlarmCtrl(c *gin.Context) {
	var req service.AlarmReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	alarm, err := service.AddAlarmServ(req)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusCreated, alarm)
}

// DeleteAlarmByIDCtrl handles DELETE /alarms/:id
func DeleteAlarmByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid alarm ID")
		return
	}

	if err := service.DeleteAlarmServByID(id); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "alarm deleted")
}

// UpdateAlarmCtrl handles PUT /alarms/:id
func UpdateAlarmCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid alarm ID")
		return
	}

	var req service.AlarmReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	if err := service.UpdateAlarmServ(id, req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "alarm updated")
}

// LoginAlarmCtrl handles POST /alarms/:id/login
func LoginAlarmCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid alarm ID")
		return
	}

	alarm, err := service.LoginAlarmServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, alarm)
}

// RefreshAlarmEventTokenCtrl handles POST /alarms/:id/event-token/refresh (public)
// Allows the alarm source to refresh its own event token using the current token
func RefreshAlarmEventTokenCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid alarm ID")
		return
	}

	eventToken := strings.TrimSpace(c.GetHeader("X-Alarm-Token"))
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

	if eventToken == "" {
		respondError(c, model.ErrUnauthorized)
		return
	}

	if err := service.ValidateAlarmEventTokenServ(eventToken); err != nil {
		respondError(c, err)
		return
	}

	alarm, err := service.RegenerateAlarmEventTokenServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, alarm)
}

// RegenerateAlarmEventTokenCtrl handles POST /alarms/:id/event-token
func RegenerateAlarmEventTokenCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid alarm ID")
		return
	}

	alarm, err := service.RegenerateAlarmEventTokenServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, alarm)
}

// SetupAlarmMediaTypeCtrl handles POST /alarms/:id/setup-media
func SetupAlarmMediaTypeCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid alarm ID")
		return
	}

	if err := service.SetupAlarmMediaTypeServ(uint(id)); err != nil {
		respondError(c, err)
		return
	}

	// Get webhook URL from config to show in response
	ip := viper.GetString("system.ip_address")
	port := viper.GetInt("system.port")
	if ip == "" {
		ip = "localhost"
	}
	if port == 0 {
		port = 8080
	}
	webhookURL := fmt.Sprintf("http://%s:%d/api/v1/alerts/webhook", ip, port)

	respondSuccess(c, http.StatusOK, gin.H{
		"message":         "Media type setup successfully",
		"webhook_url":     webhookURL,
		"test_url":        webhookURL + "/health",
		"troubleshooting": "If Zabbix cannot reach the webhook, ensure the IP address is accessible from Zabbix server",
		"test_command":    fmt.Sprintf("curl -X GET %s/health", webhookURL),
	})
}
