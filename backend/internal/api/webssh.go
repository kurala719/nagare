package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"nagare/internal/repository"
	"nagare/internal/service/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow cross-origin for development
	},
}

type WindowSize struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

type Message struct {
	Type string `json:"type"`
	Data string `json:"data"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

// HandleWebSSH handles WebSocket connections for WebSSH
func HandleWebSSH(c *gin.Context) {
	hostIDStr := c.Param("id")
	if hostIDStr == "" {
		respondBadRequest(c, "host id is required")
		return
	}

	hostID, err := strconv.ParseUint(hostIDStr, 10, 32)
	if err != nil {
		respondBadRequest(c, "invalid host id")
		return
	}

	host, err := repository.GetHostByIDDAO(uint(hostID))
	if err != nil {
		respondError(c, err)
		return
	}

	if host.SSHUser == "" || host.SSHPassword == "" {
		respondBadRequest(c, "SSH credentials not configured for this host")
		return
	}

	password, err := utils.Decrypt(host.SSHPassword)
	if err != nil {
		respondError(c, fmt.Errorf("failed to decrypt SSH password: %w", err))
		return
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to websocket: %v", err)
		return
	}
	defer ws.Close()

	// SSH configuration
	sshConfig := &ssh.ClientConfig{
		User: host.SSHUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host.IPAddr, host.SSHPort)
	if host.SSHPort == 0 {
		addr = fmt.Sprintf("%s:22", host.IPAddr)
	}

	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Connection failed: %v", err)))
		return
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to create session: %v", err)))
		return
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		return
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		return
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return
	}

	// Request pseudo terminal
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	cols, _ := strconv.Atoi(c.DefaultQuery("cols", "80"))
	rows, _ := strconv.Atoi(c.DefaultQuery("rows", "24"))

	if err := session.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to request PTY: %v", err)))
		return
	}

	if err := session.Shell(); err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to start shell: %v", err)))
		return
	}

	// Pipe SSH output to WebSocket
	go func() {
		defer ws.Close()
		io.Copy(wsWriter{ws}, io.MultiReader(stdout, stderr))
	}()

	// Pipe WebSocket input to SSH
	for {
		messageType, p, err := ws.ReadMessage()
		if err != nil {
			return
		}

		if messageType == websocket.TextMessage {
			var msg Message
			if err := json.Unmarshal(p, &msg); err == nil {
				if msg.Type == "resize" {
					session.WindowChange(msg.Rows, msg.Cols)
					continue
				}
				if msg.Type == "data" {
					stdin.Write([]byte(msg.Data))
				}
			} else {
				// Fallback for raw data if not JSON
				stdin.Write(p)
			}
		} else {
			stdin.Write(p)
		}
	}
}

type wsWriter struct {
	*websocket.Conn
}

func (w wsWriter) Write(p []byte) (n int, err error) {
	err = w.WriteMessage(websocket.BinaryMessage, p)
	return len(p), err
}
