package router

import (
	"nagare/internal/api"

	"github.com/gin-gonic/gin"
)

func setupMonitoringDomainRoutes(rg *gin.RouterGroup) {
	setupMonitorRoutes(rg)
	setupGroupRoutes(rg)
	setupHostRoutes(rg)
	setupItemRoutes(rg)
}

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

		// Host Sync operations
		monitors.POST("/:id/hosts/pull", api.PrivilegesMiddleware(2), api.PullHostsFromMonitorCtrl)
		monitors.POST("/:id/hosts/:hid/pull", api.PrivilegesMiddleware(2), api.PullHostFromMonitorCtrl)
		monitors.POST("/:id/hosts/push", api.PrivilegesMiddleware(2), api.PushHostsFromMonitorCtrl)
		monitors.POST("/:id/hosts/:hid/push", api.PrivilegesMiddleware(2), api.PushHostToMonitorCtrl)
	}
}

func setupGroupRoutes(rg *gin.RouterGroup) {
	groups := rg.Group("/groups")
	{
		// Routes with privilege level 1
		groupsRead := groups.Group("", api.PrivilegesMiddleware(1))
		groupsRead.GET("", api.SearchGroupsCtrl)
		groupsRead.GET("/:id", api.GetGroupByIDCtrl)
		groupsRead.GET("/:id/detail", api.GetGroupDetailCtrl)

		// Routes with privilege level 2
		groupsWrite := groups.Group("", api.PrivilegesMiddleware(2))
		groupsWrite.POST("", api.AddGroupCtrl)
		groupsWrite.PUT("/:id", api.UpdateGroupCtrl)
		groupsWrite.DELETE("/:id", api.DeleteGroupByIDCtrl)
		groupsWrite.POST("/check", api.CheckAllGroupsStatusCtrl)
		groupsWrite.POST("/:id/check", api.CheckGroupStatusCtrl)
	}
}

func setupHostRoutes(rg *gin.RouterGroup) {
	hosts := rg.Group("/hosts")
	{
		// Privilege level 1
		hosts.GET("", api.PrivilegesMiddleware(1), api.SearchHostsCtrl)
		hosts.GET("/:id", api.PrivilegesMiddleware(1), api.GetHostByIDCtrl)

		// Privilege level 2
		hosts.POST("", api.PrivilegesMiddleware(2), api.AddHostCtrl)
		hosts.PUT("/:id", api.PrivilegesMiddleware(2), api.UpdateHostCtrl)
		hosts.DELETE("/:id", api.PrivilegesMiddleware(2), api.DeleteHostByIDCtrl)
	}
}

func setupItemRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 1
	itemsRead := rg.Group("/items", api.PrivilegesMiddleware(1))
	itemsRead.GET("", api.SearchItemsCtrl)
	itemsRead.GET("/:id", api.GetItemByIDCtrl)

	// Routes with privilege level 2
	itemsWrite := rg.Group("/items", api.PrivilegesMiddleware(2))
	itemsWrite.POST("", api.AddItemCtrl)
	itemsWrite.PUT("/:id", api.UpdateItemCtrl)
	itemsWrite.DELETE("/:id", api.DeleteItemByIDCtrl)
	itemsWrite.POST("/hosts/:hid/import", api.AddItemsByHostIDFromMonitorCtrl)
	itemsWrite.POST("/_test/generate-history", api.GenerateTestHistoryCtrl) // DEVELOPMENT: Test data generation
}
