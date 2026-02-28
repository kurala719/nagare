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
	setupAllRoutes(apiGroup)

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
	// --- 1. ALERT GROUP ---
	alertGroup := rg.Group("/alert")
	setupAlarmRoutes(alertGroup)
	setupAlertRoutes(alertGroup)
	setupTriggerRoutes(alertGroup)

	// --- 2. MONITORING GROUP ---
	monitoringGroup := rg.Group("/monitoring")
	setupMonitorRoutes(monitoringGroup)
	setupGroupRoutes(monitoringGroup)
	setupHostRoutes(monitoringGroup)
	setupItemRoutes(monitoringGroup)
	monitoringGroup.POST("/snmp-poll-direct/:id", api.TestSNMPCtrl)

	// --- 3. MAINTENANCE GROUP ---
	maintenanceGroup := rg.Group("/maintenance")
	maintenanceGroup.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	setupIMRoutes(maintenanceGroup)

	// --- 4. SYSTEM GROUP ---
	systemGroup := rg.Group("/system")
	setupLogRoutes(systemGroup)
	setupAuditLogRoutes(systemGroup)
	setupRetentionRoutes(systemGroup)
	setupConfigurationRoutes(systemGroup)
	setupSystemRoutes(systemGroup)
	setupMcpRoutes(systemGroup)

	// --- 5. DELIVERY GROUP ---
	deliveryGroup := rg.Group("/delivery")
	setupActionRoutes(deliveryGroup)
	setupMediaRoutes(deliveryGroup)
	setupSiteMessageRoutes(deliveryGroup)
	setupQQWhitelistRoutes(deliveryGroup)

	// --- 6. ANALYSIS GROUP ---
	analysisGroup := rg.Group("/analysis")
	setupAnalyticsRoutes(analysisGroup)
	setupReportRoutes(analysisGroup)
	setupPacketAnalysisRoutes(analysisGroup)
	setupChaosRoutes(analysisGroup)

	// --- 7. USERS GROUP ---
	usersGroup := rg.Group("/users")
	setupUserRoutes(usersGroup)
	setupUserInformationRoutes(usersGroup)
	setupPublicRoutes(usersGroup)

	// --- 8. AI SERVICE GROUP ---
	aiServiceGroup := rg.Group("/ai")
	setupProviderRoutes(aiServiceGroup)
	setupKnowledgeBaseRoutes(aiServiceGroup)
	setupChatRoutes(aiServiceGroup)
}
