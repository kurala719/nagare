package router

import (
	"github.com/gin-gonic/gin"
	"nagare/internal/api"
)

func setupMonitorRoutes(rg *gin.RouterGroup) {
	monitors := rg.Group("/monitors")
	{
		// Public event token refresh - no auth required, uses event token
		monitors.POST("/:id/event-token/refresh", api.RefreshMonitorEventTokenCtrl)

		// Privilege level 1
		monitors.GET("", api.PrivilegesMiddleware(1), api.SearchMonitorsCtrl)
		monitors.GET("/:id", api.PrivilegesMiddleware(1), api.GetMonitorByIDCtrl)

		// Privilege level 2
		monitors.POST("", api.PrivilegesMiddleware(2), api.AddMonitorCtrl)
		monitors.DELETE("/:id", api.PrivilegesMiddleware(2), api.DeleteMonitorByIDCtrl)
		monitors.PUT("/:id", api.PrivilegesMiddleware(2), api.UpdateMonitorCtrl)
		monitors.POST("/:id/login", api.PrivilegesMiddleware(2), api.LoginMonitorCtrl)
		monitors.POST("/:id/event-token", api.PrivilegesMiddleware(2), api.RegenerateMonitorEventTokenCtrl)
		monitors.POST("/check", api.PrivilegesMiddleware(2), api.CheckAllMonitorsStatusCtrl)
		monitors.POST("/:id/check", api.PrivilegesMiddleware(2), api.CheckMonitorStatusCtrl)
	}
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
		alarms.POST("/:id/event-token", api.PrivilegesMiddleware(2), api.RegenerateAlarmEventTokenCtrl)
		alarms.POST("/:id/setup-media", api.PrivilegesMiddleware(2), api.SetupAlarmMediaTypeCtrl)
	}
}

func setupHostRoutes(rg *gin.RouterGroup) {
	hosts := rg.Group("/hosts")
	{
		// Privilege level 1
		hosts.GET("", api.PrivilegesMiddleware(1), api.SearchHostsCtrl)
		hosts.GET("/:id", api.PrivilegesMiddleware(1), api.GetHostByIDCtrl)
		hosts.GET("/:id/history", api.PrivilegesMiddleware(1), api.GetHostHistoryCtrl)
		hosts.POST("/:id/consult", api.PrivilegesMiddleware(1), api.ConsultHostCtrl)
		hosts.GET("/:id/ssh", api.PrivilegesMiddleware(1), api.HandleWebSSH)

		// Privilege level 2
		hosts.POST("", api.PrivilegesMiddleware(2), api.AddHostCtrl)
		hosts.PUT("/:id", api.PrivilegesMiddleware(2), api.UpdateHostCtrl)
		hosts.DELETE("/:id", api.PrivilegesMiddleware(2), api.DeleteHostByIDCtrl)
	}

	// Generic terminal route
	terminal := rg.Group("/terminal", api.PrivilegesMiddleware(1))
	terminal.GET("/ssh", api.HandleWebSSH)

	// Monitor hosts routes with privilege level 2
	monitorHosts := rg.Group("/monitors/:id/hosts", api.PrivilegesMiddleware(2))
	monitorHosts.POST("/pull", api.PullHostsFromMonitorCtrl)
	monitorHosts.POST("/pull-async", api.PullHostsAsyncCtrl)
	monitorHosts.POST("/:hid/pull", api.PullHostFromMonitorCtrl)
	monitorHosts.POST("/push", api.PushHostsFromMonitorCtrl)
	monitorHosts.POST("/:hid/push", api.PushHostToMonitorCtrl)
}

func setupItemRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 1
	itemsRead := rg.Group("/items", api.PrivilegesMiddleware(1))
	itemsRead.GET("", api.SearchItemsCtrl)
	itemsRead.GET("/:id", api.GetItemByIDCtrl)
	itemsRead.GET("/:id/history", api.GetItemHistoryCtrl)
	itemsRead.POST("/:id/consult", api.ConsultItemCtrl)

	// Routes with privilege level 2
	itemsWrite := rg.Group("/items", api.PrivilegesMiddleware(2))
	itemsWrite.POST("", api.AddItemCtrl)
	itemsWrite.PUT("/:id", api.UpdateItemCtrl)
	itemsWrite.DELETE("/:id", api.DeleteItemByIDCtrl)
	itemsWrite.POST("/hosts/:hid/import", api.AddItemsByHostIDFromMonitorCtrl)

	// Monitor items routes with privilege level 2
	monitorItems := rg.Group("/monitors/:id/items", api.PrivilegesMiddleware(2))
	monitorItems.POST("/pull", api.PullItemsFromMonitorCtrl)
	monitorItems.POST("/push", api.PushItemsFromMonitorCtrl)

	// Monitor host items routes with privilege level 2
	monitorHostItems := rg.Group("/monitors/:id/hosts/:hid/items", api.PrivilegesMiddleware(2))
	monitorHostItems.POST("/pull", api.PullItemsOfHostFromMonitorCtrl)
	monitorHostItems.POST("/pull-async", api.PullItemsAsyncCtrl)
	monitorHostItems.POST("/:item_id/pull", api.PullItemOfHostFromMonitorCtrl)
	monitorHostItems.POST("/push", api.PushItemsFromHostCtrl)
	monitorHostItems.POST("/:item_id/push", api.PushItemToMonitorCtrl)
}
