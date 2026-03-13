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
	// Routes with privilege level 3 - admin only
	config := rg.Group("/config", api.PrivilegesMiddleware(3))
	config.GET("", api.GetMainConfigCtrl)
	config.PUT("", api.ModifyMainConfigCtrl)
	config.DELETE("", api.ResetConfigCtrl)

	configSnapshots := rg.Group("/config-snapshots", api.PrivilegesMiddleware(3))
	configSnapshots.POST("", api.SaveConfigCtrl)

	configReloads := rg.Group("/config-reloads", api.PrivilegesMiddleware(3))
	configReloads.POST("", api.LoadConfigCtrl)
}

func setupLogRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 3
	logs := rg.Group("/logs", api.PrivilegesMiddleware(3))
	logs.GET("/system", api.GetSystemLogsCtrl)

	// Admin clear logs - privilege level 3
	logsAdmin := rg.Group("/logs", api.PrivilegesMiddleware(3))
	logsAdmin.DELETE("/:type", api.ClearLogsCtrl)
}

func setupAuditLogRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 3
	auditLogs := rg.Group("/audit-logs", api.PrivilegesMiddleware(3))
	auditLogs.GET("", api.SearchAuditLogsCtrl)
}

func setupRetentionRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 3 - admin only
	retention := rg.Group("/retention", api.PrivilegesMiddleware(3))
	retention.GET("/policies", api.GetRetentionPoliciesCtrl)
	retention.PUT("/policies", api.UpdateRetentionPolicyCtrl)
	retention.POST("/jobs", api.PerformCleanupCtrl)
}
