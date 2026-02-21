package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nagare/internal/model"
	"nagare/internal/repository"
	"nagare/internal/repository/monitors"
	"nagare/internal/service/utils"
)

// ItemReq represents an item request
type ItemReq struct {
	Name           string     `json:"name"`
	HID            uint       `json:"hid"`
	Value          string     `json:"value"`
	ValueType      string     `json:"value_type"`
	Type           string     `json:"type"`
	Enabled        int        `json:"enabled"`
	ItemID         string     `json:"itemid"`
	ExternalHostID string     `json:"hostid"`
	Units          string     `json:"units"`
	Comment        string     `json:"comment"`
	LastSyncAt     *time.Time `json:"last_sync_at,omitempty"`
	ExternalSource string     `json:"external_source,omitempty"`
}

// ItemResp represents an item response
type ItemResp struct {
	ID             uint       `json:"id"`
	Name           string     `json:"name"`
	HID            uint       `json:"hid"`
	Value          string     `json:"value"`
	Units          string     `json:"units"`
	Enabled        int        `json:"enabled"`
	Status         int        `json:"status"`
	StatusDesc     string     `json:"status_description"`
	Comment        string     `json:"comment"`
	LastSyncAt     *time.Time `json:"last_sync_at"`
	ExternalSource string     `json:"external_source"`
}

// SyncResult is defined in host.go - using the same type here

// GetAllItemServ retrieves all items
func GetAllItemServ() ([]ItemResp, error) {
	items, err := repository.GetItems()
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	result := make([]ItemResp, 0, len(items))
	for _, item := range items {
		result = append(result, itemToResp(item))
	}
	return result, nil
}

// SearchItemsServ retrieves items by filter
func SearchItemsServ(filter model.ItemFilter) ([]ItemResp, error) {
	items, err := repository.SearchItemsDAO(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search items: %w", err)
	}
	result := make([]ItemResp, 0, len(items))
	for _, item := range items {
		result = append(result, itemToResp(item))
	}
	return result, nil
}

// CountItemsServ returns total count for items by filter
func CountItemsServ(filter model.ItemFilter) (int64, error) {
	return repository.CountItemsDAO(filter)
}

// GetItemByIDServ retrieves an item by ID
func GetItemByIDServ(id uint) (ItemResp, error) {
	item, err := repository.GetItemByIDDAO(id)
	if err != nil {
		return ItemResp{}, fmt.Errorf("failed to get item: %w", err)
	}
	return itemToResp(item), nil
}

// AddItemServ creates a new item
func AddItemServ(req ItemReq) (ItemResp, error) {
	hid := req.HID
	if hid == 0 {
		hid = req.HID
	}
	valueType := req.ValueType
	if valueType == "" {
		valueType = req.Type
	}

	item := model.Item{
		Name:           req.Name,
		HID:            hid,
		ItemID:         req.ItemID,
		ExternalHostID: req.ExternalHostID,
		ValueType:      valueType,
		LastValue:      req.Value,
		Units:          req.Units,
		Enabled:        req.Enabled,
		Comment:        req.Comment,
	}
	var host model.Host
	if loadedHost, err := repository.GetHostByIDDAO(hid); err == nil {
		host = loadedHost
		item.Status = determineItemStatus(item)
	} else {
		item.Status = determineItemStatus(item)
	}

	if err := repository.AddItemDAO(item); err != nil {
		return ItemResp{}, fmt.Errorf("failed to add item: %w", err)
	}
	if host.ID > 0 && host.MonitorID > 0 {
		if err := PushItemToMonitorServ(host.MonitorID, host.ID, item.ID); err == nil {
			if refreshed, err := repository.GetItemByIDDAO(item.ID); err == nil {
				item = refreshed
			}
		}
	}

	return itemToResp(item), nil
}

// UpdateItemServ updates an existing item
func UpdateItemServ(id uint, req ItemReq) error {
	existing, err := repository.GetItemByIDDAO(id)
	if err != nil {
		return err
	}

	hid := req.HID
	if hid == 0 {
		hid = existing.HID
	}
	valueType := req.ValueType
	if valueType == "" {
		valueType = req.ValueType
	}
	updated := model.Item{
		Name:              req.Name,
		HID:               hid,
		ItemID:            req.ItemID,
		ExternalHostID:    req.ExternalHostID,
		ValueType:         valueType,
		LastValue:         req.Value,
		Units:             req.Units,
		Enabled:           req.Enabled,
		Comment:           req.Comment,
		LastSyncAt:        existing.LastSyncAt,
		ExternalSource:    existing.ExternalSource,
		Status:            existing.Status,
		StatusDescription: existing.StatusDescription,
	}
	if req.LastSyncAt != nil {
		updated.LastSyncAt = req.LastSyncAt
	}
	if req.ExternalSource != "" {
		updated.ExternalSource = req.ExternalSource
	}
	// Preserve status and description unless enabled state changed
	if req.Enabled != existing.Enabled {
		updated.Status = determineItemStatus(updated)
		updated.StatusDescription = ""
	}
	if err := repository.UpdateItemDAO(id, updated); err != nil {
		return err
	}
	if refreshed, err := repository.GetItemByIDDAO(id); err == nil {
		recordItemHistory(refreshed, time.Now().UTC())
		ExecuteTriggersForItem(refreshed)
	}
	_, _ = recomputeItemStatus(id)
	return nil
}

