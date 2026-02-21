package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// DefaultQQBaseURL is the local NapCat base URL
const DefaultQQBaseURL = "http://127.0.0.1:3000"

// QQProvider sends QQ messages via NapCat (OneBot 11)
type QQProvider struct {
	BaseURL string
	Client  *http.Client
}

// NewQQProvider creates a QQ provider
func NewQQProvider(baseURL string) *QQProvider {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = DefaultQQBaseURL
	}
	return &QQProvider{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 5 * time.Second},
	}
}

type qqSendResponse struct {
	Status  string `json:"status"`
	Retcode int    `json:"retcode"`
	Message string `json:"message"`
	Wording string `json:"wording"`
}

// SendMessage sends a QQ message to a user or group.
// Target formats: group:123456 | group_id=123456 | g:123456 | user:123456 | user_id=123456 | u:123456 | 123456 (default private)
func (p *QQProvider) SendMessage(ctx context.Context, target, message string) error {
	baseURL, messageType, userID, groupID, err := parseQQTarget(target)
	if err != nil {
		return err
	}

	// Try WebSocket first if connected and no specific baseURL is provided in target
	if strings.TrimSpace(baseURL) == "" && GlobalQQWSManager.IsConnected() {
		return GlobalQQWSManager.SendMessage(ctx, messageType, userID, groupID, message)
	}

	if strings.TrimSpace(baseURL) == "" {
		baseURL = p.BaseURL
	}
	payload := map[string]interface{}{
		"message":     message,
		"auto_escape": false,
	}
	if messageType == "group" {
		id, _ := strconv.ParseInt(groupID, 10, 64)
		payload["group_id"] = id
		payload["message_type"] = "group"
	} else {
		id, _ := strconv.ParseInt(userID, 10, 64)
		payload["user_id"] = id
		payload["message_type"] = "private"
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/send_msg", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("qq api status %d", resp.StatusCode)
	}
	var apiResp qqSendResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return err
	}
	if strings.ToLower(apiResp.Status) != "ok" || apiResp.Retcode != 0 {
		msg := apiResp.Message
		if msg == "" {
			msg = apiResp.Wording
		}
		if msg == "" {
			msg = "qq api failed"
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}

func parseQQTarget(target string) (baseURL, messageType, userID, groupID string, err error) {
	value := strings.TrimSpace(target)
	if value == "" {
		return "", "", "", "", fmt.Errorf("qq target is empty")
	}
	fields := strings.Fields(value)
	if len(fields) > 1 {
		candidate := strings.TrimSpace(fields[0])
		lowerCandidate := strings.ToLower(candidate)
		if !strings.HasPrefix(lowerCandidate, "group:") &&
			!strings.HasPrefix(lowerCandidate, "group_id:") &&
			!strings.HasPrefix(lowerCandidate, "group_id=") &&
			!strings.HasPrefix(lowerCandidate, "g:") &&
			!strings.HasPrefix(lowerCandidate, "private:") &&
			!strings.HasPrefix(lowerCandidate, "user:") &&
			!strings.HasPrefix(lowerCandidate, "user_id:") &&
			!strings.HasPrefix(lowerCandidate, "user_id=") &&
			!strings.HasPrefix(lowerCandidate, "u:") {
			baseURL = normalizeQQBaseURL(candidate)
			value = strings.TrimSpace(strings.Join(fields[1:], " "))
		}
	}
	if value == "" {
		return baseURL, "", "", "", fmt.Errorf("qq target is empty")
	}
	lower := strings.ToLower(value)
	switch {
	case strings.HasPrefix(lower, "group:"):
		groupID = strings.TrimSpace(value[len("group:"):])
		messageType = "group"
	case strings.HasPrefix(lower, "group_id:"):
		groupID = strings.TrimSpace(value[len("group_id:"):])
		messageType = "group"
	case strings.HasPrefix(lower, "group_id="):
		groupID = strings.TrimSpace(value[len("group_id="):])
		messageType = "group"
	case strings.HasPrefix(lower, "g:"):
		groupID = strings.TrimSpace(value[len("g:"):])
		messageType = "group"
	case strings.HasPrefix(lower, "private:"):
		userID = strings.TrimSpace(value[len("private:"):])
		messageType = "private"
	case strings.HasPrefix(lower, "user:"):
		userID = strings.TrimSpace(value[len("user:"):])
		messageType = "private"
	case strings.HasPrefix(lower, "user_id:"):
		userID = strings.TrimSpace(value[len("user_id:"):])
		messageType = "private"
	case strings.HasPrefix(lower, "user_id="):
		userID = strings.TrimSpace(value[len("user_id="):])
		messageType = "private"
	case strings.HasPrefix(lower, "u:"):
		userID = strings.TrimSpace(value[len("u:"):])
		messageType = "private"
	default:
		userID = value
		messageType = "private"
	}
	if messageType == "group" && groupID == "" {
		return baseURL, "", "", "", fmt.Errorf("qq group id is empty")
	}
	if messageType == "private" && userID == "" {
		return baseURL, "", "", "", fmt.Errorf("qq user id is empty")
	}
	return baseURL, messageType, userID, groupID, nil
}

func normalizeQQBaseURL(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(trimmed), "http://") || strings.HasPrefix(strings.ToLower(trimmed), "https://") {
		return trimmed
	}
	return "http://" + trimmed
}
