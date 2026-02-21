package media

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// OneBotAction represents a OneBot 11 action request
type OneBotAction struct {
	Action string      `json:"action"`
	Params interface{} `json:"params"`
	Echo   string      `json:"echo,omitempty"`
}

// OneBotResponse represents a OneBot 11 action response
type OneBotResponse struct {
	Status  string      `json:"status"`
	Retcode int         `json:"retcode"`
	Data    interface{} `json:"data"`
	Echo    string      `json:"echo,omitempty"`
}

// OneBotEvent represents a generic OneBot 11 event
type OneBotEvent struct {
	PostType    string          `json:"post_type"`
	MessageType string          `json:"message_type,omitempty"`
	UserID      int64           `json:"user_id,omitempty"`
	GroupID     int64           `json:"group_id,omitempty"`
	Message     json.RawMessage `json:"message,omitempty"`
	RawMessage  string          `json:"raw_message,omitempty"`
	SelfID      int64           `json:"self_id,omitempty"`
}

// QQCommandHandler is a function that handles an incoming QQ message
type QQCommandHandler func(message string, qqID string, isGroup bool) (reply string, err error)

// QQWebSocketManager manages the WebSocket connection to NapCat
type QQWebSocketManager struct {
	conn           *websocket.Conn
	accessToken    string
	mu             sync.RWMutex
	echoMap        map[string]chan OneBotResponse
	CommandHandler QQCommandHandler
}

var GlobalQQWSManager = &QQWebSocketManager{
	echoMap: make(map[string]chan OneBotResponse),
}

var qqUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleReverseWS handles NapCat reverse WebSocket connection
func (m *QQWebSocketManager) HandleReverseWS(c *gin.Context) {
	conn, err := qqUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[QQ-WS] Failed to upgrade: %v", err)
		return
	}

	// Optional: Validate access token from header if configured
	if m.accessToken != "" {
		token := c.GetHeader("Authorization")
		if token == "" {
			token = c.Query("access_token")
		}
		if !strings.HasSuffix(token, m.accessToken) {
			log.Printf("[QQ-WS] Unauthorized reverse connection attempt")
			conn.Close()
			return
		}
	}

	m.mu.Lock()
	if m.conn != nil {
		m.conn.Close()
	}
	m.conn = conn
	m.mu.Unlock()

	log.Printf("[QQ-WS] NapCat connected via Reverse WebSocket")
	m.Listen(conn, "Reverse")
}

// ConnectPositiveWS connects to NapCat as a client
func (m *QQWebSocketManager) ConnectPositiveWS(url string, accessToken string) error {
	m.mu.Lock()
	m.accessToken = accessToken
	m.mu.Unlock()

	header := http.Header{}
	if accessToken != "" {
		header.Set("Authorization", "Bearer "+accessToken)
	}

	conn, _, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		return fmt.Errorf("failed to dial positive websocket: %w", err)
	}

	m.mu.Lock()
	if m.conn != nil {
		m.conn.Close()
	}
	m.conn = conn
	m.mu.Unlock()

	log.Printf("[QQ-WS] Connected to NapCat via Positive WebSocket: %s", url)
	go m.Listen(conn, "Positive")
	return nil
}

// Listen starts reading messages from the connection
func (m *QQWebSocketManager) Listen(conn *websocket.Conn, mode string) {
	defer func() {
		m.mu.Lock()
		if m.conn == conn {
			m.conn = nil
		}
		m.mu.Unlock()
		conn.Close()
		log.Printf("[QQ-WS] NapCat %s disconnected", mode)
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[QQ-WS] %s read error: %v", err)
			break
		}

		go m.handleIncomingMessage(message)
	}
}

