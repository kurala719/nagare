package service

import (
	"context"
	"fmt"

	"nagare/internal/adapter/external/monitors"
	"nagare/internal/adapter/repository"

	"github.com/spf13/viper"
)

// SetupAlarmMediaServ configures the external alarm source to push to this Nagare instance.
func SetupAlarmMediaServ(ctx context.Context, alarmID uint) error {
	alarm, err := repository.GetAlarmByIDDAO(alarmID)
	if err != nil {
		return fmt.Errorf("failed to get alarm: %w", err)
	}

	// We only support automated setup for Zabbix right now
	if alarm.Type != 1 { // 1 = Zabbix
		return fmt.Errorf("automated media setup is only supported for Zabbix alarms")
	}

	cfg := monitors.Config{
		Auth: monitors.AuthConfig{
			URL:      alarm.URL,
			Username: alarm.Username,
			Password: alarm.Password,
			Token:    alarm.AuthToken,
		},
		Timeout: 30,
	}

	provider, err := monitors.NewZabbixProvider(cfg)
	if err != nil {
		return fmt.Errorf("failed to create zabbix provider: %w", err)
	}

	// Determine our own public URL or IP to give to Zabbix
	baseURL := viper.GetString("system.external_url")
	if baseURL == "" {
		// fallback to ip_address and port
		ip := viper.GetString("system.ip_address")
		port := viper.GetInt("system.port")
		if ip == "" {
			ip = "127.0.0.1" // This won't work for real external Zabbix, but it's a fallback
		}
		if port == 0 {
			port = 8080
		}
		baseURL = fmt.Sprintf("http://%s:%d", ip, port)
	}

	webhookURL := fmt.Sprintf("%s/api/v1/alerts/webhook", baseURL)

	err = provider.SetupWebhookMedia(ctx, webhookURL, alarm.EventToken)
	if err != nil {
		return fmt.Errorf("failed to setup zabbix webhook: %w", err)
	}

	return nil
}
