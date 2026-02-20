package router

import (
	"nagare/internal/adapter/handler"
	"nagare/internal/mcp"

	"github.com/gin-gonic/gin"
)

func setupSystemRoutes(rg *gin.RouterGroup) {
	system := rg.Group("/system")
	system.GET("/health", handler.GetHealthScoreCtrl)
	system.GET("/health/history", handler.GetNetworkStatusHistoryCtrl)
	system.GET("/metrics", handler.GetNetworkMetricsCtrl)
	system.GET("/status", handler.GetSystemStatusCtrl)
}

func setupConfigurationRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 3 - admin only
	config := rg.Group("/config", handler.PrivilegesMiddleware(3))
	config.GET("", handler.GetMainConfigCtrl)
	config.PUT("", handler.ModifyMainConfigCtrl)
	config.POST("/save", handler.SaveConfigCtrl)
	config.POST("/reload", handler.LoadConfigCtrl)
	config.POST("/reset", handler.ResetConfigCtrl)
}

func setupLogRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 2
	logs := rg.Group("/logs", handler.PrivilegesMiddleware(2))
	logs.GET("/system", handler.GetSystemLogsCtrl)
	logs.GET("/service", handler.GetServiceLogsCtrl)

	// Admin clear logs - privilege level 3
	logsAdmin := rg.Group("/logs", handler.PrivilegesMiddleware(3))
	logsAdmin.DELETE("/:type", handler.ClearLogsCtrl)
}

func setupAuditLogRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 2 (managers/admins)
	auditLogs := rg.Group("/audit-logs", handler.PrivilegesMiddleware(2))
	auditLogs.GET("", handler.SearchAuditLogsCtrl)
}

func setupAnalyticsRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 1
	analytics := rg.Group("/analytics", handler.PrivilegesMiddleware(1))
	analytics.GET("/alerts", handler.GetAlertAnalyticsCtrl)
}

func setupMcpRoutes(rg *gin.RouterGroup) {
	// MCP routes - requires API key middleware
	mcpGroup := rg.Group("/mcp", mcp.APIKeyMiddleware())
	mcpGroup.GET("/sse", mcp.SSEHandler)
	mcpGroup.POST("/message", mcp.MessageHandler)
}

func setupRetentionRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 3 - admin only
	retention := rg.Group("/retention", handler.PrivilegesMiddleware(3))
	retention.GET("/policies", handler.GetRetentionPoliciesCtrl)
	retention.POST("/policies", handler.UpdateRetentionPolicyCtrl)
	retention.POST("/cleanup", handler.PerformCleanupCtrl)
}
