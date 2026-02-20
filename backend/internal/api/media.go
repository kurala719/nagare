package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"nagare/internal/model"
	"nagare/internal/repository/media"
	"nagare/internal/service"

	"github.com/gin-gonic/gin"
)

// GetAllMediaCtrl handles GET /media
func GetAllMediaCtrl(c *gin.Context) {
	media, err := service.GetAllMediaServ()
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, media)
}

// SearchMediaCtrl handles GET /media/search
func SearchMediaCtrl(c *gin.Context) {
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
	filter := model.MediaFilter{
		Query:     c.Query("q"),
		Status:    status,
		Type:      parseOptionalString(c, "type"),
		Limit:     limit,
		Offset:    offset,
		SortBy:    c.Query("sort"),
		SortOrder: c.Query("order"),
	}
	media, err := service.SearchMediaServ(filter)
	if err != nil {
		respondError(c, err)
		return
	}
	if withTotal != nil && *withTotal {
		total, err := service.CountMediaServ(filter)
		if err != nil {
			respondError(c, err)
			return
		}
		respondSuccess(c, http.StatusOK, gin.H{"items": media, "total": total})
		return
	}
	respondSuccess(c, http.StatusOK, media)
}

// GetMediaByIDCtrl handles GET /media/:id
func GetMediaByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid media ID")
		return
	}
	media, err := service.GetMediaByIDServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, media)
}

// AddMediaCtrl handles POST /media
func AddMediaCtrl(c *gin.Context) {
	var req service.MediaReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	media, err := service.AddMediaServ(req)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusCreated, media)
}

// UpdateMediaCtrl handles PUT /media/:id
func UpdateMediaCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid media ID")
		return
	}
	var req service.MediaReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	if err := service.UpdateMediaServ(uint(id), req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "media updated")
}

// DeleteMediaByIDCtrl handles DELETE /media/:id
func DeleteMediaByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid media ID")
		return
	}
	if err := service.DeleteMediaByIDServ(uint(id)); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "media deleted")
}

// TestMediaCtrl handles POST /media/:id/test
func TestMediaCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid media ID")
		return
	}
	if err := service.TestMediaServ(uint(id)); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "test message sent successfully")
}

// OneBotMessageEvent represents an incoming OneBot 11 message event (from QQ via NapCat)
type OneBotMessageEvent struct {
	PostType    string          `json:"post_type"`
	MessageType string          `json:"message_type"` // "private" or "group"
	UserID      int64           `json:"user_id"`
	GroupID     int64           `json:"group_id,omitempty"`
	Message     json.RawMessage `json:"message"`
	RawMessage  string          `json:"raw_message,omitempty"`
	MessageID   int64           `json:"message_id,omitempty"`
	Time        int64           `json:"time,omitempty"`
	Sender      *OneBotSender   `json:"sender,omitempty"`
}

// OneBotSender represents the sender information in OneBot 11 event
type OneBotSender struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname,omitempty"`
	Role     string `json:"role,omitempty"`
}

// OneBotMessageResponseData represents response data for OneBot 11
type OneBotMessageResponseData struct {
	MessageID int64  `json:"message_id,omitempty"`
	Reply     string `json:"reply,omitempty"`
}

// OneBotMessageResponse represents response format for OneBot 11
type OneBotMessageResponse struct {
	Status  string                     `json:"status"`
	Retcode int                        `json:"retcode"`
	Message string                     `json:"message,omitempty"`
	Data    *OneBotMessageResponseData `json:"data,omitempty"`
}

