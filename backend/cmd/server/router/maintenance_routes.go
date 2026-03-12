package router

import (
	"net/http"

	"nagare/internal/api"

	"github.com/gin-gonic/gin"
)

func setupMaintenanceDomainRoutes(rg *gin.RouterGroup) {
	rg.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	setupSSHRoutes(rg)
}

func setupSSHRoutes(rg *gin.RouterGroup) {
	ssh := rg.Group("/ssh", api.PrivilegesMiddleware(1))
	ssh.GET("/hosts/:id", api.HandleWebSSH)
	ssh.GET("/terminal", api.HandleWebSSH)
}
