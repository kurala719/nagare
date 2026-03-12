package router

import (
	"nagare/internal/api"

	"github.com/gin-gonic/gin"
)

func setupUsersDomainRoutes(rg *gin.RouterGroup) {
	setupUserRoutes(rg)
	setupUserInformationRoutes(rg)
	setupPublicRoutes(rg)
}

func setupUserRoutes(rg *gin.RouterGroup) {
	// Public account/session routes - no authentication required
	rg.POST("/sessions", api.LoginUserCtrl)
	rg.POST("/registrations", api.RegisterUserCtrl)
	rg.POST("/registration-codes", api.SendRegistrationCodeCtrl)
	rg.POST("/password-reset-applications", api.SubmitPasswordResetApplicationCtrl)
	rg.POST("/password-resets", api.PrivilegesMiddleware(1), api.ResetPasswordCtrl)

	// Registration applications - requires privilege 3
	registerApps := rg.Group("/registration-applications", api.PrivilegesMiddleware(3))
	registerApps.GET("", api.ListRegisterApplicationsCtrl)
	registerApps.POST("/:id/approvals", api.ApproveRegisterApplicationCtrl)
	registerApps.POST("/:id/rejections", api.RejectRegisterApplicationCtrl)

	// Password reset applications - requires privilege 3
	resetApps := rg.Group("/password-reset-applications", api.PrivilegesMiddleware(3))
	resetApps.GET("", api.ListPasswordResetApplicationsCtrl)
	resetApps.POST("/:id/approvals", api.ApprovePasswordResetApplicationCtrl)
	resetApps.POST("/:id/rejections", api.RejectPasswordResetApplicationCtrl)

	// Users routes
	users := rg.Group("")
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
	// Authenticated user routes - manage their own profile (privilege 1)
	authenticated := rg.Group("/profile", api.PrivilegesMiddleware(1))
	{
		authenticated.GET("", api.GetMyProfileCtrl)
		authenticated.PUT("", api.UpdateMyProfileCtrl)
		authenticated.POST("/avatar", api.UploadAvatarCtrl)
	}

	// Admin routes - manage other users' information (privilege 3)
	admin := rg.Group("/profiles", api.PrivilegesMiddleware(3))
	{
		admin.GET("/:id", api.GetUserByIDCtrl)
		admin.PUT("/:id", api.UpdateUserCtrl)
	}
}

func setupPublicRoutes(rg *gin.RouterGroup) {
	rg.GET("/status", api.GetPublicStatusSummaryCtrl)
}
