package handler

import (
	"net/http"
	"strconv"

	"nagare/internal/core/domain"
	"nagare/internal/core/service"

	"github.com/gin-gonic/gin"
)

// ListRegisterApplicationsCtrl handles GET /user/register-application/search
func ListRegisterApplicationsCtrl(c *gin.Context) {
	status, err := parseOptionalInt(c, "status")
	if err != nil {
		respondBadRequest(c, "invalid status")
		return
	}
	withTotal, _ := parseOptionalBool(c, "with_total")
	limit := 100
	if l, err := parseOptionalInt(c, "limit"); err == nil && l != nil {
		limit = *l
	}
	offset := 0
	if o, err := parseOptionalInt(c, "offset"); err == nil && o != nil {
		offset = *o
	}
	filter := domain.RegisterApplicationFilter{
		Query:     c.Query("q"),
		Status:    status,
		Limit:     limit,
		Offset:    offset,
		SortBy:    c.Query("sort"),
		SortOrder: c.Query("order"),
	}
	apps, err := service.ListRegisterApplicationsServ(filter)
	if err != nil {
		respondError(c, err)
		return
	}
	if withTotal != nil && *withTotal {
		total, err := service.CountRegisterApplicationsServ(filter)
		if err != nil {
			respondError(c, err)
			return
		}
		respondSuccess(c, http.StatusOK, gin.H{"items": apps, "total": total})
		return
	}
	respondSuccess(c, http.StatusOK, apps)
}

// ApproveRegisterApplicationCtrl handles PUT /user/register-application/:id/approve
func ApproveRegisterApplicationCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid application ID")
		return
	}
	username, ok := c.Get("username")
	if !ok {
		respondError(c, domain.ErrUnauthorized)
		return
	}
	if err := service.ApproveRegisterApplicationServ(uint(id), username.(string)); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "application approved")
}

// RejectRegisterApplicationCtrl handles PUT /user/register-application/:id/reject
func RejectRegisterApplicationCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid application ID")
		return
	}
	username, ok := c.Get("username")
	if !ok {
		respondError(c, domain.ErrUnauthorized)
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	if err := service.RejectRegisterApplicationServ(uint(id), username.(string), req.Reason); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "application rejected")
}
