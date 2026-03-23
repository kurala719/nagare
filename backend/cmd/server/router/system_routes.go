package router

import (
	"nagare/internal/api"

	"github.com/gin-gonic/gin"
)

func setupSystemDomainRoutes(rg *gin.RouterGroup) {
	setupLogRoutes(rg)
	setupAuditLogRoutes(rg)
	setupRetentionRoutes(rg)
	setupConfigurationRoutes(rg)
	setupSystemRoutes(rg)
}

func setupSystemRoutes(rg *gin.RouterGroup) {
	rg.GET("/health", api.GetHealthScoreCtrl)
	rg.GET("/metrics", api.GetNetworkMetricsCtrl)
	rg.GET("/status", api.GetSystemStatusCtrl)
}

func setupConfigurationRoutes(rg *gin.RouterGroup) {
	config := rg.Group("/config", api.PrivilegesMiddleware(3))
	config.GET("", api.GetMainConfigCtrl)
	config.PUT("", api.ModifyMainConfigCtrl)
	config.DELETE("", api.ResetConfigCtrl)
	config.GET("/mcp/clients", api.GetMCPClientStatusCtrl)
	config.POST("/mcp/test", api.TestMCPClientCtrl)

	configSnapshots := rg.Group("/config-snapshots", api.PrivilegesMiddleware(3))
	configSnapshots.POST("", api.SaveConfigCtrl)

	configReloads := rg.Group("/config-reloads", api.PrivilegesMiddleware(3))
	configReloads.POST("", api.LoadConfigCtrl)
}

func setupLogRoutes(rg *gin.RouterGroup) {
	logs := rg.Group("/logs", api.PrivilegesMiddleware(3))
	logs.GET("/system", api.GetSystemLogsCtrl)
	logs.DELETE("/:type", api.ClearLogsCtrl)
}

func setupAuditLogRoutes(rg *gin.RouterGroup) {
	auditLogs := rg.Group("/audit-logs", api.PrivilegesMiddleware(3))
	auditLogs.GET("", api.SearchAuditLogsCtrl)
}

func setupRetentionRoutes(rg *gin.RouterGroup) {
	retention := rg.Group("/retention", api.PrivilegesMiddleware(3))
	retention.GET("/policies", api.GetRetentionPoliciesCtrl)
	retention.PUT("/policies", api.UpdateRetentionPolicyCtrl)
	retention.POST("/jobs", api.PerformCleanupCtrl)
}
