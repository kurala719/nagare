package handler

import (
	"net/http"

	"nagare/internal/adapter/repository"
	"nagare/internal/core/service"

	"github.com/gin-gonic/gin"
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
	repository.SetConfigValue("system", req.System)
	repository.SetConfigValue("database", req.Database)
	repository.SetConfigValue("sync", req.Sync)
	repository.SetConfigValue("status_check", req.StatusCheck)
	repository.SetConfigValue("mcp", req.MCP)
	repository.SetConfigValue("ai", req.AI)
	repository.SetConfigValue("media_rate_limit", req.MediaRateLimit)

	if err := repository.SaveConfig(); err != nil {
		respondError(c, err)
		return
	}
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
	respondSuccessMessage(c, http.StatusOK, "configuration loaded")
}

// ResetConfigCtrl resets the configuration to defaults
func ResetConfigCtrl(c *gin.Context) {
	if err := service.ResetConfigServ(); err != nil {
		respondError(c, err)
		return
	}
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
