package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"nagare/internal/model"
	"nagare/internal/repository"
	mediaSvc "nagare/internal/repository/media"
)

var ErrMediaSendSkipped = errors.New("media send skipped")

// ActionReq represents an action request
type ActionReq struct {
	Name        string `json:"name" binding:"required"`
	MediaID     uint   `json:"media_id"`
	Enabled     int    `json:"enabled"`
	Description string `json:"description"`
	// Filter conditions
	SeverityMin *int   `json:"severity_min"`
	AlertStatus *int   `json:"alert_status"`
	UserIDs     []uint `json:"user_ids"`
}

// ActionResp represents an action response
type ActionResp struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	MediaID     uint   `json:"media_id"`
	Enabled     int    `json:"enabled"`
	Status      int    `json:"status"`
	Description string `json:"description"`
	// Filter conditions
	SeverityMin *int           `json:"severity_min"`
	AlertStatus *int           `json:"alert_status"`
	Users       []UserResponse `json:"users,omitempty"`
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
		Enabled:     req.Enabled,
		Description: req.Description,
		SeverityMin: req.SeverityMin,
		AlertStatus: req.AlertStatus,
	}

	// Bind users
	for _, uid := range req.UserIDs {
		user, err := repository.GetUserByIDDAO(int(uid))
		if err == nil {
			action.Users = append(action.Users, user)
		}
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
		Enabled:     req.Enabled,
		Description: req.Description,
		Status:      existing.Status,
		SeverityMin: req.SeverityMin,
		AlertStatus: req.AlertStatus,
	}
	updated.ID = id

	for _, uid := range req.UserIDs {
		user, err := repository.GetUserByIDDAO(int(uid))
		if err == nil {
			updated.Users = append(updated.Users, user)
		}
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
	var usersResp []UserResponse
	for _, u := range action.Users {
		usersResp = append(usersResp, userToResp(u))
	}

	return ActionResp{
		ID:          int(action.ID),
		Name:        action.Name,
		MediaID:     action.MediaID,
		Enabled:     action.Enabled,
		Status:      action.Status,
		Description: action.Description,
		SeverityMin: action.SeverityMin,
		AlertStatus: action.AlertStatus,
		Users:       usersResp,
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
		"alert_id":        alert.ID,
		"alert_message":   alert.Message,
		"total_actions":   len(actions),
		"enabled_actions": enabledCount,
	}, nil, "")

	// Safeguard: Only process alerts created in the last 10 minutes to avoid accidental historical storms
	// (Unless it's a test alert or similar)
	if time.Since(alert.CreatedAt) > 10*time.Minute {
		LogService("info", "action evaluation skipped: alert too old", map[string]interface{}{
			"alert_id":   alert.ID,
			"created_at": alert.CreatedAt,
			"age":        time.Since(alert.CreatedAt).String(),
		}, nil, "")
		return
	}

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
				"action_id":   action.ID,
				"action_name": action.Name,
				"alert_id":    alert.ID,
			}, nil, "")

			// Get Media
			media, err := repository.GetMediaByIDDAO(action.MediaID)
			if err != nil {
				LogService("error", "failed to load media for action", map[string]interface{}{
					"action_id": action.ID,
					"media_id":  action.MediaID,
					"error":     err.Error(),
				}, nil, "")
				continue
			}
			if media.Enabled == 0 {
				LogService("warn", "media disabled for action", map[string]interface{}{
					"action_id":  action.ID,
					"media_id":   action.MediaID,
					"media_name": media.Name,
				}, nil, "")
				continue
			}

			lowerType := strings.ToLower(media.Type)
			endpointOnlyQQTarget := (lowerType == "qq" || lowerType == "qrobot") && isQQEndpointOnlyTargetForAction(media.Target)

			// Execute default target
			if media.Target != "" {
				if endpointOnlyQQTarget {
					LogService("info", "action default qq target skipped (endpoint-only target requires user-bound recipients)", map[string]interface{}{
						"action_id": action.ID,
						"media_id":  media.ID,
						"target":    media.Target,
					}, nil, "")
				} else if err := ExecuteAction(action, media, replacements); err != nil {
					if errors.Is(err, ErrMediaSendSkipped) {
						LogService("info", "action execution skipped", map[string]interface{}{
							"action_id": action.ID,
							"media_id":  media.ID,
							"target":    media.Target,
							"reason":    err.Error(),
						}, nil, "")
					} else {
						LogService("error", "action execution failed", map[string]interface{}{
							"action_id": action.ID,
							"media_id":  media.ID,
							"target":    media.Target,
							"error":     err.Error(),
						}, nil, "")
					}
				} else {
					LogService("info", "action execution succeeded", map[string]interface{}{
						"action_id": action.ID,
						"media_id":  media.ID,
						"target":    media.Target,
					}, nil, "")
				}
			} else {
				LogService("debug", "action default target empty", map[string]interface{}{
					"action_id": action.ID,
					"media_id":  media.ID,
				}, nil, "")
			}

			// Also send to specifically associated users
			for _, user := range action.Users {
				userTarget := ""
				if (lowerType == "qq" || lowerType == "qrobot") && user.QQ != "" {
					if endpointOnlyQQTarget {
						userTarget = strings.TrimSpace(media.Target) + " user:" + user.QQ
					} else {
						userTarget = "user:" + user.QQ
					}
				} else if (lowerType == "smtp" || lowerType == "email") && user.Email != "" {
					userTarget = user.Email
				}

				if userTarget != "" {
					userMedia := media
					userMedia.Target = userTarget
					LogService("debug", "sending alert to associated user", map[string]interface{}{
						"action_id": action.ID,
						"user_id":   user.ID,
						"target":    userTarget,
					}, nil, "")
					if err := ExecuteAction(action, userMedia, replacements); err != nil {
						if errors.Is(err, ErrMediaSendSkipped) {
							LogService("info", "user action execution skipped", map[string]interface{}{
								"action_id": action.ID,
								"user_id":   user.ID,
								"target":    userTarget,
								"reason":    err.Error(),
							}, nil, "")
						} else {
							LogService("error", "user action execution failed", map[string]interface{}{
								"action_id": action.ID,
								"user_id":   user.ID,
								"target":    userTarget,
								"error":     err.Error(),
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
	msg := "Alert: {{message}}"
	msg = renderMessageTemplate(msg, replacements)
	msg = appendAlertDetails(msg, replacements)
	return sendMediaMessage(media, msg)
}

type alertMatchContext struct {
	alert     model.Alert
	host      *model.Host
	item      *model.Item
	monitorID uint
}

func buildAlertMatchContext(alert model.Alert) alertMatchContext {
	ctx := alertMatchContext{alert: alert}
	if alert.ItemID != nil && *alert.ItemID > 0 {
		if item, err := repository.GetItemByIDDAO(*alert.ItemID); err == nil {
			ctx.item = &item
		}
	}
	if ctx.host == nil && ctx.item != nil && ctx.item.HostID > 0 {
		if host, err := repository.GetHostByIDDAO(ctx.item.HostID); err == nil {
			ctx.host = &host
		}
	}
	if ctx.host != nil {
		if grp, err := repository.GetGroupByIDDAO(ctx.host.GroupID); err == nil {
			ctx.monitorID = grp.MonitorID
		}
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

	return true
}

func buildAlertReplacements(ctx alertMatchContext) map[string]string {
	alert := ctx.alert
	var hostIDStr string = "0"
	if ctx.host != nil {
		hostIDStr = fmt.Sprintf("%d", ctx.host.ID)
	}

	return map[string]string{
		"{{alert_id}}":       fmt.Sprintf("%d", alert.ID),
		"{{message}}":        alert.Message,
		"{{severity}}":       fmt.Sprintf("%d", alert.Severity),
		"{{severity_label}}": severityLabel(alert.Severity),
		"{{status}}":         fmt.Sprintf("%d", alert.Status),
		"{{host_id}}":        hostIDStr,
		"{{item_id}}":        fmt.Sprintf("%d", alert.ItemID),
		"{{monitor_id}}":     fmt.Sprintf("%d", ctx.monitorID),
		"{{group_id}}":       "0",
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

	lowerType := strings.ToLower(resolvedType)

	if ok, wait := allowMediaSend(media); !ok {
		LogService("info", "send message skipped (rate limit)", map[string]interface{}{
			"media":        media.Type,
			"media_id":     media.ID,
			"target":       media.Target,
			"wait_seconds": int(wait.Seconds()),
			"skip_trigger": true,
		}, nil, "")
		return fmt.Errorf("%w: rate limit, retry in %ds", ErrMediaSendSkipped, int(wait.Seconds()))
	}
	if err := mediaSvc.GetService().SendMessage(context.Background(), lowerType, media.Target, msg); err != nil {
		LogService("error", "send message failed", map[string]interface{}{"media": media.Type, "target": media.Target, "error": err.Error(), "skip_trigger": true}, nil, "")
		return err
	}
	LogService("info", "send message", map[string]interface{}{"media": media.Type, "target": media.Target, "message": msg, "skip_trigger": true}, nil, "")
	return nil
}

// TestActionServ manually triggers an action execution for testing purposes
func TestActionServ(id uint) error {
	action, err := repository.GetActionByIDDAO(id)
	if err != nil {
		return err
	}

	if action.MediaID == 0 {
		return fmt.Errorf("action has no delivery media defined")
	}

	media, err := repository.GetMediaByIDDAO(action.MediaID)
	if err != nil {
		return err
	}

	testMsg := "Nagare Test Action Message from Action: " + action.Name
	var lastErr error
	var sentCount int
	var skippedCount int
	lowerType := strings.ToLower(media.Type)
	endpointOnlyQQTarget := (lowerType == "qq" || lowerType == "qrobot") && isQQEndpointOnlyTargetForAction(media.Target)

	// Send to specifically associated users
	for _, user := range action.Users {
		userTarget := ""
		if (lowerType == "qq" || lowerType == "qrobot") && user.QQ != "" {
			if endpointOnlyQQTarget {
				userTarget = strings.TrimSpace(media.Target) + " user:" + user.QQ
			} else {
				userTarget = "user:" + user.QQ
			}
		} else if (lowerType == "smtp" || lowerType == "email") && user.Email != "" {
			userTarget = user.Email
		}

		if userTarget != "" {
			userMedia := media
			userMedia.Target = userTarget
			if err := sendMediaMessage(userMedia, testMsg); err != nil {
				if errors.Is(err, ErrMediaSendSkipped) {
					skippedCount++
				} else {
					lastErr = err
				}
			} else {
				sentCount++
			}
		}
	}

	// Send to default target if configured
	if media.Target != "" {
		if endpointOnlyQQTarget {
			LogService("info", "test action default qq target skipped (endpoint-only target requires user-bound recipients)", map[string]interface{}{
				"action_id": action.ID,
				"media_id":  media.ID,
				"target":    media.Target,
			}, nil, "")
		} else {
			if err := sendMediaMessage(media, testMsg); err != nil {
				if errors.Is(err, ErrMediaSendSkipped) {
					skippedCount++
				} else {
					lastErr = err
				}
			} else {
				sentCount++
			}
		}
	}

	if sentCount == 0 && skippedCount > 0 && lastErr == nil {
		return fmt.Errorf("message sending skipped due to rate limit")
	}

	if sentCount == 0 && lastErr == nil {
		return fmt.Errorf("no valid delivery targets found for this action")
	}

	return lastErr
}
func resolveMediaTypeKeyForSend(media model.Media) string {
	return strings.TrimSpace(media.Type)
}

func isQQEndpointOnlyTargetForAction(target string) bool {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		return false
	}
	return !strings.ContainsAny(trimmed, " \t\r\n")
}

func severityLabel(severity int) string {
	switch severity {
	case 5:
		return "Disaster"
	case 4:
		return "High"
	case 3:
		return "Average"
	case 2:
		return "Warning"
	case 1:
		return "Information"
	default:
		return "Not Classified"
	}
}
