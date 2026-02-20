package router

import (
	"github.com/gin-gonic/gin"
	"nagare/internal/api"
)

func setupUserRoutes(rg *gin.RouterGroup) {
	// Public auth routes - no authentication required
	auth := rg.Group("/auth")
	{
		auth.POST("/login", api.LoginUserCtrl)
		auth.POST("/register", api.RegisterUserCtrl)

		// Reset password - requires privilege 1
		authProtected := auth.Group("", api.PrivilegesMiddleware(1))
		authProtected.POST("/reset", api.ResetPasswordCtrl)
	}

	// Register applications - requires privilege 3
	registerApps := rg.Group("/register-applications", api.PrivilegesMiddleware(3))
	registerApps.GET("", api.ListRegisterApplicationsCtrl)
	registerApps.PUT("/:id/approve", api.ApproveRegisterApplicationCtrl)
	registerApps.PUT("/:id/reject", api.RejectRegisterApplicationCtrl)

	// Legacy register applications - requires privilege 3
	registerAppsLegacy := rg.Group("/register-application", api.PrivilegesMiddleware(3))
	registerAppsLegacy.GET("", api.ListRegisterApplicationsCtrl)
	registerAppsLegacy.PUT("/:id/approve", api.ApproveRegisterApplicationCtrl)
	registerAppsLegacy.PUT("/:id/reject", api.RejectRegisterApplicationCtrl)

	// Users routes
	users := rg.Group("/users")
	{
		// requires privilege 2 for read
		usersRead := users.Group("", api.PrivilegesMiddleware(2))
		usersRead.GET("", api.SearchUsersCtrl)
		usersRead.GET("/:id", api.GetUserByIDCtrl)

		// requires privilege 3 for write
		usersWrite := users.Group("", api.PrivilegesMiddleware(3))
		usersWrite.POST("", api.AddUserCtrl)
		usersWrite.DELETE("/:id", api.DeleteUserByIDCtrl)
		usersWrite.PUT("/:id", api.UpdateUserCtrl)
	}
}

func setupUserInformationRoutes(rg *gin.RouterGroup) {
	// Authenticated user routes - manage their own information (privilege 1)
	authenticated := rg.Group("/user-info", api.PrivilegesMiddleware(1))
	authenticated.GET("/me", api.GetUserInformationCtrl)
	authenticated.POST("/me", api.CreateUserInformationCtrl)
	authenticated.PUT("/me", api.UpdateUserInformationCtrl)
	authenticated.DELETE("/me", api.DeleteUserInformationCtrl)

	// Admin routes - manage other users' information (privilege 3)
	admin := rg.Group("/user-info", api.PrivilegesMiddleware(3))
	admin.GET("/users/:user_id", api.GetUserInformationByUserIDCtrl)
	admin.PUT("/users/:user_id", api.UpdateUserInformationByUserIDCtrl)
}
