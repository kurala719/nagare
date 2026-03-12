package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"nagare/internal/model"
	"nagare/internal/repository"
	"nagare/internal/repository/media"

	"github.com/gorilla/websocket"
)

// MediaReq represents a media request
type MediaReq struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Target      string `json:"target" binding:"required"`
	Enabled     int    `json:"enabled"`
	Description string `json:"description"`
}

// MediaResp represents a media response
type MediaResp struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Target      string `json:"target"`
	Enabled     int    `json:"enabled"`
	Status      int    `json:"status"`
	Description string `json:"description"`
}

func GetAllMediaServ() ([]MediaResp, error) {
	media, err := repository.GetAllMediaDAO()
	if err != nil {
		return nil, fmt.Errorf("failed to get media: %w", err)
	}
	result := make([]MediaResp, 0, len(media))
	for _, m := range media {
		result = append(result, mediaToResp(m))
	}
	return result, nil
}

func SearchMediaServ(filter model.MediaFilter) ([]MediaResp, error) {
	media, err := repository.SearchMediaDAO(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search media: %w", err)
	}
	result := make([]MediaResp, 0, len(media))
	for _, m := range media {
		result = append(result, mediaToResp(m))
	}
	return result, nil
}

// CountMediaServ returns total count for media by filter
func CountMediaServ(filter model.MediaFilter) (int64, error) {
	return repository.CountMediaDAO(filter)
}

func GetMediaByIDServ(id uint) (MediaResp, error) {
	media, err := repository.GetMediaByIDDAO(id)
	if err != nil {
		return MediaResp{}, fmt.Errorf("failed to get media: %w", err)
	}
	return mediaToResp(media), nil
}

func TestMediaServ(id uint) error {
	media, err := repository.GetMediaByIDDAO(id)
	if err != nil {
		LogService("error", "test media failed to load media", map[string]interface{}{"media_id": id, "error": err.Error()}, nil, "")
		return fmt.Errorf("failed to load media: %w", err)
	}
	if media.Enabled == 0 {
		LogService("warn", "test media skipped - media disabled", map[string]interface{}{"media_id": id, "media_type": media.Type, "media_name": media.Name}, nil, "")
		return fmt.Errorf("media is disabled")
	}

	testMessage := "Nagare Media Test Connection: Your notification system is working correctly."

	if strings.EqualFold(strings.TrimSpace(media.Type), "qq") && isQQEndpointOnlyTarget(media.Target) {
		if err := testQQWebSocketEndpoint(media.Target); err != nil {
			LogService("error", "test media failed: qq websocket endpoint unreachable", map[string]interface{}{
				"media_id":   id,
				"media_type": media.Type,
				"media_name": media.Name,
				"target":     media.Target,
				"error":      err.Error(),
			}, nil, "")
			return fmt.Errorf("failed to connect qq websocket endpoint: %w", err)
		}

		LogService("info", "test media succeeded (qq endpoint connectivity)", map[string]interface{}{
			"media_id":   id,
			"media_type": media.Type,
			"media_name": media.Name,
			"target":     media.Target,
		}, nil, "")
		return nil
	}

	LogService("info", "test media starting", map[string]interface{}{
		"media_id":   id,
		"media_type": media.Type,
		"media_name": media.Name,
		"target":     media.Target,
	}, nil, "")

	err = SendIMReply(media.Type, media.Target, testMessage)
	if err != nil {
		LogService("error", "test media failed", map[string]interface{}{
			"media_id":   id,
			"media_type": media.Type,
			"media_name": media.Name,
			"target":     media.Target,
			"error":      err.Error(),
		}, nil, "")
		return fmt.Errorf("failed to send test message via %s: %w", media.Type, err)
	}

	LogService("info", "test media succeeded", map[string]interface{}{
		"media_id":   id,
		"media_type": media.Type,
		"media_name": media.Name,
		"target":     media.Target,
	}, nil, "")
	return nil
}

func isQQEndpointOnlyTarget(target string) bool {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		return false
	}
	return len(strings.Fields(trimmed)) == 1
}

func normalizeQQWebSocketEndpoint(target string) string {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		return ""
	}
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "ws://") || strings.HasPrefix(lower, "wss://") {
		return trimmed
	}
	if strings.HasPrefix(lower, "http://") {
		return "ws://" + strings.TrimPrefix(trimmed, "http://")
	}
	if strings.HasPrefix(lower, "https://") {
		return "wss://" + strings.TrimPrefix(trimmed, "https://")
	}
	return "ws://" + trimmed
}

func testQQWebSocketEndpoint(target string) error {
	endpoint := normalizeQQWebSocketEndpoint(target)
	if endpoint == "" {
		return fmt.Errorf("empty websocket endpoint")
	}

	dialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 5 * time.Second,
	}

	conn, _, err := dialer.Dial(endpoint, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "test done"), time.Now().Add(1*time.Second))
	return nil
}

func AddMediaServ(req MediaReq) (MediaResp, error) {
	media := model.Media{
		Name:        req.Name,
		Type:        req.Type,
		Target:      req.Target,
		Enabled:     req.Enabled,
		Status:      determineMediaStatus(model.Media{Enabled: req.Enabled, Type: req.Type, Target: req.Target}),
		Description: req.Description,
	}
	if err := repository.AddMediaDAO(media); err != nil {
		return MediaResp{}, fmt.Errorf("failed to add media: %w", err)
	}
	return mediaToResp(media), nil
}

