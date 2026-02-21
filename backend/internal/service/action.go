package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nagare/internal/model"
	"nagare/internal/repository"
	mediaSvc "nagare/internal/repository/media"
)

// ActionReq represents an action request
type ActionReq struct {
	Name        string `json:"name" binding:"required"`
	MediaID     uint   `json:"media_id"`
	Template    string `json:"template"`
	Enabled     int    `json:"enabled"`
	Description string `json:"description"`
	// Filter conditions
	SeverityMin *int  `json:"severity_min"`
	TriggerID   *uint `json:"trigger_id"`
	HostID      *uint `json:"host_id"`
	GroupID     *uint `json:"group_id"`
	AlertStatus *int  `json:"alert_status"`
}

// ActionResp represents an action response
type ActionResp struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	MediaID     uint   `json:"media_id"`
	Template    string `json:"template"`
	Enabled     int    `json:"enabled"`
	Status      int    `json:"status"`
	Description string `json:"description"`
	// Filter conditions
	SeverityMin *int  `json:"severity_min"`
	TriggerID   *uint `json:"trigger_id"`
	HostID      *uint `json:"host_id"`
	GroupID     *uint `json:"group_id"`
	AlertStatus *int  `json:"alert_status"`
}

func GetAllActionsServ() ([]ActionResp, error) {
	actions, err := repository.GetAllActionsDAO()
	if err != nil {
		return nil, fmt.Errorf("failed to get actions: %w", err)
	}
	result := make([]ActionResp, 0, len(actions))
	for _, a := range actions {
		result = append(result, actionToResp(a))
	}
	return result, nil
}

func SearchActionsServ(filter model.ActionFilter) ([]ActionResp, error) {
	actions, err := repository.SearchActionsDAO(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search actions: %w", err)
	}
	result := make([]ActionResp, 0, len(actions))
	for _, a := range actions {
		result = append(result, actionToResp(a))
	}
	return result, nil
}

// CountActionsServ returns total count for actions by filter
func CountActionsServ(filter model.ActionFilter) (int64, error) {
	return repository.CountActionsDAO(filter)
}

func GetActionByIDServ(id uint) (ActionResp, error) {
	action, err := repository.GetActionByIDDAO(id)
	if err != nil {
		return ActionResp{}, fmt.Errorf("failed to get action: %w", err)
	}
	return actionToResp(action), nil
}

func AddActionServ(req ActionReq) (ActionResp, error) {
	action := model.Action{
		Name:        req.Name,
		MediaID:     req.MediaID,
		Template:    req.Template,
		Enabled:     req.Enabled,
		Description: req.Description,
		SeverityMin: req.SeverityMin,
		TriggerID:   req.TriggerID,
		HostID:      req.HostID,
		GroupID:     req.GroupID,
		AlertStatus: req.AlertStatus,
	}
	if media, err := repository.GetMediaByIDDAO(req.MediaID); err == nil {
		action.Status = determineActionStatus(action, media)
	} else {
		action.Status = determineActionStatus(action, model.Media{})
	}
	if err := repository.AddActionDAO(action); err != nil {
		return ActionResp{}, fmt.Errorf("failed to add action: %w", err)
	}
	return actionToResp(action), nil
}

func UpdateActionServ(id uint, req ActionReq) error {
	existing, err := repository.GetActionByIDDAO(id)
	if err != nil {
		return err
	}
	updated := model.Action{
		Name:        req.Name,
		MediaID:     req.MediaID,
		Template:    req.Template,
		Enabled:     req.Enabled,
		Description: req.Description,
		Status:      existing.Status,
		SeverityMin: req.SeverityMin,
		TriggerID:   req.TriggerID,
		HostID:      req.HostID,
		GroupID:     req.GroupID,
		AlertStatus: req.AlertStatus,
	}
	// Preserve status unless enabled state or media changed
	if req.Enabled != existing.Enabled || req.MediaID != existing.MediaID {
		if media, err := repository.GetMediaByIDDAO(req.MediaID); err == nil {
			updated.Status = determineActionStatus(updated, media)
		} else {
			updated.Status = determineActionStatus(updated, model.Media{})
		}
	}
	if err := repository.UpdateActionDAO(id, updated); err != nil {
		return err
	}
	_, _ = recomputeActionStatus(id)
	return nil
}

func DeleteActionByIDServ(id uint) error {
	return repository.DeleteActionByIDDAO(id)
}

func actionToResp(action model.Action) ActionResp {
	return ActionResp{
		ID:          int(action.ID),
		Name:        action.Name,
		MediaID:     action.MediaID,
		Template:    action.Template,
		Enabled:     action.Enabled,
		Status:      action.Status,
		Description: action.Description,
		SeverityMin: action.SeverityMin,
		TriggerID:   action.TriggerID,
		HostID:      action.HostID,
		GroupID:     action.GroupID,
		AlertStatus: action.AlertStatus,
	}
}

// ExecuteActionsForAlert evaluates all active actions against the alert and executes matching ones
func ExecuteActionsForAlert(alert model.Alert) {
	// Fetch all enabled actions
	// Ideally we would have a repository method to get enabled actions only
	actions, err := repository.GetAllActionsDAO()
	if err != nil {
		LogService("error", "failed to load actions for alert execution", map[string]interface{}{"error": err.Error()}, nil, "")
		return
	}

	// Prepare context for matching
	context := buildAlertMatchContext(alert)
	replacements := buildAlertReplacements(context)

	for _, action := range actions {
		if action.Enabled == 0 {
			continue
		}
		if matchActionFilter(action, context) {
			// Get Media
			media, err := repository.GetMediaByIDDAO(action.MediaID)
			if err != nil || media.Enabled == 0 {
				continue
			}
			// Execute
			_ = ExecuteAction(action, media, replacements)
		}
	}
}

