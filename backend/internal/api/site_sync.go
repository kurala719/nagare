package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"nagare/internal/service"
)

// PullSitesFromMonitorCtrl handles the request to pull sites from a monitor
func PullSitesFromMonitorCtrl(c *gin.Context) {
	idStr := c.Param("id")
	mid, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "invalid monitor id"})
		return
	}

	result, err := service.PullSitesFromMonitorServ(uint(mid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, APIResponse{Success: true, Data: result})
}

// PushSiteToMonitorCtrl handles the request to push a site to a monitor
func PushSiteToMonitorCtrl(c *gin.Context) {
	idStr := c.Param("id")
	mid, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "invalid monitor id"})
		return
	}

	siteIdStr := c.Param("sid")
	sid, err := strconv.ParseUint(siteIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "invalid site id"})
		return
	}

	err = service.PushSiteToMonitorServ(uint(mid), uint(sid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, APIResponse{Success: true, Message: "site pushed successfully"})
}
