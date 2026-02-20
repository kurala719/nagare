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
	// Public WebSocket endpoint (auth via token in query handled by middleware)
	rg.GET("/site-messages/ws", api.PrivilegesMiddleware(1), api.HandleSiteMessageWS)

	// Protected routes
	messages := rg.Group("/site-messages", api.PrivilegesMiddleware(1))
	messages.GET("", api.GetSiteMessagesCtrl)
	messages.GET("/unread-count", api.GetUnreadSiteMessagesCountCtrl)
	messages.PUT("/:id/read", api.MarkSiteMessageAsReadCtrl)
	messages.PUT("/read-all", api.MarkAllSiteMessagesAsReadCtrl)
	messages.DELETE("/:id", api.DeleteSiteMessageCtrl)
}

func setupReportRoutes(rg *gin.RouterGroup) {
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

func setupQQWhitelistRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 2
	whitelist := rg.Group("/qq-whitelist", api.PrivilegesMiddleware(2))
	whitelist.GET("", api.GetQQWhitelistCtrl)
	whitelist.POST("", api.AddQQWhitelistCtrl)
	whitelist.PUT("/:id", api.UpdateQQWhitelistCtrl)
	whitelist.DELETE("/:id", api.DeleteQQWhitelistCtrl)
}
