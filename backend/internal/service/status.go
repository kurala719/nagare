package service

import (
	"fmt"

	"nagare/internal/model"
	"nagare/internal/repository"
)

func determineMonitorStatus(m model.Monitor) int {
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
}

func determineAlarmStatus(a model.Alarm) int {
	if a.Enabled == 0 {
		return 0
	}
	if a.StatusDescription != "" {
		return 2
	}
	if a.AuthToken != "" {
		return 1
	}
	if a.Username == "" && a.Password == "" {
		return 0
	}
	return 1
}

func determineProviderStatus(p model.Provider) int {
	if p.Enabled == 0 {
		return 0
	}
	if p.APIKey == "" {
		return 2
	}
	return 1
}

func determineHostStatus(h model.Host, monitorStatus int, groupStatus int) int {
	if h.Enabled == 0 || h.Status == 0 || groupStatus == 0 {
		return 0
	}
	if h.StatusDescription != "" || h.Status == 2 || groupStatus == 2 {
		return 2
	}
	if (h.Status == 1 || h.Status == 3) && (groupStatus == 1 || groupStatus == 3) {
		return 1
	}
	return h.Status
}

func determineItemStatus(i model.Item, hostStatus int) int {
	if i.Enabled == 0 || hostStatus == 0 {
		return 0
	}
	if i.StatusDescription != "" || i.Status == 2 || hostStatus == 2 {
		return 2
	}
	if hostStatus == 1 || hostStatus == 3 {
		return 1
	}
	return i.Status
}

func determineMediaStatus(m model.Media) int {
	if m.Enabled == 0 {
		return 0
	}
	if m.Type == "" || m.Target == "" {
		return 2
	}
	return 1
}

func determineActionStatus(a model.Action, media model.Media) int {
	if a.Enabled == 0 {
		return 0
	}
	if a.MediaID == 0 || media.ID == 0 {
		return 2
	}
	if media.Status == 2 {
		return 2
	}
	return 1
}

func determineTriggerStatus(t model.Trigger) int {
	if t.Enabled == 0 {
		return 0
	}

	return 1
}

func determineGroupStatus(group model.Group, monitorStatus int) int {
	if group.Enabled == 0 || monitorStatus == 0 {
		return 0
	}
	if group.StatusDescription != "" || group.Status == 2 || monitorStatus == 2 {
		return 2
	}
	if monitorStatus == 1 || monitorStatus == 3 {
		return 1
	}
	return 0
}

func setMonitorStatusSyncing(mid uint) {
	_ = repository.UpdateMonitorStatusDAO(mid, 3)
}

func setMonitorStatusError(mid uint) {
	_ = repository.UpdateMonitorStatusDAO(mid, 2)
}

func setMonitorStatusErrorWithReason(mid uint, reason string) {
	if reason == "" {
		setMonitorStatusError(mid)
		return
	}
	_ = repository.UpdateMonitorStatusAndDescriptionDAO(mid, 2, reason)
	_ = recomputeMonitorRelated(mid)
}

func setMonitorRelatedError(mid uint, reason string) {
	if reason == "" {
		reason = "monitor is down"
	}
	setMonitorStatusErrorWithReason(mid, reason)
	hosts, err := repository.SearchHostsDAO(model.HostFilter{MID: &mid})
	if err != nil {
		return
	}
	for _, host := range hosts {
		setHostStatusErrorWithReason(host.ID, reason)
		items, err := repository.GetItemsByHIDDAO(host.ID)
		if err != nil {
			continue
		}
		for _, item := range items {
			_ = repository.UpdateItemStatusAndDescriptionDAO(item.ID, 2, reason)
		}
	}

	// Propagate error to groups
	groups, err := repository.SearchGroupsDAO(model.GroupFilter{MonitorID: &mid})
	if err == nil {
		for _, group := range groups {
			_, _ = recomputeGroupStatus(group.ID)
		}
	}
}

func setHostStatusSyncing(hid uint) {
	_ = repository.UpdateHostStatusDAO(hid, 3)
}

func setHostStatusError(hid uint) {
	_ = repository.UpdateHostStatusDAO(hid, 2)
}

func setHostStatusErrorWithReason(hid uint, reason string) {
	if reason == "" {
		setHostStatusError(hid)
		return
	}
	_ = repository.UpdateHostStatusAndDescriptionDAO(hid, 2, reason)
}

