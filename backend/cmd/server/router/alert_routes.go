package router

import (
	"nagare/internal/api"

	"github.com/gin-gonic/gin"
)

func setupAlertDomainRoutes(rg *gin.RouterGroup) {
	setupAlarmRoutes(rg)
	setupAlertRoutes(rg)
	setupTriggerRoutes(rg)
}

func setupAlarmRoutes(rg *gin.RouterGroup) {
	alarms := rg.Group("/alarms")
	{
		// Public event token refresh - no auth required, uses event token
		alarms.POST("/:id/event-token/refresh", api.RefreshAlarmEventTokenCtrl)

		// Privilege level 1
		alarms.GET("", api.PrivilegesMiddleware(1), api.SearchAlarmsCtrl)
		alarms.GET("/:id", api.PrivilegesMiddleware(1), api.GetAlarmByIDCtrl)

		// Privilege level 2
		alarms.POST("", api.PrivilegesMiddleware(2), api.AddAlarmCtrl)
		alarms.DELETE("/:id", api.PrivilegesMiddleware(2), api.DeleteAlarmByIDCtrl)
		alarms.PUT("/:id", api.PrivilegesMiddleware(2), api.UpdateAlarmCtrl)
		alarms.POST("/:id/login", api.PrivilegesMiddleware(2), api.LoginAlarmCtrl)
		alarms.POST(":id/setup-media", api.PrivilegesMiddleware(2), api.SetupAlarmMediaCtrl)
		alarms.POST("/:id/event-token", api.PrivilegesMiddleware(2), api.RegenerateAlarmEventTokenCtrl)
	}
}

func setupTriggerRoutes(rg *gin.RouterGroup) {
	triggers := rg.Group("/triggers")
	{
		// Routes with privilege level 1
		triggersRead := triggers.Group("", api.PrivilegesMiddleware(1))
		triggersRead.GET("", api.SearchTriggersCtrl)
		triggersRead.GET("/:id", api.GetTriggerByIDCtrl)

		// Routes with privilege level 2
		triggersWrite := triggers.Group("", api.PrivilegesMiddleware(2))
		triggersWrite.POST("", api.AddTriggerCtrl)
		triggersWrite.PUT("/:id", api.UpdateTriggerCtrl)
		triggersWrite.DELETE("/:id", api.DeleteTriggerByIDCtrl)
	}
}

func setupAlertRoutes(rg *gin.RouterGroup) {
	// Webhook endpoints - public, no auth required
	alerts := rg.Group("/alerts")
	alerts.POST("/webhook", api.AlertWebhookCtrl)
	alerts.GET("/webhook/health", api.WebhookHealthCtrl)

	// Routes with privilege level 1
	alertsRead := rg.Group("/alerts", api.PrivilegesMiddleware(1))
	alertsRead.GET("", api.SearchAlertsCtrl)
	alertsRead.GET("/:id", api.GetAlertByIDCtrl)
	alertsRead.GET("/score", api.GetAlertScoreCtrl)

	// Routes with privilege level 2
	alertsWrite := rg.Group("/alerts", api.PrivilegesMiddleware(2))
	alertsWrite.POST("", api.AddAlertCtrl)
	alertsWrite.DELETE("/:id", api.DeleteAlertByIDCtrl)
	alertsWrite.PUT("/:id", api.UpdateAlertCtrl)
	alertsWrite.POST("/generate-test", api.GenerateTestAlertsCtrl)
}
