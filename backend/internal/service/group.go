package service

import (
	"fmt"
	"sync"

	"nagare/internal/model"
	"nagare/internal/repository"
)

// GroupReq represents a group request
type GroupReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Enabled     int    `json:"enabled"`
}

// GroupResp represents a group response
type GroupResp struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     int    `json:"enabled"`
	Status      int    `json:"status"`
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
		Query:  filter.Query,
		Status: filter.Status,
		Limit:  filter.Limit,
		Offset: filter.Offset,
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
		Query:  filter.Query,
		Status: filter.Status,
		Limit:  filter.Limit,
		Offset: filter.Offset,
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
	group.Status = determineGroupStatus(group, nil)
	if err := repository.AddGroupDAO(group); err != nil {
		return GroupResp{}, fmt.Errorf("failed to add group: %w", err)
	}
	return groupToResp(group), nil
}

// UpdateGroupServ updates a group by ID
func UpdateGroupServ(id uint, req GroupReq) error {
	hosts, err := repository.SearchHostsDAO(model.HostFilter{GroupID: &id})
	if err != nil {
		return err
	}
	updated := model.Group{
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
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
	_, err := repository.GetGroupByIDDAO(id)
	if err != nil {
		setGroupStatusError(id)
		LogService("error", "pull group failed to load group", map[string]interface{}{"group_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get group: %w", err)
	}
	hosts, err := repository.SearchHostsDAO(model.HostFilter{GroupID: &id})
	if err != nil {
		setGroupStatusError(id)
		LogService("error", "pull group failed to load hosts", map[string]interface{}{"group_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get group hosts: %w", err)
	}
	result := SyncResult{}
	for _, host := range hosts {
		if host.MonitorID == 0 {
			setGroupStatusError(id)
			LogService("error", "pull group skipped host without monitor", map[string]interface{}{"group_id": id, "host_id": host.ID}, nil, "")
			result.Failed++
			continue
		}
		pullHostRes, err := PullHostFromMonitorServ(host.MonitorID, host.ID)
		if err != nil {
			setGroupStatusError(id)
			LogService("error", "pull group failed to pull host", map[string]interface{}{"group_id": id, "host_id": host.ID, "monitor_id": host.MonitorID, "error": err.Error()}, nil, "")
			result.Failed++
			continue
		}
		result.Added += pullHostRes.Added
		result.Updated += pullHostRes.Updated
		result.Failed += pullHostRes.Failed
		result.Total += pullHostRes.Total

		pullItemRes, err := PullItemsFromHostServ(host.MonitorID, host.ID)
		if err != nil {
			setGroupStatusError(id)
			LogService("error", "pull group failed to pull items", map[string]interface{}{"group_id": id, "host_id": host.ID, "monitor_id": host.MonitorID, "error": err.Error()}, nil, "")
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

func PushGroupToMonitorsServ(id uint) (SyncResult, error) {
	_, err := repository.GetGroupByIDDAO(id)
	if err != nil {
		setGroupStatusError(id)
		LogService("error", "push group failed to load group", map[string]interface{}{"group_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get group: %w", err)
	}
	hosts, err := repository.SearchHostsDAO(model.HostFilter{GroupID: &id})
	if err != nil {
		setGroupStatusError(id)
		LogService("error", "push group failed to load hosts", map[string]interface{}{"group_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get group hosts: %w", err)
	}
	result := SyncResult{}
	for _, host := range hosts {
		if host.MonitorID == 0 {
			setGroupStatusError(id)
			LogService("error", "push group skipped host without monitor", map[string]interface{}{"group_id": id, "host_id": host.ID}, nil, "")
			result.Failed++
			continue
		}
		if err := PushHostToMonitorServ(host.MonitorID, host.ID); err != nil {
			setGroupStatusError(id)
			LogService("error", "push group failed to push host", map[string]interface{}{"group_id": id, "host_id": host.ID, "monitor_id": host.MonitorID, "error": err.Error()}, nil, "")
			result.Failed++
			continue
		}
		result.Added++
		result.Total++

		pushItemsRes, err := PushItemsFromHostServ(host.MonitorID, host.ID)
		if err != nil {
			setGroupStatusError(id)
			LogService("error", "push group failed to push items", map[string]interface{}{"group_id": id, "host_id": host.ID, "monitor_id": host.MonitorID, "error": err.Error()}, nil, "")
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

func groupToResp(group model.Group) GroupResp {
	return GroupResp{
		ID:          int(group.ID),
		Name:        group.Name,
		Description: group.Description,
		Enabled:     group.Enabled,
		Status:      group.Status,
	}
}
