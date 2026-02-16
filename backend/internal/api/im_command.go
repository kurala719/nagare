package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"nagare/internal/service"
)

type imCommandRequest struct {
	Message   string `json:"message"`
	MediaType string `json:"media_type"`
	Target    string `json:"target"`
}

// IMCommandCtrl handles POST /im/command
func IMCommandCtrl(c *gin.Context) {
	var req imCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	result, err := service.HandleIMCommand(req.Message)
	if err != nil {
		respondError(c, err)
		return
	}
	if err := service.SendIMReply(req.MediaType, req.Target, result.Reply); err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, result)
}
