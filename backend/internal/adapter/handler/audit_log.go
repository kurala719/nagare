package handler

import (
	"net/http"
	"strconv"

	"nagare/internal/core/service"

	"github.com/gin-gonic/gin"
)

// SearchAuditLogsCtrl returns audit logs
func SearchAuditLogsCtrl(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	query := c.Query("q")
	if query == "" {
		query = c.Query("query")
	}

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	logs, total, err := service.SearchAuditLogsServ(limit, offset, query)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{
		"items": logs,
		"total": total,
	})
}
