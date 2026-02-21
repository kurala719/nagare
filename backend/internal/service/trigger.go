package service

import (
	"fmt"
	"strconv"
	"strings"

	"nagare/internal/model"
	"nagare/internal/repository"
)

// TriggerReq represents a trigger request
type TriggerReq struct {
	Name                  string   `json:"name" binding:"required"`
	Entity                string   `json:"entity"`
	SeverityMin           int      `json:"severity_min"`
	AlertID               *uint    `json:"alert_id"`
	AlertStatus           *int     `json:"alert_status"`
	AlertGroupID          *uint    `json:"alert_group_id"`
	AlertMonitorID        *uint    `json:"alert_monitor_id"`
	AlertHostID           *uint    `json:"alert_host_id"`
	AlertItemID           *uint    `json:"alert_item_id"`
	AlertQuery            string   `json:"alert_query"`
	LogType               string   `json:"log_type"`
	LogSeverity           *int     `json:"log_severity"`
	LogQuery              string   `json:"log_query"`
	ItemStatus            *int     `json:"item_status"`
	ItemValueThreshold    *float64 `json:"item_value_threshold"`
	ItemValueThresholdMax *float64 `json:"item_value_threshold_max"`
	ItemValueOperator     string   `json:"item_value_operator"`
	Enabled               int      `json:"enabled"`
}

// TriggerResp represents a trigger response
type TriggerResp struct {
	ID                    int      `json:"id"`
	Name                  string   `json:"name"`
	Entity                string   `json:"entity"`
	SeverityMin           int      `json:"severity_min"`
	AlertID               *uint    `json:"alert_id"`
	AlertStatus           *int     `json:"alert_status"`
	AlertGroupID          *uint    `json:"alert_group_id"`
	AlertMonitorID        *uint    `json:"alert_monitor_id"`
	AlertHostID           *uint    `json:"alert_host_id"`
	AlertItemID           *uint    `json:"alert_item_id"`
	AlertQuery            string   `json:"alert_query"`
	LogType               string   `json:"log_type"`
	LogSeverity           *int     `json:"log_severity"`
	LogQuery              string   `json:"log_query"`
	ItemStatus            *int     `json:"item_status"`
	ItemValueThreshold    *float64 `json:"item_value_threshold"`
	ItemValueThresholdMax *float64 `json:"item_value_threshold_max"`
	ItemValueOperator     string   `json:"item_value_operator"`
	Enabled               int      `json:"enabled"`
	Status                int      `json:"status"`
}

func GetAllTriggersServ() ([]TriggerResp, error) {
	triggers, err := repository.GetAllTriggersDAO()
	if err != nil {
		return nil, fmt.Errorf("failed to get triggers: %w", err)
	}
	result := make([]TriggerResp, 0, len(triggers))
	for _, t := range triggers {
		result = append(result, triggerToResp(t))
	}
	return result, nil
}

func SearchTriggersServ(filter model.TriggerFilter) ([]TriggerResp, error) {
	triggers, err := repository.SearchTriggersDAO(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search triggers: %w", err)
	}
	result := make([]TriggerResp, 0, len(triggers))
	for _, t := range triggers {
		result = append(result, triggerToResp(t))
	}
	return result, nil
}

// CountTriggersServ returns total count for triggers by filter
func CountTriggersServ(filter model.TriggerFilter) (int64, error) {
	return repository.CountTriggersDAO(filter)
}

func GetTriggerByIDServ(id uint) (TriggerResp, error) {
	trigger, err := repository.GetTriggerByIDDAO(id)
	if err != nil {
		return TriggerResp{}, fmt.Errorf("failed to get trigger: %w", err)
	}
	return triggerToResp(trigger), nil
}

