package router

import (
	"nagare/internal/api"

	"github.com/gin-gonic/gin"
)

func setupDeliveryDomainRoutes(rg *gin.RouterGroup) {
	setupActionRoutes(rg)
	setupMediaRoutes(rg)
	setupSiteMessageRoutes(rg)
}

func setupActionRoutes(rg *gin.RouterGroup) {
	actions := rg.Group("/actions")
	{
		// Routes with privilege level 2
		actionsRead := actions.Group("", api.PrivilegesMiddleware(2))
		actionsRead.GET("", api.SearchActionsCtrl)
		actionsRead.GET("/:id", api.GetActionByIDCtrl)

		// Routes with privilege level 2
		actionsWrite := actions.Group("", api.PrivilegesMiddleware(2))
		actionsWrite.POST("", api.AddActionCtrl)
		actionsWrite.PUT("/:id", api.UpdateActionCtrl)
		actionsWrite.DELETE("/:id", api.DeleteActionByIDCtrl)
		actionsWrite.POST("/:id/test-runs", api.TestActionCtrl)
	}
}

func setupMediaRoutes(rg *gin.RouterGroup) {
	// Webhook endpoint MUST be first, before any authenticated routes
	media := rg.Group("/media")
	media.POST("/qq/messages", api.HandleQQMessageCtrl)
	media.GET("/qq/socket-sessions", api.HandleQQWebSocket)

	// Routes with privilege level 2
	mediaRead := rg.Group("/media", api.PrivilegesMiddleware(2))
	mediaRead.GET("", api.SearchMediaCtrl)
	mediaRead.GET("/:id", api.GetMediaByIDCtrl)

	// Routes with privilege level 2
	mediaWrite := rg.Group("/media", api.PrivilegesMiddleware(2))
	mediaWrite.POST("", api.AddMediaCtrl)
	mediaWrite.PUT("/:id", api.UpdateMediaCtrl)
	mediaWrite.DELETE("/:id", api.DeleteMediaByIDCtrl)
	mediaWrite.POST("/:id/test-runs", api.TestMediaCtrl)
}

func setupSiteMessageRoutes(rg *gin.RouterGroup) {
	messages := rg.Group("/site-messages")
	{
		// Public WebSocket endpoint (auth via token in query handled by middleware)
		messages.GET("/socket-sessions", api.PrivilegesMiddleware(1), api.HandleSiteMessageWS)

		// Protected routes
		msgProtected := messages.Group("", api.PrivilegesMiddleware(1))
		msgProtected.GET("", api.GetSiteMessagesCtrl)
		msgProtected.GET("/unread/count", api.GetUnreadSiteMessagesCountCtrl)
		msgProtected.PUT("/:id/read-status", api.MarkSiteMessageAsReadCtrl)
		msgProtected.PUT("/read-status", api.MarkAllSiteMessagesAsReadCtrl)
		msgProtected.DELETE("/:id", api.DeleteSiteMessageCtrl)
	}
}
