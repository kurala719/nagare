package router

import (
	"nagare/internal/adapter/handler"

	"github.com/gin-gonic/gin"
)

func setupMonitorRoutes(rg *gin.RouterGroup) {
	monitors := rg.Group("/monitors")
	{
		// Public event token refresh - no auth required, uses event token
		monitors.POST("/:id/event-token/refresh", handler.RefreshMonitorEventTokenCtrl)

		// Privilege level 1
		monitors.GET("", handler.PrivilegesMiddleware(1), handler.SearchMonitorsCtrl)
		monitors.GET("/:id", handler.PrivilegesMiddleware(1), handler.GetMonitorByIDCtrl)

		// Privilege level 2
		monitors.POST("", handler.PrivilegesMiddleware(2), handler.AddMonitorCtrl)
		monitors.DELETE("/:id", handler.PrivilegesMiddleware(2), handler.DeleteMonitorByIDCtrl)
		monitors.PUT("/:id", handler.PrivilegesMiddleware(2), handler.UpdateMonitorCtrl)
		monitors.POST("/:id/login", handler.PrivilegesMiddleware(2), handler.LoginMonitorCtrl)
		monitors.POST("/:id/event-token", handler.PrivilegesMiddleware(2), handler.RegenerateMonitorEventTokenCtrl)
		monitors.POST("/check", handler.PrivilegesMiddleware(2), handler.CheckAllMonitorsStatusCtrl)
		monitors.POST("/:id/check", handler.PrivilegesMiddleware(2), handler.CheckMonitorStatusCtrl)
	}
}

func setupAlarmRoutes(rg *gin.RouterGroup) {
	alarms := rg.Group("/alarms")
	{
		// Public event token refresh - no auth required, uses event token
		alarms.POST("/:id/event-token/refresh", handler.RefreshAlarmEventTokenCtrl)

		// Privilege level 1
		alarms.GET("", handler.PrivilegesMiddleware(1), handler.SearchAlarmsCtrl)
		alarms.GET("/:id", handler.PrivilegesMiddleware(1), handler.GetAlarmByIDCtrl)

		// Privilege level 2
		alarms.POST("", handler.PrivilegesMiddleware(2), handler.AddAlarmCtrl)
		alarms.DELETE("/:id", handler.PrivilegesMiddleware(2), handler.DeleteAlarmByIDCtrl)
		alarms.PUT("/:id", handler.PrivilegesMiddleware(2), handler.UpdateAlarmCtrl)
		alarms.POST("/:id/login", handler.PrivilegesMiddleware(2), handler.LoginAlarmCtrl)
		alarms.POST("/:id/event-token", handler.PrivilegesMiddleware(2), handler.RegenerateAlarmEventTokenCtrl)
		alarms.POST("/:id/setup-media", handler.PrivilegesMiddleware(2), handler.SetupAlarmMediaCtrl)
	}
}

func setupHostRoutes(rg *gin.RouterGroup) {
	hosts := rg.Group("/hosts")
	{
		// Privilege level 1
		hosts.GET("", handler.PrivilegesMiddleware(1), handler.SearchHostsCtrl)
		hosts.GET("/:id", handler.PrivilegesMiddleware(1), handler.GetHostByIDCtrl)
		hosts.GET("/:id/history", handler.PrivilegesMiddleware(1), handler.GetHostHistoryCtrl)
		hosts.POST("/:id/consult", handler.PrivilegesMiddleware(1), handler.ConsultHostCtrl)
		hosts.GET("/:id/ssh", handler.PrivilegesMiddleware(1), handler.HandleWebSSH)

		// Privilege level 2
		hosts.POST("", handler.PrivilegesMiddleware(2), handler.AddHostCtrl)
		hosts.PUT("/:id", handler.PrivilegesMiddleware(2), handler.UpdateHostCtrl)
		hosts.DELETE("/:id", handler.PrivilegesMiddleware(2), handler.DeleteHostByIDCtrl)
	}

	// Generic terminal route
	terminal := rg.Group("/terminal", handler.PrivilegesMiddleware(1))
	terminal.GET("/ssh", handler.HandleWebSSH)
}

func setupItemRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 1
	itemsRead := rg.Group("/items", handler.PrivilegesMiddleware(1))
	itemsRead.GET("", handler.SearchItemsCtrl)
	itemsRead.GET("/:id", handler.GetItemByIDCtrl)
	itemsRead.GET("/:id/history", handler.GetItemHistoryCtrl)
	itemsRead.POST("/:id/consult", handler.ConsultItemCtrl)

	// Routes with privilege level 2
	itemsWrite := rg.Group("/items", handler.PrivilegesMiddleware(2))
	itemsWrite.POST("", handler.AddItemCtrl)
	itemsWrite.PUT("/:id", handler.UpdateItemCtrl)
	itemsWrite.DELETE("/:id", handler.DeleteItemByIDCtrl)
	itemsWrite.POST("/hosts/:hid/import", handler.AddItemsByHostIDFromMonitorCtrl)
}