func AddTriggerServ(req TriggerReq) (TriggerResp, error) {
	trigger := model.Trigger{
		Name:                  req.Name,
		Entity:                normalizeTriggerEntity(req.Entity),
		SeverityMin:           req.SeverityMin,
		AlertID:               req.AlertID,
		AlertStatus:           req.AlertStatus,
		AlertGroupID:          req.AlertGroupID,
		AlertMonitorID:        req.AlertMonitorID,
		AlertHostID:           req.AlertHostID,
		AlertItemID:           req.AlertItemID,
		AlertQuery:            req.AlertQuery,
		LogType:               req.LogType,
		LogSeverity:           req.LogSeverity,
		LogQuery:              req.LogQuery,
		ItemStatus:            req.ItemStatus,
		ItemValueThreshold:    req.ItemValueThreshold,
		ItemValueThresholdMax: req.ItemValueThresholdMax,
		ItemValueOperator:     req.ItemValueOperator,
		Enabled:               req.Enabled,
		Status:                1, // Default active if enabled
	}
	if req.Enabled == 0 {
		trigger.Status = 0
	}

	if err := repository.AddTriggerDAO(trigger); err != nil {
		return TriggerResp{}, fmt.Errorf("failed to add trigger: %w", err)
	}
	return triggerToResp(trigger), nil
}

func UpdateTriggerServ(id uint, req TriggerReq) error {
	existing, err := repository.GetTriggerByIDDAO(id)
	if err != nil {
		return err
	}
	updated := model.Trigger{
		Name:                  req.Name,
		Entity:                normalizeTriggerEntity(req.Entity),
		SeverityMin:           req.SeverityMin,
		AlertID:               req.AlertID,
		AlertStatus:           req.AlertStatus,
		AlertGroupID:          req.AlertGroupID,
		AlertMonitorID:        req.AlertMonitorID,
		AlertHostID:           req.AlertHostID,
		AlertItemID:           req.AlertItemID,
		AlertQuery:            req.AlertQuery,
		LogType:               req.LogType,
		LogSeverity:           req.LogSeverity,
		LogQuery:              req.LogQuery,
		ItemStatus:            req.ItemStatus,
		ItemValueThreshold:    req.ItemValueThreshold,
		ItemValueThresholdMax: req.ItemValueThresholdMax,
		ItemValueOperator:     req.ItemValueOperator,
		Enabled:               req.Enabled,
		Status:                existing.Status,
	}
	
	// Update status based on enabled state
	if req.Enabled != existing.Enabled {
		if req.Enabled == 1 {
			updated.Status = 1
		} else {
			updated.Status = 0
		}
	}

	if err := repository.UpdateTriggerDAO(id, updated); err != nil {
		return err
	}
	// No recompute needed as status is simple
	return nil
}

func DeleteTriggerByIDServ(id uint) error {
	return repository.DeleteTriggerByIDDAO(id)
}

func triggerToResp(trigger model.Trigger) TriggerResp {
	return TriggerResp{
		ID:                    int(trigger.ID),
		Name:                  trigger.Name,
		Entity:                trigger.Entity,
		SeverityMin:           trigger.SeverityMin,
		AlertID:               trigger.AlertID,
		AlertStatus:           trigger.AlertStatus,
		AlertGroupID:          trigger.AlertGroupID,
		AlertMonitorID:        trigger.AlertMonitorID,
		AlertHostID:           trigger.AlertHostID,
		AlertItemID:           trigger.AlertItemID,
		AlertQuery:            trigger.AlertQuery,
		LogType:               trigger.LogType,
		LogSeverity:           trigger.LogSeverity,
		LogQuery:              trigger.LogQuery,
		ItemStatus:            trigger.ItemStatus,
		ItemValueThreshold:    trigger.ItemValueThreshold,
		ItemValueThresholdMax: trigger.ItemValueThresholdMax,
		ItemValueOperator:     trigger.ItemValueOperator,
		Enabled:               trigger.Enabled,
		Status:                trigger.Status,
	}
}

// ExecuteTriggersForItem runs matching triggers for an item update
func ExecuteTriggersForItem(item model.Item) {
	triggers, err := repository.GetActiveTriggersForEntityDAO("item")
	if err != nil {
		return
	}
	for _, trigger := range triggers {
		if !matchItemTrigger(trigger, item) {
			continue
		}
		// Generate alert if item trigger matches
		generateAlertFromItemTrigger(trigger, item)
	}
}

