package service

import (
	"context"
	"fmt"
	"time"

	"nagare/internal/model"
	"nagare/internal/repository"
	"nagare/internal/repository/monitors"
)

// HostReq represents a host request
type HostReq struct {
	Name        string `json:"name" binding:"required"`
	MID         uint   `json:"m_id"`
	GroupID     uint   `json:"group_id"`
	Hostid      string `json:"hostid"`
	Description string `json:"description"`
	Enabled     int    `json:"enabled"`
	IPAddr      string `json:"ip_addr"`
	Comment     string `json:"comment"`
}

// HostResp represents a host response
type HostResp struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	MID         uint   `json:"m_id"`
	GroupID     uint   `json:"group_id"`
	Hostid      string `json:"hostid"`
	Description string `json:"description"`
	Enabled     int    `json:"enabled"`
	Status      int    `json:"status"`
	StatusDesc  string `json:"status_description"`
	IPAddr      string `json:"ip_addr"`
	Comment     string `json:"comment"`
}

// GetAllHostsServ retrieves all hosts
func GetAllHostsServ() ([]HostResp, error) {
	hosts, err := repository.GetAllHostsDAO()
	if err != nil {
		return nil, fmt.Errorf("failed to get hosts: %w", err)
	}

	result := make([]HostResp, 0, len(hosts))
	for _, h := range hosts {
		result = append(result, hostToResp(h))
	}
	return result, nil
}

// SearchHostsServ retrieves hosts by filter
func SearchHostsServ(filter model.HostFilter) ([]HostResp, error) {
	hosts, err := repository.SearchHostsDAO(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search hosts: %w", err)
	}
	result := make([]HostResp, 0, len(hosts))
	for _, h := range hosts {
		result = append(result, hostToResp(h))
	}
	return result, nil
}

// CountHostsServ returns total count for hosts by filter
func CountHostsServ(filter model.HostFilter) (int64, error) {
	return repository.CountHostsDAO(filter)
}

// GetHostByIDServ retrieves a host by ID
func GetHostByIDServ(id uint) (HostResp, error) {
	host, err := repository.GetHostByIDDAO(id)
	if err != nil {
		return HostResp{}, fmt.Errorf("failed to get host: %w", err)
	}
	return hostToResp(host), nil
}

// DeleteHostsByMIDServ deletes all hosts by monitor ID
func DeleteHostsByMIDServ(mid uint) error {
	return repository.DeleteHostByMIDDAO(mid)
}

// AddHostServ creates a new host
func AddHostServ(h HostReq) (HostResp, error) {
	newHost := model.Host{
		Name:        h.Name,
		Hostid:      h.Hostid,
		MonitorID:   h.MID,
		GroupID:     h.GroupID,
		Description: h.Description,
		Enabled:     h.Enabled,
		IPAddr:      h.IPAddr,
		Comment:     h.Comment,
	}
	if h.MID > 0 {
		if monitor, err := repository.GetMonitorByIDDAO(h.MID); err == nil {
			newHost.Status = determineHostStatus(newHost, monitor)
		}
	} else {
		newHost.Status = determineHostStatus(newHost, model.Monitor{Enabled: 1, Status: 1})
	}

	if err := repository.AddHostDAO(newHost); err != nil {
		return HostResp{}, fmt.Errorf("failed to add host: %w", err)
	}
	if newHost.MonitorID > 0 {
		if err := PushHostToMonitorServ(newHost.MonitorID, newHost.ID); err == nil {
			if refreshed, err := repository.GetHostByIDDAO(newHost.ID); err == nil {
				newHost = refreshed
			}
		}
	}
	return HostResp{
		ID:          int(newHost.ID),
		Name:        newHost.Name,
		MID:         newHost.MonitorID,
		GroupID:     newHost.GroupID,
		Hostid:      newHost.Hostid,
		Description: newHost.Description,
		Enabled:     newHost.Enabled,
		Status:      newHost.Status,
		StatusDesc:  newHost.StatusDescription,
		IPAddr:      newHost.IPAddr,
		Comment:     newHost.Comment,
	}, nil
}

