package router

import (
	"nagare/internal/api"

	"github.com/gin-gonic/gin"
)

func setupAnalysisDomainRoutes(rg *gin.RouterGroup) {
	setupAnalyticsRoutes(rg)
	setupHistoryRoutes(rg)
	setupReportRoutes(rg)
	setupChaosRoutes(rg)
}

func setupAnalyticsRoutes(rg *gin.RouterGroup) {
	alerts := rg.Group("/alerts", api.PrivilegesMiddleware(1))
	alerts.GET("/analytics", api.GetAlertAnalyticsCtrl)
}

func setupHistoryRoutes(rg *gin.RouterGroup) {
	history := rg.Group("", api.PrivilegesMiddleware(1))
	history.GET("/monitors/:id/history", api.GetMonitorHistoryCtrl)
	history.GET("/groups/:id/history", api.GetGroupHistoryCtrl)
	history.GET("/hosts/:id/history", api.GetHostHistoryCtrl)
	history.GET("/items/:id/history", api.GetItemHistoryCtrl)
	history.GET("/system/health/history", api.GetNetworkStatusHistoryCtrl)
}

func setupReportRoutes(rg *gin.RouterGroup) {
	reports := rg.Group("/reports", api.PrivilegesMiddleware(2))
	{
		reports.GET("", api.ListReportsCtrl)
		reports.GET("/:id", api.GetReportCtrl)
		reports.GET("/:id/content", api.GetReportContentCtrl)
		reports.GET("/:id/file", api.DownloadReportCtrl)
		reports.GET("/configuration", api.GetReportConfigCtrl)
		reports.PUT("/configuration", api.UpdateReportConfigCtrl)
		reports.POST("/generations/daily", api.GenerateDailyReportCtrl)
		reports.POST("/generations/weekly", api.GenerateWeeklyReportCtrl)
		reports.POST("/generations/monthly", api.GenerateMonthlyReportCtrl)
		reports.POST("/generations/custom", api.GenerateCustomReportCtrl)
		reports.DELETE("/:id", api.DeleteReportCtrl)
	}
}

func setupChaosRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 2 (managers/admins)
	alertStorms := rg.Group("/alert-storms", api.PrivilegesMiddleware(2))
	alertStorms.POST("", api.TriggerAlertStormCtrl)
}
