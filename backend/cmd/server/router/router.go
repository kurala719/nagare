package router

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"nagare/internal/service"
	"nagare/internal/mcp"
	"nagare/internal/api"
)

type RouteGroup interface {
	Group(string, ...gin.HandlerFunc) *gin.RouterGroup
}

// InitRouter initializes and starts the HTTP router
func InitRouter() {
	r := gin.Default()
	r.RedirectTrailingSlash = false
	r.Use(api.RequestIDMiddleware())
	r.Use(api.AccessLogMiddleware())

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, api.APIResponse{
			Success: false,
			Error:   "resource not found",
		})
	})
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, api.APIResponse{
			Success: false,
			Error:   "method not allowed",
		})
	})

	// Setup all routes
	api := r.Group("/api/v1")
	setupAllRoutes(api)
	setupMcpRoutes(r)

	port := viper.GetInt("system.port")
	if port == 0 {
		port = 8080
	}

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	service.LogSystem("info", "server starting", map[string]interface{}{"port": port}, nil, "")
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			service.LogSystem("error", "failed to start server", map[string]interface{}{"error": err.Error()}, nil, "")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		service.LogSystem("error", "server shutdown error", map[string]interface{}{"error": err.Error()}, nil, "")
	}
}

func setupAllRoutes(rg RouteGroup) {
	setupConfigurationRoutes(rg)
	setupMonitorRoutes(rg)
	setupSiteRoutes(rg)
	setupHostRoutes(rg)
	setupAlertRoutes(rg)
	setupQueueRoutes(rg)
	setupSystemRoutes(rg)
	setupIMRoutes(rg)
	setupMediaTypeRoutes(rg)
	setupMediaRoutes(rg)
	setupActionRoutes(rg)
	setupTriggerRoutes(rg)
	setupLogRoutes(rg)
	setupItemRoutes(rg)
	setupChatRoutes(rg)
	setupProviderRoutes(rg)
	setupUserRoutes(rg)
	setupUserInformationRoutes(rg)
}

func setupConfigurationRoutes(rg RouteGroup) {
	// Routes with privilege level 3 - admin only
	config := rg.Group("/config", api.PrivilegesMiddleware(3))
	config.GET("/", api.GetMainConfigCtrl)
	config.PUT("/", api.ModifyMainConfigCtrl)
	config.POST("/save", api.SaveConfigCtrl)
	config.POST("/reload", api.LoadConfigCtrl)
}

func setupMonitorRoutes(rg RouteGroup) {
	// Public event token refresh - no auth required, uses event token
	monitorsPublic := rg.Group("/monitors")
	monitorsPublic.POST("/:id/event-token/refresh", api.RefreshMonitorEventTokenCtrl)

	// Routes with privilege level 1
	monitorsRead := rg.Group("/monitors", api.PrivilegesMiddleware(1))
	monitorsRead.GET("/", api.SearchMonitorsCtrl)
	monitorsRead.GET("/:id", api.GetMonitorByIDCtrl)

	// Routes with privilege level 2
	monitorsWrite := rg.Group("/monitors", api.PrivilegesMiddleware(2))
	monitorsWrite.POST("/", api.AddMonitorCtrl)
	monitorsWrite.DELETE("/:id", api.DeleteMonitorByIDCtrl)
	monitorsWrite.PUT("/:id", api.UpdateMonitorCtrl)
	monitorsWrite.POST("/:id/login", api.LoginMonitorCtrl)
	monitorsWrite.POST("/:id/event-token", api.RegenerateMonitorEventTokenCtrl)
	monitorsWrite.POST("/check", api.CheckAllMonitorsStatusCtrl)
	monitorsWrite.POST("/:id/check", api.CheckMonitorStatusCtrl)
}

func setupSiteRoutes(rg RouteGroup) {
	// Routes with privilege level 1
	sitesRead := rg.Group("/sites", api.PrivilegesMiddleware(1))
	sitesRead.GET("/", api.SearchSitesCtrl)
	sitesRead.GET("/:id", api.GetSiteByIDCtrl)
	sitesRead.GET("/:id/detail", api.GetSiteDetailCtrl)

	// Routes with privilege level 2
	sitesWrite := rg.Group("/sites", api.PrivilegesMiddleware(2))
	sitesWrite.POST("/", api.AddSiteCtrl)
	sitesWrite.PUT("/:id", api.UpdateSiteCtrl)
	sitesWrite.DELETE("/:id", api.DeleteSiteByIDCtrl)
	sitesWrite.POST("/:id/pull", api.PullSiteFromMonitorsCtrl)
	sitesWrite.POST("/:id/push", api.PushSiteToMonitorsCtrl)
	sitesWrite.POST("/check", api.CheckAllSitesStatusCtrl)
	sitesWrite.POST("/:id/check", api.CheckSiteStatusCtrl)
}

