package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nagare/internal/model"
	"nagare/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

type UserRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Privileges   int    `json:"privileges"`
	Status       *int   `json:"status"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Avatar       string `json:"avatar"`
	Address      string `json:"address"`
	Introduction string `json:"introduction"`
	Nickname     string `json:"nickname"`
}

type UserResponse struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	Privileges   int    `json:"privileges"`
	Status       int    `json:"status"`
	Role         string `json:"role"`
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
	Email    string `json:"email"`
	Code     string `json:"code"`
}

type ResetPasswordRequest struct {
	Username    string `json:"username"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type CustomedClaims struct {
	UID        uint   `json:"uid"`
	Username   string `json:"username"`
	Privileges int    `json:"privileges"`
	jwt.RegisteredClaims
}

// ============= User Service Functions =============

func GetAllUsersServ() ([]UserResponse, error) {
	users, err := repository.GetAllUsersDAO()
	if err != nil {
		return nil, err
	}
	var userResponses []UserResponse
	for _, u := range users {
		userResponses = append(userResponses, userToResp(u))
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
		responses = append(responses, userToResp(u))
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
	return userToResp(user), nil
}

func GetUserByUsernameServ(username string) (UserResponse, error) {
	user, err := repository.GetUserByUsernameDAO(username)
	if err != nil {
		return UserResponse{}, err
	}
	return userToResp(user), nil
}

func AddUserServ(req UserRequest) error {
	status := 1
	if req.Status != nil {
		status = *req.Status
	}
	user := model.User{
		Username:     req.Username,
		Password:     req.Password,
		Privileges:   req.Privileges,
		Status:       status,
		Email:        req.Email,
		Phone:        req.Phone,
		Avatar:       req.Avatar,
		Address:      req.Address,
		Introduction: req.Introduction,
		Nickname:     req.Nickname,
	}
	return repository.AddUserDAO(user)
}

func DeleteUserByIDServ(id int) error {
	return repository.DeleteUserByIDDAO(id)
}

func UpdateUserServ(id int, req UserRequest) error {
	user, err := repository.GetUserByIDDAO(id)
	if err != nil {
		return err
	}

	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Password != "" {
		user.Password = req.Password
	}
	if req.Privileges != 0 {
		user.Privileges = req.Privileges
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Address != "" {
		user.Address = req.Address
	}
	if req.Introduction != "" {
		user.Introduction = req.Introduction
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}

	return repository.UpdateUserDAO(id, user)
}

func UpdateUserProfileServ(username string, req UserRequest) error {
	user, err := repository.GetUserByUsernameDAO(username)
	if err != nil {
		return err
	}

	user.Email = req.Email
	user.Phone = req.Phone
	user.Avatar = req.Avatar
	user.Address = req.Address
	user.Introduction = req.Introduction
	user.Nickname = req.Nickname

	if req.Username != "" {
		user.Username = req.Username
	}

	return repository.UpdateUserDAO(int(user.ID), user)
}

// UploadAvatarServ handles avatar file upload and returns the file URL
func UploadAvatarServ(username string, fileHeader *multipart.FileHeader) (string, error) {
	// Validate file
	if fileHeader == nil {
		return "", errors.New("file is nil")
	}
	if fileHeader.Size == 0 {
		return "", errors.New("file is empty")
	}

	// Check file size (max 5MB)
	if fileHeader.Size > 5<<20 {
		return "", errors.New("file size exceeds 5MB limit")
	}

	// Validate file type (allow only image files)
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("failed to read uploaded file: %w", err)
	}
	if n == 0 {
		return "", errors.New("file is empty")
	}
	contentType := http.DetectContentType(buffer[:n])

	if !allowedTypes[contentType] {
		return "", errors.New("invalid file type, only images are allowed")
	}

	ext := ".png"
	if contentType == "image/jpeg" || contentType == "image/jpg" {
		ext = ".jpg"
	} else if contentType == "image/gif" {
		ext = ".gif"
	} else if contentType == "image/webp" {
		ext = ".webp"
	}

	safeUsername := strings.NewReplacer("/", "_", "\\", "_", " ", "_").Replace(username)
	filename := fmt.Sprintf("avatar_%s_%d%s", safeUsername, time.Now().UnixNano(), ext)
	filePath := filepath.Join("public", "avatars", filename)

	// Create avatars directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create avatars directory: %w", err)
	}

	// Save file
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to reset file pointer: %w", err)
	}
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Return relative URL and update user profile
	avatarURL := fmt.Sprintf("/avatars/%s", filename)
	user, err := repository.GetUserByUsernameDAO(username)
	if err != nil {
		return "", err
	}
	user.Avatar = avatarURL
	if err := repository.UpdateUserDAO(int(user.ID), user); err != nil {
		return "", err
	}

	return avatarURL, nil
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
		UID:        user.ID,
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
	if req.Username == "" || req.Password == "" || req.Email == "" || req.Code == "" {
		return model.ErrInvalidInput
	}

	// Verify code
	_, err := repository.FindEmailVerificationDAO(req.Email, req.Code)
	if err != nil {
		return fmt.Errorf("invalid or expired verification code")
	}

	if _, err := repository.GetUserByUsernameDAO(req.Username); err == nil {
		return model.ErrConflict
	} else if !errors.Is(err, model.ErrNotFound) {
		return err
	}

	if err := repository.CreateRegisterApplicationDAO(model.RegisterApplication{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Status:   0,
	}); err != nil {
		return err
	}

	// Clean up code after successful registration
	// (Optional: depends on if we want to allow re-use within expiration window,
	// but here we just leave it for simplicity or explicit deletion can be added)

	_ = CreateSiteMessageServ("New Registration", fmt.Sprintf("A new user '%s' (%s) has applied for registration.", req.Username, req.Email), "system", 2, nil)
	return nil
}

func SendRegistrationCodeServ(email string) error {
	if email == "" {
		return model.ErrInvalidInput
	}

	code := generateVerificationCode(6)
	expiresAt := time.Now().Add(15 * time.Minute)

	ev := model.EmailVerification{
		Email:     email,
		Code:      code,
		ExpiresAt: expiresAt,
	}

	if err := repository.SaveEmailVerificationDAO(ev); err != nil {
		return err
	}

	subject := "Nagare Registration Verification Code"
	body := fmt.Sprintf("Your verification code is: %s\nThis code will expire in 15 minutes.", code)

	return SendEmailServ(email, subject, body)
}

func generateVerificationCode(length int) string {
	table := [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	b := make([]byte, length)
	n, err := io.ReadAtLeast(rand.Reader, b, length)
	if n != length || err != nil {
		// Fallback to simpler method if rand fails
		return "123456"
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}

func ResetPasswordServ(req ResetPasswordRequest) error {
	if req.Username == "" || req.OldPassword == "" || req.NewPassword == "" {
		return model.ErrInvalidInput
	}
	user, err := repository.GetUserByUsernameDAO(req.Username)
	if err != nil {
		return err
	}
	if user.Password != req.OldPassword {
		return model.ErrAuthenticationFailed
	}
	return repository.UpdateUserPasswordByUsernameDAO(req.Username, req.NewPassword)
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

func userToResp(u model.User) UserResponse {
	return UserResponse{
		ID:           int(u.ID),
		Username:     u.Username,
		Privileges:   u.Privileges,
		Status:       u.Status,
		Role:         privilegeToRole(u.Privileges),
		Email:        u.Email,
		Phone:        u.Phone,
		Avatar:       u.Avatar,
		Address:      u.Address,
		Introduction: u.Introduction,
		Nickname:     u.Nickname,
	}
}
