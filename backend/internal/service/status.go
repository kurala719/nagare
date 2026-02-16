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
	if m.AuthToken != "" {
		return 1
	}
	if m.Username == "" && m.Password == "" {
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

func determineHostStatus(h model.Host, monitor model.Monitor) int {
	if h.Enabled == 0 {
		return 0
	}
	if monitor.Enabled == 0 {
		return 2
	}
	if monitor.Status == 2 {
		return 2
	}
	if h.Status == 2 {
		return 2
	}
	if monitor.Status == 3 {
		return 1
	}
	if h.Hostid == "" {
		return 2
	}
	if monitor.Status == 1 {
		return 1
	}
	return 0
}

func determineItemStatus(i model.Item, host model.Host) int {
	if i.Enabled == 0 {
		return 0
	}
	if host.Enabled == 0 {
		return 0
	}
	// Item status depends on host status - propagate host errors and active state
	if host.Status == 2 {
		return 2
	}
	if host.Status == 0 {
		return 0
	}
	if i.ItemID == "" {
		return 2
	}
	if i.LastValue == "" {
		return 0
	}
	return 1
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

func determineMediaTypeStatus(mt model.MediaType) int {
	if mt.Enabled == 0 {
		return 0
	}
	if mt.Name == "" || mt.Key == "" {
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

func determineTriggerStatus(t model.Trigger, action model.Action) int {
	if t.Enabled == 0 {
		return 0
	}
	if t.Entity != "" && t.Entity != "alert" && t.Entity != "log" {
		return 2
	}
	if t.ActionID == 0 || action.ID == 0 {
		return 2
	}
	if action.Status == 2 {
		return 2
	}
	return 1
}

func determineSiteStatus(site model.Site, hosts []model.Host) int {
	if site.Enabled == 0 {
		return 0
	}
	hasActive := false
	hasSyncing := false
	for _, host := range hosts {
		switch host.Status {
		case 2:
			return 2
		case 3:
			hasSyncing = true
		case 1:
			hasActive = true
		}
	}
	if hasSyncing {
		return 3
	}
	if hasActive {
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

func setSiteStatusError(sid uint) {
	_ = repository.UpdateSiteStatusDAO(sid, 2)
}

func recomputeMonitorStatus(mid uint) (int, error) {
	monitor, err := repository.GetMonitorByIDDAO(mid)
	if err != nil {
		return 0, err
	}
	status := determineMonitorStatus(monitor)
	if status == 2 {
		if err := repository.UpdateMonitorStatusDAO(mid, status); err != nil {
			return status, err
		}
	} else {
		if err := repository.UpdateMonitorStatusAndDescriptionDAO(mid, status, ""); err != nil {
			return status, err
		}
	}
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

func recomputeSiteStatus(sid uint) (int, error) {
	site, err := repository.GetSiteByIDDAO(sid)
	if err != nil {
		return 0, err
	}
	hosts, err := repository.SearchHostsDAO(model.HostFilter{SiteID: &sid})
	if err != nil {
		return 0, err
	}
	status := determineSiteStatus(site, hosts)
	if err := repository.UpdateSiteStatusDAO(sid, status); err != nil {
		return status, err
	}
	return status, nil
}

func recomputeHostStatus(hid uint) (int, error) {
	host, err := repository.GetHostByIDDAO(hid)
	if err != nil {
		return 0, err
	}
	monitor, err := repository.GetMonitorByIDDAO(host.MonitorID)
	if err != nil {
		status := determineHostStatus(host, model.Monitor{Enabled: 0, Status: 2})
		if status == 2 {
			return status, repository.UpdateHostStatusDAO(hid, status)
		}
		return status, repository.UpdateHostStatusAndDescriptionDAO(hid, status, "")
	}
	status := determineHostStatus(host, monitor)
	if status == 2 {
		if err := repository.UpdateHostStatusDAO(hid, status); err != nil {
			return status, err
		}
	} else {
		if err := repository.UpdateHostStatusAndDescriptionDAO(hid, status, ""); err != nil {
			return status, err
		}
	}
	return status, nil
}

func recomputeItemStatus(id uint) (int, error) {
	item, err := repository.GetItemByIDDAO(id)
	if err != nil {
		return 0, err
	}
	host, err := repository.GetHostByIDDAO(item.HID)
	if err != nil {
		status := determineItemStatus(item, model.Host{Enabled: 0, Status: 2})
		if status == 2 {
			return status, repository.UpdateItemStatusDAO(id, status)
		}
		return status, repository.UpdateItemStatusAndDescriptionDAO(id, status, "")
	}
	status := determineItemStatus(item, host)
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
	if _, err := recomputeMonitorStatus(mid); err != nil {
		return err
	}
	hosts, err := repository.SearchHostsDAO(model.HostFilter{MID: &mid})
	if err != nil {
		return err
	}
	for _, host := range hosts {
		if _, err := recomputeHostStatus(host.ID); err != nil {
			return err
		}
		if err := recomputeItemsForHost(host.ID); err != nil {
			return err
		}
	}
	return nil
}

func recomputeMediaStatus(id uint) (int, error) {
	media, err := repository.GetMediaByIDDAO(id)
	if err != nil {
		return 0, err
	}
	if media.MediaTypeID > 0 {
		if mediaType, err := repository.GetMediaTypeByIDDAO(media.MediaTypeID); err == nil {
			if mediaType.Status == 2 || mediaType.Enabled == 0 {
				status := 2
				if media.Enabled == 0 {
					status = 0
				}
				if err := repository.UpdateMediaStatusDAO(id, status); err != nil {
					return status, err
				}
				return status, nil
			}
		}
	}
	status := determineMediaStatus(media)
	if err := repository.UpdateMediaStatusDAO(id, status); err != nil {
		return status, err
	}
	return status, nil
}

func recomputeMediaTypeStatus(id uint) (int, error) {
	mediaType, err := repository.GetMediaTypeByIDDAO(id)
	if err != nil {
		return 0, err
	}
	status := determineMediaTypeStatus(mediaType)
	if err := repository.UpdateMediaTypeStatusDAO(id, status); err != nil {
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
	action, err := repository.GetActionByIDDAO(trigger.ActionID)
	if err != nil {
		status := determineTriggerStatus(trigger, model.Action{})
		return status, repository.UpdateTriggerStatusDAO(id, status)
	}
	status := determineTriggerStatus(trigger, action)
	if err := repository.UpdateTriggerStatusDAO(id, status); err != nil {
		return status, err
	}
	return status, nil
}

// RecomputeActionAndTriggerStatuses refreshes stored status values for actions and triggers.
func RecomputeActionAndTriggerStatuses() error {
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
