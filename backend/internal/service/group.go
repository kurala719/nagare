package service

import (
	"fmt"
	"sync"
	"time"

	"nagare/internal/model"
	"nagare/internal/repository"
)

// GroupReq represents a group request
type GroupReq struct {
	Name           string     `json:"name" binding:"required"`
	Description    string     `json:"description"`
	Enabled        int        `json:"enabled"`
	MonitorID      *uint      `json:"monitor_id,omitempty"`
	ExternalID     *string    `json:"external_id,omitempty"`
	LastSyncAt     *time.Time `json:"last_sync_at,omitempty"`
	ExternalSource *string    `json:"external_source,omitempty"`
	PushToMonitor  bool       `json:"push_to_monitor"`
}

// GroupResp represents a group response
type GroupResp struct {
	ID             int        `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Enabled        int        `json:"enabled"`
	Status         int        `json:"status"`
	MonitorID      uint       `json:"monitor_id"`
	ExternalID     string     `json:"external_id"`
	LastSyncAt     *time.Time `json:"last_sync_at"`
	ExternalSource string     `json:"external_source"`
	HealthScore    int        `json:"health_score"`
}

// GroupSummary represents aggregated group data
type GroupSummary struct {
	TotalHosts   int `json:"total_hosts"`
	ActiveHosts  int `json:"active_hosts"`
	ErrorHosts   int `json:"error_hosts"`
	SyncingHosts int `json:"syncing_hosts"`
	TotalItems   int `json:"total_items"`
}

// GroupDetailResp represents detailed group data
type GroupDetailResp struct {
	Group   GroupResp    `json:"group"`
	Summary GroupSummary `json:"summary"`
	Hosts   []HostResp   `json:"hosts"`
}

// GetAllGroupsServ retrieves all groups
func GetAllGroupsServ() ([]GroupResp, error) {
	groups, err := repository.GetAllGroupsDAO()
	if err != nil {
		return nil, fmt.Errorf("failed to get groups: %w", err)
	}
	result := make([]GroupResp, 0, len(groups))
	for _, s := range groups {
		result = append(result, groupToResp(s))
	}
	return result, nil
}

// SearchGroupsServ retrieves groups by filter
func SearchGroupsServ(filter model.GroupFilter) ([]GroupResp, error) {
	groupFilter := model.GroupFilter{
		Query:     filter.Query,
		Status:    filter.Status,
		MonitorID: filter.MonitorID,
		Limit:     filter.Limit,
		Offset:    filter.Offset,
		SortBy:    filter.SortBy,
		SortOrder: filter.SortOrder,
	}
	groups, err := repository.SearchGroupsDAO(groupFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to search groups: %w", err)
	}
	result := make([]GroupResp, 0, len(groups))
	for _, s := range groups {
		result = append(result, groupToResp(s))
	}
	return result, nil
}

// CountGroupsServ returns total count for groups by filter
func CountGroupsServ(filter model.GroupFilter) (int64, error) {
	groupFilter := model.GroupFilter{
		Query:     filter.Query,
		Status:    filter.Status,
		MonitorID: filter.MonitorID,
		Limit:     filter.Limit,
		Offset:    filter.Offset,
	}
	return repository.CountGroupsDAO(groupFilter)
}

// GetGroupByIDServ retrieves a group by ID
func GetGroupByIDServ(id uint) (GroupResp, error) {
	group, err := repository.GetGroupByIDDAO(id)
	if err != nil {
		return GroupResp{}, fmt.Errorf("failed to get group: %w", err)
	}
	return groupToResp(group), nil
}

// AddGroupServ creates a new group
func AddGroupServ(req GroupReq) (GroupResp, error) {
	group := model.Group{
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
	}
	if req.MonitorID != nil {
		group.MonitorID = *req.MonitorID
	}
	if req.ExternalID != nil {
		group.ExternalID = *req.ExternalID
	}

	monitorStatus := 1
	if group.MonitorID > 0 {
		if m, err := repository.GetMonitorByIDDAO(group.MonitorID); err == nil {
			monitorStatus = determineMonitorStatus(m)
		}
	}

	group.Status = determineGroupStatus(group, monitorStatus)
	if err := repository.AddGroupDAO(group); err != nil {
		return GroupResp{}, fmt.Errorf("failed to add group: %w", err)
	}
	// Fetch ID and recompute
	groups, err := repository.SearchGroupsDAO(model.GroupFilter{Query: group.Name})
	if err == nil && len(groups) > 0 {
		created := groups[len(groups)-1]
		
		// Auto-push to monitor if MID is set AND PushToMonitor is true
		if created.MonitorID > 0 && req.PushToMonitor {
			_ = PushGroupToMonitorServ(created.MonitorID, created.ID)
		}
		
		_, _ = recomputeGroupStatus(created.ID)
		group, _ = repository.GetGroupByIDDAO(created.ID)
	}
	return groupToResp(group), nil
}

// UpdateGroupServ updates a group by ID
func UpdateGroupServ(id uint, req GroupReq) error {
	// Get existing group first
	existing, err := repository.GetGroupByIDDAO(id)
	if err != nil {
		return err
	}

	// Update fields from request
	updated := model.Group{
		Name:           req.Name,
		Description:    req.Description,
		Enabled:        req.Enabled,
		MonitorID:      existing.MonitorID,      // Preserve existing value
		ExternalID:     existing.ExternalID,     // Preserve existing value
		LastSyncAt:     existing.LastSyncAt,     // Preserve existing value
		ExternalSource: existing.ExternalSource, // Preserve existing value
		Status:         existing.Status,
		HealthScore:    existing.HealthScore,
	}

	// Only update if provided in request
	if req.MonitorID != nil {
		updated.MonitorID = *req.MonitorID
	}
	if req.ExternalID != nil {
		updated.ExternalID = *req.ExternalID
	}
	if req.LastSyncAt != nil {
		updated.LastSyncAt = req.LastSyncAt
	}
	if req.ExternalSource != nil {
		updated.ExternalSource = *req.ExternalSource
	}

	// Preserve status unless enabled state changed
	if req.Enabled != existing.Enabled {
		monitorStatus := 1
		if updated.MonitorID > 0 {
			if m, err := repository.GetMonitorByIDDAO(updated.MonitorID); err == nil {
				monitorStatus = determineMonitorStatus(m)
			}
		}
		updated.Status = determineGroupStatus(updated, monitorStatus)
	}
	if err := repository.UpdateGroupDAO(id, updated); err != nil {
		return err
	}
	
	// Auto-push to monitor if PushToMonitor is true
	if updated.MonitorID > 0 && req.PushToMonitor {
		_ = PushGroupToMonitorServ(updated.MonitorID, id)
	} else if updated.MonitorID > 0 && (updated.Name != existing.Name || updated.Description != existing.Description) {
		// Existing logic: still push if name/desc changed? 
        // User asked to SEPARATE it. So maybe I should REMOVE this auto-push logic if they didn't explicitly ask for it.
        // Actually, if they asked to separate it, it means they want to control it.
        // I will remove the auto-push on name change and rely on the flag.
	}
	
	_, _ = recomputeGroupStatus(id)
	return nil
}

// DeleteGroupByIDServ deletes a group by ID and all its associated hosts and items
func DeleteGroupByIDServ(id uint) error {
	// 1. Get all hosts in this group to perform cascading delete
	hosts, err := repository.SearchHostsDAO(model.HostFilter{GroupID: &id})
	if err == nil {
		for _, h := range hosts {
			// Call service-level delete to handle Host -> Item cascade
			if err := DeleteHostByIDServ(h.ID); err != nil {
				LogService("error", "failed to delete host during group cascading delete", map[string]interface{}{"group_id": id, "host_id": h.ID, "error": err.Error()}, nil, "")
			}
		}
	}

	// 2. Delete the group itself
	return repository.DeleteGroupByIDDAO(id)
}

// DeleteGroupsByMIDServ deletes all groups by monitor ID and their associated hosts/items
func DeleteGroupsByMIDServ(mid uint) error {
	groups, err := repository.SearchGroupsDAO(model.GroupFilter{MonitorID: &mid})
	if err == nil {
		for _, g := range groups {
			if err := DeleteGroupByIDServ(g.ID); err != nil {
				LogService("error", "failed to delete group during monitor cascading delete", map[string]interface{}{"monitor_id": mid, "group_id": g.ID, "error": err.Error()}, nil, "")
			}
		}
	}
	return repository.DeleteGroupsByMIDDAO(mid)
}

// GetGroupDetailServ returns group with summary and hosts
func GetGroupDetailServ(id uint) (GroupDetailResp, error) {
	group, err := repository.GetGroupByIDDAO(id)
	if err != nil {
		return GroupDetailResp{}, fmt.Errorf("failed to get group: %w", err)
	}
	hosts, err := repository.SearchHostsDAO(model.HostFilter{GroupID: &id})
	if err != nil {
		return GroupDetailResp{}, fmt.Errorf("failed to get group hosts: %w", err)
	}

	summary := GroupSummary{}
	respHosts := make([]HostResp, len(hosts))
	var mu sync.Mutex

	limit := configuredLimit("group.detail_concurrency", 10)
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

	return GroupDetailResp{
		Group:   groupToResp(group),
		Summary: summary,
		Hosts:   respHosts,
	}, nil
}

func PullGroupFromMonitorsServ(id uint) (SyncResult, error) {
	group, err := repository.GetGroupByIDDAO(id)
	if err != nil {
		setGroupStatusError(id)
		LogService("error", "pull group failed to load group", map[string]interface{}{"group_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get group: %w", err)
	}

	result := SyncResult{Total: 1}

	// Pull group metadata first if monitor is set
	if group.MonitorID > 0 {
		if err := PullGroupFromMonitorServ(group.MonitorID, id); err != nil {
			setGroupStatusError(id)
			LogService("error", "pull group failed to pull group entity", map[string]interface{}{"group_id": id, "monitor_id": group.MonitorID, "error": err.Error()}, nil, "")
			result.Failed = 1
			return result, err
		}
		result.Updated = 1
	} else {
		return result, fmt.Errorf("group is not associated with a monitor")
	}

	_, _ = recomputeGroupStatus(id)
	return result, nil
}

func PushGroupToMonitorsServ(id uint) (SyncResult, error) {
	group, err := repository.GetGroupByIDDAO(id)
	if err != nil {
		setGroupStatusError(id)
		LogService("error", "push group failed to load group", map[string]interface{}{"group_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get group: %w", err)
	}

	result := SyncResult{Total: 1}

	// Push group metadata if monitor is set
	if group.MonitorID > 0 {
		if err := PushGroupToMonitorServ(group.MonitorID, id); err != nil {
			setGroupStatusError(id)
			LogService("error", "push group failed to push group entity", map[string]interface{}{"group_id": id, "monitor_id": group.MonitorID, "error": err.Error()}, nil, "")
			result.Failed = 1
			return result, err
		}
		result.Updated = 1
	} else {
		return result, fmt.Errorf("group is not associated with a monitor")
	}

	return result, nil
}

// PullGroupConfigServ pulls group configuration from monitor
func PullGroupConfigServ(id uint) error {
	group, err := repository.GetGroupByIDDAO(id)
	if err != nil {
		return err
	}
	if group.MonitorID == 0 {
		return fmt.Errorf("group is not associated with a monitor")
	}
	return PullGroupFromMonitorServ(group.MonitorID, id)
}

// PushGroupConfigServ pushes group configuration to monitor
func PushGroupConfigServ(id uint) error {
	group, err := repository.GetGroupByIDDAO(id)
	if err != nil {
		return err
	}
	if group.MonitorID == 0 {
		return fmt.Errorf("group is not associated with a monitor")
	}
	return PushGroupToMonitorServ(group.MonitorID, id)
}

func groupToResp(group model.Group) GroupResp {
	return GroupResp{
		ID:             int(group.ID),
		Name:           group.Name,
		Description:    group.Description,
		Enabled:        group.Enabled,
		Status:         group.Status,
		MonitorID:      group.MonitorID,
		ExternalID:     group.ExternalID,
		LastSyncAt:     group.LastSyncAt,
		ExternalSource: group.ExternalSource,
		HealthScore:    group.HealthScore,
	}
}
