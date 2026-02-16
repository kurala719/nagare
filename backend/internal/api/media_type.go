package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"nagare/internal/service"
	"nagare/internal/model"
)

// GetAllMediaTypesCtrl handles GET /media-type
func GetAllMediaTypesCtrl(c *gin.Context) {
	mediaTypes, err := service.GetAllMediaTypesServ()
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, mediaTypes)
}

// SearchMediaTypesCtrl handles GET /media-type/search
func SearchMediaTypesCtrl(c *gin.Context) {
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
	filter := model.MediaTypeFilter{
		Query:     c.Query("q"),
		Status:    status,
		Limit:     limit,
		Offset:    offset,
		SortBy:    c.Query("sort"),
		SortOrder: c.Query("order"),
	}
	mediaTypes, err := service.SearchMediaTypesServ(filter)
	if err != nil {
		respondError(c, err)
		return
	}
	if withTotal != nil && *withTotal {
		total, err := service.CountMediaTypesServ(filter)
		if err != nil {
			respondError(c, err)
			return
		}
		respondSuccess(c, http.StatusOK, gin.H{"items": mediaTypes, "total": total})
		return
	}
	respondSuccess(c, http.StatusOK, mediaTypes)
}

// GetMediaTypeByIDCtrl handles GET /media-type/:id
func GetMediaTypeByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid media type ID")
		return
	}
	mediaType, err := service.GetMediaTypeByIDServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, mediaType)
}

// AddMediaTypeCtrl handles POST /media-type
func AddMediaTypeCtrl(c *gin.Context) {
	var req service.MediaTypeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	mediaType, err := service.AddMediaTypeServ(req)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusCreated, mediaType)
}

// UpdateMediaTypeCtrl handles PUT /media-type/:id
func UpdateMediaTypeCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid media type ID")
		return
	}
	var req service.MediaTypeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	if err := service.UpdateMediaTypeServ(uint(id), req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "media type updated")
}

// DeleteMediaTypeByIDCtrl handles DELETE /media-type/:id
func DeleteMediaTypeByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid media type ID")
		return
	}
	if err := service.DeleteMediaTypeByIDServ(uint(id)); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "media type deleted")
}
