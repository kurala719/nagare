package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"nagare/internal/service"
	"nagare/internal/model"
)

// GetSystemLogsCtrl handles GET /log/system
func GetSystemLogsCtrl(c *gin.Context) {
	withTotal, _ := parseOptionalBool(c, "with_total")
	logs, total, err := searchLogs(c, service.LogTypeSystem)
	if err != nil {
		respondError(c, err)
		return
	}
	if withTotal != nil && *withTotal {
		respondSuccess(c, http.StatusOK, gin.H{"items": logs, "total": total})
		return
	}
	respondSuccess(c, http.StatusOK, logs)
}

// GetServiceLogsCtrl handles GET /log/service
func GetServiceLogsCtrl(c *gin.Context) {
	withTotal, _ := parseOptionalBool(c, "with_total")
	logs, total, err := searchLogs(c, service.LogTypeService)
	if err != nil {
		respondError(c, err)
		return
	}
	if withTotal != nil && *withTotal {
		respondSuccess(c, http.StatusOK, gin.H{"items": logs, "total": total})
		return
	}
	respondSuccess(c, http.StatusOK, logs)
}

// ClearLogsCtrl handles DELETE /logs/:type
func ClearLogsCtrl(c *gin.Context) {
	logType := c.Param("type")
	if logType != service.LogTypeSystem && logType != service.LogTypeService {
		respondBadRequest(c, "invalid log type")
		return
	}

	count, err := service.ClearLogsServ(logType)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{
		"message": "logs cleared",
		"count":   count,
	})
}

func searchLogs(c *gin.Context, logType string) ([]service.LogResp, int64, error) {
	severityPtr := parseLogSeverityParam(c.Query("severity"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "200"))
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	filter := model.LogFilter{
		Type:      logType,
		Severity:  severityPtr,
		Query:     normalizeLogQueryParam(c.Query("q")),
		Limit:     limit,
		Offset:    offset,
		SortBy:    c.Query("sort"),
		SortOrder: c.Query("order"),
	}
	logs, err := service.SearchLogsServ(filter)
	if err != nil {
		return nil, 0, err
	}
	total, err := service.CountLogsServ(filter)
	if err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func normalizeLogQueryParam(value string) string {
	switch value {
	case "", "undefined", "null", "NULL", "Undefined", "Null":
		return ""
	default:
		return value
	}
}

func parseLogSeverityParam(value string) *int {
	trimmed := normalizeLogQueryParam(value)
	if trimmed == "" {
		return nil
	}
	if n, err := strconv.Atoi(trimmed); err == nil {
		return &n
	}
	switch strings.ToLower(trimmed) {
	case "warn", "warning":
		v := 1
		return &v
	case "error", "err":
		v := 2
		return &v
	case "info":
		v := 0
		return &v
	default:
		return nil
	}
}
