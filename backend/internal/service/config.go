package service

import (
	"context"
	"log"
	"time"

	"nagare/internal/repository"
	"nagare/internal/repository/media"
)

// GetAllConfigServ returns all configuration settings
func GetAllConfigServ() map[string]interface{} {
	return repository.GetAllConfig()
}

// GetConfigServ retrieves a specific configuration value
func GetConfigServ(key string) interface{} {
	return repository.GetConfigValue(key)
}

// GetMainConfigServ returns the main configuration
func GetMainConfigServ() (repository.ConfigResponse, error) {
	return repository.GetMainConfig()
}

// ModifyConfigServ updates a configuration value and saves
func ModifyConfigServ(key string, value interface{}) error {
	repository.SetConfigValue(key, value)
	return repository.SaveConfig()
}

// InitConfigServ initializes configuration from path
func InitConfigServ(path string) error {
	return repository.InitConfig(path)
}

// LoadConfigServ reloads the configuration
func LoadConfigServ() error {
	return repository.LoadConfig()
}

// ResetConfigServ resets the configuration to defaults
func ResetConfigServ() error {
	return repository.ResetConfig()
}

// SaveConfigServ persists configuration to disk
func SaveConfigServ() error {
	return repository.SaveConfig()
}

func siteMessageMinAlertSeverity() int {
	val := GetConfigServ("site_message.min_alert_severity")
	if val == nil {
		return 0
	}
	if i, ok := val.(int); ok {
		return i
	}
	// Try float64 because JSON unmarshals numbers as floats
	if f, ok := val.(float64); ok {
		return int(f)
	}
	return 0
}

func siteMessageMinLogSeverity() int {
	val := GetConfigServ("site_message.min_log_severity")
	if val == nil {
		return 1
	}
	if i, ok := val.(int); ok {
		return i
	}
	if f, ok := val.(float64); ok {
		return int(f)
	}
	return 1
}

var (
	qqWSCancel context.CancelFunc
)

// InitQQWSServ initializes the QQ WebSocket connection based on configuration
func InitQQWSServ() {
	config, err := GetMainConfigServ()
	if err != nil {
		log.Printf("[QQ-WS] Failed to get config for initialization: %v", err)
		return
	}

	// Update manager's internal config anyway
	media.GlobalQQWSManager.UpdateConfig(config.QQ.AccessToken)

	if !config.QQ.Enabled || config.QQ.Mode != "positive" {
		StopQQWSServ()
		return
	}

	// Small delay if we just stopped it
	if qqWSCancel != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	qqWSCancel = cancel

	go func() {
		log.Printf("[QQ-WS] Starting positive reconnection loop")
		for {
			select {
			case <-ctx.Done():
				log.Printf("[QQ-WS] Positive reconnection loop stopped")
				return
			default:
				if !media.GlobalQQWSManager.IsConnected() {
					log.Printf("[QQ-WS] Attempting positive connection to %s", config.QQ.PositiveURL)
					err := media.GlobalQQWSManager.ConnectPositiveWS(config.QQ.PositiveURL, config.QQ.AccessToken)
					if err != nil {
						log.Printf("[QQ-WS] Positive connection failed: %v, retrying in 10s...", err)
					}
				}
				time.Sleep(10 * time.Second)
			}
		}
	}()
}

// StopQQWSServ stops the Positive WebSocket reconnection loop
func StopQQWSServ() {
	if qqWSCancel != nil {
		qqWSCancel()
		qqWSCancel = nil
	}
}

// RestartQQWSServ stops and restarts the QQ WebSocket service
func RestartQQWSServ() {
	StopQQWSServ()
	// Wait a bit for goroutine to exit
	time.Sleep(200 * time.Millisecond)
	InitQQWSServ()
}
