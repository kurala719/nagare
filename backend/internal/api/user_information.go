package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"nagare/internal/service"
	"nagare/internal/model"
)

// GetUserInformationCtrl retrieves user information for the authenticated user
func GetUserInformationCtrl(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		respondError(c, model.ErrUnauthorized)
		return
	}
	userInfo, err := service.GetUserInformationByUsernameServ(username.(string))
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			respondSuccess(c, http.StatusOK, service.UserInformationResponse{})
			return
		}
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, userInfo)
}

// CreateUserInformationCtrl creates user information for the authenticated user
func CreateUserInformationCtrl(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		respondError(c, model.ErrUnauthorized)
		return
	}
	var req service.UserInformationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	if err := service.CreateUserInformationByUsernameServ(username.(string), req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusCreated, "user information created")
}

// UpdateUserInformationCtrl updates user information for the authenticated user
func UpdateUserInformationCtrl(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		respondError(c, model.ErrUnauthorized)
		return
	}
	var req service.UserInformationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	if err := service.UpdateUserInformationByUsernameServ(username.(string), req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "user information updated")
}

// DeleteUserInformationCtrl deletes user information for the authenticated user
func DeleteUserInformationCtrl(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		respondError(c, model.ErrUnauthorized)
		return
	}
	if err := service.DeleteUserInformationByUsernameServ(username.(string)); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "user information deleted")
}

// GetUserInformationByUserIDCtrl retrieves user information by user ID (admin only)
func GetUserInformationByUserIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		respondBadRequest(c, "invalid user ID")
		return
	}
	userInfo, err := service.GetUserInformationByUserIDServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, userInfo)
}

// UpdateUserInformationByUserIDCtrl updates user information by user ID (superadmin only)
func UpdateUserInformationByUserIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		respondBadRequest(c, "invalid user ID")
		return
	}
	var req service.UserInformationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	if err := service.UpdateUserInformationByUserIDServ(uint(id), req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "user information updated")
}
