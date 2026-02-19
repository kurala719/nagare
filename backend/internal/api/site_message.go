package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"nagare/internal/service"
)

// GetSiteMessagesCtrl handles GET /site-messages
func GetSiteMessagesCtrl(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	unreadOnlyQuery := c.DefaultQuery("unread_only", "0")
	unreadOnly := unreadOnlyQuery == "1" || unreadOnlyQuery == "true"
	
	// Get user ID from context if authenticated
	var userID *uint
	if val, ok := c.Get("uid"); ok {
		if id, ok := val.(uint); ok {
			userID = &id
		}
	}

	messages, err := service.GetSiteMessagesServ(userID, unreadOnly, limit, offset)
	if err != nil {
		respondError(c, err)
		return
	}

	var total int64
	if unreadOnly {
		total, _ = service.GetUnreadCountServ(userID)
	} else {
		total, _ = service.GetTotalMessagesCountServ(userID)
	}

	respondSuccess(c, http.StatusOK, gin.H{
		"items": messages,
		"total": total,
	})
}

// MarkSiteMessageAsReadCtrl handles PUT /site-messages/:id/read
func MarkSiteMessageAsReadCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid message ID")
		return
	}

	if err := service.MarkSiteMessageAsReadServ(uint(id)); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "message marked as read")
}

// MarkAllSiteMessagesAsReadCtrl handles PUT /site-messages/read-all
func MarkAllSiteMessagesAsReadCtrl(c *gin.Context) {
	var userID *uint
	if val, ok := c.Get("uid"); ok {
		if id, ok := val.(uint); ok {
			userID = &id
		}
	}

	if err := service.MarkAllSiteMessagesAsReadServ(userID); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "all messages marked as read")
}

// GetUnreadSiteMessagesCountCtrl handles GET /site-messages/unread-count
func GetUnreadSiteMessagesCountCtrl(c *gin.Context) {
	var userID *uint
	if val, ok := c.Get("uid"); ok {
		if id, ok := val.(uint); ok {
			userID = &id
		}
	}

	count, err := service.GetUnreadCountServ(userID)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, gin.H{"count": count})
}

// DeleteSiteMessageCtrl handles DELETE /site-messages/:id
func DeleteSiteMessageCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid message ID")
		return
	}

	if err := service.DeleteSiteMessageServ(uint(id)); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "message deleted")
}

// HandleSiteMessageWS handles WebSocket connections for notifications
func HandleSiteMessageWS(c *gin.Context) {
	var uid uint = 0
	if val, ok := c.Get("uid"); ok {
		switch v := val.(type) {
		case uint:
			uid = v
		case float64:
			uid = uint(v)
		case int:
			uid = uint(v)
		}
	}
	service.ServeWs(service.GlobalHub, c.Writer, c.Request, uid)
}