// generateAlertFromItemTrigger creates an alert when an item trigger matches
func generateAlertFromItemTrigger(trigger model.Trigger, item model.Item) {
	// Deduplication: Check for existing active alerts for this item
	hostID := int(item.HID)
	itemID := int(item.ID)
	status := 0
	
	activeAlerts, err := repository.SearchAlertsDAO(model.AlertFilter{
		HostID: &hostID,
		ItemID: &itemID,
		Status: &status,
	})

	if err == nil && len(activeAlerts) > 0 {
		// An active alert already exists for this item, suppress duplicate
		return
	}

	// Build alert message with item information
	host, _ := repository.GetHostByIDDAO(item.HID)
	hostName := "Unknown"
	if host.ID > 0 {
		hostName = host.Name
	}

	message := fmt.Sprintf("Item %s on host %s has value %s%s",
		item.Name, hostName, item.LastValue, item.Units)

	// Determine severity from trigger settings
	severity := trigger.SeverityMin
	if severity == 0 {
		severity = 1 // Default to warning level
	}

	// Create the alert
	alertReq := AlertReq{
		Message:  message,
		Severity: severity,
		HostID:   item.HID,
		ItemID:   item.ID,
		Comment: fmt.Sprintf("Triggered by %s: %s operator %v", trigger.Name,
			describeItemTriggerCondition(trigger), trigger.ItemValueThreshold),
	}

	_ = AddAlertServ(alertReq)
	LogService("info", "alert generated from item trigger", map[string]interface{}{
		"trigger_id":   trigger.ID,
		"trigger_name": trigger.Name,
		"item_id":      item.ID,
		"item_name":    item.Name,
		"item_value":   item.LastValue,
		"host_id":      item.HID,
		"host_name":    hostName,
	}, nil, "")
}

// describeItemTriggerCondition creates a human-readable description of the trigger condition
func describeItemTriggerCondition(trigger model.Trigger) string {
	if trigger.ItemValueThreshold == nil {
		return "status check"
	}

	operator := strings.TrimSpace(trigger.ItemValueOperator)
	if operator == "" {
		operator = ">"
	}

	if operator == "between" || operator == "outside" {
		if trigger.ItemValueThresholdMax != nil {
			return fmt.Sprintf("%s between %.2f and %.2f",
				operator, *trigger.ItemValueThreshold, *trigger.ItemValueThresholdMax)
		}
	}

	return fmt.Sprintf("%s %.2f", operator, *trigger.ItemValueThreshold)
}

func matchItemTrigger(trigger model.Trigger, item model.Item) bool {
	entity := normalizeTriggerEntity(trigger.Entity)
	if entity != "item" {
		return false
	}
	if trigger.AlertItemID != nil && item.ID != *trigger.AlertItemID {
		return false
	}
	if trigger.AlertHostID != nil && item.HID != *trigger.AlertHostID {
		return false
	}
	if trigger.ItemStatus != nil && item.Status != *trigger.ItemStatus {
		return false
	}
	if trigger.ItemValueThreshold != nil {
		val, err := strconv.ParseFloat(item.LastValue, 64)
		if err != nil {
			return false
		}
		threshold := *trigger.ItemValueThreshold
		operator := strings.TrimSpace(trigger.ItemValueOperator)
		switch operator {
		case ">":
			if !(val > threshold) {
				return false
			}
		case ">=":
			if !(val >= threshold) {
				return false
			}
		case "<":
			if !(val < threshold) {
				return false
			}
		case "<=":
			if !(val <= threshold) {
				return false
			}
		case "=", "==":
			if !(val == threshold) {
				return false
			}
		case "!=":
			if !(val != threshold) {
				return false
			}
		case "between", "outside":
			if trigger.ItemValueThresholdMax == nil {
				return false
			}
			maxThreshold := *trigger.ItemValueThresholdMax
			minThreshold := threshold
			if minThreshold > maxThreshold {
				minThreshold, maxThreshold = maxThreshold, minThreshold
			}
			if operator == "between" {
				if val < minThreshold || val > maxThreshold {
					return false
				}
			} else {
				if val >= minThreshold && val <= maxThreshold {
					return false
				}
			}
		default:
			return false
		}
	}
	return true
}

func normalizeTriggerEntity(entity string) string {
	value := strings.ToLower(strings.TrimSpace(entity))
	// Default to "item" now
	if value == "" {
		return "item"
	}
	return value
}
