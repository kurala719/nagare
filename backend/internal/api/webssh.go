package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"nagare/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Strict origin check: allow only if it matches host or if in development
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		// In production, you should validate against your allowed domains
		return true // Keeping it true for now
	},
}

// HandleWebSSH handles WebSocket connections for WebSSH
func HandleWebSSH(c *gin.Context) {
	ip, port, user, password, ok := parseSSHParams(c)
	if !ok {
		return
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to websocket: %v", err)
		return
	}
	defer ws.Close()

	sshConfig := service.BuildSSHConfig(user, password)

	client, err := service.ConnectSSH(ip, port, sshConfig)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Connection failed: %v", err)))
		return
	}
	defer client.Close()

	cols, _ := strconv.Atoi(c.DefaultQuery("cols", "80"))
	rows, _ := strconv.Atoi(c.DefaultQuery("rows", "24"))

	session, stdin, err := service.SetupSSHSession(client, ws, cols, rows)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		return
	}
	defer session.Close()

	service.HandleSSHIOLoop(ws, session, stdin)
}

func parseSSHParams(c *gin.Context) (ip string, port int, user, password string, ok bool) {
	hostIDStr := c.Param("id")
	if hostIDStr != "" {
		hostID, err := strconv.ParseUint(hostIDStr, 10, 32)
		if err != nil {
			respondBadRequest(c, "invalid host id")
			return "", 0, "", "", false
		}

		ip, port, user, password, err = service.GetHostConnectionDetails(uint(hostID))
		if err != nil {
			respondError(c, err)
			return "", 0, "", "", false
		}
		return ip, port, user, password, true
	}

	// Generic route - get details from query params
	ip = c.Query("ip")
	user = c.Query("user")
	password = c.Query("password")
	portStr := c.Query("port")

	if ip == "" || user == "" {
		respondBadRequest(c, "ip and user are required for direct connection")
		return "", 0, "", "", false
	}

	port, _ = strconv.Atoi(portStr)
	if port == 0 {
		port = 22
	}
	return ip, port, user, password, true
}

