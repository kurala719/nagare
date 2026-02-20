package router

import (
	"github.com/gin-gonic/gin"
	"nagare/internal/api"
)

func setupMonitorRoutes(rg *gin.RouterGroup) {
	// Public event token refresh - no auth required, uses event token
	monitorsPublic := rg.Group("/monitors")
	monitorsPublic.POST("/:id/event-token/refresh", api.RefreshMonitorEventTokenCtrl)

	// Routes with privilege level 1
	monitorsRead := rg.Group("/monitors", api.PrivilegesMiddleware(1))
	monitorsRead.GET("", api.SearchMonitorsCtrl)
	monitorsRead.GET("/:id", api.GetMonitorByIDCtrl)

	// Routes with privilege level 2
	monitorsWrite := rg.Group("/monitors", api.PrivilegesMiddleware(2))
	monitorsWrite.POST("", api.AddMonitorCtrl)
	monitorsWrite.DELETE("/:id", api.DeleteMonitorByIDCtrl)
	monitorsWrite.PUT("/:id", api.UpdateMonitorCtrl)
	monitorsWrite.POST("/:id/login", api.LoginMonitorCtrl)
	monitorsWrite.POST("/:id/event-token", api.RegenerateMonitorEventTokenCtrl)
	monitorsWrite.POST("/check", api.CheckAllMonitorsStatusCtrl)
	monitorsWrite.POST("/:id/check", api.CheckMonitorStatusCtrl)
}

func setupAlarmRoutes(rg *gin.RouterGroup) {
	// Public event token refresh - no auth required, uses event token
	alarmsPublic := rg.Group("/alarms")
	alarmsPublic.POST("/:id/event-token/refresh", api.RefreshAlarmEventTokenCtrl)

	// Routes with privilege level 1
	alarmsRead := rg.Group("/alarms", api.PrivilegesMiddleware(1))
	alarmsRead.GET("", api.SearchAlarmsCtrl)
	alarmsRead.GET("/:id", api.GetAlarmByIDCtrl)

	// Routes with privilege level 2
	alarmsWrite := rg.Group("/alarms", api.PrivilegesMiddleware(2))
	alarmsWrite.POST("", api.AddAlarmCtrl)
	alarmsWrite.DELETE("/:id", api.DeleteAlarmByIDCtrl)
	alarmsWrite.PUT("/:id", api.UpdateAlarmCtrl)
	alarmsWrite.POST("/:id/login", api.LoginAlarmCtrl)
	alarmsWrite.POST("/:id/event-token", api.RegenerateAlarmEventTokenCtrl)
	alarmsWrite.POST("/:id/setup-media", api.SetupAlarmMediaTypeCtrl)
}

func setupHostRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 1
	hostsRead := rg.Group("/hosts", api.PrivilegesMiddleware(1))
	hostsRead.GET("", api.SearchHostsCtrl)
	hostsRead.GET("/:id", api.GetHostByIDCtrl)
	hostsRead.GET("/:id/history", api.GetHostHistoryCtrl)
	hostsRead.POST("/:id/consult", api.ConsultHostCtrl)
	hostsRead.GET("/:id/ssh", api.HandleWebSSH)

	// Generic terminal route
	terminal := rg.Group("/terminal", api.PrivilegesMiddleware(1))
	terminal.GET("/ssh", api.HandleWebSSH)

	// Routes with privilege level 2
	hostsWrite := rg.Group("/hosts", api.PrivilegesMiddleware(2))
	hostsWrite.POST("", api.AddHostCtrl)
	hostsWrite.PUT("/:id", api.UpdateHostCtrl)
	hostsWrite.DELETE("/:id", api.DeleteHostByIDCtrl)

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
