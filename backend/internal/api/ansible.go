package api

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"nagare/internal/service"
)

// ============= Inventory Controller =============

func GetAnsibleInventoryCtrl(c *gin.Context) {
	inventory, err := service.GetAnsibleDynamicInventory()
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, inventory)
}

// ============= Playbook Controllers =============

func CreatePlaybookCtrl(c *gin.Context) {
	// Log raw body for debugging 400 errors
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore for ShouldBindJSON
	
	service.LogService("debug", "create playbook raw body", map[string]interface{}{
		"body": string(bodyBytes),
	}, nil, c.ClientIP())

	var req service.PlaybookReq
	if err := c.ShouldBindJSON(&req); err != nil {
		service.LogService("warn", "create playbook binding error", map[string]interface{}{
			"error": err.Error(),
			"body_preview": string(bodyBytes),
		}, nil, c.ClientIP())
		
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid playbook data: " + err.Error(),
		})
		return
	}
	pb, err := service.CreatePlaybookServ(req)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusCreated, pb)
}

func ListPlaybooksCtrl(c *gin.Context) {
	q := c.Query("q")
	// Frontend sends limit/offset, for now we just return all matching
	pbs, err := service.ListPlaybooksServ(q)
	if err != nil {
		respondError(c, err)
		return
	}
	
	respondSuccess(c, http.StatusOK, gin.H{
		"items": pbs,
		"total": len(pbs),
	})
}

func GetPlaybookCtrl(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	pb, err := service.GetPlaybookServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, pb)
}

func UpdatePlaybookCtrl(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	// Log raw body for debugging 400 errors
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore for ShouldBindJSON

	var req service.PlaybookReq
	if err := c.ShouldBindJSON(&req); err != nil {
		service.LogService("warn", "update playbook binding error", map[string]interface{}{
			"error": err.Error(), 
			"id": id,
			"body_preview": string(bodyBytes),
		}, nil, c.ClientIP())
		
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid playbook data: " + err.Error(),
		})
		return
	}
	if err := service.UpdatePlaybookServ(uint(id), req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "playbook updated")
}

func DeletePlaybookCtrl(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := service.DeletePlaybookServ(uint(id)); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "playbook deleted")
}

// ============= Job Controllers =============

func RunPlaybookCtrl(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		HostFilter string `json:"host_filter"`
	}
	_ = c.ShouldBindJSON(&req)

	var userID *uint
	if val, ok := c.Get("uid"); ok {
		if id, ok := val.(uint); ok {
			userID = &id
		}
	}

	jobID, err := service.RunAnsiblePlaybookServ(uint(id), req.HostFilter, userID)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusAccepted, gin.H{"job_id": jobID})
}

func GetAnsibleJobCtrl(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	job, err := service.GetAnsibleJobServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, job)
}

func ListAnsibleJobsCtrl(c *gin.Context) {
	playbookID, _ := strconv.Atoi(c.DefaultQuery("playbook_id", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	jobs, err := service.ListAnsibleJobsServ(uint(playbookID), limit)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, jobs)
}

func RecommendPlaybookCtrl(c *gin.Context) {
	var req struct {
		Context string `json:"context"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	content, err := service.RecommendAnsiblePlaybookServ(req.Context)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, gin.H{"content": content})
}
