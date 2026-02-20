package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"nagare/internal/model"
	"nagare/internal/repository"
)

// TriggerReq represents a trigger request
type TriggerReq struct {
	Name                  string   `json:"name" binding:"required"`
	Entity                string   `json:"entity"`
	SeverityMin           int      `json:"severity_min"`
	ActionID              uint     `json:"action_id"`
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
	ActionID              uint     `json:"action_id"`
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
	if req.ActionID == 0 {
		return TriggerResp{}, fmt.Errorf("action_id is required")
	}
	trigger := model.Trigger{
		Name:                  req.Name,
		Entity:                normalizeTriggerEntity(req.Entity),
		SeverityMin:           req.SeverityMin,
		ActionID:              req.ActionID,
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
	}
	if action, err := repository.GetActionByIDDAO(req.ActionID); err == nil {
		trigger.Status = determineTriggerStatus(trigger, action)
	} else {
		trigger.Status = determineTriggerStatus(trigger, model.Action{})
	}
	if err := repository.AddTriggerDAO(trigger); err != nil {
		return TriggerResp{}, fmt.Errorf("failed to add trigger: %w", err)
	}
	return triggerToResp(trigger), nil
}

func UpdateTriggerServ(id uint, req TriggerReq) error {
	if req.ActionID == 0 {
		return fmt.Errorf("action_id is required")
	}
	existing, err := repository.GetTriggerByIDDAO(id)
	if err != nil {
		return err
	}
	updated := model.Trigger{
		Name:                  req.Name,
		Entity:                normalizeTriggerEntity(req.Entity),
		SeverityMin:           req.SeverityMin,
		ActionID:              req.ActionID,
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
	// Preserve status unless enabled state or action changed
	if req.Enabled != existing.Enabled || req.ActionID != existing.ActionID {
		if action, err := repository.GetActionByIDDAO(req.ActionID); err == nil {
			updated.Status = determineTriggerStatus(updated, action)
		} else {
			updated.Status = determineTriggerStatus(updated, model.Action{})
		}
	}
	if err := repository.UpdateTriggerDAO(id, updated); err != nil {
		return err
	}
	_, _ = recomputeTriggerStatus(id)
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
		ActionID:              trigger.ActionID,
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

// ExecuteTriggersForAlert runs matching triggers and sends messages via media
func ExecuteTriggersForAlert(alert model.Alert) {
	context := buildAlertMatchContext(alert)
	replacements := buildAlertReplacements(context)
	execTriggersForAlert(context, replacements)
}

// AlertEvent represents a non-alert event (e.g. sync) that can trigger actions
type AlertEvent struct {
	Severity  int
	Status    int
	Message   string
	HostID    uint
	ItemID    uint
	MonitorID uint
	Entity    string
	Added     int
	Updated   int
	Failed    int
	Total     int
}

// ExecuteTriggersForEvent runs matching triggers for a custom event
func ExecuteTriggersForEvent(event AlertEvent) {
	replacements := map[string]string{
		"{{message}}":    event.Message,
		"{{severity}}":   fmt.Sprintf("%d", event.Severity),
		"{{status}}":     fmt.Sprintf("%d", event.Status),
		"{{host_id}}":    fmt.Sprintf("%d", event.HostID),
		"{{item_id}}":    fmt.Sprintf("%d", event.ItemID),
		"{{monitor_id}}": fmt.Sprintf("%d", event.MonitorID),
		"{{entity}}":     event.Entity,
		"{{added}}":      fmt.Sprintf("%d", event.Added),
		"{{updated}}":    fmt.Sprintf("%d", event.Updated),
		"{{failed}}":     fmt.Sprintf("%d", event.Failed),
		"{{total}}":      fmt.Sprintf("%d", event.Total),
	}
	execTriggersForEvent(event.Severity, replacements)
}

// ExecuteTriggersForLog runs matching triggers for a log entry
func ExecuteTriggersForLog(entry model.LogEntry) {
	replacements := buildLogReplacements(entry)
	execTriggersForLog(entry, replacements)
}

// ExecuteTriggersForItem runs matching triggers for an item update
func ExecuteTriggersForItem(item model.Item) {
	replacements := buildItemReplacements(item)
	execTriggersForItem(item, replacements)
}

func buildAlertReplacements(ctx alertMatchContext) map[string]string {
	alert := ctx.alert
	return map[string]string{
		"{{alert_id}}":       fmt.Sprintf("%d", alert.ID),
		"{{message}}":        alert.Message,
		"{{severity}}":       fmt.Sprintf("%d", alert.Severity),
		"{{severity_label}}": severityLabel(alert.Severity),
		"{{status}}":         fmt.Sprintf("%d", alert.Status),
		"{{host_id}}":        fmt.Sprintf("%d", alert.HostID),
		"{{item_id}}":        fmt.Sprintf("%d", alert.ItemID),
		"{{monitor_id}}":     fmt.Sprintf("%d", ctx.monitorID),
		"{{group_id}}":       fmt.Sprintf("%d", ctx.groupID),
		"{{analysis}}":       alert.Comment,
		"{{created_at}}":     alert.CreatedAt.Format(time.RFC3339),
	}
}

func buildLogReplacements(entry model.LogEntry) map[string]string {
	return map[string]string{
		"{{log_id}}":         fmt.Sprintf("%d", entry.ID),
		"{{message}}":        entry.Message,
		"{{level}}":          logSeverityLabel(entry.Severity),
		"{{severity}}":       fmt.Sprintf("%d", entry.Severity),
		"{{severity_label}}": logSeverityLabel(entry.Severity),
		"{{type}}":           entry.Type,
		"{{context}}":        entry.Context,
		"{{created_at}}":     entry.CreatedAt.Format(time.RFC3339),
	}
}

func buildItemReplacements(item model.Item) map[string]string {
	hostName := ""
	if host, err := repository.GetHostByIDDAO(item.HID); err == nil {
		hostName = host.Name
	}
	return map[string]string{
		"{{item_id}}":    fmt.Sprintf("%d", item.ID),
		"{{name}}":       item.Name,
		"{{value}}":      item.LastValue,
		"{{units}}":      item.Units,
		"{{status}}":     fmt.Sprintf("%d", item.Status),
		"{{host_id}}":    fmt.Sprintf("%d", item.HID),
		"{{host_name}}":  hostName,
		"{{created_at}}": time.Now().UTC().Format(time.RFC3339),
	}
}

func execTriggersForAlert(ctx alertMatchContext, replacements map[string]string) {
	triggers, err := repository.GetActiveTriggersForEntityDAO("alert")
	if err != nil {
		return
	}
	for _, trigger := range triggers {
		if !matchAlertTrigger(trigger, ctx) {
			continue
		}
		invokeAlertTriggerAction(trigger, replacements)
	}
}

func execTriggersForEvent(severity int, replacements map[string]string) {
	triggers, err := repository.GetActiveTriggersForEntityDAO("alert")
	if err != nil {
		return
	}
	for _, trigger := range triggers {
		if trigger.Enabled == 0 || trigger.SeverityMin > severity {
			continue
		}
		invokeAlertTriggerAction(trigger, replacements)
	}
}

func execTriggersForLog(entry model.LogEntry, replacements map[string]string) {
	triggers, err := repository.GetActiveTriggersForEntityDAO("log")
	if err != nil {
		return
	}
	for _, trigger := range triggers {
		if !matchLogTrigger(trigger, entry) {
			continue
		}
		invokeLogTriggerAction(trigger, replacements)
	}
}

func execTriggersForItem(item model.Item, replacements map[string]string) {
	triggers, err := repository.GetActiveTriggersForEntityDAO("item")
	if err != nil {
		return
	}
	for _, trigger := range triggers {
		if !matchItemTrigger(trigger, item) {
			continue
		}
		invokeItemTriggerAction(trigger, replacements)
	}
}

func invokeAlertTriggerAction(trigger model.Trigger, replacements map[string]string) {
	if trigger.Enabled == 0 {
		return
	}
	action, err := repository.GetActionByIDDAO(trigger.ActionID)
	if err != nil || action.Enabled == 0 {
		return
	}
	media, err := repository.GetMediaByIDDAO(action.MediaID)
	if err != nil || media.Enabled == 0 {
		return
	}
	_ = ExecuteAction(action, media, replacements)
}

func invokeLogTriggerAction(trigger model.Trigger, replacements map[string]string) {
	if trigger.Enabled == 0 {
		return
	}
	action, err := repository.GetActionByIDDAO(trigger.ActionID)
	if err != nil || action.Enabled == 0 {
		return
	}
	media, err := repository.GetMediaByIDDAO(action.MediaID)
	if err != nil || media.Enabled == 0 {
		return
	}
	_ = ExecuteLogAction(action, media, replacements)
}

func invokeItemTriggerAction(trigger model.Trigger, replacements map[string]string) {
	if trigger.Enabled == 0 {
		return
	}
	action, err := repository.GetActionByIDDAO(trigger.ActionID)
	if err != nil || action.Enabled == 0 {
		return
	}
	media, err := repository.GetMediaByIDDAO(action.MediaID)
	if err != nil || media.Enabled == 0 {
		return
	}
	_ = ExecuteItemAction(action, media, replacements)
}

type alertMatchContext struct {
	alert     model.Alert
	host      *model.Host
	item      *model.Item
	monitorID uint
	groupID   uint
}

func buildAlertMatchContext(alert model.Alert) alertMatchContext {
	ctx := alertMatchContext{alert: alert}
	if alert.ItemID > 0 {
		if item, err := repository.GetItemByIDDAO(alert.ItemID); err == nil {
			ctx.item = &item
		}
	}
	if alert.HostID > 0 {
		if host, err := repository.GetHostByIDDAO(alert.HostID); err == nil {
			ctx.host = &host
		}
	}
	if ctx.host == nil && ctx.item != nil && ctx.item.HID > 0 {
		if host, err := repository.GetHostByIDDAO(ctx.item.HID); err == nil {
			ctx.host = &host
		}
	}
	if ctx.host != nil {
		ctx.monitorID = ctx.host.MonitorID
		ctx.groupID = ctx.host.GroupID
	}
	return ctx
}

func matchAlertTrigger(trigger model.Trigger, ctx alertMatchContext) bool {
	entity := normalizeTriggerEntity(trigger.Entity)
	if entity != "" && entity != "alert" {
		return false
	}
	alert := ctx.alert
	if trigger.AlertID != nil && alert.ID != *trigger.AlertID {
		return false
	}
	if trigger.SeverityMin > 0 && alert.Severity < trigger.SeverityMin {
		return false
	}
	if trigger.AlertStatus != nil && alert.Status != *trigger.AlertStatus {
		return false
	}
	if trigger.AlertMonitorID != nil && ctx.monitorID != *trigger.AlertMonitorID {
		return false
	}
	if trigger.AlertGroupID != nil && ctx.groupID != *trigger.AlertGroupID {
		return false
	}
	hostID := alert.HostID
	if ctx.host != nil {
		hostID = ctx.host.ID
	}
	if trigger.AlertHostID != nil && hostID != *trigger.AlertHostID {
		return false
	}
	if trigger.AlertItemID != nil && alert.ItemID != *trigger.AlertItemID {
		return false
	}
	if trigger.AlertQuery != "" && !strings.Contains(strings.ToLower(alert.Message), strings.ToLower(trigger.AlertQuery)) {
		return false
	}
	return true
}

func matchLogTrigger(trigger model.Trigger, entry model.LogEntry) bool {
	entity := normalizeTriggerEntity(trigger.Entity)
	if entity != "log" {
		return false
	}
	if trigger.LogType != "" && !strings.EqualFold(trigger.LogType, entry.Type) {
		return false
	}
	if trigger.LogSeverity != nil && entry.Severity != *trigger.LogSeverity {
		return false
	}
	if trigger.LogQuery != "" {
		q := strings.ToLower(trigger.LogQuery)
		if !strings.Contains(strings.ToLower(entry.Message), q) && !strings.Contains(strings.ToLower(entry.Context), q) {
			return false
		}
	}
	return true
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
	if value == "" {
		return "alert"
	}
	if value != "alert" && value != "log" && value != "item" {
		return "alert"
	}
	return value
}

func severityLabel(severity int) string {
	if severity >= 3 {
		return "Critical"
	}
	if severity == 2 {
		return "Warning"
	}
	return "Normal"
}

func logSeverityLabel(severity int) string {
	switch severity {
	case 2:
		return "Error"
	case 1:
		return "Warn"
	default:
		return "Info"
	}
}
