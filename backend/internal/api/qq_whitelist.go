package api

import (
	"net/http"
	"strconv"

	"nagare/internal/service"

	"github.com/gin-gonic/gin"
)

// GetQQWhitelistCtrl handles GET /qq-whitelist
func GetQQWhitelistCtrl(c *gin.Context) {
	whitelistType, err := parseOptionalInt(c, "type")
	if err != nil {
		respondBadRequest(c, "invalid type")
		return
	}
	enabled, err := parseOptionalInt(c, "enabled")
	if err != nil {
		respondBadRequest(c, "invalid enabled")
		return
	}

	limit := 100
	if l, err := parseOptionalInt(c, "limit"); err == nil && l != nil {
		limit = *l
	}
	offset := 0
	if o, err := parseOptionalInt(c, "offset"); err == nil && o != nil {
		offset = *o
	}

	whitelist, err := service.ListQQWhitelistServ(whitelistType, enabled, limit, offset)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, whitelist)
}

// AddQQWhitelistCtrl handles POST /qq-whitelist
func AddQQWhitelistCtrl(c *gin.Context) {
	var req service.QQWhitelistReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	result, err := service.AddQQWhitelistServ(req)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, http.StatusCreated, result)
}

// UpdateQQWhitelistCtrl handles PUT /qq-whitelist/:id
func UpdateQQWhitelistCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid whitelist ID")
		return
	}

	var req service.QQWhitelistReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	if err := service.UpdateQQWhitelistServ(uint(id), req); err != nil {
		respondError(c, err)
		return
	}

	respondSuccessMessage(c, http.StatusOK, "whitelist updated")
}

// DeleteQQWhitelistCtrl handles DELETE /qq-whitelist/:id
func DeleteQQWhitelistCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid whitelist ID")
		return
	}

	if err := service.DeleteQQWhitelistServ(uint(id)); err != nil {
		respondError(c, err)
		return
	}

	respondSuccessMessage(c, http.StatusOK, "whitelist deleted")
}
