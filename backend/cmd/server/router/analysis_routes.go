package router

import (
	"nagare/internal/api"

	"github.com/gin-gonic/gin"
)

func setupAnalysisDomainRoutes(rg *gin.RouterGroup) {
	setupAnalyticsRoutes(rg)
	setupHistoryRoutes(rg)
	setupReportRoutes(rg)
	setupReportConfigRoutes(rg)
	setupReportGenerationRoutes(rg)
	setupReportDownloadRoutes(rg)
	setupChaosRoutes(rg)
}

func setupAnalyticsRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 1
	analytics := rg.Group("/analytics", api.PrivilegesMiddleware(1))
	analytics.GET("/alerts", api.GetAlertAnalyticsCtrl)
}

func setupHistoryRoutes(rg *gin.RouterGroup) {
	history := rg.Group("/history", api.PrivilegesMiddleware(1))
	history.GET("/monitors/:id", api.GetMonitorHistoryCtrl)
	history.GET("/groups/:id", api.GetGroupHistoryCtrl)
	history.GET("/hosts/:id", api.GetHostHistoryCtrl)
	history.GET("/items/:id", api.GetItemHistoryCtrl)
	history.GET("/system/health", api.GetNetworkStatusHistoryCtrl)
}

func setupReportRoutes(rg *gin.RouterGroup) {
	reports := rg.Group("/reports", api.PrivilegesMiddleware(2))
	{
		reports.GET("", api.ListReportsCtrl)
		reports.GET("/:id", api.GetReportCtrl)
		reports.GET("/:id/content", api.GetReportContentCtrl)
		reports.DELETE("/:id", api.DeleteReportCtrl)
	}
}

func setupReportConfigRoutes(rg *gin.RouterGroup) {
	reportConfig := rg.Group("/report-config", api.PrivilegesMiddleware(2))
	reportConfig.GET("", api.GetReportConfigCtrl)
	reportConfig.PUT("", api.UpdateReportConfigCtrl)
}

func setupReportGenerationRoutes(rg *gin.RouterGroup) {
	reportGeneration := rg.Group("/report-generation", api.PrivilegesMiddleware(2))
	reportGeneration.POST("/daily", api.GenerateDailyReportCtrl)
	reportGeneration.POST("/weekly", api.GenerateWeeklyReportCtrl)
	reportGeneration.POST("/monthly", api.GenerateMonthlyReportCtrl)
	reportGeneration.POST("/custom", api.GenerateCustomReportCtrl)
}

func setupReportDownloadRoutes(rg *gin.RouterGroup) {
	reportDownload := rg.Group("/report-download", api.PrivilegesMiddleware(2))
	reportDownload.GET("/:id", api.DownloadReportCtrl)
}

func setupChaosRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 2 (managers/admins)
	chaos := rg.Group("/chaos", api.PrivilegesMiddleware(2))
	chaos.POST("/alert-storm", api.TriggerAlertStormCtrl)
}
