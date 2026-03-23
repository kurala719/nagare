package router

import (
	"nagare/internal/api"

	"github.com/gin-gonic/gin"
)

func setupAIDomainRoutes(rg *gin.RouterGroup) {
	setupAISettingsRoutes(rg)
	setupProviderRoutes(rg)
	setupKnowledgeBaseRoutes(rg)
	setupChatRoutes(rg)
	setupConsultRoutes(rg)
	setupPacketAnalysisRoutes(rg)
}

func setupAISettingsRoutes(rg *gin.RouterGroup) {
	settings := rg.Group("/settings", api.PrivilegesMiddleware(1))
	settings.GET("", api.GetAIConfigCtrl)
}

func setupChatRoutes(rg *gin.RouterGroup) {
	chats := rg.Group("/chats", api.PrivilegesMiddleware(1))
	chats.GET("", api.SearchChatsCtrl)
	chats.POST("", api.SendChatCtrl)
}

func setupKnowledgeBaseRoutes(rg *gin.RouterGroup) {
	kb := rg.Group("/knowledge-base", api.PrivilegesMiddleware(2))
	kb.GET("", api.GetAllKnowledgeBaseCtrl)
	kb.GET("/:id", api.GetKnowledgeBaseByIDCtrl)
	kb.POST("", api.AddKnowledgeBaseCtrl)
	kb.PUT("/:id", api.UpdateKnowledgeBaseCtrl)
	kb.DELETE("/:id", api.DeleteKnowledgeBaseCtrl)
}

func setupProviderRoutes(rg *gin.RouterGroup) {
	providersRead := rg.Group("/providers", api.PrivilegesMiddleware(1))
	providersRead.GET("", api.SearchProvidersCtrl)
	providersRead.GET("/:id", api.GetProviderByIDCtrl)

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
	targets := []struct {
		path    string
		handler gin.HandlerFunc
	}{
		{path: "/alerts", handler: api.ConsultAlertCtrl},
		{path: "/hosts", handler: api.ConsultHostCtrl},
		{path: "/items", handler: api.ConsultItemCtrl},
	}

	for _, target := range targets {
		group := rg.Group(target.path, api.PrivilegesMiddleware(1))
		group.POST("/:id/consultations", target.handler)
	}
}

func setupPacketAnalysisRoutes(rg *gin.RouterGroup) {
	packets := rg.Group("/packet-analyses", api.PrivilegesMiddleware(2))
	packets.GET("", api.ListPacketAnalysesCtrl)
	packets.POST("", api.UploadPacketCtrl)
	packets.DELETE("/:id", api.DeletePacketAnalysisCtrl)
	packets.POST("/:id/runs", api.StartPacketAnalysisCtrl)
}
