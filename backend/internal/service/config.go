package service

import (
	"nagare/internal/database"
	"nagare/internal/repository"
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
	repository.LogConfigChange = func(path string) {
		LogSystem("info", "configuration file changed", map[string]interface{}{"path": path}, nil, "")
	}

	// Register observers for hot reload
	repository.RegisterConfigObserver(func() {
		LogSystem("info", "restarting services after configuration change", nil, nil, "")
		RestartAutoSync()
		RestartStatusChecks()
		if err := database.ReapplyPoolSettings(); err != nil {
			LogSystem("error", "failed to reapply database pool settings", map[string]interface{}{"error": err.Error()}, nil, "")
		}
	})

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
	if !repository.IsConfigSet("site_message.min_alert_severity") {
		return 0 // Default: show all alerts
	}
	return repository.GetConfigInt("site_message.min_alert_severity")
}

func siteMessageMinLogSeverity() int {
	if !repository.IsConfigSet("site_message.min_log_severity") {
		return 1 // Default: show warnings and errors
	}
	return repository.GetConfigInt("site_message.min_log_severity")
}
