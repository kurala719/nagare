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
	SeverityMin *int   `json:"severity_min"`
	TriggerID   *uint  `json:"trigger_id"`
	HostID      *uint  `json:"host_id"`
	GroupID     *uint  `json:"group_id"`
	AlertStatus *int   `json:"alert_status"`
	UserIDs     []uint `json:"user_ids"`
}

// ActionResp represents an action response
type ActionResp struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	MediaID     uint           `json:"media_id"`
	Template    string         `json:"template"`
	Enabled     int            `json:"enabled"`
	Status      int            `json:"status"`
	Description string         `json:"description"`
	// Filter conditions
	SeverityMin *int           `json:"severity_min"`
	TriggerID   *uint          `json:"trigger_id"`
	HostID      *uint          `json:"host_id"`
	GroupID     *uint          `json:"group_id"`
	AlertStatus *int           `json:"alert_status"`
	Users       []UserResponse `json:"users"`
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
	users := make([]model.User, 0)
	for _, uid := range req.UserIDs {
		if u, err := repository.GetUserByIDDAO(int(uid)); err == nil {
			users = append(users, u)
		}
	}

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
		Users:       users,
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

	users := make([]model.User, 0)
	for _, uid := range req.UserIDs {
		if u, err := repository.GetUserByIDDAO(int(uid)); err == nil {
			users = append(users, u)
		}
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
		Users:       users,
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
	userResps := make([]UserResponse, 0, len(action.Users))
	for _, u := range action.Users {
		userResps = append(userResps, userToResp(u))
	}

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
		Users:       userResps,
	}
}

// ExecuteActionsForAlert evaluates all active actions against the alert and executes matching ones
func ExecuteActionsForAlert(alert model.Alert) {
	// Fetch all enabled actions
	actions, err := repository.GetAllActionsDAO()
	if err != nil {
		LogService("error", "failed to load actions for alert execution", map[string]interface{}{"error": err.Error()}, nil, "")
		return
	}

	enabledCount := 0
	for _, a := range actions {
		if a.Enabled == 1 {
			enabledCount++
		}
	}

	LogService("info", "evaluating actions for alert", map[string]interface{}{
		"alert_id": alert.ID,
		"total_actions": len(actions),
		"enabled_actions": enabledCount,
	}, nil, "")

	// Prepare context for matching
	matchCtx := buildAlertMatchContext(alert)
	replacements := buildAlertReplacements(matchCtx)

	for _, action := range actions {
		if action.Enabled == 0 {
			LogService("debug", "action skipped: disabled", map[string]interface{}{"action_id": action.ID, "action_name": action.Name}, nil, "")
			continue
		}
		if matchActionFilter(action, matchCtx) {
			LogService("info", "action matched for alert", map[string]interface{}{
				"action_id": action.ID,
				"action_name": action.Name,
				"alert_id": alert.ID,
			}, nil, "")

			// Get Media
			media, err := repository.GetMediaByIDDAO(action.MediaID)
			if err != nil {
				LogService("error", "failed to load media for action", map[string]interface{}{
					"action_id": action.ID,
					"media_id": action.MediaID,
					"error": err.Error(),
				}, nil, "")
				continue
			}
			if media.Enabled == 0 {
				LogService("warn", "media disabled for action", map[string]interface{}{
					"action_id": action.ID,
					"media_id": action.MediaID,
					"media_name": media.Name,
				}, nil, "")
				continue
			}

			// Execute default target
			if err := ExecuteAction(action, media, replacements); err != nil {
				LogService("error", "action execution failed", map[string]interface{}{
					"action_id": action.ID,
					"media_id": media.ID,
					"target": media.Target,
					"error": err.Error(),
				}, nil, "")
			} else {
				LogService("info", "action execution succeeded", map[string]interface{}{
					"action_id": action.ID,
					"media_id": media.ID,
					"target": media.Target,
				}, nil, "")
			}

			// Also send to specifically associated users if it's a QQ media
			lowerType := strings.ToLower(media.Type)
			if (lowerType == "qq" || lowerType == "qrobot") && len(action.Users) > 0 {
				for _, user := range action.Users {
					if user.QQ != "" {
						userMedia := media
						userMedia.Target = "user:" + user.QQ
						LogService("debug", "sending alert to associated user", map[string]interface{}{
							"action_id": action.ID,
							"user_id": user.ID,
							"qq": user.QQ,
						}, nil, "")
						if err := ExecuteAction(action, userMedia, replacements); err != nil {
							LogService("error", "user action execution failed", map[string]interface{}{
								"action_id": action.ID,
								"user_id": user.ID,
								"qq": user.QQ,
								"error": err.Error(),
							}, nil, "")
						}
					}
				}
			}
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
		LogService("debug", "action filter mismatch: severity", map[string]interface{}{"action_id": action.ID, "alert_severity": alert.Severity, "min_severity": *action.SeverityMin}, nil, "")
		return false
	}
	
	// Status Check
	// If AlertStatus is nil or -1, ignore status filter. 
	// (Note: default 0 means only active alerts)
	if action.AlertStatus != nil && *action.AlertStatus != -1 && alert.Status != *action.AlertStatus {
		LogService("debug", "action filter mismatch: status", map[string]interface{}{"action_id": action.ID, "alert_status": alert.Status, "filter_status": *action.AlertStatus}, nil, "")
		return false
	}
	
	// Host Check
	// Ignore if nil or 0
	if action.HostID != nil && *action.HostID > 0 {
		// Alert must be associated with this host
		hostID := alert.HostID
		if ctx.host != nil {
			hostID = ctx.host.ID
		}
		if hostID != *action.HostID {
			LogService("debug", "action filter mismatch: host", map[string]interface{}{"action_id": action.ID, "alert_host_id": hostID, "filter_host_id": *action.HostID}, nil, "")
			return false
		}
	}
	
	// Group Check
	// Ignore if nil or 0
	if action.GroupID != nil && *action.GroupID > 0 {
		if ctx.groupID != *action.GroupID {
			LogService("debug", "action filter mismatch: group", map[string]interface{}{"action_id": action.ID, "ctx_group_id": ctx.groupID, "filter_group_id": *action.GroupID}, nil, "")
			return false
		}
	}
	
	// Trigger ID Check
	if action.TriggerID != nil && *action.TriggerID > 0 {
		if alert.AlarmID != *action.TriggerID {
			LogService("debug", "action filter mismatch: trigger/alarm", map[string]interface{}{"action_id": action.ID, "alert_alarm_id": alert.AlarmID, "filter_trigger_id": *action.TriggerID}, nil, "")
			return false
		}
	}
	
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
			err := fmt.Errorf("QQ ID %s (group=%v) is not in whitelist or authorized", qqID, isGroup)
			LogService("info", "send message skipped (QQ alert whitelist)", map[string]interface{}{
				"media":        media.Type,
				"media_id":     media.ID,
				"target":       media.Target,
				"qq_id":        qqID,
				"is_group":     isGroup,
				"skip_trigger": true,
				"error":        err.Error(),
			}, nil, "")
			return err
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
