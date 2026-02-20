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

// MonitorReq represents a monitor request
type MonitorReq struct {
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

// MonitorResp represents a monitor response
type MonitorResp struct {
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
	HealthScore int    `json:"health_score"`
}

func generateMonitorEventToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GetAllMonitorsServ retrieves all monitors
func GetAllMonitorsServ() ([]MonitorResp, error) {
	monitors, err := repository.GetAllMonitorsDAO()
	if err != nil {
		return nil, fmt.Errorf("failed to get monitors: %w", err)
	}

	result := make([]MonitorResp, 0, len(monitors))
	for _, m := range monitors {
		result = append(result, monitorToResp(m))
	}
	return result, nil
}

// SearchMonitorsServ retrieves monitors by filter
func SearchMonitorsServ(filter domain.MonitorFilter) ([]MonitorResp, error) {
	monitors, err := repository.SearchMonitorsDAO(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search monitors: %w", err)
	}
	result := make([]MonitorResp, 0, len(monitors))
	for _, m := range monitors {
		result = append(result, monitorToResp(m))
	}
	return result, nil
}

// CountMonitorsServ returns total count for monitors by filter
func CountMonitorsServ(filter domain.MonitorFilter) (int64, error) {
	return repository.CountMonitorsDAO(filter)
}

// GetMonitorByIDServ retrieves a monitor by ID
func GetMonitorByIDServ(id uint) (MonitorResp, error) {
	monitor, err := repository.GetMonitorByIDDAO(id)
	if err != nil {
		return MonitorResp{}, fmt.Errorf("failed to get monitor: %w", err)
	}
	return monitorToResp(monitor), nil
}

// TestMonitorStatusServ performs a live status check for a monitor.
func TestMonitorStatusServ(id uint) (StatusCheckResult, error) {
	result, err := CheckMonitorStatusServ(id)
	if err != nil {
		return StatusCheckResult{}, err
	}
	if result.Status == 1 {
		_ = recomputeMonitorRelated(id)
	}
	return result, nil
}

// AddMonitorServ creates a new monitor and optionally attempts to login
func AddMonitorServ(m MonitorReq) (MonitorResp, error) {
	eventToken := strings.TrimSpace(m.EventToken)
	if eventToken == "" {
		generated, err := generateMonitorEventToken()
		if err != nil {
			return MonitorResp{}, fmt.Errorf("failed to generate event token: %w", err)
		}
		eventToken = generated
	}
	monitor := domain.Monitor{
		Name:        m.Name,
		URL:         m.URL,
		Username:    m.Username,
		Password:    m.Password,
		AuthToken:   m.AuthToken,
		EventToken:  eventToken,
		Description: m.Description,
		Type:        m.Type,
		Enabled:     m.Enabled,
		Status:      determineMonitorStatus(domain.Monitor{Enabled: m.Enabled, AuthToken: m.AuthToken, Username: m.Username, Password: m.Password}),
	}

	if err := repository.AddMonitorDAO(monitor); err != nil {
		return MonitorResp{}, fmt.Errorf("failed to create monitor: %w", err)
	}

	// Retrieve the newly created monitor (to get its ID)
	monitors, err := repository.SearchMonitorsDAO(domain.MonitorFilter{
		Query: m.Name,
	})
	if err != nil || len(monitors) == 0 {
		return MonitorResp{}, fmt.Errorf("failed to retrieve created monitor")
	}

	createdMonitor := monitors[len(monitors)-1] // Get the most recent one
	_, _ = recomputeMonitorStatus(createdMonitor.ID)
	createdMonitor, _ = repository.GetMonitorByIDDAO(createdMonitor.ID)
	result := monitorToResp(createdMonitor)

	// If credentials are provided, attempt to login automatically
	if m.Username != "" && m.Password != "" {
		loggedInMonitor, err := LoginMonitorServ(createdMonitor.ID)
		if err != nil {
			// Log the error but don't fail the creation
			LogService("warn", "auto-login failed for monitor", map[string]interface{}{"monitor_id": createdMonitor.ID, "error": err.Error()}, nil, "")
		} else {
			result = loggedInMonitor
		}
	}

	return result, nil
}

// DeleteMonitorServByID deletes a monitor by ID
func DeleteMonitorServByID(id int) error {
	return repository.DeleteMonitorByIDDAO(id)
}

// UpdateMonitorServ updates an existing monitor
func UpdateMonitorServ(id int, m MonitorReq) error {
	existing, err := GetMonitorByIDServ(uint(id))
	if err != nil {
		return err
	}
	eventToken := strings.TrimSpace(m.EventToken)
	if eventToken == "" {
		eventToken = existing.EventToken
	}
	updated := domain.Monitor{
		Name:              m.Name,
		URL:               m.URL,
		Username:          m.Username,
		Password:          m.Password,
		AuthToken:         m.AuthToken,
		EventToken:        eventToken,
		Description:       m.Description,
		Type:              m.Type,
		Enabled:           m.Enabled,
		Status:            existing.Status,
		StatusDescription: existing.StatusDesc,
		HealthScore:       existing.HealthScore,
	}
	// Preserve status and description unless enabled state changed
	if m.Enabled != existing.Enabled {
		updated.Status = determineMonitorStatus(domain.Monitor{Enabled: m.Enabled, AuthToken: m.AuthToken, Username: m.Username, Password: m.Password})
		updated.StatusDescription = ""
	}
	if err := repository.UpdateMonitorDAO(id, updated); err != nil {
		return err
	}
	_, _ = recomputeMonitorStatus(uint(id))
	return recomputeMonitorRelated(uint(id))
}

// RegenerateMonitorEventTokenServ creates a new event token for a monitor
func RegenerateMonitorEventTokenServ(id uint) (MonitorResp, error) {
	newToken, err := generateMonitorEventToken()
	if err != nil {
		return MonitorResp{}, fmt.Errorf("failed to generate event token: %w", err)
	}
	if err := repository.UpdateMonitorEventTokenDAO(id, newToken); err != nil {
		return MonitorResp{}, fmt.Errorf("failed to save event token: %w", err)
	}
	updatedMonitor, err := GetMonitorByIDServ(id)
	if err != nil {
		return MonitorResp{}, fmt.Errorf("failed to retrieve updated monitor: %w", err)
	}
	return updatedMonitor, nil
}

// ValidateMonitorEventTokenServ validates inbound event tokens for webhook use
func ValidateMonitorEventTokenServ(eventToken string) error {
	if strings.TrimSpace(eventToken) == "" {
		return domain.ErrUnauthorized
	}
	_, err := repository.GetMonitorByEventTokenDAO(eventToken)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrUnauthorized
		}
		return err
	}
	return nil
}

