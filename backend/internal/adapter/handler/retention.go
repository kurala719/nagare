package handler

import (
	"net/http"

	"nagare/internal/core/domain"
	"nagare/internal/core/service"

	"github.com/gin-gonic/gin"
)

// GetRetentionPoliciesCtrl handles GET /retention/policies
func GetRetentionPoliciesCtrl(c *gin.Context) {
	policies, err := service.GetRetentionPoliciesServ()
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, policies)
}

// UpdateRetentionPolicyCtrl handles POST /retention/policies
func UpdateRetentionPolicyCtrl(c *gin.Context) {
	var policy domain.RetentionPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	if err := service.UpdateRetentionPolicyServ(policy); err != nil {
		respondError(c, err)
		return
	}

	respondSuccessMessage(c, http.StatusOK, "Retention policy updated successfully")
}

// PerformCleanupCtrl handles POST /retention/cleanup
func PerformCleanupCtrl(c *gin.Context) {
	results, err := service.PerformDataRetentionCleanupServ()
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, results)
}
