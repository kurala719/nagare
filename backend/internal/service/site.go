package service

import (
	"fmt"
	"sync"

	"nagare/internal/model"
	"nagare/internal/repository"
)

// SiteReq represents a site request
type SiteReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Enabled     int    `json:"enabled"`
}

// SiteResp represents a site response
type SiteResp struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     int    `json:"enabled"`
	Status      int    `json:"status"`
}

// SiteSummary represents aggregated site data
type SiteSummary struct {
	TotalHosts   int `json:"total_hosts"`
	ActiveHosts  int `json:"active_hosts"`
	ErrorHosts   int `json:"error_hosts"`
	SyncingHosts int `json:"syncing_hosts"`
	TotalItems   int `json:"total_items"`
}

// SiteDetailResp represents detailed site data
type SiteDetailResp struct {
	Site    SiteResp    `json:"site"`
	Summary SiteSummary `json:"summary"`
	Hosts   []HostResp  `json:"hosts"`
}

// GetAllSitesServ retrieves all sites
func GetAllSitesServ() ([]SiteResp, error) {
	sites, err := repository.GetAllSitesDAO()
	if err != nil {
		return nil, fmt.Errorf("failed to get sites: %w", err)
	}
	result := make([]SiteResp, 0, len(sites))
	for _, s := range sites {
		result = append(result, siteToResp(s))
	}
	return result, nil
}

// SearchSitesServ retrieves sites by filter
func SearchSitesServ(filter model.SiteFilter) ([]SiteResp, error) {
	sites, err := repository.SearchSitesDAO(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search sites: %w", err)
	}
	result := make([]SiteResp, 0, len(sites))
	for _, s := range sites {
		result = append(result, siteToResp(s))
	}
	return result, nil
}

// CountSitesServ returns total count for sites by filter
func CountSitesServ(filter model.SiteFilter) (int64, error) {
	return repository.CountSitesDAO(filter)
}

// GetSiteByIDServ retrieves a site by ID
func GetSiteByIDServ(id uint) (SiteResp, error) {
	site, err := repository.GetSiteByIDDAO(id)
	if err != nil {
		return SiteResp{}, fmt.Errorf("failed to get site: %w", err)
	}
	return siteToResp(site), nil
}

// AddSiteServ creates a new site
func AddSiteServ(req SiteReq) (SiteResp, error) {
	site := model.Site{
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
	}
	site.Status = determineSiteStatus(site, nil)
	if err := repository.AddSiteDAO(site); err != nil {
		return SiteResp{}, fmt.Errorf("failed to add site: %w", err)
	}
	return siteToResp(site), nil
}

// UpdateSiteServ updates a site by ID
func UpdateSiteServ(id uint, req SiteReq) error {
	hosts, err := repository.SearchHostsDAO(model.HostFilter{SiteID: &id})
	if err != nil {
		return err
	}
	updated := model.Site{
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
	}
	updated.Status = determineSiteStatus(updated, hosts)
	if err := repository.UpdateSiteDAO(id, updated); err != nil {
		return err
	}
	return nil
}

// DeleteSiteByIDServ deletes a site by ID
func DeleteSiteByIDServ(id uint) error {
	return repository.DeleteSiteByIDDAO(id)
}

// GetSiteDetailServ returns site with summary and hosts
func GetSiteDetailServ(id uint) (SiteDetailResp, error) {
	site, err := repository.GetSiteByIDDAO(id)
	if err != nil {
		return SiteDetailResp{}, fmt.Errorf("failed to get site: %w", err)
	}
	hosts, err := repository.SearchHostsDAO(model.HostFilter{SiteID: &id})
	if err != nil {
		return SiteDetailResp{}, fmt.Errorf("failed to get site hosts: %w", err)
	}

	summary := SiteSummary{}
	respHosts := make([]HostResp, len(hosts))
	var mu sync.Mutex

	limit := configuredLimit("site.detail_concurrency", 10)
	runWithLimit(len(hosts), limit, func(i int) {
		h := hosts[i]
		respHosts[i] = hostToResp(h)

		items, err := repository.GetItemsByHIDDAO(h.ID)
		itemCount := 0
		if err == nil {
			itemCount = len(items)
		}

		mu.Lock()
		summary.TotalHosts++
		switch h.Status {
		case 1:
			summary.ActiveHosts++
		case 2:
			summary.ErrorHosts++
		case 3:
			summary.SyncingHosts++
		}
		summary.TotalItems += itemCount
		mu.Unlock()
	})

	return SiteDetailResp{
		Site:    siteToResp(site),
		Summary: summary,
		Hosts:   respHosts,
	}, nil
}

