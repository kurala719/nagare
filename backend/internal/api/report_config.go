package api

import (
	"net/http"

	"nagare/internal/service"

	"github.com/gin-gonic/gin"
)

// GetReportConfigCtrl handles GET /reports/config
func GetReportConfigCtrl(c *gin.Context) {
	config, err := service.GetReportConfigServ()
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, config)
}

// UpdateReportConfigCtrl handles PUT /reports/config
func UpdateReportConfigCtrl(c *gin.Context) {
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	if err := service.UpdateReportConfigServ(updates); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "report configuration updated")
}
