package router

import (
	"nagare/internal/adapter/handler"

	"github.com/gin-gonic/gin"
)

func setupUserRoutes(rg *gin.RouterGroup) {
	// Public auth routes - no authentication required
	auth := rg.Group("/auth")
	{
		auth.POST("/login", handler.LoginUserCtrl)
		auth.POST("/register", handler.RegisterUserCtrl)
		auth.POST("/send-code", handler.SendRegistrationCodeCtrl)
		auth.POST("/reset-request", handler.SubmitPasswordResetApplicationCtrl)

		// Reset password - requires privilege 1
		authProtected := auth.Group("", handler.PrivilegesMiddleware(1))
		authProtected.POST("/reset", handler.ResetPasswordCtrl)
	}

	// Register applications - requires privilege 3
	registerApps := rg.Group("/register-applications", handler.PrivilegesMiddleware(3))
	registerApps.GET("", handler.ListRegisterApplicationsCtrl)
	registerApps.PUT("/:id/approve", handler.ApproveRegisterApplicationCtrl)
	registerApps.PUT("/:id/reject", handler.RejectRegisterApplicationCtrl)

	// Password reset applications - requires privilege 3
	resetApps := rg.Group("/reset-applications", handler.PrivilegesMiddleware(3))
	resetApps.GET("", handler.ListPasswordResetApplicationsCtrl)
	resetApps.PUT("/:id/approve", handler.ApprovePasswordResetApplicationCtrl)
	resetApps.PUT("/:id/reject", handler.RejectPasswordResetApplicationCtrl)

	// Legacy register applications - requires privilege 3
	registerAppsLegacy := rg.Group("/register-application", handler.PrivilegesMiddleware(3))
	registerAppsLegacy.GET("", handler.ListRegisterApplicationsCtrl)
	registerAppsLegacy.PUT("/:id/approve", handler.ApproveRegisterApplicationCtrl)
	registerAppsLegacy.PUT("/:id/reject", handler.RejectRegisterApplicationCtrl)

	// Users routes
	users := rg.Group("/users")
	{
		// requires privilege 2 for read
		usersRead := users.Group("", handler.PrivilegesMiddleware(2))
		usersRead.GET("", handler.SearchUsersCtrl)
		usersRead.GET("/:id", handler.GetUserByIDCtrl)

		// requires privilege 3 for write
		usersWrite := users.Group("", handler.PrivilegesMiddleware(3))
		usersWrite.POST("", handler.AddUserCtrl)
		usersWrite.DELETE("/:id", handler.DeleteUserByIDCtrl)
		usersWrite.PUT("/:id", handler.UpdateUserCtrl)
	}
}

func setupUserInformationRoutes(rg *gin.RouterGroup) {
	// Authenticated user routes - manage their own profile (privilege 1)
	authenticated := rg.Group("/user-info", handler.PrivilegesMiddleware(1))
	{
		authenticated.GET("/me", handler.GetMyProfileCtrl)
		authenticated.PUT("/me", handler.UpdateMyProfileCtrl)
		authenticated.POST("/me", handler.UpdateMyProfileCtrl) // Map POST to update for compatibility
	}

	// Admin routes - manage other users' information (privilege 3)
	admin := rg.Group("/user-info", handler.PrivilegesMiddleware(3))
	{
		admin.GET("/users/:id", handler.GetUserByIDCtrl)
		admin.PUT("/users/:id", handler.UpdateUserCtrl)
	}
}
