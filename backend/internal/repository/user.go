package repository

import (
	"errors"

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
		query = query.Where("username LIKE ?", "%"+filter.Query+"%")
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
		"created_at": "created_at",
		"updated_at": "updated_at",
		"id":         "id",
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
		query = query.Where("username LIKE ?", "%"+filter.Query+"%")
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

// UpdateUserDAO updates a user by ID (only auth fields)
func UpdateUserDAO(id int, user model.User) error {
	updates := map[string]interface{}{
		"username":   user.Username,
		"privileges": user.Privileges,
		"status":     user.Status,
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

// ============= UserInformation DAO Functions =============

// GetUserInformationByUserIDDAO retrieves user information by user ID
func GetUserInformationByUserIDDAO(userID uint) (model.UserInformation, error) {
	var userInfo model.UserInformation
	err := database.DB.Where("user_id = ?", userID).First(&userInfo).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return userInfo, model.ErrNotFound
	}
	return userInfo, err
}

// CreateUserInformationDAO creates new user information
func CreateUserInformationDAO(userInfo model.UserInformation) error {
	return database.DB.Create(&userInfo).Error
}

// UpdateUserInformationDAO updates user information by user ID
func UpdateUserInformationDAO(userID uint, userInfo model.UserInformation) error {
	updates := map[string]interface{}{
		"email":        userInfo.Email,
		"phone":        userInfo.Phone,
		"avatar":       userInfo.Avatar,
		"address":      userInfo.Address,
		"introduction": userInfo.Introduction,
		"nickname":     userInfo.Nickname,
	}
	result := database.DB.Model(&model.UserInformation{}).Where("user_id = ?", userID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return nil
}

// DeleteUserInformationByUserIDDAO deletes user information by user ID
func DeleteUserInformationByUserIDDAO(userID uint) error {
	return database.DB.Where("user_id = ?", userID).Delete(&model.UserInformation{}).Error
}
