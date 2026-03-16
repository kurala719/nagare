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
	actionsPrivileged := actions.Group("", api.PrivilegesMiddleware(2))
	actionsPrivileged.GET("", api.SearchActionsCtrl)
	actionsPrivileged.GET("/:id", api.GetActionByIDCtrl)
	actionsPrivileged.POST("", api.AddActionCtrl)
	actionsPrivileged.PUT("/:id", api.UpdateActionCtrl)
	actionsPrivileged.DELETE("/:id", api.DeleteActionByIDCtrl)
	actionsPrivileged.POST("/:id/test-runs", api.TestActionCtrl)
}

func setupMediaRoutes(rg *gin.RouterGroup) {
	media := rg.Group("/media")
	media.POST("/qq/messages", api.HandleQQMessageCtrl)
	media.GET("/qq/socket-sessions", api.HandleQQWebSocket)

	mediaPrivileged := rg.Group("/media", api.PrivilegesMiddleware(2))
	mediaPrivileged.GET("", api.SearchMediaCtrl)
	mediaPrivileged.GET("/:id", api.GetMediaByIDCtrl)
	mediaPrivileged.POST("", api.AddMediaCtrl)
	mediaPrivileged.PUT("/:id", api.UpdateMediaCtrl)
	mediaPrivileged.DELETE("/:id", api.DeleteMediaByIDCtrl)
	mediaPrivileged.POST("/:id/test-runs", api.TestMediaCtrl)
}

func setupSiteMessageRoutes(rg *gin.RouterGroup) {
	messages := rg.Group("/site-messages")
	messages.GET("/socket-sessions", api.PrivilegesMiddleware(1), api.HandleSiteMessageWS)

	msgProtected := messages.Group("", api.PrivilegesMiddleware(1))
	msgProtected.GET("", api.GetSiteMessagesCtrl)
	msgProtected.GET("/unread/count", api.GetUnreadSiteMessagesCountCtrl)
	msgProtected.PUT("/:id/read-status", api.MarkSiteMessageAsReadCtrl)
	msgProtected.PUT("/read-status", api.MarkAllSiteMessagesAsReadCtrl)
	msgProtected.DELETE("/:id", api.DeleteSiteMessageCtrl)
}
