package service

import (
	"context"
	"fmt"
	"time"

	"nagare/internal/model"
	"nagare/internal/repository"
	"nagare/internal/repository/monitors"
)

// PullSitesFromMonitorServ pulls sites (host groups) from a monitor
func PullSitesFromMonitorServ(mid uint) (SyncResult, error) {
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

	// 4. Sync sites
	for _, group := range groups {
		// Check if site exists by external ID and monitor ID
		site, err := repository.GetSiteByExternalIDDAO(group.ID, mid)
		if err == nil {
			// Update existing site
			site.Name = group.Name
			if err := repository.UpdateSiteDAO(site.ID, site); err == nil {
				result.Updated++
			} else {
				result.Failed++
			}
		} else {
			// Create new site
			newSite := model.Site{
				Name:        group.Name,
				Description: "Imported from " + monitor.Name,
				Enabled:     1,
				Status:      1,
				MonitorID:   mid,
				ExternalID:  group.ID,
			}
			if err := repository.AddSiteDAO(newSite); err == nil {
				result.Added++
			} else {
				result.Failed++
			}
		}
	}

	return result, nil
}

// PushSiteToMonitorServ pushes a site to a monitor (create or update host group)
func PushSiteToMonitorServ(mid uint, siteID uint) error {
	// 1. Get site and monitor
	site, err := repository.GetSiteByIDDAO(siteID)
	if err != nil {
		return fmt.Errorf("failed to get site: %w", err)
	}

	monitor, err := repository.GetMonitorByIDDAO(mid)
	if err != nil {
		return fmt.Errorf("failed to get monitor: %w", err)
	}

	if monitor.Status != 1 {
		return fmt.Errorf("monitor is not active")
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
	if site.ExternalID != "" && site.MonitorID == mid {
		// Update existing host group
		err := client.UpdateHostGroup(ctx, site.ExternalID, site.Name)
		if err != nil {
			return fmt.Errorf("failed to update host group: %w", err)
		}
	} else {
		// Check if group exists by name to avoid duplicates
		groupID, err := client.GetHostGroupByName(ctx, site.Name)
		if err == nil && groupID != "" {
			// Link existing group
			site.ExternalID = groupID
			site.MonitorID = mid
			_ = repository.UpdateSiteDAO(site.ID, site)
		} else {
			// Create new host group
			groupID, err = client.CreateHostGroup(ctx, site.Name)
			if err != nil {
				return fmt.Errorf("failed to create host group: %w", err)
			}
			site.ExternalID = groupID
			site.MonitorID = mid
			_ = repository.UpdateSiteDAO(site.ID, site)
		}
	}

	return nil
}