func UpdateMediaServ(id uint, req MediaReq) error {
	existing, err := repository.GetMediaByIDDAO(id)
	if err != nil {
		return err
	}
	updated := model.Media{
		Model:       existing.Model,
		Name:        req.Name,
		Type:        req.Type,
		Target:      req.Target,
		Enabled:     req.Enabled,
		Status:      existing.Status,
		Description: req.Description,
	}
	// Preserve status unless enabled state, type or target changed
	if req.Enabled != existing.Enabled || req.Type != existing.Type || req.Target != existing.Target {
		updated.Status = determineMediaStatus(model.Media{Enabled: req.Enabled, Type: req.Type, Target: req.Target})
	}
	if err := repository.UpdateMediaDAO(id, updated); err != nil {
		return err
	}
	_, _ = recomputeMediaStatus(id)
	if updated.Type == "qq" || existing.Type == "qq" {
		RestartQQWSServ()
	}
	return nil
}

func DeleteMediaByIDServ(id uint) error {
	existing, _ := repository.GetMediaByIDDAO(id)
	err := repository.DeleteMediaByIDDAO(id)
	if existing.Type == "qq" {
		RestartQQWSServ()
	}
	return err
}

func mediaToResp(media model.Media) MediaResp {
	return MediaResp{
		ID:          int(media.ID),
		Name:        media.Name,
		Type:        media.Type,
		Target:      media.Target,
		Enabled:     media.Enabled,
		Status:      media.Status,
		Description: media.Description,
	}
}

// BackfillMediaParamsAndTargetsServ is now a no-op as MediaType is removed
func BackfillMediaParamsAndTargetsServ() (int, int, error) {
	return 0, 0, nil
}

var (
	qqWSMu     sync.Mutex
	qqWSCancel context.CancelFunc
)

// InitQQWSServ initializes the QQ WebSocket connection based on Media configuration
func InitQQWSServ() {
	// Find the enabled QQ media (do not gate on status to avoid stale-status deadlock)
	t := "qq"
	mediaList, err := repository.SearchMediaDAO(model.MediaFilter{Type: &t, Limit: 50})
	if err != nil || len(mediaList) == 0 {
		StopQQWSServ()
		return
	}

	var qqMedia model.Media
	foundEnabled := false
	for _, m := range mediaList {
		if m.Enabled == 1 {
			qqMedia = m
			foundEnabled = true
			break
		}
	}
	if !foundEnabled {
		StopQQWSServ()
		return
	}
	_, _ = recomputeMediaStatus(qqMedia.ID)
	target := qqMedia.Target

	// Target could be ws://url?access_token=token
	var positiveURL, accessToken string
	if strings.HasPrefix(target, "ws://") || strings.HasPrefix(target, "wss://") {
		u, err := url.Parse(target)
		if err == nil {
			accessToken = u.Query().Get("access_token")
			// Remove access token from URL for connection if intended (NapCat usually takes it in header or query)
			positiveURL = target
		} else {
			positiveURL = target
		}
	} else {
		// Default to reverse mode or invalid Positive URL
		StopQQWSServ()
		return
	}

	// Update manager's internal config
	media.GlobalQQWSManager.UpdateConfig(accessToken)

	qqWSMu.Lock()
	if qqWSCancel != nil {
		qqWSMu.Unlock()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	qqWSCancel = cancel
	qqWSMu.Unlock()

	go func() {
		log.Printf("[QQ-WS] Starting positive reconnection loop for Media ID %d", qqMedia.ID)
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		attemptConnect := func() {
			if !media.GlobalQQWSManager.IsConnected() {
				_ = repository.UpdateMediaStatusDAO(qqMedia.ID, 2)
				log.Printf("[QQ-WS] Attempting positive connection to %s", positiveURL)
				err := media.GlobalQQWSManager.ConnectPositiveWS(positiveURL, accessToken)
				if err != nil {
					log.Printf("[QQ-WS] Positive connection failed: %v, retrying in 10s...", err)
					_ = repository.UpdateMediaStatusDAO(qqMedia.ID, 2)
					return
				}
			}
			_ = repository.UpdateMediaStatusDAO(qqMedia.ID, 1)
		}

		attemptConnect()
		for {
			select {
			case <-ctx.Done():
				log.Printf("[QQ-WS] Positive reconnection loop stopped")
				return
			case <-ticker.C:
				attemptConnect()
			}
		}
	}()
}

// StopQQWSServ stops the Positive WebSocket reconnection loop
func StopQQWSServ() {
	qqWSMu.Lock()
	cancel := qqWSCancel
	qqWSCancel = nil
	qqWSMu.Unlock()
	if cancel != nil {
		cancel()
	}
	if media.GlobalQQWSManager.IsConnected() {
		// Close the underlying connection to trigger a fast disconnect
		// Not strictly required since cancel takes care of the loop, but good for cleanup.
		// A dedicated close method is better, but this will do if missing.
	}
}

// RestartQQWSServ stops and restarts the QQ WebSocket service
func RestartQQWSServ() {
	StopQQWSServ()
	InitQQWSServ()
}
