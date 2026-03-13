package router

import (
	"nagare/internal/api"
	"nagare/internal/mcp"

	"github.com/gin-gonic/gin"
)

func setupAIDomainRoutes(rg *gin.RouterGroup) {
	setupAISettingsRoutes(rg)
	setupProviderRoutes(rg)
	setupKnowledgeBaseRoutes(rg)
	setupChatRoutes(rg)
	setupConsultRoutes(rg)
	setupPacketAnalysisRoutes(rg)
	setupMcpRoutes(rg)
}

func setupAISettingsRoutes(rg *gin.RouterGroup) {
	settings := rg.Group("/settings", api.PrivilegesMiddleware(1))
	settings.GET("", api.GetAIConfigCtrl)
}

func setupChatRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 1
	chats := rg.Group("/chats", api.PrivilegesMiddleware(1))
	chats.GET("", api.SearchChatsCtrl)
	chats.POST("", api.SendChatCtrl)
}

func setupKnowledgeBaseRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 2
	kbRead := rg.Group("/knowledge-base", api.PrivilegesMiddleware(2))
	kbRead.GET("", api.GetAllKnowledgeBaseCtrl)
	kbRead.GET("/:id", api.GetKnowledgeBaseByIDCtrl)

	// Routes with privilege level 2
	kbWrite := rg.Group("/knowledge-base", api.PrivilegesMiddleware(2))
	kbWrite.POST("", api.AddKnowledgeBaseCtrl)
	kbWrite.PUT("/:id", api.UpdateKnowledgeBaseCtrl)
	kbWrite.DELETE("/:id", api.DeleteKnowledgeBaseCtrl)
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
	providersWrite.POST("/checks", api.CheckAllProvidersStatusCtrl)
	providersWrite.POST("/:id/checks", api.CheckProviderStatusCtrl)
	providersWrite.POST("/:id/models", api.FetchProviderModelsCtrl)
	providersWrite.POST("/models", api.FetchModelsDirectCtrl)
}

func setupConsultRoutes(rg *gin.RouterGroup) {
	consultationAlerts := rg.Group("/alerts", api.PrivilegesMiddleware(1))
	consultationAlerts.POST("/:id/consultations", api.ConsultAlertCtrl)

	consultationHosts := rg.Group("/hosts", api.PrivilegesMiddleware(1))
	consultationHosts.POST("/:id/consultations", api.ConsultHostCtrl)

	consultationItems := rg.Group("/items", api.PrivilegesMiddleware(1))
	consultationItems.POST("/:id/consultations", api.ConsultItemCtrl)
}

func setupPacketAnalysisRoutes(rg *gin.RouterGroup) {
	packets := rg.Group("/packet-analyses", api.PrivilegesMiddleware(2))
	{
		packets.GET("", api.ListPacketAnalysesCtrl)
		packets.POST("", api.UploadPacketCtrl)
		packets.DELETE("/:id", api.DeletePacketAnalysisCtrl)
		packets.POST("/:id/runs", api.StartPacketAnalysisCtrl)
	}
}

func setupMcpRoutes(rg *gin.RouterGroup) {
	// MCP routes - requires API key middleware
	mcpGroup := rg.Group("/mcp", mcp.APIKeyMiddleware())
	mcpGroup.GET("/sessions", mcp.SSEHandler)
	mcpGroup.POST("/messages", mcp.MessageHandler)
}
