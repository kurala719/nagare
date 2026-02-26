package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"nagare/internal/service"
	"nagare/internal/repository"
)

// GetConfigCtrl returns all configuration settings
func GetConfigCtrl(c *gin.Context) {
	configs := service.GetAllConfigServ()
	respondSuccess(c, http.StatusOK, configs)
}

// GetMainConfigCtrl returns the main configuration
func GetMainConfigCtrl(c *gin.Context) {
	config, err := service.GetMainConfigServ()
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, config)
}

// ModifyMainConfigCtrl modifies the main configuration
func ModifyMainConfigCtrl(c *gin.Context) {
	var req repository.ConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	// Set individual fields to ensure Viper tracks them correctly for Get calls
	repository.SetConfigValue("system.system_name", req.System.SystemName)
	repository.SetConfigValue("system.ip_address", req.System.IPAddress)
	repository.SetConfigValue("system.port", req.System.Port)
	repository.SetConfigValue("system.availability", req.System.Availability)

	repository.SetConfigValue("database.host", req.Database.Host)
	repository.SetConfigValue("database.port", req.Database.Port)
	repository.SetConfigValue("database.username", req.Database.Username)
	repository.SetConfigValue("database.password", req.Database.Password)
	repository.SetConfigValue("database.version", req.Database.Version)
	repository.SetConfigValue("database.database_name", req.Database.DatabaseName)

	repository.SetConfigValue("sync.enabled", req.Sync.Enabled)
	repository.SetConfigValue("sync.interval_seconds", req.Sync.IntervalSeconds)
	repository.SetConfigValue("sync.concurrency", req.Sync.Concurrency)

	repository.SetConfigValue("status_check.enabled", req.StatusCheck.Enabled)
	repository.SetConfigValue("status_check.provider_enabled", req.StatusCheck.ProviderEnabled)
	repository.SetConfigValue("status_check.interval_seconds", req.StatusCheck.IntervalSeconds)
	repository.SetConfigValue("status_check.concurrency", req.StatusCheck.Concurrency)

	repository.SetConfigValue("mcp.enabled", req.MCP.Enabled)
	repository.SetConfigValue("mcp.api_key", req.MCP.APIKey)
	repository.SetConfigValue("mcp.max_concurrency", req.MCP.MaxConcurrency)

	repository.SetConfigValue("ai.analysis_enabled", req.AI.AnalysisEnabled)
	repository.SetConfigValue("ai.notification_guard_enabled", req.AI.NotificationGuardEnabled)
	repository.SetConfigValue("ai.provider_id", req.AI.ProviderID)
	repository.SetConfigValue("ai.model", req.AI.Model)
	repository.SetConfigValue("ai.analysis_timeout_seconds", req.AI.AnalysisTimeoutSeconds)
	repository.SetConfigValue("ai.analysis_min_severity", req.AI.AnalysisMinSeverity)

	repository.SetConfigValue("gmail.enabled", req.Gmail.Enabled)
	repository.SetConfigValue("gmail.credentials_file", req.Gmail.CredentialsFile)
	repository.SetConfigValue("gmail.token_file", req.Gmail.TokenFile)
	repository.SetConfigValue("gmail.from", req.Gmail.From)

	repository.SetConfigValue("smtp.enabled", req.SMTP.Enabled)
	repository.SetConfigValue("smtp.host", req.SMTP.Host)
	repository.SetConfigValue("smtp.port", req.SMTP.Port)
	repository.SetConfigValue("smtp.username", req.SMTP.Username)
	repository.SetConfigValue("smtp.password", req.SMTP.Password)
	repository.SetConfigValue("smtp.from", req.SMTP.From)

	repository.SetConfigValue("qq.enabled", req.QQ.Enabled)
	repository.SetConfigValue("qq.mode", req.QQ.Mode)
	repository.SetConfigValue("qq.positive_url", req.QQ.PositiveURL)
	repository.SetConfigValue("qq.access_token", req.QQ.AccessToken)

	repository.SetConfigValue("site_message.min_alert_severity", req.SiteMessage.MinAlertSeverity)
	repository.SetConfigValue("site_message.min_log_severity", req.SiteMessage.MinLogSeverity)

	repository.SetConfigValue("media_rate_limit.global_interval_seconds", req.MediaRateLimit.GlobalIntervalSeconds)
	repository.SetConfigValue("media_rate_limit.protocol_interval_seconds", req.MediaRateLimit.ProtocolIntervalSeconds)
	repository.SetConfigValue("media_rate_limit.media_interval_seconds", req.MediaRateLimit.MediaIntervalSeconds)

	repository.SetConfigValue("external", req.External)

	if err := repository.SaveConfig(); err != nil {
		respondError(c, err)
		return
	}

	// Trigger hot reload for background services
	repository.NotifyConfigObservers()

	respondSuccessMessage(c, http.StatusOK, "configuration updated")
}

// ModifyConfig modifies arbitrary configuration values
func ModifyConfig(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	for key, value := range req {
		repository.SetConfigValue(key, value)
	}
	if err := repository.SaveConfig(); err != nil {
		respondError(c, err)
		return
	}
	// Trigger hot reload
	repository.NotifyConfigObservers()
	respondSuccessMessage(c, http.StatusOK, "configuration updated")
}

// SaveConfigCtrl saves the current configuration
func SaveConfigCtrl(c *gin.Context) {
	if err := repository.SaveConfig(); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "configuration saved")
}

// LoadConfigCtrl reloads the configuration
func LoadConfigCtrl(c *gin.Context) {
	if err := repository.LoadConfig(); err != nil {
		respondError(c, err)
		return
	}
	// Trigger hot reload after manual reload
	repository.NotifyConfigObservers()
	respondSuccessMessage(c, http.StatusOK, "configuration loaded")
}

// ResetConfigCtrl resets the configuration to defaults
func ResetConfigCtrl(c *gin.Context) {
	if err := service.ResetConfigServ(); err != nil {
		respondError(c, err)
		return
	}
	// Trigger hot reload
	repository.NotifyConfigObservers()
	respondSuccessMessage(c, http.StatusOK, "configuration reset to defaults")
}

// InitConfigCtrl initializes configuration from a path
func InitConfigCtrl(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	if err := repository.InitConfig(req.Path); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "configuration initialized")
}
