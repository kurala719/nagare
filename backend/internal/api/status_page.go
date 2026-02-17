package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"nagare/internal/service"
)

// GetPublicStatusSummaryCtrl handles the public status page request
func GetPublicStatusSummaryCtrl(c *gin.Context) {
	summary, err := service.GetPublicStatusSummaryServ()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    summary,
	})
}
