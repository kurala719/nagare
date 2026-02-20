package router

import (
	"nagare/internal/adapter/handler"

	"github.com/gin-gonic/gin"
)

func setupIMRoutes(rg *gin.RouterGroup) {
	im := rg.Group("/im")
	im.POST("/command", handler.IMCommandCtrl)
}

func setupChatRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 1
	chats := rg.Group("/chats", handler.PrivilegesMiddleware(1))
	chats.GET("", handler.SearchChatsCtrl)
	chats.POST("", handler.SendChatCtrl)
}

func setupSiteMessageRoutes(rg *gin.RouterGroup) {
	messages := rg.Group("/site-messages")
	{
		// Public WebSocket endpoint (auth via token in query handled by middleware)
		messages.GET("/ws", handler.PrivilegesMiddleware(1), handler.HandleSiteMessageWS)

		// Protected routes
		msgProtected := messages.Group("", handler.PrivilegesMiddleware(1))
		msgProtected.GET("", handler.GetSiteMessagesCtrl)
		msgProtected.GET("/unread-count", handler.GetUnreadSiteMessagesCountCtrl)
		msgProtected.PUT("/:id/read", handler.MarkSiteMessageAsReadCtrl)
		msgProtected.PUT("/read-all", handler.MarkAllSiteMessagesAsReadCtrl)
		msgProtected.DELETE("/:id", handler.DeleteSiteMessageCtrl)
	}
}

func setupReportRoutes(rg *gin.RouterGroup) {
	reports := rg.Group("/reports", handler.PrivilegesMiddleware(2))
	{
		reports.GET("", handler.ListReportsCtrl)
		reports.GET("/:id", handler.GetReportCtrl)
		reports.GET("/:id/content", handler.GetReportContentCtrl)
		reports.GET("/config", handler.GetReportConfigCtrl)
		reports.PUT("/config", handler.UpdateReportConfigCtrl)
		reports.POST("/generate/weekly", handler.GenerateWeeklyReportCtrl)
		reports.POST("/generate/monthly", handler.GenerateMonthlyReportCtrl)
		reports.DELETE("/:id", handler.DeleteReportCtrl)
		reports.GET("/:id/download", handler.DownloadReportCtrl)
	}
}

func setupQQWhitelistRoutes(rg *gin.RouterGroup) {
	whitelist := rg.Group("/qq-whitelist", handler.PrivilegesMiddleware(2))
	{
		whitelist.GET("", handler.GetQQWhitelistCtrl)
		whitelist.POST("", handler.AddQQWhitelistCtrl)
		whitelist.PUT("/:id", handler.UpdateQQWhitelistCtrl)
		whitelist.DELETE("/:id", handler.DeleteQQWhitelistCtrl)
	}
}
