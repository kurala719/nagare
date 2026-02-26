package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
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
	QQ           string `json:"qq"`
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
	QQ           string `json:"qq"`
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
	if req.Username != "" && !isValidUsername(req.Username) {
		return model.ErrInvalidUsername
	}
	if req.Email != "" && !isValidEmail(req.Email) {
		return model.ErrInvalidEmail
	}
	if req.Password != "" && !isStrongPassword(req.Password) {
		return model.ErrWeakPassword
	}
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
		QQ:           req.QQ,
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
		if !isStrongPassword(req.Password) {
			return model.ErrWeakPassword
		}
		user.Password = req.Password
	}
	if req.Privileges != 0 {
		user.Privileges = req.Privileges
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	if req.Email != "" {
		if !isValidEmail(req.Email) {
			return model.ErrInvalidEmail
		}
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
	if req.QQ != "" {
		user.QQ = req.QQ
	}

	return repository.UpdateUserDAO(id, user)
}

func UpdateUserProfileServ(username string, req UserRequest) error {
	user, err := repository.GetUserByUsernameDAO(username)
	if err != nil {
		return err
	}
	oldAvatar := strings.TrimSpace(user.Avatar)
	if req.Email != "" && !isValidEmail(req.Email) {
		return model.ErrInvalidEmail
	}

	user.Email = req.Email
	user.Phone = req.Phone
	newAvatar := strings.TrimSpace(req.Avatar)
	if newAvatar != "" {
		user.Avatar = newAvatar
	}
	user.Address = req.Address
	user.Introduction = req.Introduction
	user.Nickname = req.Nickname
	user.QQ = req.QQ

	if req.Username != "" {
		user.Username = req.Username
	}

	if err := repository.UpdateUserDAO(int(user.ID), user); err != nil {
		return err
	}

	if newAvatar != "" && newAvatar != oldAvatar {
		_ = deleteLocalAvatar(oldAvatar)
	}

	return nil
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
	switch contentType {
	case "image/jpeg", "image/jpg":
		ext = ".jpg"
	case "image/gif":
		ext = ".gif"
	case "image/webp":
		ext = ".webp"
	}

	safeUsername := strings.NewReplacer("/", "_", "\\", "_", " ", "_").Replace(username)
	filename := fmt.Sprintf("avatar_%s_%d%s", safeUsername, time.Now().UnixNano(), ext)
	filePath := filepath.Join("public", "avatars", filename)

	// Create avatars directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create avatars directory: %w", err)
	}

	// Save file (resize if possible)
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to reset file pointer: %w", err)
	}
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	resized := false
	if contentType != "image/webp" {
		img, format, err := image.Decode(file)
		if err == nil {
			img = resizeToMax(img, 512)
			switch format {
			case "jpeg":
				if err := jpeg.Encode(out, img, &jpeg.Options{Quality: 85}); err == nil {
					resized = true
				}
			case "png":
				if err := png.Encode(out, img); err == nil {
					resized = true
				}
			case "gif":
				if err := gif.Encode(out, img, nil); err == nil {
					resized = true
				}
			}
		}
	}

	if !resized {
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return "", fmt.Errorf("failed to reset file pointer: %w", err)
		}
		if _, err := io.Copy(out, file); err != nil {
			return "", fmt.Errorf("failed to save file: %w", err)
		}
	}

	// Return relative URL
	avatarURL := fmt.Sprintf("/avatars/%s", filename)
	return avatarURL, nil
}