// UpdateHostServ updates a host
func UpdateHostServ(id uint, h HostReq) error {
	existing, err := repository.GetHostByIDDAO(id)
	if err != nil {
		return err
	}
	monitorID := h.MID
	if monitorID == 0 {
		monitorID = existing.MonitorID
	}
	updated := model.Host{
		Name:        h.Name,
		Hostid:      h.Hostid,
		MonitorID:   monitorID,
		GroupID:     h.GroupID,
		Description: h.Description,
		Enabled:     h.Enabled,
		IPAddr:      h.IPAddr,
		Comment:     h.Comment,
	}
	if monitorID > 0 {
		if monitor, err := repository.GetMonitorByIDDAO(monitorID); err == nil {
			updated.Status = determineHostStatus(updated, monitor)
		}
	} else {
		updated.Status = determineHostStatus(updated, model.Monitor{Enabled: 1, Status: 1})
	}
	if err := repository.UpdateHostDAO(id, updated); err != nil {
		return err
	}
	if refreshed, err := repository.GetHostByIDDAO(id); err == nil {
		recordHostHistory(refreshed, time.Now().UTC())
	}
	return recomputeItemsForHost(id)
}

// DeleteHostByIDServ deletes a host by ID
func DeleteHostByIDServ(id uint) error {
	return repository.DeleteHostByIDDAO(id)
}

// GetHostsFromMonitorServ retrieves hosts from an external monitor
func GetHostsFromMonitorServ(mid uint) ([]HostResp, error) {
	monitor, err := GetMonitorByIDServ(mid)
	if err != nil {
		return nil, fmt.Errorf("failed to get monitor: %w", err)
	}

	client, err := createMonitorClient(monitor)
	if err != nil {
		return nil, fmt.Errorf("failed to create monitor client: %w", err)
	}

	// Use existing auth token if available
	if monitor.AuthToken != "" {
		client.SetAuthToken(monitor.AuthToken)
	} else {
		if err := client.Authenticate(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to authenticate with monitor: %w", err)
		}
	}

	monitorHosts, err := client.GetHosts(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get hosts from monitor: %w", err)
	}

	hosts := make([]HostResp, 0, len(monitorHosts))
	for _, h := range monitorHosts {
		activeAvailable := ""
		if h.Metadata != nil {
			if value, ok := h.Metadata["active_available"]; ok {
				activeAvailable = value
			}
		}
		status := mapMonitorHostStatus(h.Status, activeAvailable)
		hosts = append(hosts, HostResp{
			Name:        h.Name,
			Hostid:      h.ID,
			Description: h.Description,
			IPAddr:      h.IPAddress,
			Enabled:     1,
			Status:      status,
			StatusDesc:  "",
		})
	}
	return hosts, nil
}

func mapMonitorHostStatus(status string, activeAvailable string) int {
	// Check Zabbix active_available first: 0=unknown, 1=available, 2=not_available
	if activeAvailable == "2" {
		return 2
	}

	if status == "up" {
		return 1
	}
	if status == "down" {
		return 2
	}
	return 0
}

// createMonitorClient creates a monitor client from a MonitorResp
func createMonitorClient(monitor MonitorResp) (*monitors.Client, error) {
	cfg := monitors.Config{
		Name: monitor.Name,
		Type: monitors.ParseMonitorType(monitor.Type),
		Auth: monitors.AuthConfig{
			URL:      monitor.URL,
			Username: monitor.Username,
			Password: monitor.Password,
			Token:    monitor.AuthToken,
		},
		Timeout: 30,
	}
	return monitors.NewClient(cfg)
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	Added   int `json:"added"`
	Updated int `json:"updated"`
	Failed  int `json:"failed"`
	Total   int `json:"total"`
}

// hostToResp converts a domain Host to HostResp
func hostToResp(h model.Host) HostResp {
	return HostResp{
		ID:          int(h.ID),
		Name:        h.Name,
		MID:         h.MonitorID,
		GroupID:     h.GroupID,
		Hostid:      h.Hostid,
		Description: h.Description,
		Enabled:     h.Enabled,
		Status:      h.Status,
		StatusDesc:  h.StatusDescription,
		IPAddr:      h.IPAddr,
		Comment:     h.Comment,
	}
}

func PullHostsFromMonitorServ(mid uint) (SyncResult, error) {
	return pullHostsFromMonitorServ(mid, true)
}

