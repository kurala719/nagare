package router

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nagare/internal/api"
	"nagare/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// InitRouter initializes and starts the HTTP router
func InitRouter() {
	gin.SetMode(gin.DebugMode) // Enable debug mode to see route registration
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(api.CORSMiddleware()) // Add CORS support

	r.RedirectTrailingSlash = false
	r.Use(api.RequestIDMiddleware())
	r.Use(api.AccessLogMiddleware())
	r.Static("/avatars", "./public/avatars")

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, api.APIResponse{
			Success: false,
			Error:   "api route not found",
		})
	})
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, api.APIResponse{
			Success: false,
			Error:   "method not allowed",
		})
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Nagare Backend is running",
			"version": "1.0.1",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// Direct SNMP poll route (moved inside setupAllRoutes)
	// r.POST("/api/v1/snmp-poll-direct/:id", api.TestSNMPCtrl)

	// Setup all routes
	apiGroup := r.Group("/api/v1")
	apiGroup.Use(api.AuditLogMiddleware())
	apiGroup.POST("/snmp-poll-direct/:id", api.TestSNMPCtrl)
	setupAllRoutes(apiGroup)
	setupMcpRoutes(apiGroup)

	// Start WebSocket Hub
	go service.GlobalHub.Run()

	port := viper.GetInt("system.port")
	if port == 0 {
		port = 8080
	}

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	service.LogSystem("info", "server starting", map[string]interface{}{"port": port}, nil, "")
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			service.LogSystem("error", "failed to start server", map[string]interface{}{"error": err.Error()}, nil, "")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		service.LogSystem("error", "server shutdown error", map[string]interface{}{"error": err.Error()}, nil, "")
	}
}

func setupAllRoutes(rg *gin.RouterGroup) {
	setupConfigurationRoutes(rg)
	setupMonitorRoutes(rg)
	setupAlarmRoutes(rg)
	setupGroupRoutes(rg)
	setupHostRoutes(rg)
	setupAlertRoutes(rg)
	setupSystemRoutes(rg)
	setupPublicRoutes(rg)
	setupIMRoutes(rg)
	setupMediaRoutes(rg)
	setupActionRoutes(rg)
	setupTriggerRoutes(rg)
	setupLogRoutes(rg)
	setupAuditLogRoutes(rg)
	setupAnalyticsRoutes(rg)
	setupChaosRoutes(rg)
	setupItemRoutes(rg)
	setupChatRoutes(rg)
	setupProviderRoutes(rg)
	setupKnowledgeBaseRoutes(rg)
	setupUserRoutes(rg)
	setupUserInformationRoutes(rg)
	setupQQWhitelistRoutes(rg)
	setupReportRoutes(rg)
	setupSiteMessageRoutes(rg)
	setupAnsibleRoutes(rg)
	setupRetentionRoutes(rg)
}
