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

	"nagare/internal/adapter/handler"
	"nagare/internal/core/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// InitRouter initializes and starts the HTTP router
func InitRouter() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.RedirectTrailingSlash = true
	r.Use(handler.RequestIDMiddleware())
	r.Use(handler.AccessLogMiddleware())

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, handler.APIResponse{
			Success: false,
			Error:   "resource not found",
		})
	})
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, handler.APIResponse{
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

	// Direct SNMP poll route
	r.POST("/api/v1/snmp-poll-direct/:id", handler.TestSNMPCtrl)

	// Setup all routes
	apiGroup := r.Group("/api/v1")
	apiGroup.Use(handler.AuditLogMiddleware())
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
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
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
	setupQueueRoutes(rg)
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
