package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"nagare/internal/service"
	"nagare/internal/model"
)

// GetAllGroupsCtrl handles GET /groups
func GetAllGroupsCtrl(c *gin.Context) {
	groups, err := service.GetAllGroupsServ()
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, groups)
}

// SearchGroupsCtrl handles GET /groups
func SearchGroupsCtrl(c *gin.Context) {
	status, err := parseOptionalInt(c, "status")
	if err != nil {
		respondBadRequest(c, "invalid status")
		return
	}
	monitorIDInt, err := parseOptionalInt(c, "monitor_id")
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}
	var monitorID *uint
	if monitorIDInt != nil {
		val := uint(*monitorIDInt)
		monitorID = &val
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
	filter := model.GroupFilter{
		Query:     c.Query("q"),
		Status:    status,
		MonitorID: monitorID,
		Limit:     limit,
		Offset:    offset,
		SortBy:    c.Query("sort"),
		SortOrder: c.Query("order"),
	}
	groups, err := service.SearchGroupsServ(filter)
	if err != nil {
		respondError(c, err)
		return
	}
	if withTotal != nil && *withTotal {
		total, err := service.CountGroupsServ(filter)
		if err != nil {
			respondError(c, err)
			return
		}
		respondSuccess(c, http.StatusOK, gin.H{"items": groups, "total": total})
		return
	}
	respondSuccess(c, http.StatusOK, groups)
}

// GetGroupByIDCtrl handles GET /groups/:id
func GetGroupByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid group ID")
		return
	}
	group, err := service.GetGroupByIDServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, group)
}

// GetGroupDetailCtrl handles GET /groups/:id/detail
func GetGroupDetailCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid group ID")
		return
	}
	detail, err := service.GetGroupDetailServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, detail)
}

// AddGroupCtrl handles POST /groups
func AddGroupCtrl(c *gin.Context) {
	var req service.GroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	group, err := service.AddGroupServ(req)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusCreated, group)
}

// UpdateGroupCtrl handles PUT /groups/:id
func UpdateGroupCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid group ID")
		return
	}
	var req service.GroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	if err := service.UpdateGroupServ(uint(id), req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "group updated")
}

// DeleteGroupByIDCtrl handles DELETE /groups/:id
func DeleteGroupByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid group ID")
		return
	}
	pushToMonitor := c.Query("push") == "true"
	if err := service.DeleteGroupByIDServ(uint(id), pushToMonitor); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "group deleted")
}

// PullGroupFromMonitorsCtrl handles POST /groups/:id/pull
func PullGroupFromMonitorsCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid group ID")
		return
	}
	result, err := service.PullGroupFromMonitorsServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, result)
}

// PushGroupToMonitorsCtrl handles POST /groups/:id/push
func PushGroupToMonitorsCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid group ID")
		return
	}
	result, err := service.PushGroupToMonitorsServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, result)
}

// CheckGroupStatusCtrl handles POST /groups/:id/check
func CheckGroupStatusCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid group ID")
		return
	}

	result, err := service.CheckGroupStatusServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, result)
}

// CheckAllGroupsStatusCtrl handles POST /groups/check
func CheckAllGroupsStatusCtrl(c *gin.Context) {
	results := service.CheckAllGroupsStatusServ()
	respondSuccess(c, http.StatusOK, results)
}
