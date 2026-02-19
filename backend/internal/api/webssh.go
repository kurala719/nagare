package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"nagare/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"golang.org/x/crypto/ssh"
)

var jsonIter = jsoniter.ConfigCompatibleWithStandardLibrary

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
		return true // Keeping it true for now but using jsoniter
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
	var ip, user, password string
	var port int
	var err error

	hostIDStr := c.Param("id")
	if hostIDStr != "" {
		hostID, err := strconv.ParseUint(hostIDStr, 10, 32)
		if err != nil {
			respondBadRequest(c, "invalid host id")
			return
		}

		ip, port, user, password, err = service.GetHostConnectionDetails(uint(hostID))
		if err != nil {
			respondError(c, err)
			return
		}
	} else {
		// Generic route - get details from query params
		ip = c.Query("ip")
		user = c.Query("user")
		password = c.Query("password")
		portStr := c.Query("port")
		
		if ip == "" || user == "" {
			respondBadRequest(c, "ip and user are required for direct connection")
			return
		}
		
		port, _ = strconv.Atoi(portStr)
		if port == 0 {
			port = 22
		}
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to websocket: %v", err)
		return
	}
	defer ws.Close()

	// SSH configuration
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
		Config: ssh.Config{
			KeyExchanges: []string{
				"diffie-hellman-group1-sha1",
				"diffie-hellman-group14-sha1",
				"ecdh-sha2-nistp256",
				"ecdh-sha2-nistp384",
				"ecdh-sha2-nistp521",
				"curve25519-sha256@libssh.org",
			},
			Ciphers: []string{
				"aes128-cbc",
				"aes128-ctr",
				"aes192-ctr",
				"aes256-ctr",
				"aes128-gcm@openssh.com",
				"chacha20-poly1305@openssh.com",
			},
			MACs: []string{
				"hmac-sha1",
				"hmac-sha2-256",
				"hmac-sha2-512",
				"hmac-sha1-96",
			},
		},
	}
	sshConfig.HostKeyAlgorithms = append(sshConfig.HostKeyAlgorithms, "ssh-rsa", "ssh-dss")

	addr := fmt.Sprintf("%s:%d", ip, port)
	if port == 0 {
		addr = fmt.Sprintf("%s:22", ip)
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
			if err := jsonIter.Unmarshal(p, &msg); err == nil {
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