func deleteLocalAvatar(avatarURL string) error {
	if avatarURL == "" {
		return nil
	}
	trimmed := strings.TrimSpace(avatarURL)
	if trimmed == "" {
		return nil
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		parsed, err := url.Parse(trimmed)
		if err != nil {
			return nil
		}
		trimmed = parsed.Path
	}

	trimmed = strings.TrimPrefix(trimmed, "/")
	if !strings.HasPrefix(trimmed, "avatars/") {
		return nil
	}
	relPath := strings.TrimPrefix(trimmed, "avatars/")
	if relPath == "" {
		return nil
	}

	baseDir := filepath.Clean(filepath.Join("public", "avatars"))
	filePath := filepath.Clean(filepath.Join(baseDir, relPath))
	if !strings.HasPrefix(filePath, baseDir) {
		return nil
	}

	if err := os.Remove(filePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func resizeToMax(img image.Image, maxDim int) image.Image {
	if maxDim <= 0 {
		return img
	}
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= maxDim && height <= maxDim {
		return img
	}

	maxSide := float64(width)
	if height > width {
		maxSide = float64(height)
	}
	scale := float64(maxDim) / maxSide
	newWidth := int(math.Max(1, math.Round(float64(width)*scale)))
	newHeight := int(math.Max(1, math.Round(float64(height)*scale)))

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	for y := 0; y < newHeight; y++ {
		srcY := int(math.Min(float64(height-1), math.Floor(float64(y)/scale)))
		for x := 0; x < newWidth; x++ {
			srcX := int(math.Min(float64(width-1), math.Floor(float64(x)/scale)))
			dst.Set(x, y, img.At(bounds.Min.X+srcX, bounds.Min.Y+srcY))
		}
	}

	return dst
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
	if req.Username == "" || req.Password == "" {
		return model.ErrInvalidInput
	}
	if !isValidUsername(req.Username) {
		return model.ErrInvalidUsername
	}
	if !isStrongPassword(req.Password) {
		return model.ErrWeakPassword
	}

	// Verify code only if email is provided
	if req.Email != "" {
		if !isValidEmail(req.Email) {
			return model.ErrInvalidEmail
		}
		if req.Code == "" {
			return fmt.Errorf("verification code is required when email is provided")
		}
		_, err := repository.FindEmailVerificationDAO(req.Email, req.Code)
		if err != nil {
			return fmt.Errorf("invalid or expired verification code")
		}
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

	msg := fmt.Sprintf("A new user '%s' has applied for registration.", req.Username)
	if req.Email != "" {
		msg = fmt.Sprintf("A new user '%s' (%s) has applied for registration.", req.Username, req.Email)
	}
	LogSystem("error", fmt.Sprintf("New Registration: %s", msg), map[string]interface{}{"username": req.Username, "email": req.Email}, nil, "")
	return nil
}

func SendRegistrationCodeServ(email string) error {
	if email == "" {
		return model.ErrInvalidInput
	}
	if !isValidEmail(email) {
		return model.ErrInvalidEmail
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
	if !isStrongPassword(req.NewPassword) {
		return model.ErrWeakPassword
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

func isValidUsername(username string) bool {
	trimmed := strings.TrimSpace(username)
	if len(trimmed) < 3 || len(trimmed) > 32 {
		return false
	}
	pattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return pattern.MatchString(trimmed)
}

func isValidEmail(email string) bool {
	trimmed := strings.TrimSpace(email)
	if trimmed == "" {
		return false
	}
	// Simple, safe email syntax check.
	if len(trimmed) > 254 {
		return false
	}
	pattern := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	return pattern.MatchString(trimmed)
}

func isStrongPassword(password string) bool {
	trimmed := strings.TrimSpace(password)
	if len(trimmed) < 8 {
		return false
	}
	if strings.Contains(trimmed, " ") {
		return false
	}
	classes := 0
	if regexp.MustCompile(`[a-z]`).MatchString(trimmed) {
		classes++
	}
	if regexp.MustCompile(`[A-Z]`).MatchString(trimmed) {
		classes++
	}
	if regexp.MustCompile(`[0-9]`).MatchString(trimmed) {
		classes++
	}
	if regexp.MustCompile(`[^A-Za-z0-9]`).MatchString(trimmed) {
		classes++
	}
	return classes >= 3
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
		QQ:           u.QQ,
	}
}
