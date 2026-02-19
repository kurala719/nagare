package service

import (
	"context"
	"fmt"
	"time"

	"nagare/internal/model"
	"nagare/internal/repository"
	"nagare/internal/repository/monitors"
)

// PullGroupsFromMonitorServ pulls groups (host groups) from a monitor
func PullGroupsFromMonitorServ(mid uint) (SyncResult, error) {
	result := SyncResult{}

	// 1. Get monitor
	monitor, err := repository.GetMonitorByIDDAO(mid)
	if err != nil {
		return result, fmt.Errorf("failed to get monitor: %w", err)
	}

	if monitor.Status != 1 {
		return result, fmt.Errorf("monitor is not active")
	}

	// 2. Create monitor client
	cfg := monitors.Config{
		Name:    monitor.Name,
		Type:    monitors.ParseMonitorType(monitor.Type),
		Timeout: 10,
		Auth: monitors.AuthConfig{
			URL:      monitor.URL,
			Username: monitor.Username,
			Password: monitor.Password,
			Token:    monitor.AuthToken,
		},
	}

	client, err := monitors.NewClient(cfg)
	if err != nil {
		return result, fmt.Errorf("failed to create monitor client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Authenticate if needed
	if monitor.AuthToken == "" {
		if err := client.Authenticate(ctx); err != nil {
			return result, fmt.Errorf("failed to authenticate with monitor: %w", err)
		}
		// Update auth token in DB
		monitor.AuthToken = client.GetAuthToken()
		_ = repository.UpdateMonitorDAO(int(mid), monitor)
	}

	// 3. Get host groups from monitor
	groups, err := client.GetHostGroupsDetails(ctx)
	if err != nil {
		return result, fmt.Errorf("failed to get host groups: %w", err)
	}

	result.Total = len(groups)

	// Pre-fetch local groups for name matching
	allGroups, err := repository.GetAllGroupsDAO()
	localGroupsByName := make(map[string]model.Group)
	if err == nil {
		for _, g := range allGroups {
			if g.MonitorID == mid || g.MonitorID == 0 {
				localGroupsByName[g.Name] = g
			}
		}
	}

	// 4. Sync groups
	for _, hostGroup := range groups {
		// Check if group exists by external ID and monitor ID
		group, err := repository.GetGroupByExternalIDDAO(hostGroup.ID, mid)
		if err == nil {
			// Update existing group
			group.Name = hostGroup.Name
			if err := repository.UpdateGroupDAO(group.ID, group); err == nil {
				result.Updated++
			} else {
				result.Failed++
			}
		} else {
			// Check if group exists by name
			if existing, ok := localGroupsByName[hostGroup.Name]; ok {
				existing.ExternalID = hostGroup.ID
				existing.MonitorID = mid
				if err := repository.UpdateGroupDAO(existing.ID, existing); err == nil {
					result.Updated++
				} else {
					result.Failed++
				}
				continue
			}

			// Create new group
			newGroup := model.Group{
				Name:        hostGroup.Name,
				Description: "Imported from " + monitor.Name,
				Enabled:     1,
				Status:      1,
				MonitorID:   mid,
				ExternalID:  hostGroup.ID,
			}
			if err := repository.AddGroupDAO(newGroup); err == nil {
				result.Added++
			} else {
				result.Failed++
			}
		}
	}

	return result, nil
}

// PushGroupToMonitorServ pushes a group to a monitor (create or update host group)
func PushGroupToMonitorServ(mid uint, groupID uint) error {
	// 1. Get group and monitor
	group, err := repository.GetGroupByIDDAO(groupID)
	if err != nil {
		return fmt.Errorf("failed to get group: %w", err)
	}

	monitor, err := repository.GetMonitorByIDDAO(mid)
	if err != nil {
		return fmt.Errorf("failed to get monitor: %w", err)
	}

	if monitor.Status == 2 {
		return fmt.Errorf("monitor is in error state")
	}

	// 2. Create monitor client
	cfg := monitors.Config{
		Name:    monitor.Name,
		Type:    monitors.ParseMonitorType(monitor.Type),
		Timeout: 10,
		Auth: monitors.AuthConfig{
			URL:      monitor.URL,
			Username: monitor.Username,
			Password: monitor.Password,
			Token:    monitor.AuthToken,
		},
	}

	client, err := monitors.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create monitor client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if monitor.AuthToken == "" {
		if err := client.Authenticate(ctx); err != nil {
			return fmt.Errorf("failed to authenticate with monitor: %w", err)
		}
		monitor.AuthToken = client.GetAuthToken()
		_ = repository.UpdateMonitorDAO(int(mid), monitor)
	}

	// 3. Push Logic
	if group.ExternalID != "" && group.MonitorID == mid {
		// Update existing host group
		err := client.UpdateHostGroup(ctx, group.ExternalID, group.Name)
		if err != nil {
			return fmt.Errorf("failed to update host group: %w", err)
		}
	} else {
		// Check if group exists by name to avoid duplicates
		extGroupID, err := client.GetHostGroupByName(ctx, group.Name)
		if err == nil && extGroupID != "" {
			// Link existing group
			group.ExternalID = extGroupID
			group.MonitorID = mid
			_ = repository.UpdateGroupDAO(group.ID, group)
		} else {
			// Create new host group
			extGroupID, err = client.CreateHostGroup(ctx, group.Name)
			if err != nil {
				return fmt.Errorf("failed to create host group: %w", err)
			}
			group.ExternalID = extGroupID
			group.MonitorID = mid
			_ = repository.UpdateGroupDAO(group.ID, group)
		}
	}

	return nil
}

// PullGroupFromMonitorServ pulls a single group from a monitor
func PullGroupFromMonitorServ(mid uint, groupID uint) error {
	// 1. Get group and monitor
	group, err := repository.GetGroupByIDDAO(groupID)
	if err != nil {
		return fmt.Errorf("failed to get group: %w", err)
	}

	monitor, err := repository.GetMonitorByIDDAO(mid)
	if err != nil {
		return fmt.Errorf("failed to get monitor: %w", err)
	}

	if monitor.Status == 2 {
		return fmt.Errorf("monitor is in error state")
	}

	// 2. Create monitor client
	cfg := monitors.Config{
		Name:    monitor.Name,
		Type:    monitors.ParseMonitorType(monitor.Type),
		Timeout: 10,
		Auth: monitors.AuthConfig{
			URL:      monitor.URL,
			Username: monitor.Username,
			Password: monitor.Password,
			Token:    monitor.AuthToken,
		},
	}

	client, err := monitors.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create monitor client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if monitor.AuthToken == "" {
		if err := client.Authenticate(ctx); err != nil {
			return fmt.Errorf("failed to authenticate with monitor: %w", err)
		}
		monitor.AuthToken = client.GetAuthToken()
		_ = repository.UpdateMonitorDAO(int(mid), monitor)
	}

	// 3. Get host group details
	// We fetch all to ensure we can find by name if ID mismatch, or by ID
	groups, err := client.GetHostGroupsDetails(ctx)
	if err != nil {
		return fmt.Errorf("failed to get host groups: %w", err)
	}

	now := time.Now().UTC()
	for _, g := range groups {
		if (group.ExternalID != "" && g.ID == group.ExternalID) || g.Name == group.Name {
			group.Name = g.Name
			group.ExternalID = g.ID
			group.LastSyncAt = &now
			group.ExternalSource = monitor.Name
			return repository.UpdateGroupDAO(group.ID, group)
		}
	}

	return fmt.Errorf("group not found on monitor")
}