// DeleteItemByIDServ deletes an item by ID
func DeleteItemByIDServ(id uint) error {
	return repository.DeleteItemByIDDAO(id)
}

// GetItemsByHostIDFromMonitorServ retrieves items from an external monitor for a host
func GetItemsByHostIDFromMonitorServ(hid uint) ([]ItemResp, error) {
	host, err := repository.GetHostByIDDAO(hid)
	if err != nil {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	monitor, err := repository.GetMonitorByIDDAO(host.MonitorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get monitor: %w", err)
	}

	client, err := createMonitorClientFromDomain(monitor)
	if err != nil {
		return nil, fmt.Errorf("failed to create monitor client: %w", err)
	}

	// Use existing auth token if available
	if monitor.AuthToken != "" {
		client.SetAuthToken(monitor.AuthToken)
	} else if monitor.ID != 1 {
		if err := client.Authenticate(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to authenticate with monitor: %w", err)
		}
	}

	monitorItems, err := client.GetItems(context.Background(), host.Hostid)
	if err != nil {
		return nil, fmt.Errorf("failed to get items from monitor: %w", err)
	}

	result := make([]ItemResp, 0, len(monitorItems))
	for _, item := range monitorItems {
		enabled, status := mapMonitorItemStatus(item.Status)
		result = append(result, ItemResp{
			Name:       item.Name,
			HID:        hid,
			Value:      utils.ParseItemValue(item.Value, item.Units),
			Enabled:    enabled,
			Status:     status,
			StatusDesc: "",
		})
	}
	return result, nil
}

// AddItemsByHostIDFromMonitorServ adds items from an external monitor to the database
func AddItemsByHostIDFromMonitorServ(hid uint) error {
	host, err := repository.GetHostByIDDAO(hid)
	if err != nil {
		setHostStatusError(hid)
		LogService("error", "import items failed to load host", map[string]interface{}{"host_id": hid, "error": err.Error()}, nil, "")
		return fmt.Errorf("failed to get host: %w", err)
	}

	monitor, err := repository.GetMonitorByIDDAO(host.MonitorID)
	if err != nil {
		setMonitorStatusError(host.MonitorID)
		setHostStatusError(hid)
		LogService("error", "import items failed to load monitor", map[string]interface{}{"monitor_id": host.MonitorID, "host_id": hid, "error": err.Error()}, nil, "")
		return fmt.Errorf("failed to get monitor: %w", err)
	}

	client, err := createMonitorClientFromDomain(monitor)
	if err != nil {
		setMonitorStatusError(host.MonitorID)
		setHostStatusError(hid)
		LogService("error", "import items failed to create monitor client", map[string]interface{}{"monitor_id": host.MonitorID, "host_id": hid, "error": err.Error()}, nil, "")
		return fmt.Errorf("failed to create monitor client: %w", err)
	}

	// Use existing auth token if available
	if monitor.AuthToken != "" {
		client.SetAuthToken(monitor.AuthToken)
	} else if monitor.ID != 1 {
		if err := client.Authenticate(context.Background()); err != nil {
			setMonitorStatusError(host.MonitorID)
			setHostStatusError(hid)
			LogService("error", "import items failed to authenticate", map[string]interface{}{"monitor_id": host.MonitorID, "host_id": hid, "error": err.Error()}, nil, "")
			return fmt.Errorf("failed to authenticate with monitor: %w", err)
		}
	}

	monitorItems, err := client.GetItems(context.Background(), host.Hostid)
	if err != nil {
		setMonitorStatusError(host.MonitorID)
		setHostStatusError(hid)
		LogService("error", "import items failed to fetch items", map[string]interface{}{"monitor_id": host.MonitorID, "host_id": hid, "error": err.Error()}, nil, "")
		return fmt.Errorf("failed to get items from monitor: %w", err)
	}

	for _, item := range monitorItems {
		enabled, status := mapMonitorItemStatus(item.Status)
		if err := repository.AddItemDAO(model.Item{
			Name:           item.Name,
			HID:            hid,
			ItemID:         item.ID,
			ExternalHostID: item.HostID,
			ValueType:      item.ValueType,
			LastValue:      item.Value,
			Units:          item.Units,
			Enabled:        enabled,
			Status:         status,
		}); err != nil {
			setHostStatusError(hid)
			LogService("error", "import items failed to add item", map[string]interface{}{"monitor_id": host.MonitorID, "host_id": hid, "item_name": item.Name, "error": err.Error()}, nil, "")
			return fmt.Errorf("failed to add item %s: %w", item.Name, err)
		}
	}
	return nil
}

// createMonitorClientFromDomain creates a monitor client from a model.Monitor
func createMonitorClientFromDomain(monitor model.Monitor) (*monitors.Client, error) {
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

// GetItemsByHIDServ retrieves all items for a specific host
func GetItemsByHIDServ(hid uint) ([]ItemResp, error) {
	items, err := repository.GetItemsByHIDDAO(hid)
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	result := make([]ItemResp, 0, len(items))
	for _, item := range items {
		result = append(result, itemToResp(item))
	}
	return result, nil
}

// ConsultItemByIDServ consults AI about an item's data
func ConsultItemByIDServ(pid, id uint) (string, error) {
	item, err := repository.GetItemByIDDAO(id)
	if err != nil {
		return "", fmt.Errorf("failed to get item: %w", err)
	}

	provider, err := repository.GetProviderByIDDAO(pid)
	if err != nil {
		return "", fmt.Errorf("failed to get provider: %w", err)
	}

	content := fmt.Sprintf("The item name is %s, the last value is %s %s. Please help me analyze the meaning of this data.",
		item.Name, item.LastValue, item.Units)

	resp, err := SendChatServ(ChatReq{
		ProviderID: pid,
		Model:      provider.DefaultModel,
		Content:    content,
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// itemToResp converts a domain Item to ItemResp
func itemToResp(item model.Item) ItemResp {
	return ItemResp{
		ID:             item.ID,
		Name:           item.Name,
		HID:            item.HID,
		Value:          item.LastValue,
		Units:          item.Units,
		Enabled:        item.Enabled,
		Status:         item.Status,
		StatusDesc:     item.StatusDescription,
		Comment:        item.Comment,
		LastSyncAt:     item.LastSyncAt,
		ExternalSource: item.ExternalSource,
	}
}

func PullItemsFromMonitorServ(mid uint) (SyncResult, error) {
	return pullItemsFromMonitorServ(mid, true)
}

func pullItemsFromMonitorServ(mid uint, recordHistory bool) (SyncResult, error) {
	result := SyncResult{}
	setMonitorStatusSyncing(mid)
	if mid != 1 {
		_, _ = TestMonitorStatusServ(mid)
	}

	// Check monitor status before proceeding with host pull
	monitor, err := repository.GetMonitorByIDDAO(mid)
	if err != nil {
		setMonitorStatusError(mid)
		LogService("error", "pull items failed to load monitor", map[string]interface{}{"monitor_id": mid, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to get monitor: %w", err)
	}

	// Monitor must be active (status == 1 or syncing) to pull items
	if mid != 1 && (monitor.Status == 0 || monitor.Status == 2) {
		reason := "monitor is not active"
		if monitor.StatusDescription != "" {
			reason = monitor.StatusDescription
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
		LogService("warn", "pull items skipped due to monitor not active", map[string]interface{}{"monitor_id": mid, "monitor_status": monitor.Status, "monitor_status_description": reason}, nil, "")
		return result, fmt.Errorf("monitor is not active (status: %d)", monitor.Status)
	}

	hosts, err := repository.SearchHostsDAO(model.HostFilter{MID: &mid})
	if err != nil {
		setMonitorStatusError(mid)
		LogService("error", "pull items failed to load hosts", map[string]interface{}{"monitor_id": mid, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to get hosts: %w", err)
	}

	for _, host := range hosts {
		hostResult, err := pullItemsFromHostServ(mid, host.ID, recordHistory)
		if err != nil {
			setHostStatusErrorWithReason(host.ID, err.Error())
			LogService("error", "pull items failed for host", map[string]interface{}{"monitor_id": mid, "host_id": host.ID, "error": err.Error()}, nil, "")
			// Continue with other hosts
			continue
		}
		result.Added += hostResult.Added
		result.Updated += hostResult.Updated
		result.Failed += hostResult.Failed
		result.Total += hostResult.Total
	}

	_ = recomputeMonitorRelated(mid)
	recordNetworkStatusSnapshot(time.Now().UTC())
	SyncEvent("items", mid, 0, result)
	return result, nil
}

func PullItemsFromHostServ(mid, hid uint) (SyncResult, error) {
	return pullItemsFromHostServ(mid, hid, true)
}

func pullItemsFromHostServ(mid, hid uint, recordHistory bool) (SyncResult, error) {
	result := SyncResult{}
	setMonitorStatusSyncing(mid)
	setHostStatusSyncing(hid)

	host, err := repository.GetHostByIDDAO(hid)
	if err != nil {
		setHostStatusErrorWithReason(hid, err.Error())
		LogService("error", "pull items failed to load host", map[string]interface{}{"host_id": hid, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to get host: %w", err)
	}

	if host.MonitorID != mid {
		setHostStatusErrorWithReason(hid, "host does not belong to the specified monitor")
		LogService("error", "pull items failed due to monitor mismatch", map[string]interface{}{"host_id": hid, "monitor_id": mid}, nil, "")
		return result, fmt.Errorf("host does not belong to the specified monitor")
	}

	monitor, err := repository.GetMonitorByIDDAO(mid)
	if err != nil {
		// Fallback: If mid is 0 or not found, try to find the "Nagare Internal" monitor
		internalMonitors, sErr := repository.SearchMonitorsDAO(model.MonitorFilter{Query: "Nagare Internal"})
		if sErr == nil && len(internalMonitors) > 0 {
			monitor = internalMonitors[0]
			mid = monitor.ID
		} else {
			setHostStatusErrorWithReason(hid, err.Error())
			LogService("error", "pull items failed to load monitor", map[string]interface{}{"monitor_id": mid, "host_id": hid, "error": err.Error()}, nil, "")
			return result, fmt.Errorf("failed to get monitor: %w", err)
		}
	}

	if mid != 1 && (monitor.Status == 0 || monitor.Status == 2) {
		reason := "monitor is not active"
		if monitor.StatusDescription != "" {
			reason = monitor.StatusDescription
		}
		setHostStatusErrorWithReason(hid, reason)
		items, err := repository.GetItemsByHIDDAO(hid)
		if err == nil {
			for _, item := range items {
				_ = repository.UpdateItemStatusAndDescriptionDAO(item.ID, 2, reason)
			}
		}
		LogService("warn", "pull items skipped due to monitor not active", map[string]interface{}{"monitor_id": mid, "monitor_status": monitor.Status, "monitor_status_description": reason}, nil, "")
		return result, fmt.Errorf("monitor is not active (status: %d)", monitor.Status)
	}

	currentStatus := determineHostStatus(host, monitor)
	if currentStatus == 2 {
		reason := host.StatusDescription
		if reason == "" {
			reason = "host is not active"
		}
		setHostStatusErrorWithReason(hid, reason)
		items, err := repository.GetItemsByHIDDAO(hid)
		if err == nil {
			for _, item := range items {
				_ = repository.UpdateItemStatusAndDescriptionDAO(item.ID, 2, reason)
			}
		}
		LogService("warn", "pull items skipped due to host error", map[string]interface{}{"host_id": hid, "host_status": currentStatus}, nil, "")
		return result, fmt.Errorf("host is not active (status: %d)", currentStatus)
	}
	// Allow status 1 (Active) or 3 (Syncing) to proceed
	if currentStatus != 1 && currentStatus != 3 {
		_ = repository.UpdateHostStatusAndDescriptionDAO(hid, currentStatus, "")
		items, err := repository.GetItemsByHIDDAO(hid)
		if err == nil {
			for _, item := range items {
				_ = repository.UpdateItemStatusAndDescriptionDAO(item.ID, currentStatus, "")
			}
		}
		LogService("warn", "pull items skipped due to host not active", map[string]interface{}{"host_id": hid, "host_status": currentStatus}, nil, "")
		return result, fmt.Errorf("host is not active (status: %d)", currentStatus)
	}

	if host.Status != 1 || host.StatusDescription != "" {
		_ = repository.UpdateHostStatusAndDescriptionDAO(hid, 1, "")
	}

	client, err := createMonitorClientFromDomain(monitor)
	if err != nil {
		setHostStatusErrorWithReason(hid, err.Error())
		LogService("error", "pull items failed to create monitor client", map[string]interface{}{"monitor_id": mid, "host_id": hid, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to create monitor client: %w", err)
	}

	// Use existing auth token if available
	if monitor.AuthToken != "" {
		client.SetAuthToken(monitor.AuthToken)
	} else {
		if err := client.Authenticate(context.Background()); err != nil {
			setHostStatusErrorWithReason(hid, err.Error())
			LogService("error", "pull items failed to authenticate", map[string]interface{}{"monitor_id": mid, "host_id": hid, "error": err.Error()}, nil, "")
			return result, fmt.Errorf("failed to authenticate with monitor: %w", err)
		}
	}

	ctx := context.Background()
	targetID := host.Hostid
	mType := monitors.ParseMonitorType(monitor.Type)
	fmt.Printf("[DEBUG] pullItemsFromHostServ: monitor_type=%s (%d), host=%s, target=%s\n", mType.String(), monitor.Type, host.Name, targetID)
	
	if mType == monitors.MonitorSNMP {
		LogService("debug", "preparing SNMP poll", map[string]interface{}{"host": host.Name, "ip": host.IPAddr, "version": host.SNMPVersion}, nil, "")
		fmt.Printf("[DEBUG] SNMP Config: IP=%s, Version=%s, Community=%s\n", host.IPAddr, host.SNMPVersion, host.SNMPCommunity)
		authPass := host.SNMPV3AuthPass
		if authPass != "" {
			if decrypted, err := utils.Decrypt(authPass); err == nil {
				authPass = decrypted
			}
		}
		privPass := host.SNMPV3PrivPass
		if privPass != "" {
			if decrypted, err := utils.Decrypt(privPass); err == nil {
				privPass = decrypted
			}
		}

		snmpCfg := monitors.SnmpConfig{
			Community:       host.SNMPCommunity,
			Version:         host.SNMPVersion,
			Port:            host.SNMPPort,
			V3User:          host.SNMPV3User,
			V3AuthPass:      authPass,
			V3PrivPass:      privPass,
			V3AuthProtocol:  host.SNMPV3AuthProtocol,
			V3PrivProtocol:  host.SNMPV3PrivProtocol,
			V3SecurityLevel: host.SNMPV3SecurityLevel,
		}
		// Use IP address as target for SNMP
		targetID = host.IPAddr
		if targetID == "" {
			targetID = host.Hostid
		}
		fmt.Printf("[DEBUG] SNMP Final Target: %s\n", targetID)

		// Load existing items to pick up custom OIDs
		items, err := repository.GetItemsByHIDDAO(host.ID)
		if err == nil && len(items) > 0 {
			snmpCfg.CustomOIDs = make(map[string]string)
			for _, it := range items {
				// If ItemID looks like an OID, add it to poller
				if strings.HasPrefix(it.ItemID, "1.3.6") || strings.HasPrefix(it.ItemID, ".") {
					snmpCfg.CustomOIDs[it.ItemID] = it.Name
				}
			}
		}
		ctx = context.WithValue(ctx, "snmp_config", snmpCfg)
	}

	fmt.Printf("[DEBUG] Calling client.GetItems for %s\n", targetID)
	monitorItems, err := client.GetItems(ctx, targetID)
	if err != nil {
		fmt.Printf("[DEBUG] client.GetItems error: %v\n", err)
		LogService("error", "poller.GetItems failed", map[string]interface{}{"monitor_id": mid, "host_id": hid, "error": err.Error(), "target": targetID}, nil, "")
		setHostStatusErrorWithReason(hid, err.Error())
		return result, fmt.Errorf("failed to get items from monitor: %w", err)
	}
	fmt.Printf("[DEBUG] client.GetItems returned %d items\n", len(monitorItems))
	monitorItemIDs := make(map[string]struct{}, len(monitorItems))
	for _, mItem := range monitorItems {
		monitorItemIDs[mItem.ID] = struct{}{}
	}

	now := time.Now().UTC()
	for _, mItem := range monitorItems {
		enabled, status := mapMonitorItemStatus(mItem.Status)
		item, err := repository.GetItemByHIDAndItemIDDAO(hid, mItem.ID)

		if err != nil {
			// Check if item exists with same NAME but different ID (OID Migration case)
			existingItems, _ := repository.GetItemsByHIDDAO(hid)
			foundByName := false
			for _, ex := range existingItems {
				// Case-insensitive name matching for robustness
				if strings.EqualFold(strings.TrimSpace(ex.Name), strings.TrimSpace(mItem.Name)) && ex.ItemID != mItem.ID {
					// Delete old item if it has a legacy OID to prevent duplicates
					if strings.Contains(ex.ItemID, ".2011.5.25.31.1.1.1.1") {
						_ = repository.DeleteItemByIDDAO(ex.ID)
						foundByName = false // Trigger creation of fresh V200 record
						break
					}

					ex.ItemID = mItem.ID
					ex.LastValue = mItem.Value
					ex.Status = status
					ex.LastSyncAt = &now
					if err := repository.UpdateItemDAO(ex.ID, ex); err != nil {
						// Error handled by logging if needed, but removing debug print
					}
					item = ex
					foundByName = true
					// Mark as found in the current poll to prevent cleanup
					monitorItemIDs[mItem.ID] = struct{}{}
					break
				}
			}

			if !foundByName {
				// Item does not exist at all, add it
				newItem := model.Item{
					Name:           mItem.Name,
					HID:            hid,
					ItemID:         mItem.ID,
					ExternalHostID: mItem.HostID,
					ValueType:      mItem.ValueType,
					LastValue:      mItem.Value,
					Units:          mItem.Units,
					Enabled:        enabled,
					Status:         status,
					LastSyncAt:     &now,
					ExternalSource: monitor.Name,
				}
				if err := repository.AddItemDAO(newItem); err != nil {
					setHostStatusErrorWithReason(hid, err.Error())
					LogService("error", "pull items failed to add item", map[string]interface{}{"monitor_id": mid, "host_id": hid, "item_name": mItem.Name, "error": err.Error()}, nil, "")
					result.Failed++
					continue
				}
				if created, err := repository.GetItemByHIDAndItemIDDAO(hid, mItem.ID); err == nil {
					if recordHistory {
						sampledAt := time.Now().UTC()
						if mItem.Timestamp > 0 {
							sampledAt = time.Unix(mItem.Timestamp, 0).UTC()
						}
						recordItemHistory(created, sampledAt)
					}
					ExecuteTriggersForItem(created)
				}
				result.Added++
				continue
			}
		} else {
			// Item exists with same ID, update it
			item.Name = mItem.Name // FORCE UPDATE NAME
			item.ExternalHostID = mItem.HostID
			item.ValueType = mItem.ValueType
			item.LastValue = mItem.Value
			item.Units = mItem.Units // FORCE UPDATE UNITS
			item.Enabled = enabled
			item.Status = status
			item.LastSyncAt = &now
			item.ExternalSource = monitor.Name
			if err := repository.UpdateItemDAO(item.ID, item); err != nil {
				setItemStatusErrorWithReason(item.ID, err.Error())
				LogService("error", "pull items failed to update item", map[string]interface{}{"monitor_id": mid, "host_id": hid, "item_id": item.ID, "error": err.Error()}, nil, "")
				result.Failed++
				continue
			}
			if recordHistory {
				sampledAt := time.Now().UTC()
				if mItem.Timestamp > 0 {
					sampledAt = time.Unix(mItem.Timestamp, 0).UTC()
				}
				recordItemHistory(item, sampledAt)
			}
			ExecuteTriggersForItem(item)
			result.Updated++
		}
		result.Total++
	}

	localItems, err := repository.GetItemsByHIDDAO(hid)
	if err == nil && monitors.ParseMonitorType(monitor.Type) != monitors.MonitorSNMP {
		for _, localItem := range localItems {
			// Normalize local ItemID for comparison with monitorItemIDs (which are already normalized)
			normalizedLocalID := localItem.ItemID
			if len(normalizedLocalID) > 0 && normalizedLocalID[0] == '.' {
				normalizedLocalID = normalizedLocalID[1:]
			}

			if _, ok := monitorItemIDs[normalizedLocalID]; ok {
				continue
			}
			// Only mark as missing if it's not a special calculated item
			if strings.HasPrefix(localItem.ItemID, "calculated.") {
				continue
			}
			reason := "item not found on monitor"
			_ = repository.UpdateItemStatusAndDescriptionDAO(localItem.ID, 2, reason)
		}
	}

	// Deduplication and Cleanup for SNMP
	if monitors.ParseMonitorType(monitor.Type) == monitors.MonitorSNMP {
		allLocal, err := repository.GetItemsByHIDDAO(hid)
		if err == nil {
			nameToItems := make(map[string][]model.Item)
			for _, it := range allLocal {
				nameToItems[it.Name] = append(nameToItems[it.Name], it)
			}
			for _, its := range nameToItems {
				if len(its) > 1 {
					hasWorking := false
					for _, it := range its {
						// Item is working if it has a real value and status is Normal
						if it.LastValue != "N/A" && it.LastValue != "" && it.Status == 1 {
							hasWorking = true
							break
						}
					}
					if hasWorking {
						for _, it := range its {
							if it.LastValue == "N/A" || it.LastValue == "" || it.Status != 1 {
								_ = repository.DeleteItemByIDDAO(it.ID)
							}
						}
					}
				}
			}
		}
	}

	_, _ = recomputeHostStatus(hid)

	// Update host LastSyncAt after successful poll
	nowHost := time.Now().UTC()
	_ = repository.UpdateHostLastSyncAtDAO(hid, &nowHost)

	if host.GroupID > 0 {
		_, _ = recomputeGroupStatus(host.GroupID)
	}
	_, _ = recomputeMonitorStatus(mid)

	// Trigger automatic threshold checks for this host
	go CheckItemThresholds(hid)

	SyncEvent("items", mid, hid, result)
	return result, nil
}

func PullItemOfHostFromMonitorServ(mid, hid, id uint) (SyncResult, error) {
	item, err := repository.GetItemByIDDAO(id)
	if err != nil {
		setItemStatusErrorWithReason(id, err.Error())
		LogService("error", "pull item failed to load item", map[string]interface{}{"item_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get item: %w", err)
	}
	setMonitorStatusSyncing(mid)
	setItemStatusSyncing(id)

	host, err := repository.GetHostByIDDAO(hid)
	if err != nil {
		setHostStatusErrorWithReason(hid, err.Error())
		setItemStatusErrorWithReason(id, err.Error())
		LogService("error", "pull item failed to load host", map[string]interface{}{"item_id": id, "host_id": hid, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get host: %w", err)
	}

	if item.HID != host.ID {
		setItemStatusErrorWithReason(id, "item does not belong to the specified host")
		LogService("error", "pull item failed due to host mismatch", map[string]interface{}{"item_id": id, "host_id": hid}, nil, "")
		return SyncResult{}, fmt.Errorf("item does not belong to the specified host")
	}

	monitor, err := repository.GetMonitorByIDDAO(mid)
	if err != nil {
		setMonitorStatusError(mid)
		setItemStatusErrorWithReason(id, err.Error())
		LogService("error", "pull item failed to load monitor", map[string]interface{}{"monitor_id": mid, "item_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get monitor: %w", err)
	}

	client, err := createMonitorClientFromDomain(monitor)
	if err != nil {
		setMonitorStatusError(mid)
		setItemStatusErrorWithReason(id, err.Error())
		LogService("error", "pull item failed to create monitor client", map[string]interface{}{"monitor_id": mid, "item_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to create monitor client: %w", err)
	}

	// Use existing auth token if available
	if monitor.AuthToken != "" {
		client.SetAuthToken(monitor.AuthToken)
	} else {
		if err := client.Authenticate(context.Background()); err != nil {
			setMonitorStatusError(mid)
			setItemStatusErrorWithReason(id, err.Error())
			LogService("error", "pull item failed to authenticate", map[string]interface{}{"monitor_id": mid, "item_id": id, "error": err.Error()}, nil, "")
			return SyncResult{}, fmt.Errorf("failed to authenticate with monitor: %w", err)
		}
	}

	monitorItem, err := client.GetItemByID(context.Background(), item.ItemID)
	if err != nil {
		setMonitorStatusError(mid)
		setItemStatusErrorWithReason(id, err.Error())
		LogService("error", "pull item failed to fetch item", map[string]interface{}{"monitor_id": mid, "item_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get item from monitor: %w", err)
	}

	if monitorItem == nil {
		return SyncResult{}, fmt.Errorf("item %s not found on monitor", item.ItemID)
	}

	item.Name = monitorItem.Name
	item.ExternalHostID = monitorItem.HostID
	item.ValueType = monitorItem.ValueType
	item.LastValue = monitorItem.Value
	item.Units = monitorItem.Units
	enabled, status := mapMonitorItemStatus(monitorItem.Status)
	item.Enabled = enabled
	item.Status = status
	if err := repository.UpdateItemDAO(item.ID, item); err != nil {
		setItemStatusErrorWithReason(item.ID, err.Error())
		LogService("error", "pull item failed to update item", map[string]interface{}{"monitor_id": mid, "item_id": item.ID, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to update item: %w", err)
	}
	sampledAt := time.Now().UTC()
	if monitorItem.Timestamp > 0 {
		sampledAt = time.Unix(monitorItem.Timestamp, 0).UTC()
	}
	recordItemHistory(item, sampledAt)
	ExecuteTriggersForItem(item)
	_ = recomputeMonitorRelated(mid)
	recordNetworkStatusSnapshot(time.Now().UTC())
	result := SyncResult{
		Added:   0,
		Updated: 1,
		Failed:  0,
		Total:   1,
	}
	SyncEvent("items", mid, hid, result)
	return result, nil
}

// PushItemToMonitorServ pushes an item from local database to remote monitor
func PushItemToMonitorServ(mid, hid, id uint) error {
	item, err := repository.GetItemByIDDAO(id)
	if err != nil {
		setItemStatusErrorWithReason(id, err.Error())
		LogService("error", "push item failed to load item", map[string]interface{}{"item_id": id, "error": err.Error()}, nil, "")
		return fmt.Errorf("failed to get item: %w", err)
	}
	setItemStatusSyncing(id)

	host, err := repository.GetHostByIDDAO(hid)
	if err != nil {
		setHostStatusErrorWithReason(hid, err.Error())
		setItemStatusErrorWithReason(id, err.Error())
		LogService("error", "push item failed to load host", map[string]interface{}{"item_id": id, "host_id": hid, "error": err.Error()}, nil, "")
		return fmt.Errorf("failed to get host: %w", err)
	}

	if item.HID != host.ID {
		setItemStatusErrorWithReason(id, "item does not belong to the specified host")
		LogService("error", "push item failed due to host mismatch", map[string]interface{}{"item_id": id, "host_id": hid}, nil, "")
		return fmt.Errorf("item does not belong to the specified host")
	}

	if host.MonitorID != mid {
		setHostStatusErrorWithReason(hid, "host does not belong to the specified monitor")
		setItemStatusErrorWithReason(id, "host does not belong to the specified monitor")
		LogService("error", "push item failed due to monitor mismatch", map[string]interface{}{"item_id": id, "host_id": hid, "monitor_id": mid}, nil, "")
		return fmt.Errorf("host does not belong to the specified monitor")
	}
	if host.Hostid == "" {
		if _, err := PushHostToMonitorServ(mid, hid); err != nil {
			setHostStatusErrorWithReason(hid, err.Error())
			setItemStatusErrorWithReason(id, err.Error())
			LogService("error", "push item failed to create host", map[string]interface{}{"item_id": id, "host_id": hid, "monitor_id": mid, "error": err.Error()}, nil, "")
			return fmt.Errorf("failed to create host in monitor: %w", err)
		}
		updatedHost, err := repository.GetHostByIDDAO(hid)
		if err != nil {
			return fmt.Errorf("failed to reload host: %w", err)
		}
		host = updatedHost
	}

	monitor, err := repository.GetMonitorByIDDAO(mid)
	if err != nil {
		setMonitorStatusError(mid)
		setItemStatusErrorWithReason(id, err.Error())
		LogService("error", "push item failed to load monitor", map[string]interface{}{"monitor_id": mid, "item_id": id, "error": err.Error()}, nil, "")
		return fmt.Errorf("failed to get monitor: %w", err)
	}

	if mid != 1 && monitor.Status == 2 {
		reason := "monitor is in error state"
		if monitor.StatusDescription != "" {
			reason = monitor.StatusDescription
		}
		setItemStatusErrorWithReason(id, reason)
		return fmt.Errorf("monitor is in error state: %s", reason)
	}

	client, err := createMonitorClientFromDomain(monitor)
	if err != nil {
		setMonitorStatusError(mid)
		setItemStatusErrorWithReason(id, err.Error())
		LogService("error", "push item failed to create monitor client", map[string]interface{}{"monitor_id": mid, "item_id": id, "error": err.Error()}, nil, "")
		return fmt.Errorf("failed to create monitor client: %w", err)
	}

	// Use existing auth token if available
	if monitor.AuthToken != "" {
		client.SetAuthToken(monitor.AuthToken)
	} else {
		if err := client.Authenticate(context.Background()); err != nil {
			setMonitorStatusError(mid)
			setItemStatusErrorWithReason(id, err.Error())
			LogService("error", "push item failed to authenticate", map[string]interface{}{"monitor_id": mid, "item_id": id, "error": err.Error()}, nil, "")
			return fmt.Errorf("failed to authenticate with monitor: %w", err)
		}
	}

	key := item.Name
	valueType := item.ValueType
	units := item.Units
	monitorItem := monitors.Item{
		ID:        item.ItemID,
		HostID:    host.Hostid,
		Name:      item.Name,
		Key:       key,
		Units:     units,
		ValueType: valueType,
		Status:    map[bool]string{true: "1", false: "0"}[item.Enabled == 0],
	}
	if item.ItemID == "" {
		created, err := client.CreateItem(context.Background(), monitorItem)
		if err != nil {
			setMonitorStatusError(mid)
			setItemStatusErrorWithReason(id, err.Error())
			LogService("error", "push item failed to create item", map[string]interface{}{"monitor_id": mid, "item_id": id, "error": err.Error()}, nil, "")
			return fmt.Errorf("failed to create item in monitor: %w", err)
		}
		if created.ID != "" {
			item.ItemID = created.ID
			_ = repository.UpdateItemDAO(item.ID, item)
		}
	} else {
		if _, err := client.UpdateItem(context.Background(), monitorItem); err != nil {
			setMonitorStatusError(mid)
			setItemStatusErrorWithReason(id, err.Error())
			LogService("error", "push item failed to update item", map[string]interface{}{"monitor_id": mid, "item_id": id, "error": err.Error()}, nil, "")
			return fmt.Errorf("failed to update item in monitor: %w", err)
		}
	}
	LogService("info", "push item to monitor", map[string]interface{}{"item_name": item.Name, "item_id": item.ItemID, "host": host.Name, "monitor": monitor.Name}, nil, "")
	_, _ = recomputeItemStatus(id)
	return recomputeMonitorRelated(mid)
}

// PushItemsFromHostServ pushes all items for a host from local database to remote monitor
func PushItemsFromHostServ(mid, hid uint) (SyncResult, error) {
	result := SyncResult{}
	setHostStatusSyncing(hid)

	// Check host status before proceeding with item push
	host, err := repository.GetHostByIDDAO(hid)
	if err != nil {
		setHostStatusErrorWithReason(hid, err.Error())
		LogService("error", "push items failed to load host", map[string]interface{}{"host_id": hid, "monitor_id": mid, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to get host: %w", err)
	}

	monitor, err := repository.GetMonitorByIDDAO(mid)
	if err != nil {
		setMonitorStatusError(mid)
		setHostStatusErrorWithReason(hid, err.Error())
		LogService("error", "push items failed to load monitor", map[string]interface{}{"monitor_id": mid, "host_id": hid, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to get monitor: %w", err)
	}
	if monitor.Status == 2 {
		reason := "monitor is in error state"
		if monitor.StatusDescription != "" {
			reason = monitor.StatusDescription
		}
		setMonitorStatusErrorWithReason(mid, reason)
		setHostStatusErrorWithReason(hid, reason)
		LogService("warn", "push items skipped due to monitor error", map[string]interface{}{"monitor_id": mid, "monitor_status": monitor.Status, "monitor_status_description": reason}, nil, "")
		return result, fmt.Errorf("monitor is in error state (status: %d)", monitor.Status)
	}

	currentStatus := determineHostStatus(host, monitor)
	if currentStatus == 2 {
		reason := host.StatusDescription
		if reason == "" {
			reason = "host is in error state"
		}
		setHostStatusErrorWithReason(hid, reason)
		LogService("warn", "push items skipped due to host error", map[string]interface{}{"host_id": hid, "host_status": currentStatus, "host_status_description": reason}, nil, "")
		return result, fmt.Errorf("host is in error state (status: %d)", currentStatus)
	}

	if host.Status != currentStatus {
		_ = repository.UpdateHostStatusAndDescriptionDAO(hid, currentStatus, "")
	}

	items, err := repository.GetItemsByHIDDAO(hid)
	if err != nil {
		setHostStatusErrorWithReason(hid, err.Error())
		LogService("error", "push items failed to load items", map[string]interface{}{"host_id": hid, "monitor_id": mid, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to get items: %w", err)
	}

	result.Total = len(items)

	for _, item := range items {
		if err := PushItemToMonitorServ(mid, hid, item.ID); err != nil {
			setItemStatusErrorWithReason(item.ID, err.Error())
			LogService("error", "push items failed to push item", map[string]interface{}{"monitor_id": mid, "host_id": hid, "item_id": item.ID, "error": err.Error()}, nil, "")
			result.Failed++
			continue
		}
		result.Added++
	}

	_ = recomputeMonitorRelated(mid)
	return result, nil
}

// PushItemsFromMonitorServ pushes all items from all hosts for a monitor from local database to remote monitor
func PushItemsFromMonitorServ(mid uint) (SyncResult, error) {
	result := SyncResult{}
	setMonitorStatusSyncing(mid)
	if mid != 1 {
		_, _ = TestMonitorStatusServ(mid)
	}

	// Check monitor status before proceeding with item push
	monitor, err := repository.GetMonitorByIDDAO(mid)
	if err != nil {
		setMonitorStatusError(mid)
		LogService("error", "push items failed to load monitor", map[string]interface{}{"monitor_id": mid, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to get monitor: %w", err)
	}

	// Monitor must be active or inactive (not error) to push items
	if mid != 1 && monitor.Status == 2 {
		reason := "monitor is in error state"
		if monitor.StatusDescription != "" {
			reason = monitor.StatusDescription
		}
		setMonitorStatusErrorWithReason(mid, reason)
		LogService("warn", "push items skipped due to monitor error", map[string]interface{}{"monitor_id": mid, "monitor_status": monitor.Status, "monitor_status_description": reason}, nil, "")
		return result, fmt.Errorf("monitor is in error state (status: %d)", monitor.Status)
	}

	hosts, err := repository.SearchHostsDAO(model.HostFilter{MID: &mid})
	if err != nil {
		setMonitorStatusError(mid)
		LogService("error", "push items failed to load hosts", map[string]interface{}{"monitor_id": mid, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to get hosts: %w", err)
	}

	for _, host := range hosts {
		hostResult, err := PushItemsFromHostServ(mid, host.ID)
		if err != nil {
			setHostStatusErrorWithReason(host.ID, err.Error())
			LogService("error", "push items failed for host", map[string]interface{}{"monitor_id": mid, "host_id": host.ID, "error": err.Error()}, nil, "")
			result.Failed++
			continue
		}
		result.Added += hostResult.Added
		result.Updated += hostResult.Updated
		result.Failed += hostResult.Failed
		result.Total += hostResult.Total
	}

	_ = recomputeMonitorRelated(mid)
	return result, nil
}

func mapMonitorItemStatus(status string) (enabled int, itemStatus int) {
	enabled = 1
	itemStatus = 1

	// Convention: '0' = Enabled/Active, '1' = Disabled/Inactive
	if status == "1" || strings.ToLower(status) == "disabled" {
		enabled = 0
		itemStatus = 0
	}

	return enabled, itemStatus
}

// Comment helpers removed: comment is reserved for human/AI notes.
