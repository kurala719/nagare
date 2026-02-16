package mcp

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const protocolVersion = "2024-11-05"

const (
	rpcCodeParseError     = -32700
	rpcCodeInvalidRequest = -32600
	rpcCodeMethodNotFound = -32601
	rpcCodeInvalidParams  = -32602
	rpcCodeInternalError  = -32603
)

type session struct {
	id   string
	ch   chan []byte
	done chan struct{}
}

type sessionStore struct {
	mu       sync.Mutex
	sessions map[string]*session
}

func newSessionStore() *sessionStore {
	return &sessionStore{sessions: make(map[string]*session)}
}

func (s *sessionStore) create() *session {
	sess := &session{id: newSessionID(), ch: make(chan []byte, 16), done: make(chan struct{})}
	s.mu.Lock()
	s.sessions[sess.id] = sess
	s.mu.Unlock()
	return sess
}

func (s *sessionStore) get(id string) (*session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[id]
	return sess, ok
}

func (s *sessionStore) remove(id string) {
	s.mu.Lock()
	sess, ok := s.sessions[id]
	if ok {
		delete(s.sessions, id)
	}
	s.mu.Unlock()
	if ok {
		close(sess.done)
	}
}

var store = newSessionStore()

type limiter struct {
	ch chan struct{}
}

func newLimiter(size int) *limiter {
	if size <= 0 {
		size = 4
	}
	return &limiter{ch: make(chan struct{}, size)}
}

func (l *limiter) Acquire() {
	l.ch <- struct{}{}
}

func (l *limiter) Release() {
	<-l.ch
}

var mcpLimiterOnce sync.Once
var mcpLimiter *limiter

func getMCPLimiter() *limiter {
	mcpLimiterOnce.Do(func() {
		size := viper.GetInt("mcp.max_concurrency")
		mcpLimiter = newLimiter(size)
	})
	return mcpLimiter
}

// APIKeyMiddleware guards MCP endpoints with a shared API key.
func APIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !viper.GetBool("mcp.enabled") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			c.Abort()
			return
		}
		expected := strings.TrimSpace(viper.GetString("mcp.api_key"))
		if expected == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "mcp api key not configured"})
			c.Abort()
			return
		}
		provided := strings.TrimSpace(c.GetHeader("X-MCP-API-Key"))
		if provided == "" {
			authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
			if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				provided = strings.TrimSpace(authHeader[7:])
			}
		}
		if provided == "" || provided != expected {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// SSEHandler establishes the MCP SSE transport.
func SSEHandler(c *gin.Context) {
	sess := store.create()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Status(http.StatusOK)

	endpoint := gin.H{
		"url":             "/mcp/message?session=" + sess.id,
		"protocolVersion": protocolVersion,
	}
	writeSSE(c, "endpoint", endpoint)

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	ctx := c.Request.Context()
	for {
		select {
		case msg := <-sess.ch:
			writeSSEBytes(c, "message", msg)
		case <-ticker.C:
			writeSSE(c, "ping", gin.H{"time": time.Now().Unix()})
		case <-ctx.Done():
			store.remove(sess.id)
			return
		}
	}
}

// MessageHandler receives JSON-RPC requests and responds via SSE.
func MessageHandler(c *gin.Context) {
	sessionID := strings.TrimSpace(c.Query("session"))
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing session"})
		return
	}
	sess, ok := store.get(sessionID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	var req rpcRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sendRPCError(sess, req.ID, rpcCodeParseError, "invalid json")
		c.Status(http.StatusAccepted)
		return
	}

	if req.JSONRPC != "2.0" || req.Method == "" {
		sendRPCError(sess, req.ID, rpcCodeInvalidRequest, "invalid request")
		c.Status(http.StatusAccepted)
		return
	}

	limiter := getMCPLimiter()
	request := req
	go func() {
		limiter.Acquire()
		defer limiter.Release()

		resp := handleRPC(request)
		if resp != nil {
			payload, _ := json.Marshal(resp)
			sendToSession(sess, payload)
		}
	}()

	c.Status(http.StatusAccepted)
}

func handleRPC(req rpcRequest) *rpcResponse {
	switch req.Method {
	case "initialize":
		return &rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: gin.H{
				"protocolVersion": protocolVersion,
				"serverInfo": gin.H{
					"name":    "Nagare MCP",
					"version": "0.1.0",
				},
				"capabilities": gin.H{
					"tools": gin.H{},
				},
			},
		}
	case "tools/list":
		return &rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: gin.H{
				"tools": listTools(),
			},
		}
	case "tools/call":
		var params toolCallParams
		if err := decodeParams(req.Params, &params); err != nil {
			return rpcErrorResponse(req.ID, rpcCodeInvalidParams, err.Error())
		}
		result, err := callTool(params.Name, params.Arguments)
		if err != nil {
			return &rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: gin.H{
					"isError": true,
					"content": []toolContent{{Type: "text", Text: err.Error()}},
				},
			}
		}
		resultJSON, _ := json.Marshal(result)
		return &rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: gin.H{
				"isError": false,
				"content": []toolContent{{Type: "text", Text: string(resultJSON)}},
			},
		}
	default:
		return rpcErrorResponse(req.ID, rpcCodeMethodNotFound, "method not found")
	}
}

func sendRPCError(sess *session, id json.RawMessage, code int, message string) {
	resp := rpcErrorResponse(id, code, message)
	payload, _ := json.Marshal(resp)
	sendToSession(sess, payload)
}

func rpcErrorResponse(id json.RawMessage, code int, message string) *rpcResponse {
	return &rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &rpcError{
			Code:    code,
			Message: message,
		},
	}
}

func writeSSE(c *gin.Context, event string, payload interface{}) {
	data, _ := json.Marshal(payload)
	writeSSEBytes(c, event, data)
}

func writeSSEBytes(c *gin.Context, event string, data []byte) {
	_, _ = c.Writer.Write([]byte("event: " + event + "\n"))
	_, _ = c.Writer.Write([]byte("data: " + string(data) + "\n\n"))
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

func sendToSession(sess *session, payload []byte) {
	select {
	case sess.ch <- payload:
	case <-sess.done:
	}
}

func newSessionID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405")))
	}
	return hex.EncodeToString(buf)
}
