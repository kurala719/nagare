package repository

import (
	"strings"

	"github.com/spf13/viper"
)

var ConfigPath string

// Config represents the main application configuration
type Config struct {
	AI             AIConfig             `yaml:"ai" json:"ai" mapstructure:"ai"`
	Gmail          GmailConfig          `yaml:"gmail" json:"gmail" mapstructure:"gmail"`
	SMTP           SMTPConfig           `yaml:"smtp" json:"smtp" mapstructure:"smtp"`
	QQ             QQConfig             `yaml:"qq" json:"qq" mapstructure:"qq"`
	SiteMessage    SiteMessageConfig    `yaml:"site_message" json:"site_message" mapstructure:"site_message"`
	MediaRateLimit MediaRateLimitConfig `yaml:"media_rate_limit" json:"media_rate_limit" mapstructure:"media_rate_limit"`
	External       []ExternalItemConfig `yaml:"external" json:"external" mapstructure:"external"`
}

// SiteMessageConfig holds internal notification settings
type SiteMessageConfig struct {
	MinAlertSeverity int `yaml:"min_alert_severity" json:"min_alert_severity" mapstructure:"min_alert_severity"`
	MinLogSeverity   int `yaml:"min_log_severity" json:"min_log_severity" mapstructure:"min_log_severity"`
}

// SMTPConfig holds SMTP server settings
type SMTPConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	Host     string `yaml:"host" json:"host" mapstructure:"host"`
	Port     int    `yaml:"port" json:"port" mapstructure:"port"`
	Username string `yaml:"username" json:"username" mapstructure:"username"`
	Password string `yaml:"password" json:"password" mapstructure:"password"`
	From     string `yaml:"from" json:"from" mapstructure:"from"`
}

// QQConfig holds OneBot/NapCat WebSocket settings
type QQConfig struct {
	Enabled     bool   `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	Mode        string `yaml:"mode" json:"mode" mapstructure:"mode"` // "reverse" or "positive"
	PositiveURL string `yaml:"positive_url" json:"positive_url" mapstructure:"positive_url"`
	AccessToken string `yaml:"access_token" json:"access_token" mapstructure:"access_token"`
}

// GmailConfig holds Gmail API settings
type GmailConfig struct {
	Enabled         bool   `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	CredentialsFile string `yaml:"credentials_file" json:"credentials_file" mapstructure:"credentials_file"`
	TokenFile       string `yaml:"token_file" json:"token_file" mapstructure:"token_file"`
	From            string `yaml:"from" json:"from" mapstructure:"from"`
}

// ExternalItemConfig defines external infrastructure items
type ExternalItemConfig struct {
	Type string `yaml:"type" json:"type" mapstructure:"type"` // monitor, alarm, provider, media
	Key  string `yaml:"key" json:"key" mapstructure:"key"`    // unique identifier
	Name string `yaml:"name" json:"name" mapstructure:"name"` // display name
	ID   int    `yaml:"id" json:"id" mapstructure:"id"`       // numeric value used in DB
}

// SystemConfig holds system-level settings
type SystemConfig struct {
	SystemName   string `yaml:"system_name" json:"system_name" mapstructure:"system_name"`
	IPAddress    string `yaml:"ip_address" json:"ip_address" mapstructure:"ip_address"`
	Port         int    `yaml:"port" json:"port" mapstructure:"port"`
	Availability bool   `yaml:"availability" json:"availability" mapstructure:"availability"`
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Version      string `yaml:"version" json:"version" mapstructure:"version"`
	Host         string `yaml:"host" json:"host" mapstructure:"host"`
	Port         int    `yaml:"port" json:"port" mapstructure:"port"`
	Username     string `yaml:"username" json:"username" mapstructure:"username"`
	Password     string `yaml:"password" json:"password" mapstructure:"password"`
	DatabaseName string `yaml:"database_name" json:"database_name" mapstructure:"database_name"`
}

// SyncConfig holds background sync settings
type SyncConfig struct {
	Enabled         bool `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	IntervalSeconds int  `yaml:"interval_seconds" json:"interval_seconds" mapstructure:"interval_seconds"`
	Concurrency     int  `yaml:"concurrency" json:"concurrency" mapstructure:"concurrency"`
}

// StatusCheckConfig holds status check settings
type StatusCheckConfig struct {
	Enabled         bool `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	ProviderEnabled bool `yaml:"provider_enabled" json:"provider_enabled" mapstructure:"provider_enabled"`
	IntervalSeconds int  `yaml:"interval_seconds" json:"interval_seconds" mapstructure:"interval_seconds"`
	Concurrency     int  `yaml:"concurrency" json:"concurrency" mapstructure:"concurrency"`
}

