package api

import (
	"net/http"

	"nagare/internal/service"

	"github.com/gin-gonic/gin"
)

// GetHealthScoreCtrl handles GET /system/health
func GetHealthScoreCtrl(c *gin.Context) {
	score, err := service.GetHealthScoreServ()
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, score)
}

// WebhookHealthCtrl handles GET /alerts/webhook/health - for testing webhook connectivity
func WebhookHealthCtrl(c *gin.Context) {
	service.LogService("info", "webhook health check hit", map[string]interface{}{
		"remote_addr": c.ClientIP(),
		"user_agent":  c.GetHeader("User-Agent"),
	}, nil, "")
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Nagare webhook endpoint is reachable",
		"service": "nagare-webhook",
		"version": "1.0",
	})
}
