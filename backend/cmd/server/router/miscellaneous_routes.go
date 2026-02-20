package router

import (
	"nagare/internal/adapter/handler"

	"github.com/gin-gonic/gin"
)

func setupGroupRoutes(rg *gin.RouterGroup) {
	groups := rg.Group("/groups")
	{
		// Routes with privilege level 1
		groupsRead := groups.Group("", handler.PrivilegesMiddleware(1))
		groupsRead.GET("", handler.SearchGroupsCtrl)
		groupsRead.GET("/:id", handler.GetGroupByIDCtrl)
		groupsRead.GET("/:id/detail", handler.GetGroupDetailCtrl)

		// Routes with privilege level 2
		groupsWrite := groups.Group("", handler.PrivilegesMiddleware(2))
		groupsWrite.POST("", handler.AddGroupCtrl)
		groupsWrite.PUT("/:id", handler.UpdateGroupCtrl)
		groupsWrite.DELETE("/:id", handler.DeleteGroupByIDCtrl)
		groupsWrite.POST("/check", handler.CheckAllGroupsStatusCtrl)
		groupsWrite.POST("/:id/check", handler.CheckGroupStatusCtrl)
	}
}

func setupTriggerRoutes(rg *gin.RouterGroup) {
	triggers := rg.Group("/triggers")
	{
		// Routes with privilege level 1
		triggersRead := triggers.Group("", handler.PrivilegesMiddleware(1))
		triggersRead.GET("", handler.SearchTriggersCtrl)
		triggersRead.GET("/:id", handler.GetTriggerByIDCtrl)

		// Routes with privilege level 2
		triggersWrite := triggers.Group("", handler.PrivilegesMiddleware(2))
		triggersWrite.POST("", handler.AddTriggerCtrl)
		triggersWrite.PUT("/:id", handler.UpdateTriggerCtrl)
		triggersWrite.DELETE("/:id", handler.DeleteTriggerByIDCtrl)
	}
}

func setupActionRoutes(rg *gin.RouterGroup) {
	actions := rg.Group("/actions")
	{
		// Routes with privilege level 1
		actionsRead := actions.Group("", handler.PrivilegesMiddleware(1))
		actionsRead.GET("", handler.SearchActionsCtrl)
		actionsRead.GET("/:id", handler.GetActionByIDCtrl)

		// Routes with privilege level 2
		actionsWrite := actions.Group("", handler.PrivilegesMiddleware(2))
		actionsWrite.POST("", handler.AddActionCtrl)
		actionsWrite.PUT("/:id", handler.UpdateActionCtrl)
		actionsWrite.DELETE("/:id", handler.DeleteActionByIDCtrl)
	}
}

func setupChaosRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 2 (managers/admins)
	chaos := rg.Group("/chaos", handler.PrivilegesMiddleware(2))
	chaos.POST("/alert-storm", handler.TriggerAlertStormCtrl)
}

func setupKnowledgeBaseRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 1
	kbRead := rg.Group("/knowledge-base", handler.PrivilegesMiddleware(1))
	kbRead.GET("", handler.GetAllKnowledgeBaseCtrl)
	kbRead.GET("/:id", handler.GetKnowledgeBaseByIDCtrl)

	// Routes with privilege level 2
	kbWrite := rg.Group("/knowledge-base", handler.PrivilegesMiddleware(2))
	kbWrite.POST("", handler.AddKnowledgeBaseCtrl)
	kbWrite.PUT("/:id", handler.UpdateKnowledgeBaseCtrl)
	kbWrite.DELETE("/:id", handler.DeleteKnowledgeBaseCtrl)
}

func setupPublicRoutes(rg *gin.RouterGroup) {
	public := rg.Group("/public")
	public.GET("/status", handler.GetPublicStatusSummaryCtrl)
}

func setupProviderRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 1
	providersRead := rg.Group("/providers", handler.PrivilegesMiddleware(1))
	providersRead.GET("", handler.SearchProvidersCtrl)
	providersRead.GET("/:id", handler.GetProviderByIDCtrl)

	// Routes with privilege level 2
	providersWrite := rg.Group("/providers", handler.PrivilegesMiddleware(2))
	providersWrite.POST("", handler.AddProviderCtrl)
	providersWrite.DELETE("/:id", handler.DeleteProviderByIDCtrl)
	providersWrite.PUT("/:id", handler.UpdateProviderCtrl)
	providersWrite.POST("/check", handler.CheckAllProvidersStatusCtrl)
	providersWrite.POST("/:id/check", handler.CheckProviderStatusCtrl)
}

func setupMediaRoutes(rg *gin.RouterGroup) {
	// Webhook endpoint MUST be first, before any authenticated routes
	media := rg.Group("/media")
	media.POST("/qq/message", handler.HandleQQMessageCtrl)
	media.GET("/qq/ws", handler.HandleQQWebSocket)

	// Routes with privilege level 1
	mediaRead := rg.Group("/media", handler.PrivilegesMiddleware(1))
	mediaRead.GET("", handler.SearchMediaCtrl)
	mediaRead.GET("/:id", handler.GetMediaByIDCtrl)

	// Routes with privilege level 2
	mediaWrite := rg.Group("/media", handler.PrivilegesMiddleware(2))
	mediaWrite.POST("", handler.AddMediaCtrl)
	mediaWrite.PUT("/:id", handler.UpdateMediaCtrl)
	mediaWrite.DELETE("/:id", handler.DeleteMediaByIDCtrl)
	mediaWrite.POST("/:id/test", handler.TestMediaCtrl)
}

func setupAlertRoutes(rg *gin.RouterGroup) {
	// Webhook endpoints - public, no auth required
	alerts := rg.Group("/alerts")
	alerts.POST("/webhook", handler.AlertWebhookCtrl)
	alerts.GET("/webhook/health", handler.WebhookHealthCtrl)

	// Routes with privilege level 1
	alertsRead := rg.Group("/alerts", handler.PrivilegesMiddleware(1))
	alertsRead.GET("", handler.SearchAlertsCtrl)
	alertsRead.GET("/:id", handler.GetAlertByIDCtrl)
	alertsRead.POST("/:id/consult", handler.ConsultAlertCtrl)
	alertsRead.GET("/score", handler.GetAlertScoreCtrl)

	// Routes with privilege level 2
	alertsWrite := rg.Group("/alerts", handler.PrivilegesMiddleware(2))
	alertsWrite.POST("", handler.AddAlertCtrl)
	alertsWrite.DELETE("/:id", handler.DeleteAlertByIDCtrl)
	alertsWrite.PUT("/:id", handler.UpdateAlertCtrl)
	alertsWrite.POST("/generate-test", handler.GenerateTestAlertsCtrl)
}

func setupQueueRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 2
	queue := rg.Group("/queue", handler.PrivilegesMiddleware(2))
	queue.GET("/stats", handler.QueueStatsCtrl)
}
