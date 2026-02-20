package service

import (
	"context"
	"time"

	"nagare/internal/model"
	"nagare/internal/repository"
	"nagare/internal/repository/monitors"

	"github.com/spf13/viper"
)

const (
	defaultStatusCheckIntervalSeconds = 300
	defaultStatusCheckConcurrency     = 4
)

// StatusCheckResult represents a status check response.
type StatusCheckResult struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Status int    `json:"status"`
	Error  string `json:"error,omitempty"`
}

// StartStatusChecks starts periodic status checks for key resources.
func StartStatusChecks() {
	enabled := viper.GetBool("status_check.enabled")
	if !enabled {
		LogSystem("info", "status checks disabled via configuration", nil, nil, "")
		return
	}

	intervalSeconds := viper.GetInt("status_check.interval_seconds")
	if intervalSeconds <= 0 {
		intervalSeconds = defaultStatusCheckIntervalSeconds
	}

	LogSystem("info", "status checks enabled", map[string]interface{}{"interval_seconds": intervalSeconds}, nil, "")

	go func() {
		ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
		defer ticker.Stop()

		_ = CheckAllMonitorsStatusServ()
		if viper.GetBool("status_check.provider_enabled") {
			_ = CheckAllProvidersStatusServ()
		}
		_ = CheckAllGroupsStatusServ()
		for range ticker.C {
			_ = CheckAllMonitorsStatusServ()
			if viper.GetBool("status_check.provider_enabled") {
				_ = CheckAllProvidersStatusServ()
			}
			_ = CheckAllGroupsStatusServ()
		}
	}()
}

// CheckMonitorStatusServ checks a single monitor's status.
func CheckMonitorStatusServ(id uint) (StatusCheckResult, error) {
	monitor, err := repository.GetMonitorByIDDAO(id)
	if err != nil {
		return StatusCheckResult{}, err
	}
	return checkMonitorStatus(monitor), nil
}

// CheckAllMonitorsStatusServ checks all monitors.
func CheckAllMonitorsStatusServ() []StatusCheckResult {
	monitorsList, err := repository.GetAllMonitorsDAO()
	if err != nil {
		LogSystem("error", "status check failed to load monitors", map[string]interface{}{"error": err.Error()}, nil, "")
		return nil
	}

	results := make([]StatusCheckResult, len(monitorsList))
	limit := configuredLimit("status_check.concurrency", defaultStatusCheckConcurrency)
	runWithLimit(len(monitorsList), limit, func(i int) {
		results[i] = checkMonitorStatus(monitorsList[i])
	})
	return results
}

