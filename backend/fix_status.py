import os

filepath = r'd:\\Nagare_Project\\nagare\\backend\\internal\\service\\status.go'
with open(filepath, 'r', encoding='utf-8') as f:
    content = f.read()

# Replace determineMonitorStatus
old_monitor = """func determineMonitorStatus(m model.Monitor) int {
	if m.Enabled == 0 {
		return 0
	}
	// Explicit error description overrides optimistic status
	if m.StatusDescription != "" {
		return 2
	}
	// SNMP monitors are active if enabled, as polling is per-host
	if m.Type == 1 { // SNMP
		return 1
	}
	// Monitor status is independent - based on its own connectivity/token
	if m.AuthToken != "" {
		return 1
	}
	if m.Username == "" && m.Password == "" {
		return 0
	}
	return 1
}"""

new_monitor = """func determineMonitorStatus(m model.Monitor) int {
	if m.Status == 3 {
		return 3
	}
	if m.Enabled == 0 {
		return 0
	}
	if m.StatusDescription != "" || m.Status == 2 {
		return 2
	}
	if m.AuthToken != "" || m.Type == 1 {
		return 1
	}
	return 0
}"""
content = content.replace(old_monitor, new_monitor)

# Replace determineHostStatus
old_host = """func determineHostStatus(h model.Host, monitorStatus int, groupStatus int) int {
	if h.Enabled == 0 {
		return 0
	}

	// If the monitor is in error, the host should be too
	if monitorStatus == 2 {
		return 2
	}

	// If the monitor is in error, the host MUST be in error
	if monitorStatus == 2 {
		return 2
	}

	// If the host has an explicit error description, keep it in error state.
	// This ensures that Nagare's internal polling failures (like auth or timeout)
	// are not overridden by stale 'available' data from Zabbix.
	// We exclude generic "host is not active" to allow recovery during sync.
	if h.StatusDescription != "" && h.StatusDescription != "host is not active" {
		return 2
	}

	// If the group is in error, the host should be too
	if groupStatus == 2 {
		return 2
	}

	// If monitor or group is inactive, host is inactive
	if monitorStatus == 0 || groupStatus == 0 {
		return 0
	}

	// Without ActiveAvailable or MonitorID on the Host, we default to Active (1)
	// when there are no errors from the Monitor or Group.
	return 1
}"""

new_host = """func determineHostStatus(h model.Host, monitorStatus int, groupStatus int) int {
	if h.Status == 3 {
		return 3
	}
	if h.Enabled == 0 || h.Status == 0 || groupStatus == 0 {
		return 0
	}
	if h.StatusDescription != "" || h.Status == 2 || groupStatus == 2 {
		return 2
	}
	if h.Status == 1 && groupStatus == 1 {
		return 1
	}
	return h.Status
}"""
content = content.replace(old_host, new_host)

# Replace determineItemStatus
old_item = """func determineItemStatus(i model.Item) int {
	if i.Enabled == 0 {
		return 0
	}
	// Item status is independent of host's overall health status
	if i.ExternalID == "" {
		return 2
	}
	// 'N/A' indicates a polling failure or missing instance for this specific device
	if i.LastValue == "" || i.LastValue == "N/A" {
		return 2
	}
	return 1
}"""

new_item = """func determineItemStatus(i model.Item, hostStatus int) int {
	if i.Status == 3 {
		return 3
	}
	if i.Enabled == 0 || hostStatus == 0 {
		return 0
	}
	if i.StatusDescription != "" || i.Status == 2 || hostStatus == 2 {
		return 2
	}
	if hostStatus == 1 {
		return 1
	}
	return i.Status
}"""
content = content.replace(old_item, new_item)

# Replace determineGroupStatus
old_group = """func determineGroupStatus(group model.Group, monitorStatus int) int {
	if group.Enabled == 0 {
		return 0
	}

	// If the monitor is in error, the group should be too
	if monitorStatus == 2 {
		return 2
	}

	// If monitor is inactive, group is inactive
	if monitorStatus == 0 {
		return 0
	}

	// Default: Enabled and no specific error, mark as Active
	return 1
}"""

new_group = """func determineGroupStatus(group model.Group, monitorStatus int) int {
	if group.Status == 3 {
		return 3
	}
	if group.Enabled == 0 || monitorStatus == 0 {
		return 0
	}
	if group.StatusDescription != "" || group.Status == 2 || monitorStatus == 2 {
		return 2
	}
	if monitorStatus == 1 {
		return 1
	}
	return 0
}"""
content = content.replace(old_group, new_group)

# Fix recomputeItemStatus calls
old_recompute = """func recomputeItemStatus(id uint) (int, error) {
	item, err := repository.GetItemByIDDAO(id)
	if err != nil {
		return 0, err
	}
	_, err = repository.GetHostByIDDAO(item.HostID)
	if err != nil {
		status := determineItemStatus(item)
		if status == 2 {
			return status, repository.UpdateItemStatusDAO(id, status)
		}
		return status, repository.UpdateItemStatusAndDescriptionDAO(id, status, "")
	}
	status := determineItemStatus(item)
	if status == 2 {
		if err := repository.UpdateItemStatusDAO(id, status); err != nil {
			return status, err
		}
	} else {
		if err := repository.UpdateItemStatusAndDescriptionDAO(id, status, ""); err != nil {
			return status, err
		}
	}
	return status, nil
}"""

new_recompute = """func recomputeItemStatus(id uint) (int, error) {
	item, err := repository.GetItemByIDDAO(id)
	if err != nil {
		return 0, err
	}
	hostStatus := 0
	if host, err := repository.GetHostByIDDAO(item.HostID); err == nil {
		hostStatus = host.Status
	}
	
	status := determineItemStatus(item, hostStatus)
	if status == 2 {
		if err := repository.UpdateItemStatusDAO(id, status); err != nil {
			return status, err
		}
	} else {
		if err := repository.UpdateItemStatusAndDescriptionDAO(id, status, ""); err != nil {
			return status, err
		}
	}
	return status, nil
}"""
content = content.replace(old_recompute, new_recompute)

with open(filepath, 'w', encoding='utf-8') as f:
    f.write(content)

print("Python script executed.")