func pullHostsFromMonitorServ(mid uint, recordHistory bool) (SyncResult, error) {
	result := SyncResult{}
	setMonitorStatusSyncing(mid)
	_, _ = TestMonitorStatusServ(mid)

	monitor, err := GetMonitorByIDServ(mid)
	if err != nil {
		setMonitorStatusError(mid)
		LogService("error", "pull hosts failed to load monitor", map[string]interface{}{"monitor_id": mid, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to get monitor: %w", err)
	}

	// Monitor must be active (status == 1 or syncing) to pull hosts
	monitorDomain, _ := repository.GetMonitorByIDDAO(mid)
	if monitorDomain.Status == 0 || monitorDomain.Status == 2 {
		reason := "monitor is not active"
		if monitorDomain.StatusDescription != "" {
			reason = monitorDomain.StatusDescription
		}
		setMonitorStatusErrorWithReason(mid, reason)

		// Mark all hosts as error with monitor down reason
		hosts, err := repository.SearchHostsDAO(model.HostFilter{MID: &mid})
		if err == nil {
			for _, host := range hosts {
				setHostStatusErrorWithReason(host.ID, reason)
				items, err := repository.GetItemsByHIDDAO(host.ID)
				if err == nil {
					for _, item := range items {
						_ = repository.UpdateItemStatusAndDescriptionDAO(item.ID, 2, reason)
					}
				}
			}
		}

		LogService("warn", "pull hosts skipped due to monitor not active", map[string]interface{}{"monitor_id": mid, "monitor_status": monitorDomain.Status, "monitor_status_description": reason}, nil, "")
		return result, fmt.Errorf("monitor is not active (status: %d)", monitorDomain.Status)
	}

	client, err := createMonitorClient(monitor)
	if err != nil {
		setMonitorStatusError(mid)
		LogService("error", "pull hosts failed to create monitor client", map[string]interface{}{"monitor_id": mid, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to create monitor client: %w", err)
	}

	if monitor.AuthToken != "" {
		client.SetAuthToken(monitor.AuthToken)
	} else {
		if err := client.Authenticate(context.Background()); err != nil {
			setMonitorStatusError(mid)
			LogService("error", "pull hosts failed to authenticate", map[string]interface{}{"monitor_id": mid, "error": err.Error()}, nil, "")
			return result, fmt.Errorf("failed to authenticate with monitor: %w", err)
		}
	}

	monitorHosts, err := client.GetHosts(context.Background())
	if err != nil {
		setMonitorStatusError(mid)
		LogService("error", "pull hosts failed to fetch hosts", map[string]interface{}{"monitor_id": mid, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to get hosts from monitor: %w", err)
	}
	monitorHostIDs := make(map[string]struct{}, len(monitorHosts))
	for _, h := range monitorHosts {
		monitorHostIDs[h.ID] = struct{}{}
	}

	result.Total = len(monitorHosts)

	for _, h := range monitorHosts {
		// Get active_available from metadata
		activeAvailable := ""
		if h.Metadata != nil {
			if value, ok := h.Metadata["active_available"]; ok {
				activeAvailable = value
			}
		}
		status := mapMonitorHostStatus(h.Status, activeAvailable)

		existingHost, err := repository.GetHostByMIDAndHostIDDAO(mid, h.ID)

		// Try to find group by groupid, defaulting to existing group ID if host exists
		var groupID uint = 0
		if err == nil {
			groupID = existingHost.GroupID
		}

		if extGroupID, ok := h.Metadata["groupid"]; ok && extGroupID != "" {
			if group, err := repository.GetGroupByExternalIDDAO(extGroupID, mid); err == nil {
				groupID = group.ID
			}
		}

		if err == nil {
			// Host exists, update it
			if err := repository.UpdateHostDAO(existingHost.ID, model.Host{
				Name:        h.Name,
				Hostid:      h.ID,
				MonitorID:   mid,
				GroupID:     groupID,
				Description: h.Description,
				Enabled:     existingHost.Enabled,
				Status:      status,
				IPAddr:      h.IPAddress,
			}); err != nil {
				setHostStatusErrorWithReason(existingHost.ID, err.Error())
				LogService("error", "pull hosts failed to update host", map[string]interface{}{"monitor_id": mid, "host_id": existingHost.ID, "error": err.Error()}, nil, "")
				result.Failed++
				continue
			}
			if recordHistory {
				if refreshed, err := repository.GetHostByIDDAO(existingHost.ID); err == nil {
					recordHostHistory(refreshed, time.Now().UTC())
				}
			}
			result.Updated++
		} else {
			// Host doesn't exist, add it
			// Try to find group by groupid for new host
			var groupID uint = 0
			if extGroupID, ok := h.Metadata["groupid"]; ok && extGroupID != "" {
				if group, err := repository.GetGroupByExternalIDDAO(extGroupID, mid); err == nil {
					groupID = group.ID
				}
			}

			if err := repository.AddHostDAO(model.Host{
				Name:        h.Name,
				Hostid:      h.ID,
				MonitorID:   mid,
				GroupID:     groupID,
				Description: h.Description,
				Enabled:     1,
				Status:      status,
				IPAddr:      h.IPAddress,
			}); err != nil {
				setMonitorStatusError(mid)
				LogService("error", "pull hosts failed to add host", map[string]interface{}{"monitor_id": mid, "host_name": h.Name, "host_external_id": h.ID, "error": err.Error()}, nil, "")
				result.Failed++
				continue
			}
			if recordHistory {
				if created, err := repository.GetHostByMIDAndHostIDDAO(mid, h.ID); err == nil {
					recordHostHistory(created, time.Now().UTC())
				}
			}
			result.Added++
		}
	}

	localHosts, err := repository.SearchHostsDAO(model.HostFilter{MID: &mid})
	if err == nil {
		for _, localHost := range localHosts {
			if _, ok := monitorHostIDs[localHost.Hostid]; ok {
				continue
			}
			reason := "host not found on monitor"
			setHostStatusErrorWithReason(localHost.ID, reason)
			items, err := repository.GetItemsByHIDDAO(localHost.ID)
			if err == nil {
				for _, item := range items {
					_ = repository.UpdateItemStatusAndDescriptionDAO(item.ID, 2, reason)
				}
			}
		}
	}
	_ = recomputeMonitorRelated(mid)
	recordNetworkStatusSnapshot(time.Now().UTC())
	SyncEvent("hosts", mid, 0, result)
	return result, nil
}

func PullHostFromMonitorServ(mid, id uint) (SyncResult, error) {
	host, err := repository.GetHostByIDDAO(id)
	if err != nil {
		setHostStatusErrorWithReason(id, err.Error())
		LogService("error", "pull host failed to load host", map[string]interface{}{"host_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get host: %w", err)
	}
	setHostStatusSyncing(id)
	if host.MonitorID != mid {
		setHostStatusErrorWithReason(id, "host does not belong to the specified monitor")
		LogService("error", "pull host failed due to monitor mismatch", map[string]interface{}{"host_id": id, "monitor_id": mid}, nil, "")
		return SyncResult{}, fmt.Errorf("host does not belong to the specified monitor")
	}

	monitor, err := GetMonitorByIDServ(mid)
	if err != nil {
		setMonitorStatusError(mid)
		setHostStatusErrorWithReason(id, err.Error())
		LogService("error", "pull host failed to load monitor", map[string]interface{}{"monitor_id": mid, "host_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get monitor: %w", err)
	}
	setMonitorStatusSyncing(mid)

	client, err := createMonitorClient(monitor)
	if err != nil {
		setMonitorStatusError(mid)
		setHostStatusErrorWithReason(id, err.Error())
		LogService("error", "pull host failed to create monitor client", map[string]interface{}{"monitor_id": mid, "host_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to create monitor client: %w", err)
	}

	if monitor.AuthToken != "" {
		client.SetAuthToken(monitor.AuthToken)
	} else {
		if err := client.Authenticate(context.Background()); err != nil {
			setMonitorStatusError(mid)
			setHostStatusErrorWithReason(id, err.Error())
			LogService("error", "pull host failed to authenticate", map[string]interface{}{"monitor_id": mid, "host_id": id, "error": err.Error()}, nil, "")
			return SyncResult{}, fmt.Errorf("failed to authenticate with monitor: %w", err)
		}
	}

	h, err := client.GetHostByID(context.Background(), host.Hostid)
	if err != nil {
		setMonitorStatusError(mid)
		setHostStatusErrorWithReason(id, err.Error())
		LogService("error", "pull host failed to fetch host", map[string]interface{}{"monitor_id": mid, "host_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get host from monitor: %w", err)
	}

	// Get active_available from metadata
	activeAvailable := ""
	if h.Metadata != nil {
		if value, ok := h.Metadata["active_available"]; ok {
			activeAvailable = value
		}
	}

	// Check if host already exists
	existingHost, err := repository.GetHostByMIDAndHostIDDAO(mid, h.ID)

	var groupID uint = 0
	if err == nil {
		groupID = existingHost.GroupID
	}

	if h.Metadata != nil {
		if extGroupID, ok := h.Metadata["groupid"]; ok && extGroupID != "" {
			if group, err := repository.GetGroupByExternalIDDAO(extGroupID, mid); err == nil {
				groupID = group.ID
			}
		}
	}

	if err == nil {
		// Host exists, update it
		if err := repository.UpdateHostDAO(existingHost.ID, model.Host{
			Name:      h.Name,
			Hostid:    h.ID,
			MonitorID: mid,
			GroupID:   groupID,
			Enabled:   existingHost.Enabled,
			Status:    mapMonitorHostStatus(h.Status, activeAvailable),
			IPAddr:    h.IPAddress,
		}); err != nil {
			setHostStatusErrorWithReason(existingHost.ID, err.Error())
			LogService("error", "pull host failed to update host", map[string]interface{}{"monitor_id": mid, "host_id": existingHost.ID, "error": err.Error()}, nil, "")
			return SyncResult{}, fmt.Errorf("failed to update host: %w", err)
		}
		if refreshed, err := repository.GetHostByIDDAO(existingHost.ID); err == nil {
			recordHostHistory(refreshed, time.Now().UTC())
		}
	} else {
		// Host doesn't exist, add it
		if err := repository.AddHostDAO(model.Host{
			Name:        h.Name,
			Hostid:      h.ID,
			MonitorID:   mid,
			GroupID:     groupID,
			Description: h.Description,
			Enabled:     1,
			Status:      mapMonitorHostStatus(h.Status, activeAvailable),
			IPAddr:      h.IPAddress,
		}); err != nil {
			setMonitorStatusError(mid)
			LogService("error", "pull host failed to add host", map[string]interface{}{"monitor_id": mid, "host_name": h.Name, "host_external_id": h.ID, "error": err.Error()}, nil, "")
			return SyncResult{}, fmt.Errorf("failed to add host: %w", err)
		}
		if created, err := repository.GetHostByMIDAndHostIDDAO(mid, h.ID); err == nil {
			recordHostHistory(created, time.Now().UTC())
		}
	}
	_ = recomputeMonitorRelated(mid)
	recordNetworkStatusSnapshot(time.Now().UTC())
	result := SyncResult{Added: 1, Total: 1}
	SyncEvent("hosts", mid, 0, result)
	return result, nil
}

// PushHostToMonitorServ pushes a host from local database to remote monitor
func PushHostToMonitorServ(mid uint, id uint) error {
	host, err := repository.GetHostByIDDAO(id)
	if err != nil {
		setHostStatusErrorWithReason(id, err.Error())
		LogService("error", "push host failed to load host", map[string]interface{}{"host_id": id, "error": err.Error()}, nil, "")
		return fmt.Errorf("failed to get host: %w", err)
	}
	setHostStatusSyncing(id)

	if host.MonitorID != mid {
		setHostStatusErrorWithReason(id, "host does not belong to the specified monitor")
		LogService("error", "push host failed due to monitor mismatch", map[string]interface{}{"host_id": id, "monitor_id": mid}, nil, "")
		return fmt.Errorf("host does not belong to the specified monitor")
	}

	monitor, err := GetMonitorByIDServ(mid)
	if err != nil {
		setMonitorStatusError(mid)
		setHostStatusErrorWithReason(id, err.Error())
		LogService("error", "push host failed to load monitor", map[string]interface{}{"monitor_id": mid, "host_id": id, "error": err.Error()}, nil, "")
		return fmt.Errorf("failed to get monitor: %w", err)
	}

	client, err := createMonitorClient(monitor)
	if err != nil {
		setMonitorStatusError(mid)
		setHostStatusErrorWithReason(id, err.Error())
		LogService("error", "push host failed to create monitor client", map[string]interface{}{"monitor_id": mid, "host_id": id, "error": err.Error()}, nil, "")
		return fmt.Errorf("failed to create monitor client: %w", err)
	}

	if monitor.AuthToken != "" {
		client.SetAuthToken(monitor.AuthToken)
	} else {
		if err := client.Authenticate(context.Background()); err != nil {
			setMonitorStatusError(mid)
			setHostStatusErrorWithReason(id, err.Error())
			LogService("error", "push host failed to authenticate", map[string]interface{}{"monitor_id": mid, "host_id": id, "error": err.Error()}, nil, "")
			return fmt.Errorf("failed to authenticate with monitor: %w", err)
		}
	}
	// Create host group (group name) then create host in monitor
	extGroupID := ""
	groupName := "Default"
	if host.GroupID > 0 {
		if group, err := repository.GetGroupByIDDAO(host.GroupID); err == nil {
			if group.Name != "" {
				groupName = group.Name
			}
			if group.ExternalID != "" && group.MonitorID == mid {
				extGroupID = group.ExternalID
			} else if group.Name != "" {
				// Fallback: Check/Create by name if ExternalID missing or mismatched
				gid, err := client.CreateHostGroup(context.Background(), group.Name)
				if err != nil {
					setMonitorStatusError(mid)
					setHostStatusErrorWithReason(id, err.Error())
					LogService("error", "push host failed to create host group", map[string]interface{}{"monitor_id": mid, "host_id": id, "group": group.Name, "error": err.Error()}, nil, "")
					return fmt.Errorf("failed to create host group: %w", err)
				}
				extGroupID = gid
				// Update group with new ExternalID
				group.ExternalID = extGroupID
				group.MonitorID = mid
				_ = repository.UpdateGroupDAO(group.ID, group)
			}
		}
	}
	// Default group if no group or group creation failed
	if extGroupID == "" {
		gid, err := client.CreateHostGroup(context.Background(), "Default")
		if err != nil {
			return fmt.Errorf("failed to create default host group: %w", err)
		}
		extGroupID = gid
	}

	monitorHost := monitors.Host{
		ID:          host.Hostid,
		Name:        host.Name,
		IPAddress:   host.IPAddr,
		Description: host.Description,
		Metadata:    map[string]string{"groupid": extGroupID},
	}
	if host.Hostid == "" {
		created, err := client.CreateHost(context.Background(), monitorHost)
		if err != nil {
			setMonitorStatusError(mid)
			setHostStatusErrorWithReason(id, err.Error())
			LogService("error", "push host failed to create host", map[string]interface{}{"monitor_id": mid, "host_id": id, "error": err.Error()}, nil, "")
			return fmt.Errorf("failed to create host in monitor: %w", err)
		}
		if created.ID != "" {
			host.Hostid = created.ID
			_ = repository.UpdateHostDAO(host.ID, host)
		}
	} else {
		if _, err := client.UpdateHost(context.Background(), monitorHost); err != nil {
			setMonitorStatusError(mid)
			setHostStatusErrorWithReason(id, err.Error())
			LogService("error", "push host failed to update host", map[string]interface{}{"monitor_id": mid, "host_id": id, "error": err.Error()}, nil, "")
			return fmt.Errorf("failed to update host in monitor: %w", err)
		}
	}
	LogService("info", "push host to monitor", map[string]interface{}{"host_name": host.Name, "host_id": host.Hostid, "monitor": monitor.Name, "group": groupName}, nil, "")
	_, _ = recomputeHostStatus(id)
	return recomputeMonitorRelated(mid)
}

// PushHostsFromMonitorServ pushes all hosts from local database to remote monitor
func PushHostsFromMonitorServ(mid uint) (SyncResult, error) {
	result := SyncResult{}
	setMonitorStatusSyncing(mid)
	_, _ = TestMonitorStatusServ(mid)

	// Check monitor status before proceeding with host push
	monitor, err := repository.GetMonitorByIDDAO(mid)
	if err != nil {
		setMonitorStatusError(mid)
		LogService("error", "push hosts failed to load monitor", map[string]interface{}{"monitor_id": mid, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to get monitor: %w", err)
	}

	// Monitor must be active (status == 1) to push hosts
	if monitor.Status == 0 || monitor.Status == 2 {
		reason := "monitor is not active"
		if monitor.StatusDescription != "" {
			reason = monitor.StatusDescription
		}
		setMonitorStatusErrorWithReason(mid, reason)
		LogService("warn", "push hosts skipped due to monitor not active", map[string]interface{}{"monitor_id": mid, "monitor_status": monitor.Status, "monitor_status_description": reason}, nil, "")
		return result, fmt.Errorf("monitor is not active (status: %d)", monitor.Status)
	}

	hosts, err := repository.SearchHostsDAO(model.HostFilter{MID: &mid})
	if err != nil {
		setMonitorStatusError(mid)
		LogService("error", "push hosts failed to load hosts", map[string]interface{}{"monitor_id": mid, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to get hosts: %w", err)
	}

	result.Total = len(hosts)

	for _, host := range hosts {
		if err := PushHostToMonitorServ(mid, host.ID); err != nil {
			setHostStatusErrorWithReason(host.ID, err.Error())
			LogService("error", "push hosts failed to push host", map[string]interface{}{"monitor_id": mid, "host_id": host.ID, "error": err.Error()}, nil, "")
			result.Failed++
			continue
		}
		result.Added++
	}

	_ = recomputeMonitorRelated(mid)
	return result, nil
}
