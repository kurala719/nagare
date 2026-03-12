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
	config.POST("/save", api.SaveConfigCtrl)
	config.POST("/reload", api.LoadConfigCtrl)
	config.POST("/reset", api.ResetConfigCtrl)
}

func setupLogRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 2
	logs := rg.Group("/logs", api.PrivilegesMiddleware(2))
	logs.GET("/system", api.GetSystemLogsCtrl)

	// Admin clear logs - privilege level 3
	logsAdmin := rg.Group("/logs", api.PrivilegesMiddleware(3))
	logsAdmin.DELETE("/:type", api.ClearLogsCtrl)
}

func setupAuditLogRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 2 (managers/admins)
	auditLogs := rg.Group("/audit-logs", api.PrivilegesMiddleware(2))
	auditLogs.GET("", api.SearchAuditLogsCtrl)
}

func setupRetentionRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 3 - admin only
	retention := rg.Group("/retention", api.PrivilegesMiddleware(3))
	retention.GET("/policies", api.GetRetentionPoliciesCtrl)
	retention.POST("/policies", api.UpdateRetentionPolicyCtrl)
	retention.POST("/cleanup", api.PerformCleanupCtrl)
}
