package handler

import (
	"net/http"
	"strconv"

	"nagare/internal/core/service"

	"github.com/gin-gonic/gin"
)

// GetNetworkMetricsCtrl handles GET /system/metrics
func GetNetworkMetricsCtrl(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "200"))
	query := c.Query("q")

	metrics, err := service.GetNetworkMetricsServ(query, limit)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, metrics)
}

// GetNetworkStatusHistoryCtrl handles GET /system/health/history
func GetNetworkStatusHistoryCtrl(c *gin.Context) {
	from, err := parseOptionalUnixTime(c, "from")
	if err != nil {
		respondBadRequest(c, "invalid from timestamp")
		return
	}
	to, err := parseOptionalUnixTime(c, "to")
	if err != nil {
		respondBadRequest(c, "invalid to timestamp")
		return
	}
	limit := 500
	if l, err := parseOptionalInt(c, "limit"); err == nil && l != nil {
		limit = *l
	}
	items, err := service.GetNetworkStatusHistoryServ(from, to, limit)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, items)
}
