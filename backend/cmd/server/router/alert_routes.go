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
	alarms.POST("/:id/event-token-refreshes", api.RefreshAlarmEventTokenCtrl)
	alarms.GET("", api.PrivilegesMiddleware(1), api.SearchAlarmsCtrl)
	alarms.GET("/:id", api.PrivilegesMiddleware(1), api.GetAlarmByIDCtrl)
	alarms.POST("", api.PrivilegesMiddleware(2), api.AddAlarmCtrl)
	alarms.DELETE("/:id", api.PrivilegesMiddleware(2), api.DeleteAlarmByIDCtrl)
	alarms.PUT("/:id", api.PrivilegesMiddleware(2), api.UpdateAlarmCtrl)
	alarms.POST("/:id/sessions", api.PrivilegesMiddleware(2), api.LoginAlarmCtrl)
	alarms.POST("/:id/media-bindings", api.PrivilegesMiddleware(2), api.SetupAlarmMediaCtrl)
	alarms.POST("/:id/event-tokens", api.PrivilegesMiddleware(2), api.RegenerateAlarmEventTokenCtrl)
}

func setupTriggerRoutes(rg *gin.RouterGroup) {
	triggers := rg.Group("/triggers")
	triggersPrivileged := triggers.Group("", api.PrivilegesMiddleware(2))
	triggersPrivileged.GET("", api.SearchTriggersCtrl)
	triggersPrivileged.GET("/:id", api.GetTriggerByIDCtrl)
	triggersPrivileged.POST("", api.AddTriggerCtrl)
	triggersPrivileged.PUT("/:id", api.UpdateTriggerCtrl)
	triggersPrivileged.DELETE("/:id", api.DeleteTriggerByIDCtrl)
}

func setupAlertRoutes(rg *gin.RouterGroup) {
	webhooks := rg.Group("/webhooks")
	webhooks.POST("", api.AlertWebhookCtrl)
	webhooks.GET("/health", api.WebhookHealthCtrl)

	alertsRead := rg.Group("/alerts", api.PrivilegesMiddleware(1))
	alertsRead.GET("", api.SearchAlertsCtrl)
	alertsRead.GET("/:id", api.GetAlertByIDCtrl)
	alertsRead.GET("/scores", api.GetAlertScoreCtrl)

	alertsWrite := rg.Group("/alerts", api.PrivilegesMiddleware(2))
	alertsWrite.POST("", api.AddAlertCtrl)
	alertsWrite.DELETE("/:id", api.DeleteAlertByIDCtrl)
	alertsWrite.PUT("/:id", api.UpdateAlertCtrl)

	testAlerts := rg.Group("/test-alerts", api.PrivilegesMiddleware(2))
	testAlerts.POST("", api.GenerateTestAlertsCtrl)
}
