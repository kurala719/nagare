package service

import (
	"errors"
	"fmt"

	"nagare/internal/model"
	"nagare/internal/repository"
)

type PasswordResetApplicationResponse struct {
	ID         uint   `json:"id"`
	UserID     uint   `json:"user_id"`
	Username   string `json:"username"`
	Status     int    `json:"status"`
	Reason     string `json:"reason"`
	ApprovedBy *uint  `json:"approved_by"`
}

// SubmitPasswordResetApplicationServ submits a new reset request
func SubmitPasswordResetApplicationServ(username, password string) error {
	user, err := repository.GetUserByUsernameDAO(username)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return fmt.Errorf("user not found")
		}
		return err
	}

	app := model.PasswordResetApplication{
		UserID:      user.ID,
		Username:    username,
		NewPassword: password,
		Status:      0, // pending
	}

	return repository.CreatePasswordResetApplicationDAO(app)
}

// ListPasswordResetApplicationsServ retrieves reset applications by filter
func ListPasswordResetApplicationsServ(filter model.RegisterApplicationFilter) ([]PasswordResetApplicationResponse, error) {
	apps, err := repository.SearchPasswordResetApplicationsDAO(filter)
	if err != nil {
		return nil, err
	}
	responses := make([]PasswordResetApplicationResponse, 0, len(apps))
	for _, a := range apps {
		responses = append(responses, PasswordResetApplicationResponse{
			ID:         a.ID,
			UserID:     a.UserID,
			Username:   a.Username,
			Status:     a.Status,
			Reason:     a.Reason,
			ApprovedBy: a.ApprovedBy,
		})
	}
	return responses, nil
}

// CountPasswordResetApplicationsServ returns total count for reset applications by filter
func CountPasswordResetApplicationsServ(filter model.RegisterApplicationFilter) (int64, error) {
	return repository.CountPasswordResetApplicationsDAO(filter)
}

// ApprovePasswordResetApplicationServ approves a reset request and updates user password
func ApprovePasswordResetApplicationServ(id uint, approverUsername string) error {
	app, err := repository.GetPasswordResetApplicationByIDDAO(id)
	if err != nil {
		return err
	}
	if app.Status != 0 {
		return model.ErrInvalidInput
	}

	if err := repository.UpdateUserPasswordByUsernameDAO(app.Username, app.NewPassword); err != nil {
		return err
	}

	approverID, err := repository.GetUserIDByUsernameDAO(approverUsername)
	if err != nil {
		return err
	}
	
	_ = CreateSiteMessageServ("Password Reset", "Your password reset request has been approved and applied.", "system", 1, &app.UserID)

	// Send notification email
	user, _ := repository.GetUserByIDDAO(int(app.UserID))
	if user.Email != "" {
		_ = SendEmailServ(user.Email, "Password Reset Approved - Nagare", 
			fmt.Sprintf("Hello %s,\n\nYour password reset request has been approved and applied successfully.", app.Username))
	}

	return repository.UpdatePasswordResetApplicationStatusDAO(app.ID, 1, &approverID, "approved")
}

// RejectPasswordResetApplicationServ rejects a reset request
func RejectPasswordResetApplicationServ(id uint, approverUsername, reason string) error {
	app, err := repository.GetPasswordResetApplicationByIDDAO(id)
	if err != nil {
		return err
	}
	if app.Status != 0 {
		return model.ErrInvalidInput
	}
	approverID, err := repository.GetUserIDByUsernameDAO(approverUsername)
	if err != nil {
		return err
	}

	// Send notification email
	user, _ := repository.GetUserByIDDAO(int(app.UserID))
	if user.Email != "" {
		msg := fmt.Sprintf("Hello %s,\n\nYour password reset request has been rejected.\nReason: %s", app.Username, reason)
		_ = SendEmailServ(user.Email, "Password Reset Rejected - Nagare", msg)
	}

	return repository.UpdatePasswordResetApplicationStatusDAO(app.ID, 2, &approverID, reason)
}
