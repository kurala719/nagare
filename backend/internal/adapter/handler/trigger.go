package handler

import (
	"net/http"
	"strconv"

	"nagare/internal/core/domain"
	"nagare/internal/core/service"

	"github.com/gin-gonic/gin"
)

// GetAllTriggersCtrl handles GET /trigger
func GetAllTriggersCtrl(c *gin.Context) {
	triggers, err := service.GetAllTriggersServ()
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, triggers)
}

// SearchTriggersCtrl handles GET /trigger/search
func SearchTriggersCtrl(c *gin.Context) {
	status, err := parseOptionalInt(c, "status")
	if err != nil {
		respondBadRequest(c, "invalid status")
		return
	}
	severityMin, err := parseOptionalInt(c, "severity_min")
	if err != nil {
		respondBadRequest(c, "invalid severity_min")
		return
	}
	withTotal, _ := parseOptionalBool(c, "with_total")
	alertID, err := parseOptionalUint(c, "alert_id")
	if err != nil {
		respondBadRequest(c, "invalid alert_id")
		return
	}
	alertMonitorID, err := parseOptionalUint(c, "alert_monitor_id")
	if err != nil {
		respondBadRequest(c, "invalid alert_monitor_id")
		return
	}
	alertGroupID, err := parseOptionalUint(c, "alert_group_id")
	if err != nil {
		respondBadRequest(c, "invalid alert_group_id")
		return
	}
	alertHostID, err := parseOptionalUint(c, "alert_host_id")
	if err != nil {
		respondBadRequest(c, "invalid alert_host_id")
		return
	}
	alertItemID, err := parseOptionalUint(c, "alert_item_id")
	if err != nil {
		respondBadRequest(c, "invalid alert_item_id")
		return
	}
	limit := 100
	if l, err := parseOptionalInt(c, "limit"); err == nil && l != nil {
		limit = *l
	}
	offset := 0
	if o, err := parseOptionalInt(c, "offset"); err == nil && o != nil {
		offset = *o
	}
	var entityPtr *string
	entity := c.Query("entity")
	if entity != "" {
		entityPtr = &entity
	}
	var actionIDPtr *uint
	if actionID, err := strconv.Atoi(c.Query("action_id")); err == nil && actionID > 0 {
		id := uint(actionID)
		actionIDPtr = &id
	}
	filter := domain.TriggerFilter{
		Query:          c.Query("q"),
		Status:         status,
		SeverityMin:    severityMin,
		Entity:         entityPtr,
		ActionID:       actionIDPtr,
		AlertID:        alertID,
		AlertMonitorID: alertMonitorID,
		AlertGroupID:   alertGroupID,
		AlertHostID:    alertHostID,
		AlertItemID:    alertItemID,
		Limit:          limit,
		Offset:         offset,
		SortBy:         c.Query("sort"),
		SortOrder:      c.Query("order"),
	}
	triggers, err := service.SearchTriggersServ(filter)
	if err != nil {
		respondError(c, err)
		return
	}
	if withTotal != nil && *withTotal {
		total, err := service.CountTriggersServ(filter)
		if err != nil {
			respondError(c, err)
			return
		}
		respondSuccess(c, http.StatusOK, gin.H{"items": triggers, "total": total})
		return
	}
	respondSuccess(c, http.StatusOK, triggers)
}

// GetTriggerByIDCtrl handles GET /trigger/:id
func GetTriggerByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid trigger ID")
		return
	}
	trigger, err := service.GetTriggerByIDServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, trigger)
}

// AddTriggerCtrl handles POST /trigger
func AddTriggerCtrl(c *gin.Context) {
	var req service.TriggerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	if req.ActionID == 0 {
		respondBadRequest(c, "action_id is required")
		return
	}
	trigger, err := service.AddTriggerServ(req)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusCreated, trigger)
}

// UpdateTriggerCtrl handles PUT /trigger/:id
func UpdateTriggerCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid trigger ID")
		return
	}
	var req service.TriggerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	if req.ActionID == 0 {
		respondBadRequest(c, "action_id is required")
		return
	}
	if err := service.UpdateTriggerServ(uint(id), req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "trigger updated")
}

// DeleteTriggerByIDCtrl handles DELETE /trigger/:id
func DeleteTriggerByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid trigger ID")
		return
	}
	if err := service.DeleteTriggerByIDServ(uint(id)); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "trigger deleted")
}
