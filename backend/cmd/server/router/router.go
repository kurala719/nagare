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
	r := gin.Default()
	
	// Add global logger to see every request reaching the backend
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		fmt.Printf("HTTP Debug: %s %s %d (%v)\n", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration)
	})

	r.RedirectTrailingSlash = true
	r.Use(api.RequestIDMiddleware())
	r.Use(api.AccessLogMiddleware())

	r.NoRoute(func(c *gin.Context) {
		fmt.Printf("HTTP Debug: 404 Not Found: %s %s\n", c.Request.Method, c.Request.URL.Path)
		c.JSON(http.StatusNotFound, api.APIResponse{
			Success: false,
			Error:   "resource not found",
		})
	})
	r.NoMethod(func(c *gin.Context) {
		fmt.Printf("HTTP Debug: 405 Method Not Allowed: %s %s\n", c.Request.Method, c.Request.URL.Path)
		c.JSON(http.StatusMethodNotAllowed, api.APIResponse{
			Success: false,
			Error:   "method not allowed",
		})
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Nagare Backend is running (DEBUG VERSION)",
			"version": "1.0.1",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// Direct SNMP test route for debugging
	fmt.Println("[Router] Registering direct SNMP test route...")
	r.POST("/api/v1/snmp-poll-direct/:id", api.TestSNMPCtrl)

	// Setup all routes
	apiGroup := r.Group("/api/v1")
	apiGroup.Use(api.AuditLogMiddleware())
	setupAllRoutes(apiGroup)
	setupMcpRoutes(apiGroup)

	// Debug: print all routes
	for _, rt := range r.Routes() {
		fmt.Printf("Route Registered: %s %s\n", rt.Method, rt.Path)
	}

	// Start WebSocket Hub	go service.GlobalHub.Run()

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
	setupMediaTypeRoutes(rg)
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

