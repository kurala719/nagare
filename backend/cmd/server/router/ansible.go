package router

import (
	"github.com/gin-gonic/gin"
	"nagare/internal/api"
)

func setupAnsibleRoutes(rg *gin.RouterGroup) {
	// Public Dynamic Inventory (used by ansible-playbook command)
	rg.GET("/ansible/inventory", api.GetAnsibleInventoryCtrl)

	// Playbooks
	pbs := rg.Group("/ansible/playbooks", api.PrivilegesMiddleware(2))
	pbs.GET("", api.ListPlaybooksCtrl)
	pbs.POST("", api.CreatePlaybookCtrl)
	pbs.GET("/:id", api.GetPlaybookCtrl)
	pbs.PUT("/:id", api.UpdatePlaybookCtrl)
	pbs.DELETE("/:id", api.DeletePlaybookCtrl)
	pbs.POST("/:id/run", api.RunPlaybookCtrl)
	pbs.POST("/recommend", api.RecommendPlaybookCtrl)

	// Jobs
	jobs := rg.Group("/ansible/jobs", api.PrivilegesMiddleware(2))
	jobs.GET("", api.ListAnsibleJobsCtrl)
	jobs.GET("/:id", api.GetAnsibleJobCtrl)
}
