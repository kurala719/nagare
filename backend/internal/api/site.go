package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"nagare/internal/service"
	"nagare/internal/model"
)

// GetAllSitesCtrl handles GET /site
func GetAllSitesCtrl(c *gin.Context) {
	sites, err := service.GetAllSitesServ()
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, sites)
}

// SearchSitesCtrl handles GET /site/search
func SearchSitesCtrl(c *gin.Context) {
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
	filter := model.SiteFilter{
		Query:     c.Query("q"),
		Status:    status,
		Limit:     limit,
		Offset:    offset,
		SortBy:    c.Query("sort"),
		SortOrder: c.Query("order"),
	}
	sites, err := service.SearchSitesServ(filter)
	if err != nil {
		respondError(c, err)
		return
	}
	if withTotal != nil && *withTotal {
		total, err := service.CountSitesServ(filter)
		if err != nil {
			respondError(c, err)
			return
		}
		respondSuccess(c, http.StatusOK, gin.H{"items": sites, "total": total})
		return
	}
	respondSuccess(c, http.StatusOK, sites)
}

// GetSiteByIDCtrl handles GET /site/:id
func GetSiteByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid site ID")
		return
	}
	site, err := service.GetSiteByIDServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, site)
}

// GetSiteDetailCtrl handles GET /site/:id/detail
func GetSiteDetailCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid site ID")
		return
	}
	detail, err := service.GetSiteDetailServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, detail)
}

// AddSiteCtrl handles POST /site
func AddSiteCtrl(c *gin.Context) {
	var req service.SiteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	site, err := service.AddSiteServ(req)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusCreated, site)
}

// UpdateSiteCtrl handles PUT /site/:id
func UpdateSiteCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid site ID")
		return
	}
	var req service.SiteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	if err := service.UpdateSiteServ(uint(id), req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "site updated")
}

// DeleteSiteByIDCtrl handles DELETE /site/:id
func DeleteSiteByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid site ID")
		return
	}
	if err := service.DeleteSiteByIDServ(uint(id)); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "site deleted")
}

// PullSiteFromMonitorsCtrl handles GET /site/:id/pull
func PullSiteFromMonitorsCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid site ID")
		return
	}
	result, err := service.PullSiteFromMonitorsServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, result)
}

// PushSiteToMonitorsCtrl handles GET /site/:id/push
func PushSiteToMonitorsCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid site ID")
		return
	}
	result, err := service.PushSiteToMonitorsServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, result)
}

// CheckSiteStatusCtrl handles POST /sites/:id/check
func CheckSiteStatusCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid site ID")
		return
	}

	result, err := service.CheckSiteStatusServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, result)
}

// CheckAllSitesStatusCtrl handles POST /sites/check
func CheckAllSitesStatusCtrl(c *gin.Context) {
	results := service.CheckAllSitesStatusServ()
	respondSuccess(c, http.StatusOK, results)
}
