package router

import (
	"nagare/internal/adapter/handler"

	"github.com/gin-gonic/gin"
)

func setupAnsibleRoutes(rg *gin.RouterGroup) {
	ansible := rg.Group("/ansible")
	{
		// Public Dynamic Inventory (used by ansible-playbook command)
		ansible.GET("/inventory", handler.GetAnsibleInventoryCtrl)

		// Playbooks
		pbs := ansible.Group("/playbooks", handler.PrivilegesMiddleware(2))
		pbs.GET("", handler.ListPlaybooksCtrl)
		pbs.POST("", handler.CreatePlaybookCtrl)
		pbs.GET("/:id", handler.GetPlaybookCtrl)
		pbs.PUT("/:id", handler.UpdatePlaybookCtrl)
		pbs.DELETE("/:id", handler.DeletePlaybookCtrl)
		pbs.POST("/:id/run", handler.RunPlaybookCtrl)
		pbs.POST("/recommend", handler.RecommendPlaybookCtrl)

		// Jobs
		jobs := ansible.Group("/jobs", handler.PrivilegesMiddleware(2))
		jobs.GET("", handler.ListAnsibleJobsCtrl)
		jobs.GET("/:id", handler.GetAnsibleJobCtrl)
	}
}
