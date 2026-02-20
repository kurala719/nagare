package handler

import (
	"net/http"
	"strconv"

	"nagare/internal/core/domain"
	"nagare/internal/core/service"

	"github.com/gin-gonic/gin"
)

// SendChatCtrl handles POST /chats
func SendChatCtrl(c *gin.Context) {
	var chatReq service.ChatReq
	if err := c.ShouldBindJSON(&chatReq); err != nil {
		service.LogService("warn", "chat binding error", map[string]interface{}{"error": err.Error()}, nil, c.ClientIP())
		respondBadRequest(c, err.Error())
		return
	}
	if val, ok := c.Get("privileges"); ok {
		if privileges, ok := val.(int); ok {
			chatReq.Privileges = privileges
		}
	}

	chatRes, err := service.SendChatServ(chatReq)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, chatRes)
}

// GetAllChatsCtrl handles GET /chats
func GetAllChatsCtrl(c *gin.Context) {
	// Parse optional pagination parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	chats, err := service.GetChatsWithPaginationServ(limit, offset)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, chats)
}

// SearchChatsCtrl handles GET /chat/search
func SearchChatsCtrl(c *gin.Context) {
	providerID, err := parseOptionalInt(c, "provider_id")
	if err != nil {
		respondBadRequest(c, "invalid provider_id")
		return
	}
	userID, err := parseOptionalInt(c, "user_id")
	if err != nil {
		respondBadRequest(c, "invalid user_id")
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	filter := domain.ChatFilter{
		Query:      c.Query("q"),
		Role:       parseOptionalString(c, "role"),
		ProviderID: providerID,
		UserID:     userID,
		Model:      parseOptionalString(c, "model"),
		Limit:      limit,
		Offset:     offset,
	}
	chats, err := service.SearchChatsServ(filter)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, chats)
}

// ConsultAlertCtrl handles POST /chat/alert/:id
func ConsultAlertCtrl(c *gin.Context) {
	alertID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid alert id")
		return
	}

	providerID, _ := strconv.Atoi(c.DefaultQuery("provider_id", "1"))
	model := c.Query("model")

	chatRes, err := service.ConsultAlertServ(uint(providerID), model, alertID)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, chatRes)
}

// ConsultItemCtrl handles POST /chat/item/:id
func ConsultItemCtrl(c *gin.Context) {
	itemID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid item id")
		return
	}

	chatRes, err := service.ConsultItemServ(uint(itemID))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, chatRes)
}

// ConsultHostCtrl handles POST /chat/host/:id
func ConsultHostCtrl(c *gin.Context) {
	hostID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid host id")
		return
	}

	providerID, _ := strconv.Atoi(c.DefaultQuery("provider_id", "1"))
	model := c.Query("model")

	chatRes, err := service.ConsultHostServ(uint(providerID), model, uint(hostID))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, chatRes)
}
