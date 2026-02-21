package service

import (
	"encoding/json"
	"errors"
	"fmt"

	"nagare/internal/model"
)

// ToolDefinition represents a tool that can be called by the LLM.
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ListTools returns all available read-only tools.
func ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "get_alerts",
			Description: "List alerts with optional filters.",
			InputSchema: schemaObject(map[string]interface{}{
				"q":        schemaString("Search in alert message."),
				"severity": schemaInt("Alert severity."),
				"status":   schemaInt("Alert status."),
				"alarm_id": schemaInt("Filter by alarm id."),
				"host_id":  schemaInt("Filter by host id."),
				"item_id":  schemaInt("Filter by item id."),
				"limit":    schemaInt("Max results (default 100)."),
				"offset":   schemaInt("Offset for pagination."),
			}),
		},
		{
			Name:        "get_hosts",
			Description: "List hosts with optional filters.",
			InputSchema: schemaObject(map[string]interface{}{
				"q":        schemaString("Search by name, hostid, ip, description."),
				"status":   schemaInt("Host status."),
				"m_id":     schemaInt("Monitor id."),
				"group_id": schemaInt("Group id."),
				"ip_addr":  schemaString("Filter by IP address."),
				"limit":    schemaInt("Max results (default 100)."),
				"offset":   schemaInt("Offset for pagination."),
			}),
		},
		{
			Name:        "get_items",
			Description: "List items with optional filters.",
			InputSchema: schemaObject(map[string]interface{}{
				"q":          schemaString("Search by name or ids."),
				"hid":        schemaInt("Host id."),
				"value_type": schemaString("Item value type."),
				"status":     schemaInt("Item status."),
				"hostid":     schemaString("External host id."),
				"itemid":     schemaString("External item id."),
				"limit":      schemaInt("Max results (default 100)."),
				"offset":     schemaInt("Offset for pagination."),
			}),
		},
		{
			Name:        "get_groups",
			Description: "List groups with optional filters.",
			InputSchema: schemaObject(map[string]interface{}{
				"q":      schemaString("Search by name or description."),
				"status": schemaInt("Group status."),
				"limit":  schemaInt("Max results (default 100)."),
				"offset": schemaInt("Offset for pagination."),
			}),
		},
		{
			Name:        "get_monitors",
			Description: "List monitors with optional filters.",
			InputSchema: schemaObject(map[string]interface{}{
				"q":      schemaString("Search by name, url, or description."),
				"type":   schemaString("Monitor type (zabbix, snmp, etc.)."),
				"status": schemaInt("Monitor status."),
				"limit":  schemaInt("Max results (default 100)."),
				"offset": schemaInt("Offset for pagination."),
			}),
		},
		{
			Name:        "get_providers",
			Description: "List providers with optional filters.",
			InputSchema: schemaObject(map[string]interface{}{
				"q":      schemaString("Search by name, url, or description."),
				"type":   schemaInt("Provider type."),
				"status": schemaInt("Provider status."),
				"limit":  schemaInt("Max results (default 100)."),
				"offset": schemaInt("Offset for pagination."),
			}),
		},
		{
			Name:        "get_actions",
			Description: "List actions with optional filters.",
			InputSchema: schemaObject(map[string]interface{}{
				"q":      schemaString("Search by name, template, or description."),
				"status": schemaInt("Action status."),
				"limit":  schemaInt("Max results (default 100)."),
				"offset": schemaInt("Offset for pagination."),
			}),
		},
		{
			Name:        "get_triggers",
			Description: "List triggers with optional filters.",
			InputSchema: schemaObject(map[string]interface{}{
				"q":                schemaString("Search by name."),
				"status":           schemaInt("Trigger status."),
				"severity_min":     schemaInt("Minimum severity."),
				"entity":           schemaString("Entity type (alert, log)."),
				"alert_id":         schemaInt("Alert id."),
				"alert_monitor_id": schemaInt("Alert monitor id."),
				"alert_group_id":   schemaInt("Alert group id."),
				"alert_host_id":    schemaInt("Alert host id."),
				"alert_item_id":    schemaInt("Alert item id."),
				"limit":            schemaInt("Max results (default 100)."),
				"offset":           schemaInt("Offset for pagination."),
			}),
		},
		{
			Name:        "get_media",
			Description: "List media with optional filters.",
			InputSchema: schemaObject(map[string]interface{}{
				"q":      schemaString("Search by name, type, target, description."),
				"status": schemaInt("Media status."),
				"type":   schemaString("Media type (email, other, qq, etc.)."),
				"limit":  schemaInt("Max results (default 100)."),
				"offset": schemaInt("Offset for pagination."),
			}),
		},
		{
			Name:        "get_logs",
			Description: "List logs with optional filters.",
			InputSchema: schemaObject(map[string]interface{}{
				"type":     schemaString("Log type (system, service)."),
				"severity": schemaInt("Log severity (0=info, 1=warn, 2=error)."),
				"q":        schemaString("Search in message or context."),
				"limit":    schemaInt("Max results (default 100)."),
				"offset":   schemaInt("Offset for pagination."),
			}),
		},
		{
			Name:        "get_users",
			Description: "List users with optional filters.",
			InputSchema: schemaObject(map[string]interface{}{
				"q":          schemaString("Search by username."),
				"privileges": schemaInt("Privileges level."),
				"status":     schemaInt("User status."),
				"limit":      schemaInt("Max results (default 100)."),
				"offset":     schemaInt("Offset for pagination."),
			}),
		},
		{
			Name:        "get_register_applications",
			Description: "List registration applications with optional filters.",
			InputSchema: schemaObject(map[string]interface{}{
				"q":      schemaString("Search by username."),
				"status": schemaInt("Application status."),
				"limit":  schemaInt("Max results (default 100)."),
				"offset": schemaInt("Offset for pagination."),
			}),
		},
		{
			Name:        "get_chats",
			Description: "List chat messages with optional filters.",
			InputSchema: schemaObject(map[string]interface{}{
				"q":           schemaString("Search by content."),
				"role":        schemaString("Role (user, assistant)."),
				"provider_id": schemaInt("Provider id."),
				"user_id":     schemaInt("User id."),
				"model":       schemaString("Model name."),
				"limit":       schemaInt("Max results (default 100)."),
				"offset":      schemaInt("Offset for pagination."),
			}),
		},
	}
}

