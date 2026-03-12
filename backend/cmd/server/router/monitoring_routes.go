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
		monitors.POST("/:id/event-token-refreshes", api.RefreshMonitorEventTokenCtrl)

		// Privilege level 1
		monitors.GET("", api.PrivilegesMiddleware(1), api.SearchMonitorsCtrl)
		monitors.GET("/:id", api.PrivilegesMiddleware(1), api.GetMonitorByIDCtrl)

		// Privilege level 2
		monitors.POST("", api.PrivilegesMiddleware(2), api.AddMonitorCtrl)
		monitors.DELETE("/:id", api.PrivilegesMiddleware(2), api.DeleteMonitorByIDCtrl)
		monitors.PUT("/:id", api.PrivilegesMiddleware(2), api.UpdateMonitorCtrl)
		monitors.POST("/:id/sessions", api.PrivilegesMiddleware(2), api.LoginMonitorCtrl)
		monitors.POST("/:id/event-tokens", api.PrivilegesMiddleware(2), api.RegenerateMonitorEventTokenCtrl)
		monitors.POST("/checks", api.PrivilegesMiddleware(2), api.CheckAllMonitorsStatusCtrl)
		monitors.POST("/:id/checks", api.PrivilegesMiddleware(2), api.CheckMonitorStatusCtrl)
		monitors.POST("/:id/group-imports", api.PrivilegesMiddleware(2), api.PullGroupsFromMonitorCtrl)

		// Host Sync operations
		monitors.POST("/:id/host-imports", api.PrivilegesMiddleware(2), api.PullHostsFromMonitorCtrl)
		monitors.POST("/:id/hosts/:hid/imports", api.PrivilegesMiddleware(2), api.PullHostFromMonitorCtrl)
		monitors.POST("/:id/host-exports", api.PrivilegesMiddleware(2), api.PushHostsFromMonitorCtrl)
		monitors.POST("/:id/hosts/:hid/exports", api.PrivilegesMiddleware(2), api.PushHostToMonitorCtrl)
		monitors.POST("/:id/groups/:gid/exports", api.PrivilegesMiddleware(2), api.PushGroupToMonitorCtrl)
		monitors.POST("/:id/item-imports", api.PrivilegesMiddleware(2), api.PullItemsFromMonitorCtrl)
		monitors.POST("/:id/hosts/:hid/item-imports", api.PrivilegesMiddleware(2), api.PullItemsOfHostFromMonitorCtrl)
		monitors.POST("/:id/hosts/:hid/items/:item_id/imports", api.PrivilegesMiddleware(2), api.PullItemOfHostFromMonitorCtrl)
		monitors.POST("/:id/item-exports", api.PrivilegesMiddleware(2), api.PushItemsFromMonitorCtrl)
		monitors.POST("/:id/hosts/:hid/item-exports", api.PrivilegesMiddleware(2), api.PushItemsFromHostCtrl)
		monitors.POST("/:id/hosts/:hid/items/:item_id/exports", api.PrivilegesMiddleware(2), api.PushItemToMonitorCtrl)
	}
}

func setupGroupRoutes(rg *gin.RouterGroup) {
	groups := rg.Group("/groups")
	{
		// Routes with privilege level 1
		groupsRead := groups.Group("", api.PrivilegesMiddleware(1))
		groupsRead.GET("", api.SearchGroupsCtrl)
		groupsRead.GET("/:id", api.GetGroupByIDCtrl)
		groupsRead.GET("/:id/details", api.GetGroupDetailCtrl)

		// Routes with privilege level 2
		groupsWrite := groups.Group("", api.PrivilegesMiddleware(2))
		groupsWrite.POST("", api.AddGroupCtrl)
		groupsWrite.PUT("/:id", api.UpdateGroupCtrl)
		groupsWrite.DELETE("/:id", api.DeleteGroupByIDCtrl)
		groupsWrite.POST("/checks", api.CheckAllGroupsStatusCtrl)
		groupsWrite.POST("/:id/checks", api.CheckGroupStatusCtrl)
		groupsWrite.POST("/:id/imports", api.PullGroupFromMonitorsCtrl)
		groupsWrite.POST("/:id/exports", api.PushGroupToMonitorsCtrl)
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
	itemsWrite.POST("/hosts/:hid/imports", api.AddItemsByHostIDFromMonitorCtrl)
	itemsWrite.POST("/history-generations", api.GenerateTestHistoryCtrl) // DEVELOPMENT: Test data generation
}
