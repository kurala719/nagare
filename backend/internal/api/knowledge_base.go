package api

import (
	"net/http"
	"strconv"

	"nagare/internal/service"

	"github.com/gin-gonic/gin"
)

// GetAllKnowledgeBaseCtrl handles GET /knowledge-base
func GetAllKnowledgeBaseCtrl(c *gin.Context) {
	q := c.Query("q")
	if q != "" {
		kbs, err := service.SearchKnowledgeBaseServ(q)
		if err != nil {
			respondError(c, err)
			return
		}
		respondSuccess(c, http.StatusOK, kbs)
		return
	}

	kbs, err := service.GetAllKnowledgeBaseServ()
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, kbs)
}

// GetKnowledgeBaseByIDCtrl handles GET /knowledge-base/:id
func GetKnowledgeBaseByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid id")
		return
	}

	kb, err := service.GetKnowledgeBaseByIDServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, kb)
}

// AddKnowledgeBaseCtrl handles POST /knowledge-base
func AddKnowledgeBaseCtrl(c *gin.Context) {
	var req service.KnowledgeBaseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	if err := service.AddKnowledgeBaseServ(req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusCreated, "knowledge base entry added")
}

// UpdateKnowledgeBaseCtrl handles PUT /knowledge-base/:id
func UpdateKnowledgeBaseCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid id")
		return
	}

	var req service.KnowledgeBaseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	if err := service.UpdateKnowledgeBaseServ(uint(id), req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "knowledge base entry updated")
}

// DeleteKnowledgeBaseCtrl handles DELETE /knowledge-base/:id
func DeleteKnowledgeBaseCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid id")
		return
	}

	if err := service.DeleteKnowledgeBaseServ(uint(id)); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "knowledge base entry deleted")
}