func PullSiteFromMonitorsServ(id uint) (SyncResult, error) {
	_, err := repository.GetSiteByIDDAO(id)
	if err != nil {
		setSiteStatusError(id)
		LogService("error", "pull site failed to load site", map[string]interface{}{"site_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get site: %w", err)
	}
	hosts, err := repository.SearchHostsDAO(model.HostFilter{SiteID: &id})
	if err != nil {
		setSiteStatusError(id)
		LogService("error", "pull site failed to load hosts", map[string]interface{}{"site_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get site hosts: %w", err)
	}
	result := SyncResult{}
	for _, host := range hosts {
		if host.MonitorID == 0 {
			setSiteStatusError(id)
			LogService("error", "pull site skipped host without monitor", map[string]interface{}{"site_id": id, "host_id": host.ID}, nil, "")
			result.Failed++
			continue
		}
		pullHostRes, err := PullHostFromMonitorServ(host.MonitorID, host.ID)
		if err != nil {
			setSiteStatusError(id)
			LogService("error", "pull site failed to pull host", map[string]interface{}{"site_id": id, "host_id": host.ID, "monitor_id": host.MonitorID, "error": err.Error()}, nil, "")
			result.Failed++
			continue
		}
		result.Added += pullHostRes.Added
		result.Updated += pullHostRes.Updated
		result.Failed += pullHostRes.Failed
		result.Total += pullHostRes.Total

		pullItemRes, err := PullItemsFromHostServ(host.MonitorID, host.ID)
		if err != nil {
			setSiteStatusError(id)
			LogService("error", "pull site failed to pull items", map[string]interface{}{"site_id": id, "host_id": host.ID, "monitor_id": host.MonitorID, "error": err.Error()}, nil, "")
			result.Failed++
			continue
		}
		result.Added += pullItemRes.Added
		result.Updated += pullItemRes.Updated
		result.Failed += pullItemRes.Failed
		result.Total += pullItemRes.Total
	}
	return result, nil
}

func PushSiteToMonitorsServ(id uint) (SyncResult, error) {
	_, err := repository.GetSiteByIDDAO(id)
	if err != nil {
		setSiteStatusError(id)
		LogService("error", "push site failed to load site", map[string]interface{}{"site_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get site: %w", err)
	}
	hosts, err := repository.SearchHostsDAO(model.HostFilter{SiteID: &id})
	if err != nil {
		setSiteStatusError(id)
		LogService("error", "push site failed to load hosts", map[string]interface{}{"site_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get site hosts: %w", err)
	}
	result := SyncResult{}
	for _, host := range hosts {
		if host.MonitorID == 0 {
			setSiteStatusError(id)
			LogService("error", "push site skipped host without monitor", map[string]interface{}{"site_id": id, "host_id": host.ID}, nil, "")
			result.Failed++
			continue
		}
		if err := PushHostToMonitorServ(host.MonitorID, host.ID); err != nil {
			setSiteStatusError(id)
			LogService("error", "push site failed to push host", map[string]interface{}{"site_id": id, "host_id": host.ID, "monitor_id": host.MonitorID, "error": err.Error()}, nil, "")
			result.Failed++
			continue
		}
		result.Added++
		result.Total++

		pushItemsRes, err := PushItemsFromHostServ(host.MonitorID, host.ID)
		if err != nil {
			setSiteStatusError(id)
			LogService("error", "push site failed to push items", map[string]interface{}{"site_id": id, "host_id": host.ID, "monitor_id": host.MonitorID, "error": err.Error()}, nil, "")
			result.Failed++
			continue
		}
		result.Added += pushItemsRes.Added
		result.Updated += pushItemsRes.Updated
		result.Failed += pushItemsRes.Failed
		result.Total += pushItemsRes.Total
	}
	return result, nil
}

func siteToResp(site model.Site) SiteResp {
	return SiteResp{
		ID:          int(site.ID),
		Name:        site.Name,
		Description: site.Description,
		Enabled:     site.Enabled,
		Status:      site.Status,
	}
}