// CallTool executes a tool by name.
func CallTool(name string, rawArgs json.RawMessage) (interface{}, error) {
	switch name {
	case "get_alerts":
		var args alertArgs
		if err := decodeParams(rawArgs, &args); err != nil {
			return nil, err
		}
		if isEmptyArgs(rawArgs) {
			return GetAllAlertsServ()
		}
		limit, offset := withDefaultLimitOffset(args.Limit, args.Offset)
		filter := model.AlertFilter{
			Query:    args.Query,
			Severity: args.Severity,
			Status:   args.Status,
			AlarmID:  args.AlarmID,
			HostID:   args.HostID,
			ItemID:   args.ItemID,
			Limit:    limit,
			Offset:   offset,
		}
		return SearchAlertsServ(filter)
	case "get_hosts":
		var args hostArgs
		if err := decodeParams(rawArgs, &args); err != nil {
			return nil, err
		}
		if isEmptyArgs(rawArgs) {
			return GetAllHostsServ()
		}
		limit, offset := withDefaultLimitOffset(args.Limit, args.Offset)
		mid, err := toUintPtr(args.MonitorID)
		if err != nil {
			return nil, err
		}
		groupID, err := toUintPtr(args.GroupID)
		if err != nil {
			return nil, err
		}
		filter := model.HostFilter{
			Query:   args.Query,
			MID:     mid,
			GroupID: groupID,
			Status:  args.Status,
			IPAddr:  args.IPAddr,
			Limit:   limit,
			Offset:  offset,
		}
		return SearchHostsServ(filter)
	case "get_items":
		var args itemArgs
		if err := decodeParams(rawArgs, &args); err != nil {
			return nil, err
		}
		if isEmptyArgs(rawArgs) {
			return GetAllItemServ()
		}
		limit, offset := withDefaultLimitOffset(args.Limit, args.Offset)
		hid, err := toUintPtr(args.HostID)
		if err != nil {
			return nil, err
		}
		filter := model.ItemFilter{
			Query:     args.Query,
			HID:       hid,
			ValueType: args.ValueType,
			Status:    args.Status,
			HostID:    args.ExternalHostID,
			ItemID:    args.ExternalItemID,
			Limit:     limit,
			Offset:    offset,
		}
		return SearchItemsServ(filter)
	case "get_groups":
		var args groupArgs
		if err := decodeParams(rawArgs, &args); err != nil {
			return nil, err
		}
		if isEmptyArgs(rawArgs) {
			return GetAllGroupsServ()
		}
		limit, offset := withDefaultLimitOffset(args.Limit, args.Offset)
		filter := model.GroupFilter{
			Query:  args.Query,
			Status: args.Status,
			Limit:  limit,
			Offset: offset,
		}
		return SearchGroupsServ(filter)
	case "get_monitors":
		var args monitorArgs
		if err := decodeParams(rawArgs, &args); err != nil {
			return nil, err
		}
		if isEmptyArgs(rawArgs) {
			return GetAllMonitorsServ()
		}
		limit, offset := withDefaultLimitOffset(args.Limit, args.Offset)
		filter := model.MonitorFilter{
			Query:  args.Query,
			Type:   args.Type,
			Status: args.Status,
			Limit:  limit,
			Offset: offset,
		}
		return SearchMonitorsServ(filter)
	case "get_providers":
		var args providerArgs
		if err := decodeParams(rawArgs, &args); err != nil {
			return nil, err
		}
		if isEmptyArgs(rawArgs) {
			return GetAllProvidersServ()
		}
		limit, offset := withDefaultLimitOffset(args.Limit, args.Offset)
		filter := model.ProviderFilter{
			Query:  args.Query,
			Type:   args.Type,
			Status: args.Status,
			Limit:  limit,
			Offset: offset,
		}
		return SearchProvidersServ(filter)
	case "get_actions":
		var args actionArgs
		if err := decodeParams(rawArgs, &args); err != nil {
			return nil, err
		}
		if isEmptyArgs(rawArgs) {
			return GetAllActionsServ()
		}
		limit, offset := withDefaultLimitOffset(args.Limit, args.Offset)
		filter := model.ActionFilter{
			Query:  args.Query,
			Status: args.Status,
			Limit:  limit,
			Offset: offset,
		}
		return SearchActionsServ(filter)
	case "get_triggers":
		var args triggerArgs
		if err := decodeParams(rawArgs, &args); err != nil {
			return nil, err
		}
		if isEmptyArgs(rawArgs) {
			return GetAllTriggersServ()
		}
		limit, offset := withDefaultLimitOffset(args.Limit, args.Offset)
		alertID, err := toUintPtr(args.AlertID)
		if err != nil {
			return nil, err
		}
		alertMonitorID, err := toUintPtr(args.AlertMonitorID)
		if err != nil {
			return nil, err
		}
		alertGroupID, err := toUintPtr(args.AlertGroupID)
		if err != nil {
			return nil, err
		}
		alertHostID, err := toUintPtr(args.AlertHostID)
		if err != nil {
			return nil, err
		}
		alertItemID, err := toUintPtr(args.AlertItemID)
		if err != nil {
			return nil, err
		}
		filter := model.TriggerFilter{
			Query:          args.Query,
			Status:         args.Status,
			SeverityMin:    args.SeverityMin,
			Entity:         args.Entity,
			AlertID:        alertID,
			AlertMonitorID: alertMonitorID,
			AlertGroupID:   alertGroupID,
			AlertHostID:    alertHostID,
			AlertItemID:    alertItemID,
			Limit:          limit,
			Offset:         offset,
		}
		return SearchTriggersServ(filter)
	case "get_media":
		var args mediaArgs
		if err := decodeParams(rawArgs, &args); err != nil {
			return nil, err
		}
		if isEmptyArgs(rawArgs) {
			return GetAllMediaServ()
		}
		limit, offset := withDefaultLimitOffset(args.Limit, args.Offset)
		filter := model.MediaFilter{
			Query:  args.Query,
			Status: args.Status,
			Type:   args.Type,
			Limit:  limit,
			Offset: offset,
		}
		return SearchMediaServ(filter)
	case "get_logs":
		var args logArgs
		if err := decodeParams(rawArgs, &args); err != nil {
			return nil, err
		}
		limit, offset := withDefaultLimitOffset(args.Limit, args.Offset)
		filter := model.LogFilter{
			Type:     args.Type,
			Severity: args.Severity,
			Query:    args.Query,
			Limit:    limit,
			Offset:   offset,
		}
		return SearchLogsServ(filter)
	case "get_users":
		var args userArgs
		if err := decodeParams(rawArgs, &args); err != nil {
			return nil, err
		}
		if isEmptyArgs(rawArgs) {
			return GetAllUsersServ()
		}
		limit, offset := withDefaultLimitOffset(args.Limit, args.Offset)
		filter := model.UserFilter{
			Query:      args.Query,
			Privileges: args.Privileges,
			Status:     args.Status,
			Limit:      limit,
			Offset:     offset,
		}
		return SearchUsersServ(filter)
	case "get_register_applications":
		var args registerArgs
		if err := decodeParams(rawArgs, &args); err != nil {
			return nil, err
		}
		if isEmptyArgs(rawArgs) {
			limit, offset := withDefaultLimitOffset(nil, nil)
			filter := model.RegisterApplicationFilter{Limit: limit, Offset: offset}
			return ListRegisterApplicationsServ(filter)
		}
		limit, offset := withDefaultLimitOffset(args.Limit, args.Offset)
		filter := model.RegisterApplicationFilter{
			Query:  args.Query,
			Status: args.Status,
			Limit:  limit,
			Offset: offset,
		}
		return ListRegisterApplicationsServ(filter)
	case "get_chats":
		var args chatArgs
		if err := decodeParams(rawArgs, &args); err != nil {
			return nil, err
		}
		if isEmptyArgs(rawArgs) {
			limit, offset := withDefaultLimitOffset(nil, nil)
			return GetChatsWithPaginationServ(limit, offset)
		}
		limit, offset := withDefaultLimitOffset(args.Limit, args.Offset)
		filter := model.ChatFilter{
			Query:      args.Query,
			Role:       args.Role,
			ProviderID: args.ProviderID,
			UserID:     args.UserID,
			Model:      args.Model,
			Limit:      limit,
			Offset:     offset,
		}
		return SearchChatsServ(filter)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

type alertArgs struct {
	Query    string `json:"q"`
	Severity *int   `json:"severity"`
	Status   *int   `json:"status"`
	AlarmID  *int   `json:"alarm_id"`
	HostID   *int   `json:"host_id"`
	ItemID   *int   `json:"item_id"`
	Limit    *int   `json:"limit"`
	Offset   *int   `json:"offset"`
}

type hostArgs struct {
	Query     string  `json:"q"`
	Status    *int    `json:"status"`
	MonitorID *int    `json:"m_id"`
	GroupID   *int    `json:"group_id"`
	IPAddr    *string `json:"ip_addr"`
	Limit     *int    `json:"limit"`
	Offset    *int    `json:"offset"`
}

type itemArgs struct {
	Query          string  `json:"q"`
	HostID         *int    `json:"hid"`
	ValueType      *string `json:"value_type"`
	Status         *int    `json:"status"`
	ExternalHostID *string `json:"hostid"`
	ExternalItemID *string `json:"itemid"`
	Limit          *int    `json:"limit"`
	Offset         *int    `json:"offset"`
}

type groupArgs struct {
	Query  string `json:"q"`
	Status *int   `json:"status"`
	Limit  *int   `json:"limit"`
	Offset *int   `json:"offset"`
}

type monitorArgs struct {
	Query  string  `json:"q"`
	Type   *string `json:"type"`
	Status *int    `json:"status"`
	Limit  *int    `json:"limit"`
	Offset *int    `json:"offset"`
}

type providerArgs struct {
	Query  string `json:"q"`
	Type   *int   `json:"type"`
	Status *int   `json:"status"`
	Limit  *int   `json:"limit"`
	Offset *int   `json:"offset"`
}

type actionArgs struct {
	Query  string `json:"q"`
	Status *int   `json:"status"`
	Limit  *int   `json:"limit"`
	Offset *int   `json:"offset"`
}

type triggerArgs struct {
	Query          string  `json:"q"`
	Status         *int    `json:"status"`
	SeverityMin    *int    `json:"severity_min"`
	Entity         *string `json:"entity"`
	AlertID        *int    `json:"alert_id"`
	AlertMonitorID *int    `json:"alert_monitor_id"`
	AlertGroupID   *int    `json:"alert_group_id"`
	AlertHostID    *int    `json:"alert_host_id"`
	AlertItemID    *int    `json:"alert_item_id"`
	Limit          *int    `json:"limit"`
	Offset         *int    `json:"offset"`
}

type mediaArgs struct {
	Query  string  `json:"q"`
	Status *int    `json:"status"`
	Type   *string `json:"type"`
	Limit  *int    `json:"limit"`
	Offset *int    `json:"offset"`
}

type logArgs struct {
	Type     string `json:"type"`
	Severity *int   `json:"severity"`
	Query    string `json:"q"`
	Limit    *int   `json:"limit"`
	Offset   *int   `json:"offset"`
}

type userArgs struct {
	Query      string `json:"q"`
	Privileges *int   `json:"privileges"`
	Status     *int   `json:"status"`
	Limit      *int   `json:"limit"`
	Offset     *int   `json:"offset"`
}

type registerArgs struct {
	Query  string `json:"q"`
	Status *int   `json:"status"`
	Limit  *int   `json:"limit"`
	Offset *int   `json:"offset"`
}

type chatArgs struct {
	Query      string  `json:"q"`
	Role       *string `json:"role"`
	ProviderID *int    `json:"provider_id"`
	UserID     *int    `json:"user_id"`
	Model      *string `json:"model"`
	Limit      *int    `json:"limit"`
	Offset     *int    `json:"offset"`
}

func decodeParams(raw json.RawMessage, target interface{}) error {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	return nil
}

func isEmptyArgs(raw json.RawMessage) bool {
	if len(raw) == 0 || string(raw) == "null" {
		return true
	}
	trimmed := string(raw)
	return trimmed == "{}"
}

func toUintPtr(value *int) (*uint, error) {
	if value == nil {
		return nil, nil
	}
	if *value < 0 {
		return nil, errors.New("value must be >= 0")
	}
	v := uint(*value)
	return &v, nil
}

func withDefaultLimitOffset(limit *int, offset *int) (int, int) {
	l := 100
	o := 0
	if limit != nil {
		l = *limit
	}
	if offset != nil {
		o = *offset
	}
	return l, o
}

func schemaObject(properties map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": false,
	}
}

func schemaString(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"description": description,
	}
}

func schemaInt(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "integer",
		"description": description,
	}
}