func setItemStatusSyncing(id uint) {
	_ = repository.UpdateItemStatusDAO(id, 3)
}

func setItemStatusError(id uint) {
	_ = repository.UpdateItemStatusDAO(id, 2)
}

func setItemStatusErrorWithReason(id uint, reason string) {
	if reason == "" {
		setItemStatusError(id)
		return
	}
	_ = repository.UpdateItemStatusAndDescriptionDAO(id, 2, reason)
}

func setProviderStatusError(pid uint) {
	_ = repository.UpdateProviderStatusDAO(pid, 2)
}

func setGroupStatusError(gid uint) {
	_ = repository.UpdateGroupStatusDAO(gid, 2)
}

func recomputeMonitorStatus(mid uint) (int, error) {
	monitor, err := repository.GetMonitorByIDDAO(mid)
	if err != nil {
		return 0, err
	}
	status := determineMonitorStatus(monitor)
	if status == 2 {
		_ = repository.UpdateMonitorStatusDAO(mid, status)
	} else {
		_ = repository.UpdateMonitorStatusAndDescriptionDAO(mid, status, "")
	}

	// Compute health score based on hosts
	hosts, err := repository.SearchHostsDAO(model.HostFilter{MID: &mid})

	score := 100
	if err == nil && len(hosts) > 0 {
		sum := 0
		count := 0
		for _, h := range hosts {
			if h.Enabled != 0 {
				sum += h.HealthScore
				count++
			}
		}
		if count > 0 {
			score = sum / count
		}
	} else if monitor.Enabled == 0 {
		score = 0
	} else if status == 2 {
		score = 0
	} else if status == 3 {
		score = 50
	}
	_ = repository.UpdateMonitorHealthScoreDAO(mid, score)

	return status, nil
}

func recomputeProviderStatus(pid uint) (int, error) {
	provider, err := repository.GetProviderByIDDAO(pid)
	if err != nil {
		return 0, err
	}
	status := determineProviderStatus(provider)
	if err := repository.UpdateProviderStatusDAO(pid, status); err != nil {
		return status, err
	}
	return status, nil
}

func recomputeGroupStatus(gid uint) (int, error) {
	group, err := repository.GetGroupByIDDAO(gid)
	if err != nil {
		return 0, err
	}

	monitorStatus := 1
	if group.MonitorID > 0 {
		if m, err := repository.GetMonitorByIDDAO(group.MonitorID); err == nil {
			monitorStatus = m.Status
		}
	}

	status := determineGroupStatus(group, monitorStatus)
	_ = repository.UpdateGroupStatusAndDescriptionDAO(gid, status, group.StatusDescription)

	// Propagate status change to hosts in this group
	hostsInGroup, err := repository.SearchHostsDAO(model.HostFilter{GroupID: &gid})
	if err == nil {
		for _, h := range hostsInGroup {
			_, _ = recomputeHostStatus(h.ID)
		}
	}

	// Compute health score based on hosts
	hosts, err := repository.SearchHostsDAO(model.HostFilter{GroupID: &gid})
	score := 100
	if len(hosts) > 0 {
		sum := 0
		count := 0
		for _, h := range hosts {
			if h.Enabled != 0 {
				sum += h.HealthScore
				count++
			}
		}
		if count > 0 {
			score = sum / count
		}
	} else if group.Enabled == 0 {
		score = 0
	}
	_ = repository.UpdateGroupHealthScoreDAO(gid, score)

	return status, nil
}

func recomputeHostStatus(hid uint) (int, error) {
	host, err := repository.GetHostByIDDAO(hid)
	if err != nil {
		return 0, err
	}

	groupStatus := 1
	monitorStatus := 1
	if host.GroupID > 0 {
		if g, err := repository.GetGroupByIDDAO(host.GroupID); err == nil {
			groupStatus = g.Status
			if m, err := repository.GetMonitorByIDDAO(g.MonitorID); err == nil {
				monitorStatus = m.Status
			}
		}
	}

	status := determineHostStatus(host, monitorStatus, groupStatus)
	_ = repository.UpdateHostStatusAndDescriptionDAO(hid, status, host.StatusDescription)

	// Compute health score based on status
	score := 100
	if host.Enabled == 0 {
		score = 0
	} else {
		switch status {
		case 0: // Inactive/Unknown
			score = 0
		case 2: // Error
			score = 0
		case 3: // Syncing
			score = 50
		case 1: // Active
			score = 100
		}
	}
	_ = repository.UpdateHostHealthScoreDAO(hid, score)

	return status, nil
}

