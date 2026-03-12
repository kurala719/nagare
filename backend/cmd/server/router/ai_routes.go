package router

import (
	"nagare/internal/api"
	"nagare/internal/mcp"

	"github.com/gin-gonic/gin"
)

func setupAIDomainRoutes(rg *gin.RouterGroup) {
	setupProviderRoutes(rg)
	setupProviderCheckRoutes(rg)
	setupProviderModelRoutes(rg)
	setupKnowledgeBaseRoutes(rg)
	setupChatRoutes(rg)
	setupConsultRoutes(rg)
	setupPacketAnalysisRoutes(rg)
	setupMcpRoutes(rg)
}

func setupChatRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 1
	chats := rg.Group("/chats", api.PrivilegesMiddleware(1))
	chats.GET("", api.SearchChatsCtrl)
	chats.POST("", api.SendChatCtrl)
}

func setupKnowledgeBaseRoutes(rg *gin.RouterGroup) {
	// Routes with privilege level 1
	kbRead := rg.Group("/knowledge-base/entries", api.PrivilegesMiddleware(1))
	kbRead.GET("", api.GetAllKnowledgeBaseCtrl)
	kbRead.GET("/:id", api.GetKnowledgeBaseByIDCtrl)

	// Routes with privilege level 2
	kbWrite := rg.Group("/knowledge-base/entries", api.PrivilegesMiddleware(2))
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
}

func setupProviderCheckRoutes(rg *gin.RouterGroup) {
	providerChecks := rg.Group("/provider-checks", api.PrivilegesMiddleware(2))
	providerChecks.POST("", api.CheckAllProvidersStatusCtrl)
	providerChecks.POST("/:id", api.CheckProviderStatusCtrl)
}

func setupProviderModelRoutes(rg *gin.RouterGroup) {
	providerModels := rg.Group("/provider-models", api.PrivilegesMiddleware(2))
	providerModels.POST("/:id", api.FetchProviderModelsCtrl)
	providerModels.POST("/direct", api.FetchModelsDirectCtrl)
}

func setupConsultRoutes(rg *gin.RouterGroup) {
	consult := rg.Group("/consult", api.PrivilegesMiddleware(1))
	consult.POST("/alerts/:id", api.ConsultAlertCtrl)
	consult.POST("/hosts/:id", api.ConsultHostCtrl)
	consult.POST("/items/:id", api.ConsultItemCtrl)
}

func setupPacketAnalysisRoutes(rg *gin.RouterGroup) {
	packets := rg.Group("/packet-analysis", api.PrivilegesMiddleware(1))
	{
		packets.GET("", api.ListPacketAnalysesCtrl)
		packets.POST("/upload", api.UploadPacketCtrl)
		packets.DELETE("/:id", api.DeletePacketAnalysisCtrl)
		packets.POST("/:id/analyze", api.StartPacketAnalysisCtrl)
	}
}

func setupMcpRoutes(rg *gin.RouterGroup) {
	// MCP routes - requires API key middleware
	mcpGroup := rg.Group("/mcp", mcp.APIKeyMiddleware())
	mcpGroup.GET("/sse", mcp.SSEHandler)
	mcpGroup.POST("/message", mcp.MessageHandler)
}
