package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"nagare/internal/service"
	"nagare/internal/model"
)

// GetAllMonitorsCtrl handles GET /monitors
func GetAllMonitorsCtrl(c *gin.Context) {
	monitors, err := service.GetAllMonitorsServ()
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, monitors)
}

// SearchMonitorsCtrl handles GET /monitor/search
func SearchMonitorsCtrl(c *gin.Context) {
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
	filter := model.MonitorFilter{
		Query:     c.Query("q"),
		Type:      parseOptionalString(c, "type"),
		Status:    status,
		Limit:     limit,
		Offset:    offset,
		SortBy:    c.Query("sort"),
		SortOrder: c.Query("order"),
	}
	monitors, err := service.SearchMonitorsServ(filter)
	if err != nil {
		respondError(c, err)
		return
	}
	if withTotal != nil && *withTotal {
		total, err := service.CountMonitorsServ(filter)
		if err != nil {
			respondError(c, err)
			return
		}
		respondSuccess(c, http.StatusOK, gin.H{"items": monitors, "total": total})
		return
	}
	respondSuccess(c, http.StatusOK, monitors)
}

// GetMonitorByIDCtrl handles GET /monitors/:id
func GetMonitorByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}

	monitor, err := service.GetMonitorByIDServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, monitor)
}

// AddMonitorCtrl handles POST /monitors
func AddMonitorCtrl(c *gin.Context) {
	var req service.MonitorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	monitor, err := service.AddMonitorServ(req)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusCreated, monitor)
}

// DeleteMonitorByIDCtrl handles DELETE /monitors/:id
func DeleteMonitorByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}

	if id == 1 {
		respondBadRequest(c, "cannot delete internal monitor")
		return
	}

	if err := service.DeleteMonitorServByID(id); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "monitor deleted")
}

// UpdateMonitorCtrl handles PUT /monitors/:id
func UpdateMonitorCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}

	if id == 1 {
		respondBadRequest(c, "cannot update internal monitor")
		return
	}

	var req service.MonitorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	if err := service.UpdateMonitorServ(id, req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "monitor updated")
}

// LoginMonitorCtrl handles POST /monitors/:id/login
func LoginMonitorCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}

	monitor, err := service.LoginMonitorServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, monitor)
}

// RefreshMonitorEventTokenCtrl handles POST /monitors/:id/event-token/refresh (public)
// Allows the monitor script to refresh its own event token using the current token
func RefreshMonitorEventTokenCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}

	// Get event token from header, Authorization bearer, or query params
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

	if eventToken == "" {
		respondError(c, model.ErrUnauthorized)
		return
	}

	// Validate the event token
	if err := service.ValidateMonitorEventTokenServ(eventToken); err != nil {
		respondError(c, err)
		return
	}

	// Regenerate the token
	monitor, err := service.RegenerateMonitorEventTokenServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, monitor)
}

// RegenerateMonitorEventTokenCtrl handles POST /monitors/:id/event-token
func RegenerateMonitorEventTokenCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}

	monitor, err := service.RegenerateMonitorEventTokenServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, monitor)
}

// CheckMonitorStatusCtrl handles POST /monitors/:id/check
func CheckMonitorStatusCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}

	result, err := service.TestMonitorStatusServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, result)
}

// CheckAllMonitorsStatusCtrl handles POST /monitors/check
func CheckAllMonitorsStatusCtrl(c *gin.Context) {
	results := service.CheckAllMonitorsStatusServ()
	respondSuccess(c, http.StatusOK, results)
}

// PullGroupsFromMonitorCtrl handles POST /monitors/:id/groups/pull
func PullGroupsFromMonitorCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}

	result, err := service.PullGroupsFromMonitorServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, result)
}