func (m *QQWebSocketManager) handleIncomingMessage(message []byte) {
	var raw map[string]interface{}
	if err := json.Unmarshal(message, &raw); err != nil {
		return
	}

	if _, ok := raw["post_type"]; ok {
		var event OneBotEvent
		if err := json.Unmarshal(message, &event); err != nil {
			return
		}
		m.handleEvent(event)
	} else if _, ok := raw["status"]; ok {
		var resp OneBotResponse
		if err := json.Unmarshal(message, &resp); err != nil {
			return
		}
		
		// Log non-zero retcodes for debugging
		if resp.Retcode != 0 {
			log.Printf("[QQ-WS] API Response Error: status=%s, retcode=%d, echo=%s", resp.Status, resp.Retcode, resp.Echo)
		}

		if resp.Echo != "" {
			m.mu.RLock()
			ch, ok := m.echoMap[resp.Echo]
			m.mu.RUnlock()
			if ok {
				ch <- resp
			}
		}
	}
}

func (m *QQWebSocketManager) handleEvent(event OneBotEvent) {
	if event.PostType != "message" {
		return
	}

	msgText := extractOneBotMessageText(event.Message, event.RawMessage)
	if msgText == "" {
		return
	}

	if m.CommandHandler != nil {
		var qqID string
		var isGroup bool
		if event.MessageType == "group" {
			qqID = strconv.FormatInt(event.GroupID, 10)
			isGroup = true
		} else {
			qqID = strconv.FormatInt(event.UserID, 10)
			isGroup = false
		}

		reply, err := m.CommandHandler(msgText, qqID, isGroup)
		if err != nil {
			log.Printf("[QQ-WS] Command handling error: %v", err)
			return
		}

		if reply != "" {
			m.sendReply(event, reply)
		}
	}
}

func (m *QQWebSocketManager) sendReply(event OneBotEvent, text string) {
	params := map[string]interface{}{
		"message": text,
	}
	if event.MessageType == "group" {
		params["group_id"] = event.GroupID
	} else {
		params["user_id"] = event.UserID
	}

	m.CallAction("send_msg", params)
}

func (m *QQWebSocketManager) CallAction(action string, params interface{}) (OneBotResponse, error) {
	m.mu.RLock()
	conn := m.conn
	m.mu.RUnlock()

	if conn == nil {
		return OneBotResponse{}, fmt.Errorf("no active QQ WebSocket connection")
	}

	echo := strconv.FormatInt(time.Now().UnixNano(), 10)
	req := OneBotAction{
		Action: action,
		Params: params,
		Echo:   echo,
	}

	ch := make(chan OneBotResponse, 1)
	m.mu.Lock()
	m.echoMap[echo] = ch
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		delete(m.echoMap, echo)
		m.mu.Unlock()
	}()

	data, _ := json.Marshal(req)
	m.mu.Lock()
	err := conn.WriteMessage(websocket.TextMessage, data)
	m.mu.Unlock()

	if err != nil {
		return OneBotResponse{}, err
	}

	select {
	case resp := <-ch:
		return resp, nil
	case <-time.After(5 * time.Second):
		return OneBotResponse{}, fmt.Errorf("timeout waiting for QQ response")
	}
}

// SendMessage sends a message via WebSocket
func (m *QQWebSocketManager) SendMessage(ctx context.Context, messageType, userID, groupID, message string) error {
	params := map[string]interface{}{
		"message": message,
	}
	if messageType == "group" {
		id, _ := strconv.ParseInt(groupID, 10, 64)
		params["group_id"] = id
		params["message_type"] = "group"
	} else {
		id, _ := strconv.ParseInt(userID, 10, 64)
		params["user_id"] = id
		params["message_type"] = "private"
	}

	resp, err := m.CallAction("send_msg", params)
	if err != nil {
		return err
	}

	if resp.Retcode != 0 {
		log.Printf("[QQ-WS] OneBot API error: retcode=%d, status=%s, message=%s", resp.Retcode, resp.Status, resp.Echo)
		return fmt.Errorf("QQ API error: %d (status: %s)", resp.Retcode, resp.Status)
	}

	return nil
}

func (m *QQWebSocketManager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.conn != nil
}

func (m *QQWebSocketManager) UpdateConfig(accessToken string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.accessToken = accessToken
}

func extractOneBotMessageText(raw json.RawMessage, fallback string) string {
	if len(raw) == 0 {
		return strings.TrimSpace(fallback)
	}

	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return strings.TrimSpace(text)
	}

	type oneBotMessageSegment struct {
		Type string            `json:"type"`
		Data map[string]string `json:"data"`
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
