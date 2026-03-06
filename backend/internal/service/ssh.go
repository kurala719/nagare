package service

import (
	"fmt"
	"io"
	"time"

	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"golang.org/x/crypto/ssh"
)

var jsonIter = jsoniter.ConfigCompatibleWithStandardLibrary

// SSHMessage represents a message sent over the WebSocket for SSH
type SSHMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

type wsWriter struct {
	*websocket.Conn
}

func (w wsWriter) Write(p []byte) (n int, err error) {
	err = w.WriteMessage(websocket.BinaryMessage, p)
	return len(p), err
}

// BuildSSHConfig creates an SSH client configuration using the provided credentials.
func BuildSSHConfig(user, password string) *ssh.ClientConfig {
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
	return sshConfig
}

// ConnectSSH dials the SSH server and returns the client.
func ConnectSSH(ip string, port int, config *ssh.ClientConfig) (*ssh.Client, error) {
	addr := fmt.Sprintf("%s:%d", ip, port)
	if port == 0 {
		addr = fmt.Sprintf("%s:22", ip)
	}
	return ssh.Dial("tcp", addr, config)
}

// SetupSSHSession opens a new SSH session, sets up the PTY, and starts the shell.
func SetupSSHSession(client *ssh.Client, ws *websocket.Conn, cols, rows int) (*ssh.Session, io.WriteCloser, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		return nil, nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		session.Close()
		return nil, nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Request pseudo terminal
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		session.Close()
		return nil, nil, fmt.Errorf("failed to request PTY: %w", err)
	}

	if err := session.Shell(); err != nil {
		session.Close()
		return nil, nil, fmt.Errorf("failed to start shell: %w", err)
	}

	// Pipe SSH output to WebSocket
	go func() {
		defer ws.Close()
		io.Copy(wsWriter{ws}, io.MultiReader(stdout, stderr))
	}()

	return session, stdin, nil
}

// HandleSSHIOLoop pipes the data from the WebSocket input to the SSH standard input.
func HandleSSHIOLoop(ws *websocket.Conn, session *ssh.Session, stdin io.WriteCloser) {
	for {
		messageType, p, err := ws.ReadMessage()
		if err != nil {
			return
		}

		if messageType == websocket.TextMessage {
			var msg SSHMessage
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
