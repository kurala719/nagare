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
	rg.POST("/sessions", api.LoginUserCtrl)
	rg.POST("/registrations", api.RegisterUserCtrl)
	rg.POST("/registration-codes", api.SendRegistrationCodeCtrl)
	rg.POST("/password-reset-applications", api.SubmitPasswordResetApplicationCtrl)
	rg.POST("/password-resets", api.PrivilegesMiddleware(1), api.ResetPasswordCtrl)

	registerApps := rg.Group("/registration-applications", api.PrivilegesMiddleware(3))
	registerApps.GET("", api.ListRegisterApplicationsCtrl)
	registerApps.POST("/:id/approvals", api.ApproveRegisterApplicationCtrl)
	registerApps.POST("/:id/rejections", api.RejectRegisterApplicationCtrl)

	resetApps := rg.Group("/password-reset-applications", api.PrivilegesMiddleware(3))
	resetApps.GET("", api.ListPasswordResetApplicationsCtrl)
	resetApps.POST("/:id/approvals", api.ApprovePasswordResetApplicationCtrl)
	resetApps.POST("/:id/rejections", api.RejectPasswordResetApplicationCtrl)

	users := rg.Group("", api.PrivilegesMiddleware(3))
	users.GET("", api.SearchUsersCtrl)
	users.GET("/:id", api.GetUserByIDCtrl)
	users.POST("", api.AddUserCtrl)
	users.DELETE("/:id", api.DeleteUserByIDCtrl)
	users.PUT("/:id", api.UpdateUserCtrl)
}

func setupUserInformationRoutes(rg *gin.RouterGroup) {
	authenticated := rg.Group("/profile", api.PrivilegesMiddleware(1))
	authenticated.GET("", api.GetMyProfileCtrl)
	authenticated.PUT("", api.UpdateMyProfileCtrl)
	authenticated.POST("/avatar", api.UploadAvatarCtrl)

	admin := rg.Group("/profiles", api.PrivilegesMiddleware(3))
	admin.GET("/:id", api.GetUserByIDCtrl)
	admin.PUT("/:id", api.UpdateUserCtrl)
}

func setupPublicRoutes(rg *gin.RouterGroup) {
	rg.GET("/status", api.GetPublicStatusSummaryCtrl)
}
