package repository

import (
	"strings"

	"github.com/spf13/viper"
)

var ConfigPath string

// Config represents the main application configuration
type Config struct {
	System         SystemConfig         `yaml:"system" json:"system"`
	Database       DatabaseConfig       `yaml:"database" json:"database"`
	Sync           SyncConfig           `yaml:"sync" json:"sync"`
	StatusCheck    StatusCheckConfig    `yaml:"status_check" json:"status_check"`
	MCP            MCPConfig            `yaml:"mcp" json:"mcp"`
	AI             AIConfig             `yaml:"ai" json:"ai"`
	MediaRateLimit MediaRateLimitConfig `yaml:"media_rate_limit" json:"media_rate_limit"`
}

// SystemConfig holds system-level settings
type SystemConfig struct {
	SystemName   string `yaml:"system_name" json:"system_name"`
	IPAddress    string `yaml:"ip_address" json:"ip_address"`
	Port         int    `yaml:"port" json:"port"`
	Availability bool   `yaml:"availability" json:"availability"`
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Version      string `yaml:"version" json:"version"`
	Host         string `yaml:"host" json:"host"`
	Port         int    `yaml:"port" json:"port"`
	Username     string `yaml:"username" json:"username"`
	Password     string `yaml:"password" json:"password"`
	DatabaseName string `yaml:"database_name" json:"database_name"`
}

// SyncConfig holds background sync settings
type SyncConfig struct {
	Enabled         bool `yaml:"enabled" json:"enabled"`
	IntervalSeconds int  `yaml:"interval_seconds" json:"interval_seconds"`
	Concurrency     int  `yaml:"concurrency" json:"concurrency"`
}

// StatusCheckConfig holds status check settings
type StatusCheckConfig struct {
	Enabled         bool `yaml:"enabled" json:"enabled"`
	ProviderEnabled bool `yaml:"provider_enabled" json:"provider_enabled"`
	IntervalSeconds int  `yaml:"interval_seconds" json:"interval_seconds"`
	Concurrency     int  `yaml:"concurrency" json:"concurrency"`
}

// MCPConfig holds MCP settings
type MCPConfig struct {
	Enabled        bool   `yaml:"enabled" json:"enabled"`
	APIKey         string `yaml:"api_key" json:"api_key"`
	MaxConcurrency int    `yaml:"max_concurrency" json:"max_concurrency"`
}

// AIConfig holds AI settings
type AIConfig struct {
	AnalysisEnabled        bool   `yaml:"analysis_enabled" json:"analysis_enabled"`
	ProviderID             int    `yaml:"provider_id" json:"provider_id"`
	Model                  string `yaml:"model" json:"model"`
	AnalysisTimeoutSeconds int    `yaml:"analysis_timeout_seconds" json:"analysis_timeout_seconds"`
	AnalysisMinSeverity    int    `yaml:"analysis_min_severity" json:"analysis_min_severity"`
}

// MediaRateLimitConfig holds notification rate limit settings
type MediaRateLimitConfig struct {
	GlobalIntervalSeconds    int `yaml:"global_interval_seconds" json:"global_interval_seconds"`
	MediaTypeIntervalSeconds int `yaml:"media_type_interval_seconds" json:"media_type_interval_seconds"`
	MediaIntervalSeconds     int `yaml:"media_interval_seconds" json:"media_interval_seconds"`
}

// ConfigRequest is used for modifying configuration
type ConfigRequest struct {
	System   SystemConfig `yaml:"system" json:"system"`
	Database struct {
		Host         string `yaml:"host" json:"host"`
		Port         int    `yaml:"port" json:"port"`
		Username     string `yaml:"username" json:"username"`
		Password     string `yaml:"password" json:"password"`
		Version      string `yaml:"version" json:"version"`
		DatabaseName string `yaml:"database_name" json:"database_name"`
	} `yaml:"database" json:"database"`
	Sync           SyncConfig           `yaml:"sync" json:"sync"`
	StatusCheck    StatusCheckConfig    `yaml:"status_check" json:"status_check"`
	MCP            MCPConfig            `yaml:"mcp" json:"mcp"`
	AI             AIConfig             `yaml:"ai" json:"ai"`
	MediaRateLimit MediaRateLimitConfig `yaml:"media_rate_limit" json:"media_rate_limit"`
}

// ConfigResponse represents the configuration response
type ConfigResponse struct {
	System         SystemConfig         `yaml:"system" json:"system"`
	Database       DatabaseConfig       `yaml:"database" json:"database"`
	Sync           SyncConfig           `yaml:"sync" json:"sync"`
	StatusCheck    StatusCheckConfig    `yaml:"status_check" json:"status_check"`
	MCP            MCPConfig            `yaml:"mcp" json:"mcp"`
	AI             AIConfig             `yaml:"ai" json:"ai"`
	MediaRateLimit MediaRateLimitConfig `yaml:"media_rate_limit" json:"media_rate_limit"`
}

// InitConfig initializes the configuration from the given path
func InitConfig(path string) error {
	ConfigPath = path
	viper.SetConfigFile(ConfigPath)

	// Enable environment variable support
	viper.SetEnvPrefix("NAGARE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("status_check.provider_enabled", true)
	return viper.ReadInConfig()
}

// LoadConfig reloads the configuration file
func LoadConfig() error {
	viper.SetConfigName("nagare_config")
	viper.AddConfigPath(ConfigPath)
	return viper.ReadInConfig()
}

// SaveConfig persists the current configuration to disk
func SaveConfig() error {
	return viper.WriteConfig()
}

// GetConfigValue retrieves a configuration value by key
func GetConfigValue(key string) interface{} {
	return viper.Get(key)
}

// SetConfigValue updates a configuration value
func SetConfigValue(key string, value interface{}) {
	viper.Set(key, value)
}

// GetAllConfig returns all configuration settings
func GetAllConfig() map[string]interface{} {
	return viper.AllSettings()
}

// GetMainConfig returns the main configuration structure
func GetMainConfig() (ConfigResponse, error) {
	var config ConfigResponse
	if err := viper.Unmarshal(&config); err != nil {
		return ConfigResponse{}, err
	}
	return config, nil
}
