package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"nagare/internal/service"
	"nagare/internal/model"
)

func LoginUserCtrl(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	token, err := service.LoginUserServ(req.Username, req.Password)
	if err != nil {
		service.LogService("warn", "login failed", map[string]interface{}{"username": req.Username}, nil, c.ClientIP())
		respondError(c, err)
		return
	}
	service.LogService("info", "login success", map[string]interface{}{"username": req.Username}, nil, c.ClientIP())
	respondSuccess(c, http.StatusOK, gin.H{"token": token})
}

func RegisterUserCtrl(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	if err := service.RegisterUserServ(req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusCreated, "registration application submitted")
}

func SendRegistrationCodeCtrl(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	if err := service.SendRegistrationCodeServ(req.Email); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "verification code sent")
}

func ResetPasswordCtrl(c *gin.Context) {
	requesterPrivileges := getRequesterPrivileges(c)
	if requesterPrivileges >= 3 {
		respondError(c, model.ErrForbidden)
		return
	}
	username, ok := c.Get("username")
	if !ok {
		respondError(c, model.ErrUnauthorized)
		return
	}
	var req service.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	req.Username = username.(string)
	if err := service.ResetPasswordServ(req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "password reset")
}

func GetAllUsersCtrl(c *gin.Context) {
	requesterPrivileges := getRequesterPrivileges(c)
	users, err := service.GetAllUsersServ()
	if err != nil {
		respondError(c, err)
		return
	}
	filtered := filterUsersByPrivileges(users, requesterPrivileges)
	respondSuccess(c, http.StatusOK, filtered)
}

// SearchUsersCtrl handles GET /users
func SearchUsersCtrl(c *gin.Context) {
	requesterPrivileges := getRequesterPrivileges(c)
	privileges, _ := parseOptionalInt(c, "privileges")
	status, _ := parseOptionalInt(c, "status")
	withTotal, _ := parseOptionalBool(c, "with_total")

	limit := 100
	if l, err := parseOptionalInt(c, "limit"); err == nil && l != nil {
		limit = *l
	}
	offset := 0
	if o, err := parseOptionalInt(c, "offset"); err == nil && o != nil {
		offset = *o
	}

	filter := model.UserFilter{
		Query:               c.Query("q"),
		Privileges:          privileges,
		Status:              status,
		Limit:               limit,
		Offset:              offset,
		SortBy:              c.Query("sort"),
		SortOrder:           c.Query("order"),
		RequesterPrivileges: requesterPrivileges,
	}
	users, err := service.SearchUsersServ(filter)
	if err != nil {
		respondError(c, err)
		return
	}
	
	if withTotal != nil && *withTotal {
		total, err := service.CountUsersServ(filter)
		if err != nil {
			respondError(c, err)
			return
		}
		respondSuccess(c, http.StatusOK, gin.H{"items": users, "total": total})
		return
	}
	respondSuccess(c, http.StatusOK, users)
}

func GetUserByIDCtrl(c *gin.Context) {
	requesterPrivileges := getRequesterPrivileges(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid user ID")
		return
	}
	user, err := service.GetUserByIDServ(id)
	if err != nil {
		respondError(c, err)
		return
	}
	if !canManageUser(requesterPrivileges, user.Privileges) {
		respondError(c, model.ErrForbidden)
		return
	}
	respondSuccess(c, http.StatusOK, user)
}

func AddUserCtrl(c *gin.Context) {
	requesterPrivileges := getRequesterPrivileges(c)
	if requesterPrivileges < 3 {
		respondError(c, model.ErrForbidden)
		return
	}
	var req service.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	if req.Privileges >= requesterPrivileges {
		respondError(c, model.ErrForbidden)
		return
	}
	if err := service.AddUserServ(req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusCreated, "user created")
}

func UpdateUserCtrl(c *gin.Context) {
	requesterPrivileges := getRequesterPrivileges(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid user ID")
		return
	}
	user, err := service.GetUserByIDServ(id)
	if err != nil {
		respondError(c, err)
		return
	}
	if !canManageUser(requesterPrivileges, user.Privileges) {
		respondError(c, model.ErrForbidden)
		return
	}
	var req service.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	if req.Privileges >= requesterPrivileges {
		respondError(c, model.ErrForbidden)
		return
	}
	if err := service.UpdateUserServ(id, req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "user updated")
}

func DeleteUserByIDCtrl(c *gin.Context) {
	requesterPrivileges := getRequesterPrivileges(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid user ID")
		return
	}
	user, err := service.GetUserByIDServ(id)
	if err != nil {
		respondError(c, err)
		return
	}
	if !canManageUser(requesterPrivileges, user.Privileges) {
		respondError(c, model.ErrForbidden)
		return
	}
	if err := service.DeleteUserByIDServ(id); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "user deleted")
}

// ============= Unified Profile Controllers =============

// GetMyProfileCtrl handles GET /user-info/me (or /profile)
func GetMyProfileCtrl(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		respondError(c, model.ErrUnauthorized)
		return
	}
	user, err := service.GetUserByUsernameServ(username.(string))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, user)
}

// UpdateMyProfileCtrl handles PUT /user-info/me (or /profile)
func UpdateMyProfileCtrl(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		respondError(c, model.ErrUnauthorized)
		return
	}
	var req service.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	if err := service.UpdateUserProfileServ(username.(string), req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "profile updated")
}

func getRequesterPrivileges(c *gin.Context) int {
	if val, ok := c.Get("privileges"); ok {
		if priv, ok := val.(int); ok {
			return priv
		}
	}
	return 0
}

func canManageUser(requesterPrivileges, targetPrivileges int) bool {
	if requesterPrivileges >= 3 {
		return true
	}
	return targetPrivileges < requesterPrivileges
}

func filterUsersByPrivileges(users []service.UserResponse, requesterPrivileges int) []service.UserResponse {
	if requesterPrivileges >= 3 {
		return users
	}
	filtered := make([]service.UserResponse, 0, len(users))
	for _, user := range users {
		if user.Privileges < requesterPrivileges {
			filtered = append(filtered, user)
		}
	}
	return filtered
}
