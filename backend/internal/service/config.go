package service

import "nagare/internal/repository"

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

// SaveConfigServ persists configuration to disk
func SaveConfigServ() error {
	return repository.SaveConfig()
}
