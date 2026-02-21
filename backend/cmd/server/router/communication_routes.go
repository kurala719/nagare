package router

import (
	"github.com/gin-gonic/gin"
	"nagare/internal/api"
)

func setupIMRoutes(rg *gin.RouterGroup) {
	im := rg.Group("/im")
	im.POST("/command", api.IMCommandCtrl)
}

func setupChatRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 1
	chats := rg.Group("/chats", api.PrivilegesMiddleware(1))
	chats.GET("", api.SearchChatsCtrl)
	chats.POST("", api.SendChatCtrl)
}

func setupSiteMessageRoutes(rg *gin.RouterGroup) {
	messages := rg.Group("/site-messages")
	{
		// Public WebSocket endpoint (auth via token in query handled by middleware)
		messages.GET("/ws", api.PrivilegesMiddleware(1), api.HandleSiteMessageWS)

		// Protected routes
		msgProtected := messages.Group("", api.PrivilegesMiddleware(1))
		msgProtected.GET("", api.GetSiteMessagesCtrl)
		msgProtected.GET("/unread-count", api.GetUnreadSiteMessagesCountCtrl)
		msgProtected.PUT("/:id/read", api.MarkSiteMessageAsReadCtrl)
		msgProtected.PUT("/read-all", api.MarkAllSiteMessagesAsReadCtrl)
		msgProtected.DELETE("/:id", api.DeleteSiteMessageCtrl)
	}
}

func setupReportRoutes(rg *gin.RouterGroup) {
	reports := rg.Group("/reports", api.PrivilegesMiddleware(2))
	{
		reports.GET("", api.ListReportsCtrl)
		reports.GET("/:id", api.GetReportCtrl)
		reports.GET("/:id/content", api.GetReportContentCtrl)
		reports.GET("/config", api.GetReportConfigCtrl)
		reports.PUT("/config", api.UpdateReportConfigCtrl)
		reports.POST("/generate/weekly", api.GenerateWeeklyReportCtrl)
		reports.POST("/generate/monthly", api.GenerateMonthlyReportCtrl)
		reports.POST("/generate/custom", api.GenerateCustomReportCtrl)
		reports.DELETE("/:id", api.DeleteReportCtrl)
		reports.GET("/:id/download", api.DownloadReportCtrl)
	}
}

func setupQQWhitelistRoutes(rg *gin.RouterGroup) {
	whitelist := rg.Group("/qq-whitelist", api.PrivilegesMiddleware(2))
	{
		whitelist.GET("", api.GetQQWhitelistCtrl)
		whitelist.POST("", api.AddQQWhitelistCtrl)
		whitelist.PUT("/:id", api.UpdateQQWhitelistCtrl)
		whitelist.DELETE("/:id", api.DeleteQQWhitelistCtrl)
	}
}