func checkMonitorStatus(monitor model.Monitor) StatusCheckResult {
	result := StatusCheckResult{ID: monitor.ID, Name: monitor.Name, Status: monitor.Status}
	if monitor.Enabled == 0 {
		_ = repository.UpdateMonitorStatusAndDescriptionDAO(monitor.ID, 0, "")
		result.Status = 0
		return result
	}

	// SNMP monitors are stateless and always considered up if enabled
	if monitor.Type == 4 { // SNMP
		_ = repository.UpdateMonitorStatusAndDescriptionDAO(monitor.ID, 1, "")
		result.Status = 1
		return result
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
		setMonitorRelatedError(monitor.ID, err.Error())
		LogService("error", "monitor status check failed to create client", map[string]interface{}{"monitor_id": monitor.ID, "error": err.Error()}, nil, "")
		result.Status = 2
		result.Error = err.Error()
		return result
	}

	if monitor.AuthToken != "" {
		client.SetAuthToken(monitor.AuthToken)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := client.Authenticate(ctx); err != nil {
		setMonitorRelatedError(monitor.ID, err.Error())
		LogService("error", "monitor status check authentication failed", map[string]interface{}{"monitor_id": monitor.ID, "error": err.Error()}, nil, "")
		result.Status = 2
		result.Error = err.Error()
		return result
	}

	if token := client.GetAuthToken(); token != "" {
		_ = repository.UpdateMonitorAuthTokenDAO(monitor.ID, token)
	}
	_ = repository.UpdateMonitorStatusAndDescriptionDAO(monitor.ID, 1, "")
	_ = recomputeMonitorRelated(monitor.ID)
	result.Status = 1
	return result
}

// CheckProviderStatusServ checks a single provider's status.
func CheckProviderStatusServ(id uint) (StatusCheckResult, error) {
	provider, err := repository.GetProviderByIDDAO(id)
	if err != nil {
		return StatusCheckResult{}, err
	}
	return checkProviderStatus(provider), nil
}

// CheckAllProvidersStatusServ checks all providers.
func CheckAllProvidersStatusServ() []StatusCheckResult {
	providersList, err := repository.GetAllProvidersDAO()
	if err != nil {
		LogSystem("error", "status check failed to load providers", map[string]interface{}{"error": err.Error()}, nil, "")
		return nil
	}

	results := make([]StatusCheckResult, len(providersList))
	limit := configuredLimit("status_check.concurrency", defaultStatusCheckConcurrency)
	runWithLimit(len(providersList), limit, func(i int) {
		results[i] = checkProviderStatus(providersList[i])
	})
	return results
}

func checkProviderStatus(provider model.Provider) StatusCheckResult {
	result := StatusCheckResult{ID: provider.ID, Name: provider.Name, Status: provider.Status}
	if provider.Enabled == 0 {
		_ = repository.UpdateProviderStatusDAO(provider.ID, 0)
		result.Status = 0
		return result
	}
	if provider.APIKey == "" {
		setProviderStatusError(provider.ID)
		err := "provider API key is not configured"
		LogService("error", "provider status check failed", map[string]interface{}{"provider_id": provider.ID, "error": err}, nil, "")
		result.Status = 2
		result.Error = err
		return result
	}

	client, model, err := createLLMClient(provider.ID, provider.DefaultModel)
	if err != nil {
		setProviderStatusError(provider.ID)
		LogService("error", "provider status check failed to create client", map[string]interface{}{"provider_id": provider.ID, "error": err.Error()}, nil, "")
		result.Status = 2
		result.Error = err.Error()
		return result
	}

	if model == "" {
		models := client.AvailableModels()
		if len(models) == 0 {
			err := "no available models for provider"
			setProviderStatusError(provider.ID)
			LogService("error", "provider status check failed", map[string]interface{}{"provider_id": provider.ID, "error": err}, nil, "")
			result.Status = 2
			result.Error = err
			return result
		}
		model = models[0]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if _, err := client.SimpleChat(ctx, model, "ping"); err != nil {
		setProviderStatusError(provider.ID)
		LogService("error", "provider status check failed", map[string]interface{}{"provider_id": provider.ID, "error": err.Error()}, nil, "")
		result.Status = 2
		result.Error = err.Error()
		return result
	}

	_ = repository.UpdateProviderStatusDAO(provider.ID, 1)
	result.Status = 1
	return result
}

// CheckGroupStatusServ checks a single group's status.
func CheckGroupStatusServ(id uint) (StatusCheckResult, error) {
	group, err := repository.GetGroupByIDDAO(id)
	if err != nil {
		return StatusCheckResult{}, err
	}
	return checkGroupStatus(group), nil
}

// CheckAllGroupsStatusServ checks all groups.
func CheckAllGroupsStatusServ() []StatusCheckResult {
	groupsList, err := repository.GetAllGroupsDAO()
	if err != nil {
		LogSystem("error", "status check failed to load groups", map[string]interface{}{"error": err.Error()}, nil, "")
		return nil
	}

	results := make([]StatusCheckResult, len(groupsList))
	limit := configuredLimit("status_check.concurrency", defaultStatusCheckConcurrency)
	runWithLimit(len(groupsList), limit, func(i int) {
		results[i] = checkGroupStatus(groupsList[i])
	})
	return results
}

func checkGroupStatus(group model.Group) StatusCheckResult {
	result := StatusCheckResult{ID: group.ID, Name: group.Name, Status: group.Status}
	status, err := recomputeGroupStatus(group.ID)
	if err != nil {
		setGroupStatusError(group.ID)
		LogService("error", "group status check failed", map[string]interface{}{"group_id": group.ID, "error": err.Error()}, nil, "")
		result.Status = 2
		result.Error = err.Error()
		return result
	}
	result.Status = status
	return result
}
