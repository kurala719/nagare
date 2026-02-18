package service

import (
	"fmt"
	"sync"

	"nagare/internal/model"
	"nagare/internal/repository"
)

// GroupReq represents a group request
type GroupReq struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Enabled     int     `json:"enabled"`
	MonitorID   *uint   `json:"monitor_id,omitempty"`
	ExternalID  *string `json:"external_id,omitempty"`
}

// GroupResp represents a group response
type GroupResp struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     int    `json:"enabled"`
	Status      int    `json:"status"`
	MonitorID   uint   `json:"monitor_id"`
	ExternalID  string `json:"external_id"`
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
	group.Status = determineGroupStatus(group, nil)
	if err := repository.AddGroupDAO(group); err != nil {
		return GroupResp{}, fmt.Errorf("failed to add group: %w", err)
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

	hosts, err := repository.SearchHostsDAO(model.HostFilter{GroupID: &id})
	if err != nil {
		return err
	}

	// Update fields from request
	updated := model.Group{
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
		MonitorID:   existing.MonitorID,  // Preserve existing value
		ExternalID:  existing.ExternalID, // Preserve existing value
	}

	// Only update if provided in request
	if req.MonitorID != nil {
		updated.MonitorID = *req.MonitorID
	}
	if req.ExternalID != nil {
		updated.ExternalID = *req.ExternalID
	}

	updated.Status = determineGroupStatus(updated, hosts)
	if err := repository.UpdateGroupDAO(id, updated); err != nil {
		return err
	}
	return nil
}

// DeleteGroupByIDServ deletes a group by ID
func DeleteGroupByIDServ(id uint) error {
	return repository.DeleteGroupByIDDAO(id)
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

// PullGroupHostsServ pulls hosts for a group from monitors (without pulling group entity)
func PullGroupHostsServ(id uint) (SyncResult, error) {
	group, err := repository.GetGroupByIDDAO(id)
	if err != nil {
		setGroupStatusError(id)
		LogService("error", "pull group hosts failed to load group", map[string]interface{}{"group_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get group: %w", err)
	}

	// Discover and sync all hosts belonging to this group from the monitor
	if group.MonitorID > 0 && group.ExternalID != "" {
		result, err := PullHostsOfGroupFromMonitorServ(group.MonitorID, id)
		if err != nil {
			setGroupStatusError(id)
			LogService("error", "pull group hosts failed to discover hosts", map[string]interface{}{"group_id": id, "monitor_id": group.MonitorID, "error": err.Error()}, nil, "")
			return result, err
		}
		
		_, _ = recomputeGroupStatus(id)
		return result, nil
	}

	// Fallback: If no monitor or external ID, pull existing local hosts
	hosts, err := repository.SearchHostsDAO(model.HostFilter{GroupID: &id})
	if err != nil {
		setGroupStatusError(id)
		LogService("error", "pull group hosts failed to load hosts", map[string]interface{}{"group_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get group hosts: %w", err)
	}
	result := SyncResult{}
	for _, host := range hosts {
		mid := host.MonitorID
		if mid == 0 {
			// If host has no specific monitor, try to use group's monitor if available
			if group.MonitorID > 0 {
				mid = group.MonitorID
			} else {
				setGroupStatusError(id)
				LogService("error", "pull group hosts skipped host without monitor", map[string]interface{}{"group_id": id, "host_id": host.ID}, nil, "")
				result.Failed++
				continue
			}
		}
		pullHostRes, err := PullHostFromMonitorServ(mid, host.ID)
		if err != nil {
			setGroupStatusError(id)
			LogService("error", "pull group hosts failed to pull host", map[string]interface{}{"group_id": id, "host_id": host.ID, "monitor_id": mid, "error": err.Error()}, nil, "")
			result.Failed++
			continue
		}
		result.Added += pullHostRes.Added
		result.Updated += pullHostRes.Updated
		result.Failed += pullHostRes.Failed
		result.Total += pullHostRes.Total
	}
	return result, nil
}

// PushGroupHostsServ pushes hosts for a group to monitors (without pushing group entity)
func PushGroupHostsServ(id uint) (SyncResult, error) {
	group, err := repository.GetGroupByIDDAO(id)
	if err != nil {
		setGroupStatusError(id)
		LogService("error", "push group hosts failed to load group", map[string]interface{}{"group_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get group: %w", err)
	}

	hosts, err := repository.SearchHostsDAO(model.HostFilter{GroupID: &id})
	if err != nil {
		setGroupStatusError(id)
		LogService("error", "push group hosts failed to load hosts", map[string]interface{}{"group_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get group hosts: %w", err)
	}
	result := SyncResult{}
	for _, host := range hosts {
		if host.MonitorID == 0 {
			// If host has no specific monitor, try to use group's monitor if available
			if group.MonitorID > 0 {
				host.MonitorID = group.MonitorID
			} else {
				setGroupStatusError(id)
				LogService("error", "push group hosts skipped host without monitor", map[string]interface{}{"group_id": id, "host_id": host.ID}, nil, "")
				result.Failed++
				continue
			}
		}
		hostResult, err := PushHostToMonitorServ(host.MonitorID, host.ID)
		if err != nil {
			setGroupStatusError(id)
			LogService("error", "push group hosts failed to push host", map[string]interface{}{"group_id": id, "host_id": host.ID, "monitor_id": host.MonitorID, "error": err.Error()}, nil, "")
			result.Failed++
			continue
		}
		result.Added += hostResult.Added
		result.Updated += hostResult.Updated
		result.Failed += hostResult.Failed
		result.Total += hostResult.Total
	}
	return result, nil
}

func groupToResp(group model.Group) GroupResp {
	return GroupResp{
		ID:          int(group.ID),
		Name:        group.Name,
		Description: group.Description,
		Enabled:     group.Enabled,
		Status:      group.Status,
		MonitorID:   group.MonitorID,
		ExternalID:  group.ExternalID,
	}
}