// MCPConfig holds MCP settings
type MCPConfig struct {
	Enabled        bool   `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	APIKey         string `yaml:"api_key" json:"api_key" mapstructure:"api_key"`
	MaxConcurrency int    `yaml:"max_concurrency" json:"max_concurrency" mapstructure:"max_concurrency"`
}

// AIConfig holds AI settings
type AIConfig struct {
	AnalysisEnabled          bool   `yaml:"analysis_enabled" json:"analysis_enabled" mapstructure:"analysis_enabled"`
	NotificationGuardEnabled bool   `yaml:"notification_guard_enabled" json:"notification_guard_enabled" mapstructure:"notification_guard_enabled"`
	ProviderID               int    `yaml:"provider_id" json:"provider_id" mapstructure:"provider_id"`
	Model                    string `yaml:"model" json:"model" mapstructure:"model"`
	AnalysisTimeoutSeconds   int    `yaml:"analysis_timeout_seconds" json:"analysis_timeout_seconds" mapstructure:"analysis_timeout_seconds"`
	AnalysisMinSeverity      int    `yaml:"analysis_min_severity" json:"analysis_min_severity" mapstructure:"analysis_min_severity"`
}

// MediaRateLimitConfig holds notification rate limit settings
type MediaRateLimitConfig struct {
	GlobalIntervalSeconds   int `yaml:"global_interval_seconds" json:"global_interval_seconds" mapstructure:"global_interval_seconds"`
	ProtocolIntervalSeconds int `yaml:"protocol_interval_seconds" json:"protocol_interval_seconds" mapstructure:"protocol_interval_seconds"`
	MediaIntervalSeconds    int `yaml:"media_interval_seconds" json:"media_interval_seconds" mapstructure:"media_interval_seconds"`
}

// ConfigRequest is used for modifying configuration
type ConfigRequest struct {
	System   SystemConfig `yaml:"system" json:"system" mapstructure:"system"`
	Database struct {
		Host         string `yaml:"host" json:"host" mapstructure:"host"`
		Port         int    `yaml:"port" json:"port" mapstructure:"port"`
		Username     string `yaml:"username" json:"username" mapstructure:"username"`
		Password     string `yaml:"password" json:"password" mapstructure:"password"`
		Version      string `yaml:"version" json:"version" mapstructure:"version"`
		DatabaseName string `yaml:"database_name" json:"database_name" mapstructure:"database_name"`
	} `yaml:"database" json:"database" mapstructure:"database"`
	Sync           SyncConfig           `yaml:"sync" json:"sync" mapstructure:"sync"`
	StatusCheck    StatusCheckConfig    `yaml:"status_check" json:"status_check" mapstructure:"status_check"`
	MCP            MCPConfig            `yaml:"mcp" json:"mcp" mapstructure:"mcp"`
	AI             AIConfig             `yaml:"ai" json:"ai" mapstructure:"ai"`
	Gmail          GmailConfig          `yaml:"gmail" json:"gmail" mapstructure:"gmail"`
	SMTP           SMTPConfig           `yaml:"smtp" json:"smtp" mapstructure:"smtp"`
	QQ             QQConfig             `yaml:"qq" json:"qq" mapstructure:"qq"`
	SiteMessage    SiteMessageConfig    `yaml:"site_message" json:"site_message" mapstructure:"site_message"`
	MediaRateLimit MediaRateLimitConfig `yaml:"media_rate_limit" json:"media_rate_limit" mapstructure:"media_rate_limit"`
	External       []ExternalItemConfig `yaml:"external" json:"external" mapstructure:"external"`
}

// ConfigResponse represents the configuration response
type ConfigResponse struct {
	System         SystemConfig         `yaml:"system" json:"system" mapstructure:"system"`
	Database       DatabaseConfig       `yaml:"database" json:"database" mapstructure:"database"`
	Sync           SyncConfig           `yaml:"sync" json:"sync" mapstructure:"sync"`
	StatusCheck    StatusCheckConfig    `yaml:"status_check" json:"status_check" mapstructure:"status_check"`
	MCP            MCPConfig            `yaml:"mcp" json:"mcp" mapstructure:"mcp"`
	AI             AIConfig             `yaml:"ai" json:"ai" mapstructure:"ai"`
	Gmail          GmailConfig          `yaml:"gmail" json:"gmail" mapstructure:"gmail"`
	SMTP           SMTPConfig           `yaml:"smtp" json:"smtp" mapstructure:"smtp"`
	QQ             QQConfig             `yaml:"qq" json:"qq" mapstructure:"qq"`
	SiteMessage    SiteMessageConfig    `yaml:"site_message" json:"site_message" mapstructure:"site_message"`
	MediaRateLimit MediaRateLimitConfig `yaml:"media_rate_limit" json:"media_rate_limit" mapstructure:"media_rate_limit"`
	External       []ExternalItemConfig `yaml:"external" json:"external" mapstructure:"external"`
}

// InitConfig initializes the configuration from the given path
func InitConfig(path string) error {
	ConfigPath = path
	viper.SetConfigFile(ConfigPath)

	// Enable environment variable support
	viper.SetEnvPrefix("NAGARE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("status_check.provider_enabled", false)
	
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	viper.OnConfigChange(func(e fsnotify.Event) {
		LogConfigChange(e.Name)
		NotifyConfigObservers()
	})
	viper.WatchConfig()

	return nil
}

// ConfigObserver defines a function that reacts to config changes
type ConfigObserver func()

var observers []ConfigObserver

// RegisterConfigObserver adds a listener for configuration changes
func RegisterConfigObserver(observer ConfigObserver) {
	observers = append(observers, observer)
}

// NotifyConfigObservers triggers all registered callbacks
func NotifyConfigObservers() {
	for _, observer := range observers {
		observer()
	}
}

// LogConfigChange is a placeholder for logging, will be linked to service later
var LogConfigChange = func(path string) {}

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

// ResetConfig resets the configuration to default values
func ResetConfig() error {
	// Re-apply all defaults
	viper.Set("system.system_name", "Nagare System")
	viper.Set("system.ip_address", "127.0.0.1")
	viper.Set("system.port", 8080)
	viper.Set("system.availability", true)

	viper.Set("database.version", "MYSQL 8.0")
	viper.Set("database.host", "127.0.0.1")
	viper.Set("database.port", 3306)
	viper.Set("database.username", "nagare")
	viper.Set("database.password", "")
	viper.Set("database.database_name", "nagare")

	viper.Set("sync.enabled", true)
	viper.Set("sync.interval_seconds", 60)
	viper.Set("sync.concurrency", 2)

	viper.Set("status_check.enabled", true)
	viper.Set("status_check.provider_enabled", false)
	viper.Set("status_check.interval_seconds", 60)
	viper.Set("status_check.concurrency", 4)

	viper.Set("mcp.enabled", true)
	viper.Set("mcp.api_key", "")
	viper.Set("mcp.max_concurrency", 4)

	viper.Set("ai.analysis_enabled", true)
	viper.Set("ai.notification_guard_enabled", false)
	viper.Set("ai.provider_id", 1)
	viper.Set("ai.model", "")
	viper.Set("ai.analysis_timeout_seconds", 60)
	viper.Set("ai.analysis_min_severity", 2)

	viper.Set("gmail.enabled", false)
	viper.Set("gmail.credentials_file", "configs/gmail_credentials.json")
	viper.Set("gmail.token_file", "configs/gmail_token.json")
	viper.Set("gmail.from", "")

	viper.Set("smtp.enabled", false)
	viper.Set("smtp.host", "")
	viper.Set("smtp.port", 587)
	viper.Set("smtp.username", "")
	viper.Set("smtp.password", "")
	viper.Set("smtp.from", "")

	viper.Set("qq.enabled", false)
	viper.Set("qq.mode", "reverse")
	viper.Set("qq.positive_url", "ws://localhost:3001")
	viper.Set("qq.access_token", "")

	viper.Set("site_message.min_alert_severity", 0)
	viper.Set("site_message.min_log_severity", 1)

	viper.Set("media_rate_limit.global_interval_seconds", 30)
	viper.Set("media_rate_limit.protocol_interval_seconds", 30)
	viper.Set("media_rate_limit.media_interval_seconds", 30)

	viper.Set("external", []map[string]interface{}{
		{"type": "monitor", "key": "snmp", "name": "SNMP", "id": 1},
		{"type": "monitor", "key": "zabbix", "name": "Zabbix", "id": 2},
		{"type": "monitor", "key": "other", "name": "Other", "id": 3},
		{"type": "alarm", "key": "zabbix", "name": "Zabbix", "id": 1},
		{"type": "alarm", "key": "other", "name": "Other", "id": 2},
		{"type": "provider", "key": "gemini", "name": "Gemini", "id": 1},
		{"type": "provider", "key": "openai", "name": "OpenAI", "id": 2},
		{"type": "provider", "key": "other", "name": "Other", "id": 3},
		{"type": "media", "key": "gmail", "name": "Gmail", "id": 1},
		{"type": "media", "key": "qq", "name": "QQ", "id": 2},
		{"type": "media", "key": "other", "name": "Other", "id": 3},
	})

	return SaveConfig()
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
