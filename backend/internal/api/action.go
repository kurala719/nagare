package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"nagare/internal/service"
	"nagare/internal/model"
)

// GetAllActionsCtrl handles GET /action
func GetAllActionsCtrl(c *gin.Context) {
	actions, err := service.GetAllActionsServ()
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, actions)
}

// SearchActionsCtrl handles GET /action/search
func SearchActionsCtrl(c *gin.Context) {
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
	filter := model.ActionFilter{
		Query:     c.Query("q"),
		Status:    status,
		Limit:     limit,
		Offset:    offset,
		SortBy:    c.Query("sort"),
		SortOrder: c.Query("order"),
	}
	actions, err := service.SearchActionsServ(filter)
	if err != nil {
		respondError(c, err)
		return
	}
	if withTotal != nil && *withTotal {
		total, err := service.CountActionsServ(filter)
		if err != nil {
			respondError(c, err)
			return
		}
		respondSuccess(c, http.StatusOK, gin.H{"items": actions, "total": total})
		return
	}
	respondSuccess(c, http.StatusOK, actions)
}

// GetActionByIDCtrl handles GET /action/:id
func GetActionByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid action ID")
		return
	}
	action, err := service.GetActionByIDServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, action)
}

// AddActionCtrl handles POST /action
func AddActionCtrl(c *gin.Context) {
	var req service.ActionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	action, err := service.AddActionServ(req)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusCreated, action)
}

// UpdateActionCtrl handles PUT /action/:id
func UpdateActionCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid action ID")
		return
	}
	var req service.ActionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	if err := service.UpdateActionServ(uint(id), req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "action updated")
}

// DeleteActionByIDCtrl handles DELETE /action/:id
func DeleteActionByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid action ID")
		return
	}
	if err := service.DeleteActionByIDServ(uint(id)); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "action deleted")
}