// ExecuteAction sends a message via the action's media
func ExecuteAction(action model.Action, media model.Media, replacements map[string]string) error {
	msg := action.Template
	if msg == "" {
		msg = "Alert: {{message}}"
	}
	msg = renderMessageTemplate(msg, replacements)
	msg = appendAlertDetails(msg, replacements)
	return sendMediaMessage(media, msg)
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

func matchActionFilter(action model.Action, ctx alertMatchContext) bool {
	alert := ctx.alert
	
	// Severity Check
	if action.SeverityMin != nil && alert.Severity < *action.SeverityMin {
		return false
	}
	
	// Status Check
	if action.AlertStatus != nil && alert.Status != *action.AlertStatus {
		return false
	}
	
	// Host Check
	if action.HostID != nil {
		// Alert must be associated with this host
		hostID := alert.HostID
		if ctx.host != nil {
			hostID = ctx.host.ID
		}
		if hostID != *action.HostID {
			return false
		}
	}
	
	// Group Check
	if action.GroupID != nil {
		if ctx.groupID != *action.GroupID {
			return false
		}
	}
	
	// Trigger ID Check - Not applicable as Alert model doesn't store TriggerID directly currently,
	// but we could infer it if needed. For now, skipping TriggerID match unless we add TriggerID to Alert model.
	// Since we removed ActionID from Trigger, there is no direct link back unless Alert stores "CreatedByTriggerID".
	// The current Alert model has AlarmID but not TriggerID.
	// We will skip this check for now or assume it always passes if set (feature limitation).
	
	return true
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

func renderMessageTemplate(template string, replacements map[string]string) string {
	result := template
	for k, v := range replacements {
		result = strings.ReplaceAll(result, k, v)
	}
	return result
}

func appendAlertDetails(message string, replacements map[string]string) string {
	detailsTemplate := strings.Join([]string{
		"",
		"Details:",
		"alert_id={{alert_id}}",
		"severity={{severity}} ({{severity_label}})",
		"status={{status}}",
		"monitor_id={{monitor_id}}",
		"host_id={{host_id}}",
		"item_id={{item_id}}",
		"created_at={{created_at}}",
		"group_id={{group_id}}",
		"analysis={{analysis}}",
	}, "\n")
	return strings.TrimSpace(message + "\n" + renderMessageTemplate(detailsTemplate, replacements))
}

func sendMediaMessage(media model.Media, msg string) error {
	resolvedType := resolveMediaTypeKeyForSend(media)
	if strings.TrimSpace(resolvedType) == "" {
		err := fmt.Errorf("media type key is empty")
		LogService("error", "send message failed", map[string]interface{}{"media": media.Type, "media_id": media.ID, "target": media.Target, "error": err.Error(), "skip_trigger": true}, nil, "")
		return err
	}

	// Check QQ whitelist for alert delivery
	lowerType := strings.ToLower(resolvedType)
	if lowerType == "qq" || lowerType == "qrobot" {
		qqID, isGroup := parseQQTarget(media.Target)
		if !CheckQQWhitelistForAlert(qqID, isGroup) {
			LogService("info", "send message skipped (QQ alert whitelist)", map[string]interface{}{
				"media":        media.Type,
				"media_id":     media.ID,
				"target":       media.Target,
				"qq_id":        qqID,
				"is_group":     isGroup,
				"skip_trigger": true,
			}, nil, "")
			return nil
		}
	}

	if ok, wait := allowMediaSend(media); !ok {
		LogService("info", "send message skipped (rate limit)", map[string]interface{}{
			"media":        media.Type,
			"media_id":     media.ID,
			"target":       media.Target,
			"wait_seconds": int(wait.Seconds()),
			"skip_trigger": true,
		}, nil, "")
		return nil
	}
	if err := mediaSvc.GetService().SendMessage(context.Background(), lowerType, media.Target, msg); err != nil {
		LogService("error", "send message failed", map[string]interface{}{"media": media.Type, "target": media.Target, "error": err.Error(), "skip_trigger": true}, nil, "")
		return err
	}
	LogService("info", "send message", map[string]interface{}{"media": media.Type, "target": media.Target, "message": msg, "skip_trigger": true}, nil, "")
	return nil
}

// parseQQTarget extracts QQ ID and determines if it's a group
// Target formats supported:
// - "group:123456" or "group_123456" (group)
// - "user:123456" or "user_123456" (user)
// - "123456" (legacy format, defaults to user)
func parseQQTarget(target string) (string, bool) {
	target = strings.TrimSpace(target)

	// Try colon format first
	colonParts := strings.SplitN(target, ":", 2)
	if len(colonParts) == 2 {
		prefix := strings.ToLower(colonParts[0])
		qqID := colonParts[1]
		isGroup := prefix == "group"
		return qqID, isGroup
	}

	// Try underscore format
	underscoreParts := strings.SplitN(target, "_", 2)
	if len(underscoreParts) == 2 {
		prefix := strings.ToLower(underscoreParts[0])
		qqID := underscoreParts[1]
		isGroup := prefix == "group"
		return qqID, isGroup
	}

	// Fallback: assume it's a user ID
	return target, false
}

func resolveMediaTypeKeyForSend(media model.Media) string {
	return strings.TrimSpace(media.Type)
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
