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

	"nagare/internal/api"
	"nagare/internal/mcp"
	"nagare/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
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
	setupAlarmRoutes(rg)
	setupGroupRoutes(rg)
	setupHostRoutes(rg)
	setupAlertRoutes(rg)
	setupQueueRoutes(rg)
	setupSystemRoutes(rg)
	setupPublicRoutes(rg)
	setupIMRoutes(rg)
	setupMediaTypeRoutes(rg)
	setupMediaRoutes(rg)
	setupActionRoutes(rg)
	setupTriggerRoutes(rg)
	setupLogRoutes(rg)
	setupItemRoutes(rg)
	setupChatRoutes(rg)
	setupProviderRoutes(rg)
	setupKnowledgeBaseRoutes(rg)
	setupUserRoutes(rg)
	setupUserInformationRoutes(rg)
	setupQQWhitelistRoutes(rg)
	setupReportRoutes(rg)
}

func setupKnowledgeBaseRoutes(rg RouteGroup) {
	// Routes with privilege level 1
	kbRead := rg.Group("/knowledge-base", api.PrivilegesMiddleware(1))
	kbRead.GET("", api.GetAllKnowledgeBaseCtrl)
	kbRead.GET("/:id", api.GetKnowledgeBaseByIDCtrl)

	// Routes with privilege level 2
	kbWrite := rg.Group("/knowledge-base", api.PrivilegesMiddleware(2))
	kbWrite.POST("", api.AddKnowledgeBaseCtrl)
	kbWrite.PUT("/:id", api.UpdateKnowledgeBaseCtrl)
	kbWrite.DELETE("/:id", api.DeleteKnowledgeBaseCtrl)
}

func setupPublicRoutes(rg RouteGroup) {
	public := rg.Group("/public")
	public.GET("/status", api.GetPublicStatusSummaryCtrl)
}

func setupConfigurationRoutes(rg RouteGroup) {
	// Routes with privilege level 3 - admin only
	config := rg.Group("/config", api.PrivilegesMiddleware(3))
	config.GET("/", api.GetMainConfigCtrl)
	config.PUT("/", api.ModifyMainConfigCtrl)
	config.POST("/save", api.SaveConfigCtrl)
	config.POST("/reload", api.LoadConfigCtrl)
	config.POST("/reset", api.ResetConfigCtrl)
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

func setupAlarmRoutes(rg RouteGroup) {
	// Public event token refresh - no auth required, uses event token
	alarmsPublic := rg.Group("/alarms")
	alarmsPublic.POST("/:id/event-token/refresh", api.RefreshAlarmEventTokenCtrl)

	// Routes with privilege level 1
	alarmsRead := rg.Group("/alarms", api.PrivilegesMiddleware(1))
	alarmsRead.GET("/", api.SearchAlarmsCtrl)
	alarmsRead.GET("/:id", api.GetAlarmByIDCtrl)

	// Routes with privilege level 2
	alarmsWrite := rg.Group("/alarms", api.PrivilegesMiddleware(2))
	alarmsWrite.POST("/", api.AddAlarmCtrl)
	alarmsWrite.DELETE("/:id", api.DeleteAlarmByIDCtrl)
	alarmsWrite.PUT("/:id", api.UpdateAlarmCtrl)
	alarmsWrite.POST("/:id/login", api.LoginAlarmCtrl)
	alarmsWrite.POST("/:id/event-token", api.RegenerateAlarmEventTokenCtrl)
	alarmsWrite.POST("/:id/setup-media", api.SetupAlarmMediaTypeCtrl)
}

func setupGroupRoutes(rg RouteGroup) {
	// Routes with privilege level 1
	groupsRead := rg.Group("/groups", api.PrivilegesMiddleware(1))
	groupsRead.GET("/", api.SearchGroupsCtrl)
	groupsRead.GET("/:id", api.GetGroupByIDCtrl)
	groupsRead.GET("/:id/detail", api.GetGroupDetailCtrl)

	// Routes with privilege level 2
	groupsWrite := rg.Group("/groups", api.PrivilegesMiddleware(2))
	groupsWrite.POST("/", api.AddGroupCtrl)
	groupsWrite.PUT("/:id", api.UpdateGroupCtrl)
	groupsWrite.DELETE("/:id", api.DeleteGroupByIDCtrl)
	groupsWrite.POST("/:id/pull", api.PullGroupFromMonitorsCtrl)
	groupsWrite.POST("/:id/push", api.PushGroupToMonitorsCtrl)
	groupsWrite.POST("/check", api.CheckAllGroupsStatusCtrl)
	groupsWrite.POST("/:id/check", api.CheckGroupStatusCtrl)

	// Monitor groups routes with privilege level 2
	monitorGroups := rg.Group("/monitors/:id/groups", api.PrivilegesMiddleware(2))
	monitorGroups.POST("/pull", api.PullGroupsFromMonitorCtrl)
	monitorGroups.POST("/:gid/push", api.PushGroupToMonitorCtrl)
}

func setupHostRoutes(rg RouteGroup) {
	// Routes with privilege level 1
	hostsRead := rg.Group("/hosts", api.PrivilegesMiddleware(1))
	hostsRead.GET("/", api.SearchHostsCtrl)
	hostsRead.GET("/:id", api.GetHostByIDCtrl)
	hostsRead.GET("/:id/history", api.GetHostHistoryCtrl)
	hostsRead.POST("/:id/consult", api.ConsultHostCtrl)
	hostsRead.GET("/:id/ssh", api.HandleWebSSH)

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
	// Webhook endpoints - public, no auth required
	alerts := rg.Group("/alerts")
	alerts.POST("/webhook", api.AlertWebhookCtrl)
	alerts.GET("/webhook/health", api.WebhookHealthCtrl)

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

func setupQQWhitelistRoutes(rg RouteGroup) {
	// Routes with privilege level 2
	whitelist := rg.Group("/qq-whitelist", api.PrivilegesMiddleware(2))
	whitelist.GET("/", api.GetQQWhitelistCtrl)
	whitelist.POST("/", api.AddQQWhitelistCtrl)
	whitelist.PUT("/:id", api.UpdateQQWhitelistCtrl)
	whitelist.DELETE("/:id", api.DeleteQQWhitelistCtrl)
}

func setupReportRoutes(rg RouteGroup) {
	// Routes with privilege level 2 (managers/admins)
	reports := rg.Group("/reports", api.PrivilegesMiddleware(2))
	reports.GET("", api.ListReportsCtrl)
	reports.GET("/:id", api.GetReportCtrl)
	reports.GET("/config", api.GetReportConfigCtrl)
	reports.PUT("/config", api.UpdateReportConfigCtrl)
	reports.POST("/generate/weekly", api.GenerateWeeklyReportCtrl)
	reports.POST("/generate/monthly", api.GenerateMonthlyReportCtrl)
	reports.DELETE("/:id", api.DeleteReportCtrl)
	reports.GET("/:id/download", api.DownloadReportCtrl)
}
