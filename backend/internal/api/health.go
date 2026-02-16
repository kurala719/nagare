package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"nagare/internal/service"
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