func setupHostRoutes(rg RouteGroup) {
	// Routes with privilege level 1
	hostsRead := rg.Group("/hosts", api.PrivilegesMiddleware(1))
	hostsRead.GET("/", api.SearchHostsCtrl)
	hostsRead.GET("/:id", api.GetHostByIDCtrl)
	hostsRead.GET("/:id/history", api.GetHostHistoryCtrl)
	hostsRead.POST("/:id/consult", api.ConsultHostCtrl)

	// Routes with privilege level 2
	hostsWrite := rg.Group("/hosts", api.PrivilegesMiddleware(2))
	hostsWrite.POST("/", api.AddHostCtrl)
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

func setupAlertRoutes(rg RouteGroup) {
	// Webhook endpoint - public, no auth required
	alerts := rg.Group("/alerts")
	alerts.POST("/webhook", api.AlertWebhookCtrl)

	// Routes with privilege level 1
	alertsRead := rg.Group("/alerts", api.PrivilegesMiddleware(1))
	alertsRead.GET("/", api.SearchAlertsCtrl)
	alertsRead.GET("/:id", api.GetAlertByIDCtrl)
	alertsRead.POST("/:id/consult", api.ConsultAlertCtrl)
	alertsRead.GET("/score", api.GetAlertScoreCtrl)

	// Routes with privilege level 2
	alertsWrite := rg.Group("/alerts", api.PrivilegesMiddleware(2))
	alertsWrite.POST("/", api.AddAlertCtrl)
	alertsWrite.DELETE("/:id", api.DeleteAlertByIDCtrl)
	alertsWrite.PUT("/:id", api.UpdateAlertCtrl)
	alertsWrite.POST("/generate-test", api.GenerateTestAlertsCtrl)
}

func setupQueueRoutes(rg RouteGroup) {
	// Routes with privilege level 2
	queue := rg.Group("/queue", api.PrivilegesMiddleware(2))
	queue.GET("/stats", api.QueueStatsCtrl)
}

func setupSystemRoutes(rg RouteGroup) {
	system := rg.Group("/system")
	system.GET("/health", api.GetHealthScoreCtrl)
	system.GET("/health/history", api.GetNetworkStatusHistoryCtrl)
	system.GET("/metrics", api.GetNetworkMetricsCtrl)
}

func setupIMRoutes(rg RouteGroup) {
	im := rg.Group("/im")
	im.POST("/command", api.IMCommandCtrl)
}

func setupMediaTypeRoutes(rg RouteGroup) {
	// Routes with privilege level 1
	mediaTypesRead := rg.Group("/media-types", api.PrivilegesMiddleware(1))
	mediaTypesRead.GET("/", api.SearchMediaTypesCtrl)
	mediaTypesRead.GET("/:id", api.GetMediaTypeByIDCtrl)

	// Routes with privilege level 2
	mediaTypesWrite := rg.Group("/media-types", api.PrivilegesMiddleware(2))
	mediaTypesWrite.POST("/", api.AddMediaTypeCtrl)
	mediaTypesWrite.PUT("/:id", api.UpdateMediaTypeCtrl)
	mediaTypesWrite.DELETE("/:id", api.DeleteMediaTypeByIDCtrl)
}

func setupMediaRoutes(rg RouteGroup) {
	// Webhook endpoint MUST be first, before any authenticated routes
	media := rg.Group("/media")
	media.POST("/qq/message", api.HandleQQMessageCtrl)

	// Routes with privilege level 1
	mediaRead := rg.Group("/media", api.PrivilegesMiddleware(1))
	mediaRead.GET("/", api.SearchMediaCtrl)
	mediaRead.GET("/:id", api.GetMediaByIDCtrl)

	// Routes with privilege level 2
	mediaWrite := rg.Group("/media", api.PrivilegesMiddleware(2))
	mediaWrite.POST("/", api.AddMediaCtrl)
	mediaWrite.PUT("/:id", api.UpdateMediaCtrl)
	mediaWrite.DELETE("/:id", api.DeleteMediaByIDCtrl)
}

func setupActionRoutes(rg RouteGroup) {
	// Routes with privilege level 1
	actionsRead := rg.Group("/actions", api.PrivilegesMiddleware(1))
	actionsRead.GET("/", api.SearchActionsCtrl)
	actionsRead.GET("/:id", api.GetActionByIDCtrl)

	// Routes with privilege level 2
	actionsWrite := rg.Group("/actions", api.PrivilegesMiddleware(2))
	actionsWrite.POST("/", api.AddActionCtrl)
	actionsWrite.PUT("/:id", api.UpdateActionCtrl)
	actionsWrite.DELETE("/:id", api.DeleteActionByIDCtrl)
}

func setupTriggerRoutes(rg RouteGroup) {
	// Routes with privilege level 1
	triggersRead := rg.Group("/triggers", api.PrivilegesMiddleware(1))
	triggersRead.GET("/", api.SearchTriggersCtrl)
	triggersRead.GET("/:id", api.GetTriggerByIDCtrl)

	// Routes with privilege level 2
	triggersWrite := rg.Group("/triggers", api.PrivilegesMiddleware(2))
	triggersWrite.POST("/", api.AddTriggerCtrl)
	triggersWrite.PUT("/:id", api.UpdateTriggerCtrl)
	triggersWrite.DELETE("/:id", api.DeleteTriggerByIDCtrl)
}

func setupLogRoutes(rg RouteGroup) {
	// Routes with privilege level 2
	logs := rg.Group("/logs", api.PrivilegesMiddleware(2))
	logs.GET("/system", api.GetSystemLogsCtrl)
	logs.GET("/service", api.GetServiceLogsCtrl)
}

func setupItemRoutes(rg RouteGroup) {
	// Routes with privilege level 1
	itemsRead := rg.Group("/items", api.PrivilegesMiddleware(1))
	itemsRead.GET("/", api.SearchItemsCtrl)
	itemsRead.GET("/:id", api.GetItemByIDCtrl)
	itemsRead.GET("/:id/history", api.GetItemHistoryCtrl)
	itemsRead.POST("/:id/consult", api.ConsultItemCtrl)

	// Routes with privilege level 2
	itemsWrite := rg.Group("/items", api.PrivilegesMiddleware(2))
	itemsWrite.POST("/", api.AddItemCtrl)
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

func setupChatRoutes(rg RouteGroup) {
	// Routes with privilege level 1
	chats := rg.Group("/chats", api.PrivilegesMiddleware(1))
	chats.GET("/", api.SearchChatsCtrl)
	chats.POST("/", api.SendChatCtrl)
}

func setupProviderRoutes(rg RouteGroup) {
	// Routes with privilege level 1
	providersRead := rg.Group("/providers", api.PrivilegesMiddleware(1))
	providersRead.GET("/", api.SearchProvidersCtrl)
	providersRead.GET("/:id", api.GetProviderByIDCtrl)

	// Routes with privilege level 2
	providersWrite := rg.Group("/providers", api.PrivilegesMiddleware(2))
	providersWrite.POST("/", api.AddProviderCtrl)
	providersWrite.DELETE("/:id", api.DeleteProviderByIDCtrl)
	providersWrite.PUT("/:id", api.UpdateProviderCtrl)
	providersWrite.POST("/check", api.CheckAllProvidersStatusCtrl)
	providersWrite.POST("/:id/check", api.CheckProviderStatusCtrl)
}

func setupUserRoutes(rg RouteGroup) {
	// Public auth routes - no authentication required
	auth := rg.Group("/auth")
	auth.POST("/login", api.LoginUserCtrl)
	auth.POST("/register", api.RegisterUserCtrl)

	// Reset password - requires privilege 1
	authProtected := rg.Group("/auth", api.PrivilegesMiddleware(1))
	authProtected.POST("/reset", api.ResetPasswordCtrl)

	// Register applications - requires privilege 3
	registerApps := rg.Group("/register-applications", api.PrivilegesMiddleware(3))
	registerApps.GET("/", api.ListRegisterApplicationsCtrl)
	registerApps.PUT("/:id/approve", api.ApproveRegisterApplicationCtrl)
	registerApps.PUT("/:id/reject", api.RejectRegisterApplicationCtrl)

	// Legacy register applications - requires privilege 3
	registerAppsLegacy := rg.Group("/register-application", api.PrivilegesMiddleware(3))
	registerAppsLegacy.GET("/", api.ListRegisterApplicationsCtrl)
	registerAppsLegacy.PUT("/:id/approve", api.ApproveRegisterApplicationCtrl)
	registerAppsLegacy.PUT("/:id/reject", api.RejectRegisterApplicationCtrl)

	// Users routes - requires privilege 2 for read, privilege 3 for write
	usersRead := rg.Group("/users", api.PrivilegesMiddleware(2))
	usersRead.GET("/", api.SearchUsersCtrl)
	usersRead.GET("/:id", api.GetUserByIDCtrl)

	usersWrite := rg.Group("/users", api.PrivilegesMiddleware(3))
	usersWrite.POST("/", api.AddUserCtrl)
	usersWrite.DELETE("/:id", api.DeleteUserByIDCtrl)
	usersWrite.PUT("/:id", api.UpdateUserCtrl)
}

func setupUserInformationRoutes(rg RouteGroup) {
	// Authenticated user routes - manage their own information (privilege 1)
	authenticated := rg.Group("/user-info", api.PrivilegesMiddleware(1))
	authenticated.GET("/me", api.GetUserInformationCtrl)
	authenticated.POST("/me", api.CreateUserInformationCtrl)
	authenticated.PUT("/me", api.UpdateUserInformationCtrl)
	authenticated.DELETE("/me", api.DeleteUserInformationCtrl)

	// Admin routes - manage other users' information (privilege 3)
	admin := rg.Group("/user-info", api.PrivilegesMiddleware(3))
	admin.GET("/users/:user_id", api.GetUserInformationByUserIDCtrl)
	admin.PUT("/users/:user_id", api.UpdateUserInformationByUserIDCtrl)
}

func setupMcpRoutes(rg RouteGroup) {
	// MCP routes - requires API key middleware
	mcpGroup := rg.Group("/mcp", mcp.APIKeyMiddleware())
	mcpGroup.GET("/sse", mcp.SSEHandler)
	mcpGroup.POST("/message", mcp.MessageHandler)
}
