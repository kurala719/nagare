package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"nagare/internal/model"
	"nagare/internal/repository"
)

type UserRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Privileges int    `json:"privileges"`
}

type UserResponse struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	Privileges int    `json:"privileges"`
	Status     int    `json:"status"`
	Role       string `json:"role"`
}

type UserInformationRequest struct {
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Avatar       string `json:"avatar"`
	Address      string `json:"address"`
	Introduction string `json:"introduction"`
	Nickname     string `json:"nickname"`
}

type UserInformationResponse struct {
	ID           int    `json:"id"`
	UserID       int    `json:"user_id"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Avatar       string `json:"avatar"`
	Address      string `json:"address"`
	Introduction string `json:"introduction"`
	Nickname     string `json:"nickname"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ResetPasswordRequest struct {
	Username    string `json:"username"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// ============= User Service Functions =============

func GetAllUsersServ() ([]UserResponse, error) {
	users, err := repository.GetAllUsersDAO()
	if err != nil {
		return nil, err
	}
	var userResponses []UserResponse
	for _, u := range users {
		userResponses = append(userResponses, UserResponse{
			ID:         int(u.Model.ID),
			Username:   u.Username,
			Privileges: u.Privileges,
			Status:     u.Status,
			Role:       privilegeToRole(u.Privileges),
		})
	}
	return userResponses, nil
}

// SearchUsersServ retrieves users by filter
func SearchUsersServ(filter model.UserFilter) ([]UserResponse, error) {
	users, err := repository.SearchUsersDAO(filter)
	if err != nil {
		return nil, err
	}
	responses := make([]UserResponse, 0, len(users))
	for _, u := range users {
		responses = append(responses, UserResponse{
			ID:         int(u.Model.ID),
			Username:   u.Username,
			Privileges: u.Privileges,
			Status:     u.Status,
			Role:       privilegeToRole(u.Privileges),
		})
	}
	return responses, nil
}

// CountUsersServ returns total count for users by filter
func CountUsersServ(filter model.UserFilter) (int64, error) {
	return repository.CountUsersDAO(filter)
}

func GetUserByIDServ(id int) (UserResponse, error) {
	user, err := repository.GetUserByIDDAO(id)
	if err != nil {
		return UserResponse{}, err
	}
	return UserResponse{
		ID:         int(user.Model.ID),
		Username:   user.Username,
		Privileges: user.Privileges,
		Status:     user.Status,
		Role:       privilegeToRole(user.Privileges),
	}, nil
}

func GetUserByUsernameDAO(username string) (UserResponse, error) {
	user, err := repository.GetUserByUsernameDAO(username)
	if err != nil {
		return UserResponse{}, err
	}
	return UserResponse{
		ID:         int(user.Model.ID),
		Username:   user.Username,
		Privileges: user.Privileges,
		Status:     user.Status,
		Role:       privilegeToRole(user.Privileges),
	}, nil
}

func AddUserServ(req UserRequest) error {
	user := model.User{
		Username:   req.Username,
		Password:   req.Password,
		Privileges: req.Privileges,
		Status:     1,
	}
	return repository.AddUserDAO(user)
}

func DeleteUserByIDServ(id int) error {
	return repository.DeleteUserByIDDAO(id)
}

func UpdateUserServ(id int, req UserRequest) error {
	existing, err := repository.GetUserByIDDAO(id)
	if err != nil {
		return err
	}
	user := model.User{
		Username:   req.Username,
		Password:   req.Password,
		Privileges: req.Privileges,
		Status:     existing.Status,
	}
	return repository.UpdateUserDAO(id, user)
}

type CustomedClaims struct {
	Username   string `json:"username"`
	Privileges int    `json:"privileges"`
	jwt.RegisteredClaims
}

func LoginUserServ(username, password string) (string, error) {
	user, err := repository.GetUserByUsernameDAO(username)
	if err != nil {
		return "", err
	}
	if user.Password != password {
		return "", model.ErrAuthenticationFailed
	}
	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomedClaims{
		Username:   user.Username,
		Privileges: user.Privileges,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})

	secretKey := []byte(viper.GetString("jwt.secret_key"))
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func RegisterUserServ(req RegisterRequest) error {
	if req.Username == "" || req.Password == "" {
		return model.ErrInvalidInput
	}
	if _, err := repository.GetUserByUsernameDAO(req.Username); err == nil {
		return model.ErrConflict
	} else if !errors.Is(err, model.ErrNotFound) {
		return err
	}
	if existing, err := repository.GetRegisterApplicationByUsernameDAO(req.Username); err == nil {
		if existing.Status == 0 {
			return model.ErrConflict
		}
	} else if !errors.Is(err, model.ErrNotFound) {
		return err
	}
	return repository.CreateRegisterApplicationDAO(model.RegisterApplication{
		Username: req.Username,
		Password: req.Password,
		Status:   0,
	})
}

func ResetPasswordServ(req ResetPasswordRequest) error {
	if req.Username == "" || req.OldPassword == "" || req.NewPassword == "" {
		return model.ErrInvalidInput
	}
	user, err := repository.GetUserByUsernameDAO(req.Username)
	if err != nil {
		return err
	}
	if user.Privileges >= 3 {
		return model.ErrForbidden
	}
	if user.Password != req.OldPassword {
		return model.ErrAuthenticationFailed
	}
	return repository.UpdateUserPasswordByUsernameDAO(req.Username, req.NewPassword)
}

// ============= UserInformation Service Functions =============

func GetUserInformationByUsernameServ(username string) (UserInformationResponse, error) {
	userID, err := repository.GetUserIDByUsernameDAO(username)
	if err != nil {
		return UserInformationResponse{}, err
	}
	userInfo, err := repository.GetUserInformationByUserIDDAO(userID)
	if err != nil {
		return UserInformationResponse{}, err
	}
	return UserInformationResponse{
		ID:           int(userInfo.Model.ID),
		UserID:       int(userInfo.UserID),
		Email:        userInfo.Email,
		Phone:        userInfo.Phone,
		Avatar:       userInfo.Avatar,
		Address:      userInfo.Address,
		Introduction: userInfo.Introduction,
		Nickname:     userInfo.Nickname,
	}, nil
}

func GetUserInformationByUserIDServ(userID uint) (UserInformationResponse, error) {
	userInfo, err := repository.GetUserInformationByUserIDDAO(userID)
	if err != nil {
		return UserInformationResponse{}, err
	}
	return UserInformationResponse{
		ID:           int(userInfo.Model.ID),
		UserID:       int(userInfo.UserID),
		Email:        userInfo.Email,
		Phone:        userInfo.Phone,
		Avatar:       userInfo.Avatar,
		Address:      userInfo.Address,
		Introduction: userInfo.Introduction,
		Nickname:     userInfo.Nickname,
	}, nil
}

func CreateUserInformationByUsernameServ(username string, req UserInformationRequest) error {
	userID, err := repository.GetUserIDByUsernameDAO(username)
	if err != nil {
		return err
	}
	userInfo := model.UserInformation{
		UserID:       userID,
		Email:        req.Email,
		Phone:        req.Phone,
		Avatar:       req.Avatar,
		Address:      req.Address,
		Introduction: req.Introduction,
		Nickname:     req.Nickname,
	}
	return repository.CreateUserInformationDAO(userInfo)
}

func UpdateUserInformationByUsernameServ(username string, req UserInformationRequest) error {
	userID, err := repository.GetUserIDByUsernameDAO(username)
	if err != nil {
		return err
	}
	userInfo := model.UserInformation{
		UserID:       userID,
		Email:        req.Email,
		Phone:        req.Phone,
		Avatar:       req.Avatar,
		Address:      req.Address,
		Introduction: req.Introduction,
		Nickname:     req.Nickname,
	}
	// Try to update, if not found, create
	err = repository.UpdateUserInformationDAO(userID, userInfo)
	if errors.Is(err, model.ErrNotFound) {
		return repository.CreateUserInformationDAO(userInfo)
	}
	return err
}

// UpdateUserInformationByUserIDServ allows admin to update a user's profile by user ID
func UpdateUserInformationByUserIDServ(userID uint, req UserInformationRequest) error {
	userInfo := model.UserInformation{
		UserID:       userID,
		Email:        req.Email,
		Phone:        req.Phone,
		Avatar:       req.Avatar,
		Address:      req.Address,
		Introduction: req.Introduction,
		Nickname:     req.Nickname,
	}
	if err := repository.UpdateUserInformationDAO(userID, userInfo); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return repository.CreateUserInformationDAO(userInfo)
		}
		return err
	}
	return nil
}

func DeleteUserInformationByUsernameServ(username string) error {
	userID, err := repository.GetUserIDByUsernameDAO(username)
	if err != nil {
		return err
	}
	return repository.DeleteUserInformationByUserIDDAO(userID)
}

func privilegeToRole(privileges int) string {
	switch privileges {
	case 3:
		return "superadmin"
	case 2:
		return "admin"
	case 1:
		return "user"
	default:
		return "unauthorized"
	}
}
