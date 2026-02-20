package handler

import (
	"net/http"
	"strconv"

	"nagare/internal/core/service"

	"github.com/gin-gonic/gin"
)

// QueueStatsCtrl handles GET /api/v1/queue/stats
func QueueStatsCtrl(c *gin.Context) {
	stats, err := service.GetQueueStats()
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, stats)
}

// PullHostsAsyncCtrl handles POST /api/v1/monitors/:id/hosts/pull-async
func PullHostsAsyncCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}

	taskID, err := service.PullHostsFromMonitorAsyncServ(uint(id), true)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, http.StatusAccepted, gin.H{
		"message":    "Host pull task queued",
		"task_id":    taskID,
		"monitor_id": id,
	})
}

// PullItemsAsyncCtrl handles POST /api/v1/monitors/:m_id/hosts/:h_id/items/pull-async
func PullItemsAsyncCtrl(c *gin.Context) {
	midStr := c.Param("m_id")
	hidStr := c.Param("h_id")

	mid, err := strconv.Atoi(midStr)
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}

	hid, err := strconv.Atoi(hidStr)
	if err != nil {
		respondBadRequest(c, "invalid host ID")
		return
	}

	taskID, err := service.PullItemsFromMonitorAsyncServ(uint(mid), uint(hid), true)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, http.StatusAccepted, gin.H{
		"message":    "Item pull task queued",
		"task_id":    taskID,
		"monitor_id": mid,
		"host_id":    hid,
	})
}

// GenerateTestAlertsCtrl handles POST /api/v1/alerts/generate-test
func GenerateTestAlertsCtrl(c *gin.Context) {
	count := 5
	if c, ok := c.GetQuery("count"); ok {
		if parsedCount, err := strconv.Atoi(c); err == nil {
			count = parsedCount
		}
	}

	if count <= 0 || count > 100 {
		count = 5
	}

	if err := service.GenerateTestAlerts(count); err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, http.StatusCreated, gin.H{
		"message": "Test alerts generated",
		"count":   count,
	})
}

// GetAlertScoreCtrl handles GET /api/v1/alerts/score
func GetAlertScoreCtrl(c *gin.Context) {
	score, err := service.CalculateAlertScore()
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{
		"score": score,
	})
}
