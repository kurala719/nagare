package router

import (
	"github.com/gin-gonic/gin"
	"nagare/internal/api"
	"nagare/internal/mcp"
)

func setupSystemRoutes(rg *gin.RouterGroup) {
	system := rg.Group("/system")
	system.GET("/health", api.GetHealthScoreCtrl)
	system.GET("/health/history", api.GetNetworkStatusHistoryCtrl)
	system.GET("/metrics", api.GetNetworkMetricsCtrl)
	system.GET("/status", api.GetSystemStatusCtrl)
}

func setupConfigurationRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 3 - admin only
	config := rg.Group("/config", api.PrivilegesMiddleware(3))
	config.GET("", api.GetMainConfigCtrl)
	config.PUT("", api.ModifyMainConfigCtrl)
	config.POST("/save", api.SaveConfigCtrl)
	config.POST("/reload", api.LoadConfigCtrl)
	config.POST("/reset", api.ResetConfigCtrl)
}

func setupLogRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 2
	logs := rg.Group("/logs", api.PrivilegesMiddleware(2))
	logs.GET("/system", api.GetSystemLogsCtrl)
	logs.GET("/service", api.GetServiceLogsCtrl)

	// Admin clear logs - privilege level 3
	logsAdmin := rg.Group("/logs", api.PrivilegesMiddleware(3))
	logsAdmin.DELETE("/:type", api.ClearLogsCtrl)
}

func setupAuditLogRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 2 (managers/admins)
	auditLogs := rg.Group("/audit-logs", api.PrivilegesMiddleware(2))
	auditLogs.GET("", api.SearchAuditLogsCtrl)
}

func setupAnalyticsRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 1
	analytics := rg.Group("/analytics", api.PrivilegesMiddleware(1))
	analytics.GET("/alerts", api.GetAlertAnalyticsCtrl)
}

func setupMcpRoutes(rg *gin.RouterGroup) {
	// MCP routes - requires API key middleware
	mcpGroup := rg.Group("/mcp", mcp.APIKeyMiddleware())
	mcpGroup.GET("/sse", mcp.SSEHandler)
	mcpGroup.POST("/message", mcp.MessageHandler)
}

func setupRetentionRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 3 - admin only
	retention := rg.Group("/retention", api.PrivilegesMiddleware(3))
	retention.GET("/policies", api.GetRetentionPoliciesCtrl)
	retention.POST("/policies", api.UpdateRetentionPolicyCtrl)
	retention.POST("/cleanup", api.PerformCleanupCtrl)
}
