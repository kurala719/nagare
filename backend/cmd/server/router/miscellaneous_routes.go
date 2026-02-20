package router

import (
	"github.com/gin-gonic/gin"
	"nagare/internal/api"
)

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

func setupActionRoutes(rg *gin.RouterGroup) {
	actions := rg.Group("/actions")
	{
		// Routes with privilege level 1
		actionsRead := actions.Group("", api.PrivilegesMiddleware(1))
		actionsRead.GET("", api.SearchActionsCtrl)
		actionsRead.GET("/:id", api.GetActionByIDCtrl)

		// Routes with privilege level 2
		actionsWrite := actions.Group("", api.PrivilegesMiddleware(2))
		actionsWrite.POST("", api.AddActionCtrl)
		actionsWrite.PUT("/:id", api.UpdateActionCtrl)
		actionsWrite.DELETE("/:id", api.DeleteActionByIDCtrl)
	}
}

func setupChaosRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 2 (managers/admins)
	chaos := rg.Group("/chaos", api.PrivilegesMiddleware(2))
	chaos.POST("/alert-storm", api.TriggerAlertStormCtrl)
}

func setupKnowledgeBaseRoutes(rg *gin.RouterGroup) {
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

func setupPublicRoutes(rg *gin.RouterGroup) {
	public := rg.Group("/public")
	public.GET("/status", api.GetPublicStatusSummaryCtrl)
}

func setupProviderRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 1
	providersRead := rg.Group("/providers", api.PrivilegesMiddleware(1))
	providersRead.GET("", api.SearchProvidersCtrl)
	providersRead.GET("/:id", api.GetProviderByIDCtrl)

	// Routes with privilege level 2
	providersWrite := rg.Group("/providers", api.PrivilegesMiddleware(2))
	providersWrite.POST("", api.AddProviderCtrl)
	providersWrite.DELETE("/:id", api.DeleteProviderByIDCtrl)
	providersWrite.PUT("/:id", api.UpdateProviderCtrl)
	providersWrite.POST("/check", api.CheckAllProvidersStatusCtrl)
	providersWrite.POST("/:id/check", api.CheckProviderStatusCtrl)
}

func setupMediaRoutes(rg *gin.RouterGroup) {
	// Webhook endpoint MUST be first, before any authenticated routes
	media := rg.Group("/media")
	media.POST("/qq/message", api.HandleQQMessageCtrl)

	// Routes with privilege level 1
	mediaRead := rg.Group("/media", api.PrivilegesMiddleware(1))
	mediaRead.GET("", api.SearchMediaCtrl)
	mediaRead.GET("/:id", api.GetMediaByIDCtrl)

	// Routes with privilege level 2
	mediaWrite := rg.Group("/media", api.PrivilegesMiddleware(2))
	mediaWrite.POST("", api.AddMediaCtrl)
	mediaWrite.PUT("/:id", api.UpdateMediaCtrl)
	mediaWrite.DELETE("/:id", api.DeleteMediaByIDCtrl)
}

func setupMediaTypeRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 1
	mediaTypesRead := rg.Group("/media-types", api.PrivilegesMiddleware(1))
	mediaTypesRead.GET("", api.SearchMediaTypesCtrl)
	mediaTypesRead.GET("/:id", api.GetMediaTypeByIDCtrl)

	// Routes with privilege level 2
	mediaTypesWrite := rg.Group("/media-types", api.PrivilegesMiddleware(2))
	mediaTypesWrite.POST("", api.AddMediaTypeCtrl)
	mediaTypesWrite.PUT("/:id", api.UpdateMediaTypeCtrl)
	mediaTypesWrite.DELETE("/:id", api.DeleteMediaTypeByIDCtrl)
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
	alertsRead.POST("/:id/consult", api.ConsultAlertCtrl)
	alertsRead.GET("/score", api.GetAlertScoreCtrl)

	// Routes with privilege level 2
	alertsWrite := rg.Group("/alerts", api.PrivilegesMiddleware(2))
	alertsWrite.POST("", api.AddAlertCtrl)
	alertsWrite.DELETE("/:id", api.DeleteAlertByIDCtrl)
	alertsWrite.PUT("/:id", api.UpdateAlertCtrl)
	alertsWrite.POST("/generate-test", api.GenerateTestAlertsCtrl)
}

func setupQueueRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 2
	queue := rg.Group("/queue", api.PrivilegesMiddleware(2))
	queue.GET("/stats", api.QueueStatsCtrl)
}