// HandleQQMessageCtrl handles incoming messages from QQ (OneBot 11)
// This endpoint receives messages from QQ via HTTP POST from NapCat
// Messages starting with "/" are treated as commands
// Example: POST /media/qq/message with body: {"post_type":"message","message_type":"private","user_id":123456789,"message":"/status"}
func HandleQQMessageCtrl(c *gin.Context) {
	var event OneBotMessageEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		respondBadRequest(c, "invalid OneBot message event")
		return
	}

	// Validate event
	if strings.TrimSpace(event.PostType) == "" {
		event.PostType = "message"
	}
	if event.PostType != "message" {
		respondBadRequest(c, "only message events are supported")
		return
	}

	if strings.TrimSpace(event.MessageType) == "" {
		if event.GroupID != 0 {
			event.MessageType = "group"
		} else {
			event.MessageType = "private"
		}
	}

	message := extractOneBotMessageText(event.Message, event.RawMessage)
	if message == "" {
		respondBadRequest(c, "message cannot be empty")
		return
	}

	// Determine the target for reply
	var replyTarget string
	var qqID string
	var isGroup bool
	mediaType := "qq"

	if event.MessageType == "group" {
		if event.GroupID == 0 {
			respondBadRequest(c, "group_id is required for group messages")
			return
		}
		replyTarget = "group:" + strconv.FormatInt(event.GroupID, 10)
		qqID = strconv.FormatInt(event.GroupID, 10)
		isGroup = true
	} else {
		if event.UserID == 0 {
			respondBadRequest(c, "user_id is required")
			return
		}
		replyTarget = "user:" + strconv.FormatInt(event.UserID, 10)
		qqID = strconv.FormatInt(event.UserID, 10)
		isGroup = false
	}

	// Check whitelist for command execution
	if strings.HasPrefix(strings.TrimSpace(message), "/") {
		if !service.CheckQQWhitelistForCommand(qqID, isGroup) {
			service.LogSystem("warn", "QQ command rejected: not in whitelist", map[string]interface{}{
				"qq_id":    qqID,
				"is_group": isGroup,
				"message":  message,
			}, nil, "")
			resp := OneBotMessageResponse{
				Status:  "ok",
				Retcode: 0,
				Message: "You are not authorized to execute commands.",
			}
			c.JSON(http.StatusOK, resp)
			return
		}
	}

	// Process the message as a command
	result, err := service.HandleIMCommand(message)
	if err != nil {
		// Log error but still send a response
		service.LogSystem("error", "failed to process QQ command", map[string]interface{}{
			"message": message,
			"error":   err.Error(),
		}, nil, "")
		result.Reply = "Command processing failed: " + err.Error()
	}

	// Send reply back to QQ
	if result.Reply != "" {
		err = service.SendIMReply(mediaType, replyTarget, result.Reply)
		if err != nil {
			service.LogSystem("error", "failed to send QQ reply", map[string]interface{}{
				"target": replyTarget,
				"error":  err.Error(),
			}, nil, "")
		}
	}

	// Return success response to OneBot 11
	resp := OneBotMessageResponse{
		Status:  "ok",
		Retcode: 0,
		Message: "success",
	}
	if strings.TrimSpace(result.Reply) != "" {
		resp.Message = result.Reply
		resp.Data = &OneBotMessageResponseData{Reply: result.Reply}
	}
	c.JSON(http.StatusOK, resp)
}

// HandleQQWebSocket handles Reverse WebSocket connections from NapCat (OneBot 11)
func HandleQQWebSocket(c *gin.Context) {
	media.GlobalQQWSManager.HandleReverseWS(c)
}

type oneBotMessageSegment struct {
	Type string            `json:"type"`
	Data map[string]string `json:"data"`
}

func extractOneBotMessageText(raw json.RawMessage, fallback string) string {
	if len(raw) == 0 {
		return strings.TrimSpace(fallback)
	}

	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return strings.TrimSpace(text)
	}

	var segments []oneBotMessageSegment
	if err := json.Unmarshal(raw, &segments); err == nil {
		var builder strings.Builder
		for _, segment := range segments {
			if segment.Type != "text" {
				continue
			}
			if value, ok := segment.Data["text"]; ok {
				builder.WriteString(value)
			}
		}
		return strings.TrimSpace(builder.String())
	}

	return strings.TrimSpace(fallback)
}
