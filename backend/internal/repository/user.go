package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"nagare/internal/database"
	"nagare/internal/model"
)

// GetAllUsersDAO retrieves all users from the database
func GetAllUsersDAO() ([]model.User, error) {
	var users []model.User
	if err := database.DB.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// SearchUsersDAO retrieves users by filter
func SearchUsersDAO(filter model.UserFilter) ([]model.User, error) {
	query := database.DB.Model(&model.User{})
	if filter.Query != "" {
		q := "%" + filter.Query + "%"
		query = query.Where("username LIKE ? OR email LIKE ? OR nickname LIKE ? OR phone LIKE ?", q, q, q, q)
	}
	// Security filter based on requester privileges
	if filter.RequesterPrivileges > 0 && filter.RequesterPrivileges < 3 {
		query = query.Where("privileges < ?", filter.RequesterPrivileges)
	} else if filter.RequesterPrivileges == 0 {
		return []model.User{}, nil
	}

	if filter.Privileges != nil {
		query = query.Where("privileges = ?", *filter.Privileges)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	query = applySort(query, filter.SortBy, filter.SortOrder, map[string]string{
		"name":       "username",
		"username":   "username",
		"status":     "status",
		"privileges": "privileges",
		"id":         "id",
		"nickname":   "nickname",
		"email":      "email",
		"phone":      "phone",
		"created_at": "created_at",
		"updated_at": "updated_at",
	}, "id desc")
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}
	var users []model.User
	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// CountUsersDAO returns total count for users by filter
func CountUsersDAO(filter model.UserFilter) (int64, error) {
	query := database.DB.Model(&model.User{})
	if filter.Query != "" {
		q := "%" + filter.Query + "%"
		query = query.Where("username LIKE ? OR email LIKE ? OR nickname LIKE ? OR phone LIKE ?", q, q, q, q)
	}
	// Security filter based on requester privileges
	if filter.RequesterPrivileges > 0 && filter.RequesterPrivileges < 3 {
		query = query.Where("privileges < ?", filter.RequesterPrivileges)
	} else if filter.RequesterPrivileges == 0 {
		return 0, nil
	}

	if filter.Privileges != nil {
		query = query.Where("privileges = ?", *filter.Privileges)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// GetUserByIDDAO retrieves a user by ID
func GetUserByIDDAO(id int) (model.User, error) {
	var user model.User
	err := database.DB.First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return user, model.ErrNotFound
	}
	return user, err
}

// GetUserByUsernameDAO retrieves a user by username
func GetUserByUsernameDAO(username string) (model.User, error) {
	var user model.User
	err := database.DB.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return user, model.ErrNotFound
	}
	return user, err
}

// AddUserDAO creates a new user
func AddUserDAO(user model.User) error {
	return database.DB.Create(&user).Error
}

// DeleteUserByIDDAO deletes a user by ID
func DeleteUserByIDDAO(id int) error {
	return database.DB.Delete(&model.User{}, id).Error
}

// UpdateUserDAO updates a user by ID (all fields)
func UpdateUserDAO(id int, user model.User) error {
	updates := map[string]interface{}{
		"username":     user.Username,
		"privileges":   user.Privileges,
		"status":       user.Status,
		"email":        user.Email,
		"phone":        user.Phone,
		"avatar":       user.Avatar,
		"address":      user.Address,
		"introduction": user.Introduction,
		"nickname":     user.Nickname,
	}
	if user.Password != "" {
		updates["password"] = user.Password
	}
	return database.DB.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateUserPasswordByUsernameDAO updates a user's password by username
func UpdateUserPasswordByUsernameDAO(username, newPassword string) error {
	return database.DB.Model(&model.User{}).Where("username = ?", username).Update("password", newPassword).Error
}

// GetUserIDByUsernameDAO retrieves user ID by username
func GetUserIDByUsernameDAO(username string) (uint, error) {
	var user model.User
	err := database.DB.Select("id").Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, model.ErrNotFound
	}
	return user.ID, err
}

// SaveEmailVerificationDAO saves or updates an email verification code
func SaveEmailVerificationDAO(ev model.EmailVerification) error {
	// Remove existing codes for this email first
	database.DB.Where("email = ?", ev.Email).Delete(&model.EmailVerification{})
	return database.DB.Create(&ev).Error
}

// FindEmailVerificationDAO retrieves a verification code by email and code
func FindEmailVerificationDAO(email, code string) (model.EmailVerification, error) {
	var ev model.EmailVerification
	err := database.DB.Where("email = ? AND code = ? AND expires_at > ?", email, code, time.Now()).First(&ev).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ev, model.ErrNotFound
	}
	return ev, err
}
