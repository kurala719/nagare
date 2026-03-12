package api

import (
	"net/http"
	"strconv"

	"nagare/internal/model"
	"nagare/internal/service"

	"github.com/gin-gonic/gin"
)

// SubmitPasswordResetApplicationCtrl handles POST /users/password-reset-applications
func SubmitPasswordResetApplicationCtrl(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	if err := service.SubmitPasswordResetApplicationServ(req.Username, req.Password); err != nil {
		respondError(c, err)
		return
	}

	respondSuccessMessage(c, http.StatusOK, "request submitted for audit")
}

// ListPasswordResetApplicationsCtrl handles GET /users/password-reset-applications
func ListPasswordResetApplicationsCtrl(c *gin.Context) {
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
	filter := model.RegisterApplicationFilter{
		Query:     c.Query("q"),
		Status:    status,
		Limit:     limit,
		Offset:    offset,
		SortBy:    c.Query("sort"),
		SortOrder: c.Query("order"),
	}
	apps, err := service.ListPasswordResetApplicationsServ(filter)
	if err != nil {
		respondError(c, err)
		return
	}
	if withTotal != nil && *withTotal {
		total, err := service.CountPasswordResetApplicationsServ(filter)
		if err != nil {
			respondError(c, err)
			return
		}
		respondSuccess(c, http.StatusOK, gin.H{"items": apps, "total": total})
		return
	}
	respondSuccess(c, http.StatusOK, apps)
}

// ApprovePasswordResetApplicationCtrl handles POST /users/password-reset-applications/:id/approvals
func ApprovePasswordResetApplicationCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid application ID")
		return
	}
	username, ok := c.Get("username")
	if !ok {
		respondError(c, model.ErrUnauthorized)
		return
	}
	if err := service.ApprovePasswordResetApplicationServ(uint(id), username.(string)); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "application approved")
}

// RejectPasswordResetApplicationCtrl handles POST /users/password-reset-applications/:id/rejections
func RejectPasswordResetApplicationCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid application ID")
		return
	}
	username, ok := c.Get("username")
	if !ok {
		respondError(c, model.ErrUnauthorized)
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	if err := service.RejectPasswordResetApplicationServ(uint(id), username.(string), req.Reason); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "application rejected")
}