// LoginMonitorServ authenticates with a monitor and stores the auth token
func LoginMonitorServ(id uint) (MonitorResp, error) {
	monitor, err := GetMonitorByIDServ(id)
	if err != nil {
		return MonitorResp{}, err
	}

	client, err := monitors.NewClient(monitors.Config{
		Name: monitor.Name,
		Type: monitors.ParseMonitorType(monitor.Type),
		Auth: monitors.AuthConfig{
			URL:      monitor.URL,
			Username: monitor.Username,
			Password: monitor.Password,
			Token:    monitor.AuthToken,
		},
		Timeout: 30,
	})
	if err != nil {
		return MonitorResp{}, fmt.Errorf("failed to create monitor client: %w", err)
	}

	// Authenticate with the monitor
	ctx := context.Background()
	if err := client.Authenticate(ctx); err != nil {
		return MonitorResp{}, fmt.Errorf("authentication failed: %w", err)
	}

	authToken := client.GetAuthToken()
	// SNMP and some other monitor types might not return a central auth token
	if authToken == "" && monitors.ParseMonitorType(monitor.Type) != monitors.MonitorSNMP {
		return MonitorResp{}, fmt.Errorf("authentication succeeded but no token received")
	}

	// Update the auth token in the database if received
	if authToken != "" {
		if err := repository.UpdateMonitorAuthTokenDAO(id, authToken); err != nil {
			return MonitorResp{}, fmt.Errorf("failed to save auth token: %w", err)
		}
	}
	_ = repository.UpdateMonitorStatusDAO(id, 1)

	// Retrieve and return the updated monitor
	updatedMonitor, err := GetMonitorByIDServ(id)
	if err != nil {
		return MonitorResp{}, fmt.Errorf("failed to retrieve updated monitor: %w", err)
	}

	return updatedMonitor, nil
}

// monitorToResp converts a domain Monitor to MonitorResp
func monitorToResp(m domain.Monitor) MonitorResp {
	return MonitorResp{
		ID:          int(m.ID),
		Name:        m.Name,
		URL:         m.URL,
		Username:    m.Username,
		Password:    m.Password,
		AuthToken:   m.AuthToken,
		EventToken:  m.EventToken,
		Description: m.Description,
		Type:        m.Type,
		Enabled:     m.Enabled,
		Status:      m.Status,
		StatusDesc:  m.StatusDescription,
		HealthScore: m.HealthScore,
	}
}
