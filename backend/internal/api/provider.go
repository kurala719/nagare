package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"nagare/internal/service"
	"nagare/internal/model"
)

// GetAllProvidersCtrl handles GET /providers
func GetAllProvidersCtrl(c *gin.Context) {
	providers, err := service.GetAllProvidersServ()
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, providers)
}

// SearchProvidersCtrl handles GET /provider/search
func SearchProvidersCtrl(c *gin.Context) {
	status, err := parseOptionalInt(c, "status")
	if err != nil {
		respondBadRequest(c, "invalid status")
		return
	}
	typeVal, err := parseOptionalInt(c, "type")
	if err != nil {
		respondBadRequest(c, "invalid type")
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
	filter := model.ProviderFilter{
		Query:     c.Query("q"),
		Type:      typeVal,
		Status:    status,
		Limit:     limit,
		Offset:    offset,
		SortBy:    c.Query("sort"),
		SortOrder: c.Query("order"),
	}
	providers, err := service.SearchProvidersServ(filter)
	if err != nil {
		respondError(c, err)
		return
	}
	if withTotal != nil && *withTotal {
		total, err := service.CountProvidersServ(filter)
		if err != nil {
			respondError(c, err)
			return
		}
		respondSuccess(c, http.StatusOK, gin.H{"items": providers, "total": total})
		return
	}
	respondSuccess(c, http.StatusOK, providers)
}

// GetProviderByIDCtrl handles GET /providers/:id
func GetProviderByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid provider ID")
		return
	}

	provider, err := service.GetProviderByIDServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, provider)
}

// AddProviderCtrl handles POST /providers
func AddProviderCtrl(c *gin.Context) {
	var req service.ProviderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	if err := service.AddProviderServ(req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusCreated, "provider created")
}

// DeleteProviderByIDCtrl handles DELETE /providers/:id
func DeleteProviderByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid provider ID")
		return
	}

	if err := service.DeleteProviderByIDServ(uint(id)); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "provider deleted")
}

// UpdateProviderCtrl handles PUT /providers/:id
func UpdateProviderCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid provider ID")
		return
	}

	var req service.ProviderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	if err := service.UpdateProviderServ(uint(id), req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "provider updated")
}

// CheckProviderStatusCtrl handles POST /providers/:id/check
func CheckProviderStatusCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid provider ID")
		return
	}

	result, err := service.CheckProviderStatusServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, result)
}

// CheckAllProvidersStatusCtrl handles POST /providers/check
func CheckAllProvidersStatusCtrl(c *gin.Context) {
	results := service.CheckAllProvidersStatusServ()
	respondSuccess(c, http.StatusOK, results)
}

// FetchProviderModelsCtrl handles POST /providers/:id/fetch-models
func FetchProviderModelsCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid provider ID")
		return
	}

	models, err := service.FetchProviderModelsServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, models)
}

// FetchModelsDirectCtrl handles POST /providers/fetch-models-direct
func FetchModelsDirectCtrl(c *gin.Context) {
	var req service.ProviderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	models, err := service.FetchModelsDirectServ(req)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, models)
}
