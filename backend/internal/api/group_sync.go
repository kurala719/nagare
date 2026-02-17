package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"nagare/internal/service"
)

// PullGroupsFromMonitorCtrl handles the request to pull groups from a monitor
func PullGroupsFromMonitorCtrl(c *gin.Context) {
	idStr := c.Param("id")
	mid, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "invalid monitor id"})
		return
	}

	result, err := service.PullGroupsFromMonitorServ(uint(mid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, APIResponse{Success: true, Data: result})
}

// PushGroupToMonitorCtrl handles the request to push a group to a monitor
func PushGroupToMonitorCtrl(c *gin.Context) {
	idStr := c.Param("id")
	mid, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "invalid monitor id"})
		return
	}

	groupIdStr := c.Param("gid")
	gid, err := strconv.ParseUint(groupIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "invalid group id"})
		return
	}

	err = service.PushGroupToMonitorServ(uint(mid), uint(gid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, APIResponse{Success: true, Message: "group pushed successfully"})
}