func recomputeItemStatus(id uint) (int, error) {
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
}

func recomputeItemsForHost(hid uint) error {
	items, err := repository.GetItemsByHIDDAO(hid)
	if err != nil {
		return err
	}
	for _, item := range items {
		if _, err := recomputeItemStatus(item.ID); err != nil {
			return fmt.Errorf("failed to recompute item %d: %w", item.ID, err)
		}
	}
	return nil
}

func recomputeMonitorRelated(mid uint) error {
	hosts, err := repository.SearchHostsDAO(model.HostFilter{MID: &mid})
	if err != nil {
		return err
	}

	// 1. Recompute Groups first (hosts depend on groups)
	groups, err := repository.SearchGroupsDAO(model.GroupFilter{MonitorID: &mid})
	if err == nil {
		for _, group := range groups {
			_, _ = recomputeGroupStatus(group.ID)
		}
	}
	// Catch any groups not directly under monitor but used by its hosts
	for _, host := range hosts {
		if host.GroupID > 0 {
			_, _ = recomputeGroupStatus(host.GroupID)
		}
	}

	// 2. Recompute Items and Hosts (now that groups are updated)
	for _, host := range hosts {
		_ = recomputeItemsForHost(host.ID)
		_, _ = recomputeHostStatus(host.ID)
	}

	// 3. Recompute Monitor
	_, _ = recomputeMonitorStatus(mid)

	return nil
}

func recomputeMediaStatus(id uint) (int, error) {
	media, err := repository.GetMediaByIDDAO(id)
	if err != nil {
		return 0, err
	}
	status := determineMediaStatus(media)
	if err := repository.UpdateMediaStatusDAO(id, status); err != nil {
		return status, err
	}
	return status, nil
}

func recomputeActionStatus(id uint) (int, error) {
	action, err := repository.GetActionByIDDAO(id)
	if err != nil {
		return 0, err
	}
	media, err := repository.GetMediaByIDDAO(action.MediaID)
	if err != nil {
		status := determineActionStatus(action, model.Media{})
		return status, repository.UpdateActionStatusDAO(id, status)
	}
	status := determineActionStatus(action, media)
	if err := repository.UpdateActionStatusDAO(id, status); err != nil {
		return status, err
	}
	return status, nil
}

func recomputeTriggerStatus(id uint) (int, error) {
	trigger, err := repository.GetTriggerByIDDAO(id)
	if err != nil {
		return 0, err
	}
	status := determineTriggerStatus(trigger)
	if err := repository.UpdateTriggerStatusDAO(id, status); err != nil {
		return status, err
	}
	return status, nil
}

// RecomputeAllStatuses refreshes stored status values for all entities.
func RecomputeAllStatuses() error {
	monitorsList, err := repository.GetAllMonitorsDAO()
	if err == nil {
		for _, m := range monitorsList {
			_, _ = recomputeMonitorStatus(m.ID)
		}
	}

	groups, err := repository.GetAllGroupsDAO()
	if err == nil {
		for _, g := range groups {
			_, _ = recomputeGroupStatus(g.ID)
		}
	}

	hosts, err := repository.SearchHostsDAO(model.HostFilter{})
	if err == nil {
		for _, h := range hosts {
			_, _ = recomputeHostStatus(h.ID)
			_ = recomputeItemsForHost(h.ID)
		}
	}

	actions, err := repository.GetAllActionsDAO()
	if err != nil {
		return err
	}
	for _, action := range actions {
		if _, err := recomputeActionStatus(action.ID); err != nil {
			return err
		}
	}
	triggers, err := repository.GetAllTriggersDAO()
	if err != nil {
		return err
	}
	for _, trigger := range triggers {
		if _, err := recomputeTriggerStatus(trigger.ID); err != nil {
			return err
		}
	}
	return nil
}

// RecomputeActionAndTriggerStatuses refreshes stored status values for actions and triggers.
func RecomputeActionAndTriggerStatuses() error {
	return RecomputeAllStatuses()
}
