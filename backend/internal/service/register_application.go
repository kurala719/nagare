package service

import (
	"errors"

	"nagare/internal/model"
	"nagare/internal/repository"
)

type RegisterApplicationResponse struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	Status     int    `json:"status"`
	Reason     string `json:"reason"`
	ApprovedBy *uint  `json:"approved_by"`
}

// ListRegisterApplicationsServ retrieves registration applications by filter
func ListRegisterApplicationsServ(filter model.RegisterApplicationFilter) ([]RegisterApplicationResponse, error) {
	apps, err := repository.SearchRegisterApplicationsDAO(filter)
	if err != nil {
		return nil, err
	}
	responses := make([]RegisterApplicationResponse, 0, len(apps))
	for _, a := range apps {
		responses = append(responses, RegisterApplicationResponse{
			ID:         a.ID,
			Username:   a.Username,
			Status:     a.Status,
			Reason:     a.Reason,
			ApprovedBy: a.ApprovedBy,
		})
	}
	return responses, nil
}

// CountRegisterApplicationsServ returns total count for register applications by filter
func CountRegisterApplicationsServ(filter model.RegisterApplicationFilter) (int64, error) {
	return repository.CountRegisterApplicationsDAO(filter)
}

// ApproveRegisterApplicationServ approves a registration application and creates a user
func ApproveRegisterApplicationServ(id uint, approverUsername string) error {
	app, err := repository.GetRegisterApplicationByIDDAO(id)
	if err != nil {
		return err
	}
	if app.Status != 0 {
		return model.ErrInvalidInput
	}
	if _, err := repository.GetUserByUsernameDAO(app.Username); err == nil {
		return model.ErrConflict
	} else if !errors.Is(err, model.ErrNotFound) {
		return err
	}

	user := model.User{
		Username:   app.Username,
		Password:   app.Password,
		Privileges: 1,
		Status:     1,
	}
	if err := repository.AddUserDAO(user); err != nil {
		return err
	}
	approverID, err := repository.GetUserIDByUsernameDAO(approverUsername)
	if err != nil {
		return err
	}
	return repository.UpdateRegisterApplicationStatusDAO(app.ID, 1, &approverID, "approved")
}

// RejectRegisterApplicationServ rejects a registration application
func RejectRegisterApplicationServ(id uint, approverUsername, reason string) error {
	app, err := repository.GetRegisterApplicationByIDDAO(id)
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
	return repository.UpdateRegisterApplicationStatusDAO(app.ID, 2, &approverID, reason)
}
