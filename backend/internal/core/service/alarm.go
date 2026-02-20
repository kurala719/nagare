package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"nagare/internal/adapter/external/monitors"
	"nagare/internal/adapter/repository"
	"nagare/internal/core/domain"
)

// AlarmReq represents an alarm request
type AlarmReq struct {
	Name        string `json:"name" binding:"required"`
	URL         string `json:"url" binding:"required"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	AuthToken   string `json:"auth_token"`
	EventToken  string `json:"event_token"`
	Description string `json:"description"`
	Type        int    `json:"type" binding:"required,oneof=1 2 3"`
	Enabled     int    `json:"enabled"`
}

// AlarmResp represents an alarm response
type AlarmResp struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	AuthToken   string `json:"auth_token"`
	EventToken  string `json:"event_token"`
	Description string `json:"description"`
	Type        int    `json:"type"`
	Enabled     int    `json:"enabled"`
	Status      int    `json:"status"`
	StatusDesc  string `json:"status_description"`
}

func generateAlarmEventToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GetAllAlarmsServ retrieves all alarms
func GetAllAlarmsServ() ([]AlarmResp, error) {
	alarms, err := repository.GetAllAlarmsDAO()
	if err != nil {
		return nil, fmt.Errorf("failed to get alarms: %w", err)
	}

	result := make([]AlarmResp, 0, len(alarms))
	for _, a := range alarms {
		result = append(result, alarmToResp(a))
	}
	return result, nil
}

// SearchAlarmsServ retrieves alarms by filter
func SearchAlarmsServ(filter domain.AlarmFilter) ([]AlarmResp, error) {
	alarms, err := repository.SearchAlarmsDAO(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search alarms: %w", err)
	}
	result := make([]AlarmResp, 0, len(alarms))
	for _, a := range alarms {
		result = append(result, alarmToResp(a))
	}
	return result, nil
}

// CountAlarmsServ returns total count for alarms by filter
func CountAlarmsServ(filter domain.AlarmFilter) (int64, error) {
	return repository.CountAlarmsDAO(filter)
}

// GetAlarmByIDServ retrieves an alarm by ID
func GetAlarmByIDServ(id uint) (AlarmResp, error) {
	alarm, err := repository.GetAlarmByIDDAO(id)
	if err != nil {
		return AlarmResp{}, fmt.Errorf("failed to get alarm: %w", err)
	}
	return alarmToResp(alarm), nil
}

// GetAlarmByEventTokenServ retrieves an alarm by event token
func GetAlarmByEventTokenServ(eventToken string) (domain.Alarm, error) {
	if strings.TrimSpace(eventToken) == "" {
		return domain.Alarm{}, domain.ErrUnauthorized
	}
	return repository.GetAlarmByEventTokenDAO(eventToken)
}

// AddAlarmServ creates a new alarm and optionally attempts to login
func AddAlarmServ(a AlarmReq) (AlarmResp, error) {
	eventToken := strings.TrimSpace(a.EventToken)
	if eventToken == "" {
		generated, err := generateAlarmEventToken()
		if err != nil {
			return AlarmResp{}, fmt.Errorf("failed to generate event token: %w", err)
		}
		eventToken = generated
	}

	alarm := domain.Alarm{
		Name:        a.Name,
		URL:         a.URL,
		Username:    a.Username,
		Password:    a.Password,
		AuthToken:   a.AuthToken,
		EventToken:  eventToken,
		Description: a.Description,
		Type:        a.Type,
		Enabled:     a.Enabled,
		Status:      determineAlarmStatus(domain.Alarm{Enabled: a.Enabled, AuthToken: a.AuthToken, Username: a.Username, Password: a.Password}),
	}

	if err := repository.AddAlarmDAO(alarm); err != nil {
		return AlarmResp{}, fmt.Errorf("failed to create alarm: %w", err)
	}

	alarms, err := repository.SearchAlarmsDAO(domain.AlarmFilter{Query: a.Name})
	if err != nil || len(alarms) == 0 {
		return AlarmResp{}, fmt.Errorf("failed to retrieve created alarm")
	}
	createdAlarm := alarms[len(alarms)-1]
	result := alarmToResp(createdAlarm)

	if a.Username != "" && a.Password != "" {
		loggedInAlarm, err := LoginAlarmServ(createdAlarm.ID)
		if err != nil {
			LogService("warn", "auto-login failed for alarm", map[string]interface{}{"alarm_id": createdAlarm.ID, "error": err.Error()}, nil, "")
		} else {
			result = loggedInAlarm
		}
	}

	return result, nil
}

// DeleteAlarmServByID deletes an alarm by ID
func DeleteAlarmServByID(id int) error {
	return repository.DeleteAlarmByIDDAO(id)
}

// UpdateAlarmServ updates an existing alarm
func UpdateAlarmServ(id int, a AlarmReq) error {
	existing, err := GetAlarmByIDServ(uint(id))
	if err != nil {
		return err
	}
	eventToken := strings.TrimSpace(a.EventToken)
	if eventToken == "" {
		eventToken = existing.EventToken
	}
	updated := domain.Alarm{
		Name:              a.Name,
		URL:               a.URL,
		Username:          a.Username,
		Password:          a.Password,
		AuthToken:         a.AuthToken,
		EventToken:        eventToken,
		Description:       a.Description,
		Type:              a.Type,
		Enabled:           a.Enabled,
		Status:            existing.Status,
		StatusDescription: existing.StatusDesc,
	}
	// Preserve status and description unless enabled state changed
	if a.Enabled != existing.Enabled {
		updated.Status = determineAlarmStatus(domain.Alarm{Enabled: a.Enabled, AuthToken: a.AuthToken, Username: a.Username, Password: a.Password})
		updated.StatusDescription = ""
	}
	if err := repository.UpdateAlarmDAO(id, updated); err != nil {
		return err
	}
	return nil
}

// RegenerateAlarmEventTokenServ creates a new event token for an alarm
func RegenerateAlarmEventTokenServ(id uint) (AlarmResp, error) {
	newToken, err := generateAlarmEventToken()
	if err != nil {
		return AlarmResp{}, fmt.Errorf("failed to generate event token: %w", err)
	}
	if err := repository.UpdateAlarmEventTokenDAO(id, newToken); err != nil {
		return AlarmResp{}, fmt.Errorf("failed to save event token: %w", err)
	}
	updatedAlarm, err := GetAlarmByIDServ(id)
	if err != nil {
		return AlarmResp{}, fmt.Errorf("failed to retrieve updated alarm: %w", err)
	}
	return updatedAlarm, nil
}

// ValidateAlarmEventTokenServ validates inbound event tokens for webhook use
func ValidateAlarmEventTokenServ(eventToken string) error {
	if strings.TrimSpace(eventToken) == "" {
		return domain.ErrUnauthorized
	}
	_, err := repository.GetAlarmByEventTokenDAO(eventToken)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrUnauthorized
		}
		return err
	}
	return nil
}

// LoginAlarmServ authenticates with an alarm source and stores the auth token
func LoginAlarmServ(id uint) (AlarmResp, error) {
	alarm, err := GetAlarmByIDServ(id)
	if err != nil {
		return AlarmResp{}, err
	}

	client, err := monitors.NewClient(monitors.Config{
		Name: alarm.Name,
		Type: monitors.ParseMonitorType(alarm.Type),
		Auth: monitors.AuthConfig{
			URL:      alarm.URL,
			Username: alarm.Username,
			Password: alarm.Password,
			Token:    alarm.AuthToken,
		},
		Timeout: 30,
	})
	if err != nil {
		return AlarmResp{}, fmt.Errorf("failed to create alarm client: %w", err)
	}

	ctx := context.Background()
	if err := client.Authenticate(ctx); err != nil {
		return AlarmResp{}, fmt.Errorf("authentication failed: %w", err)
	}

	authToken := client.GetAuthToken()
	if authToken == "" {
		return AlarmResp{}, fmt.Errorf("authentication succeeded but no token received")
	}

	if err := repository.UpdateAlarmAuthTokenDAO(id, authToken); err != nil {
		return AlarmResp{}, fmt.Errorf("failed to save auth token: %w", err)
	}
	_ = repository.UpdateAlarmStatusDAO(id, 1)

	updatedAlarm, err := GetAlarmByIDServ(id)
	if err != nil {
		return AlarmResp{}, fmt.Errorf("failed to retrieve updated alarm: %w", err)
	}

	return updatedAlarm, nil
}

// alarmToResp converts a domain Alarm to AlarmResp
func alarmToResp(a domain.Alarm) AlarmResp {
	return AlarmResp{
		ID:          int(a.ID),
		Name:        a.Name,
		URL:         a.URL,
		Username:    a.Username,
		Password:    a.Password,
		AuthToken:   a.AuthToken,
		EventToken:  a.EventToken,
		Description: a.Description,
		Type:        a.Type,
		Enabled:     a.Enabled,
		Status:      a.Status,
		StatusDesc:  a.StatusDescription,
	}
}
