package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	mediaSvc "nagare/internal/adapter/external/media"
	"nagare/internal/adapter/repository"
	"nagare/internal/core/domain"
)

func init() {
	mediaSvc.GlobalQQWSManager.CommandHandler = func(message string, qqID string, isGroup bool) (string, error) {
		// Whitelist check
		if strings.HasPrefix(strings.TrimSpace(message), "/") {
			if !CheckQQWhitelistForCommand(qqID, isGroup) {
				return "You are not authorized to execute commands.", nil
			}
		}

		result, err := HandleIMCommand(message)
		if err != nil {
			return "", err
		}
		return result.Reply, nil
	}
}

type IMCommandResult struct {
	Reply string `json:"reply"`
}

// IMCommandContext contains context for executing IM commands
type IMCommandContext struct {
	MediaType  string
	UserID     string
	GroupID    string
	ProviderID uint
}

type imCommand struct {
	Name        string
	Aliases     []string
	Usage       string
	Description string
	Handler     func(args []string, rawArgs string) (IMCommandResult, error)
}

// CheckQQWhitelistForCommand checks if a QQ user/group is allowed to execute commands
func CheckQQWhitelistForCommand(qqID string, isGroup bool) bool {
	result := checkQQWhitelist(qqID, isGroup, true)
	logLevel := "info"
	if result {
		logLevel = "debug"
	}
	LogService(logLevel, "QQ command whitelist decision", map[string]interface{}{
		"qq_id":    qqID,
		"is_group": isGroup,
		"allowed":  result,
	}, nil, "")
	return result
}

// CheckQQWhitelistForAlert checks if a QQ user/group is allowed to receive alerts
func CheckQQWhitelistForAlert(qqID string, isGroup bool) bool {
	return checkQQWhitelist(qqID, isGroup, false)
}

func checkQQWhitelist(qqID string, isGroup bool, isCommand bool) bool {
	if strings.TrimSpace(qqID) == "" {
		LogService("warn", "whitelist check: empty qqID", nil, nil, "")
		return false
	}

	// Determine whitelist type based on message source
	whitelistType := 0 // user
	if isGroup {
		whitelistType = 1 // group
	}

	LogService("info", "whitelist check started", map[string]interface{}{
		"qqID":          qqID,
		"isGroup":       isGroup,
		"whitelistType": whitelistType,
		"isCommand":     isCommand,
	}, nil, "")

	// Check the appropriate whitelist entry
	whitelist, err := getQQWhitelist(qqID, whitelistType)
	if err != nil {
		LogService("warn", "whitelist lookup failed", map[string]interface{}{
			"qqID":      qqID,
			"type":      whitelistType,
			"error":     err.Error(),
			"errorType": fmt.Sprintf("%T", err),
		}, nil, "")
		return false
	}

	if whitelist == nil {
		LogService("warn", "QQ ID not in whitelist (nil)", map[string]interface{}{
			"qqID": qqID,
			"type": whitelistType,
		}, nil, "")
		return false
	}

	LogService("info", "whitelist entry found", map[string]interface{}{
		"qqID":        qqID,
		"type":        whitelistType,
		"enabled":     whitelist.Enabled,
		"can_command": whitelist.CanCommand,
		"can_receive": whitelist.CanReceive,
		"nickname":    whitelist.Nickname,
	}, nil, "")

	// Check if whitelist entry is enabled
	if whitelist.Enabled == 0 {
		LogService("info", "whitelist entry disabled", map[string]interface{}{
			"qqID": qqID,
			"type": whitelistType,
		}, nil, "")
		return false
	}

	// Check appropriate permission flag
	if isCommand {
		allowed := whitelist.CanCommand == 1
		LogService("info", "whitelist command check", map[string]interface{}{
			"qqID":        qqID,
			"type":        whitelistType,
			"can_command": whitelist.CanCommand,
			"allowed":     allowed,
		}, nil, "")
		return allowed
	}

	allowed := whitelist.CanReceive == 1
	LogService("info", "whitelist alert check", map[string]interface{}{
		"qqID":        qqID,
		"type":        whitelistType,
		"can_receive": whitelist.CanReceive,
		"allowed":     allowed,
	}, nil, "")
	return allowed
}

func getQQWhitelist(qqID string, whitelistType int) (*domain.QQWhitelist, error) {
	whitelist, err := repository.GetQQWhitelistDAO(qqID, whitelistType)
	if err != nil {
		return nil, err
	}
	return &whitelist, nil
}

// HandleIMCommand processes incoming IM commands
func HandleIMCommand(message string) (IMCommandResult, error) {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return IMCommandResult{Reply: buildHelpReply()}, nil
	}

	cmd, args, rawArgs := parseIMCommand(trimmed)
	if cmd == "" {
		return IMCommandResult{Reply: buildHelpReply()}, nil
	}

	commandMap := buildIMCommandMap()
	command, ok := commandMap[cmd]
	if !ok {
		return IMCommandResult{Reply: "Unsupported command. Try /help to see available commands."}, nil
	}

	if len(args) > 0 {
		arg0 := strings.ToLower(strings.TrimSpace(args[0]))
		if arg0 == "help" || arg0 == "--help" || arg0 == "-h" {
			return IMCommandResult{Reply: fmt.Sprintf("%s\n%s", command.Usage, command.Description)}, nil
		}
	}

	return command.Handler(args, rawArgs)
}

// handleGetAlerts retrieves active alerts
func handleGetAlertsCommand(args []string, rawArgs string) (IMCommandResult, error) {
	status := intPtr(0)
	limit := 10

	options, flags := parseIMOptions(args)
	for _, flag := range flags {
		if flag == "all" {
			status = nil
		} else if flag == "active" {
			status = intPtr(0)
		} else if flag == "resolved" {
			status = intPtr(1)
		}
	}

	if val, ok := options["status"]; ok {
		if parsed, err := strconv.Atoi(val); err == nil {
			status = intPtr(parsed)
		}
	}
	if val, ok := options["limit"]; ok {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	filter := domain.AlertFilter{
		Query:  options["q"],
		Status: status,
		Limit:  limit,
		Offset: 0,
	}

	if val, ok := options["severity"]; ok {
		if parsed, err := strconv.Atoi(val); err == nil {
			filter.Severity = intPtr(parsed)
		}
	}

	alerts, err := SearchAlertsServ(filter)
	if err != nil {
		return IMCommandResult{Reply: fmt.Sprintf("Error retrieving alerts: %v", err)}, nil
	}

	if len(alerts) == 0 {
		return IMCommandResult{Reply: "No alerts found."}, nil
	}

	var reply strings.Builder
	reply.WriteString(fmt.Sprintf("Alerts (%d):\n", len(alerts)))
	for i, alert := range alerts {
		reply.WriteString(fmt.Sprintf("[%d] %s (Severity: %d)\n", i+1, alert.Message, alert.Severity))
	}

	return IMCommandResult{Reply: reply.String()}, nil
}

func handleHostsCommand(args []string, rawArgs string) (IMCommandResult, error) {
	options, _ := parseIMOptions(args)
	limit := 10
	if val, ok := options["limit"]; ok {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	filter := domain.HostFilter{
		Query:  options["q"],
		Limit:  limit,
		Offset: 0,
	}
	if val, ok := options["status"]; ok {
		if parsed, err := strconv.Atoi(val); err == nil {
			filter.Status = intPtr(parsed)
		}
	}
	if val, ok := options["monitor"]; ok {
		if parsed, err := strconv.ParseUint(val, 10, 64); err == nil {
			parsedUint := uint(parsed)
			filter.MID = &parsedUint
		}
	}
	if val, ok := options["group"]; ok {
		if parsed, err := strconv.ParseUint(val, 10, 64); err == nil {
			parsedUint := uint(parsed)
			filter.GroupID = &parsedUint
		}
	}
	if val, ok := options["ip"]; ok {
		filter.IPAddr = &val
	}

	hosts, err := SearchHostsServ(filter)
	if err != nil {
		return IMCommandResult{Reply: fmt.Sprintf("Error retrieving hosts: %v", err)}, nil
	}
	if len(hosts) == 0 {
		return IMCommandResult{Reply: "No hosts found."}, nil
	}
	var reply strings.Builder
	reply.WriteString(fmt.Sprintf("Hosts (%d):\n", len(hosts)))
	for i, host := range hosts {
		reply.WriteString(fmt.Sprintf("[%d] %s (Status: %d, IP: %s)\n", i+1, host.Name, host.Status, host.IPAddr))
	}
	return IMCommandResult{Reply: reply.String()}, nil
}

func handleMonitorsCommand(args []string, rawArgs string) (IMCommandResult, error) {
	options, _ := parseIMOptions(args)
	limit := 10
	if val, ok := options["limit"]; ok {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	filter := domain.MonitorFilter{
		Query:  options["q"],
		Limit:  limit,
		Offset: 0,
	}
	if val, ok := options["status"]; ok {
		if parsed, err := strconv.Atoi(val); err == nil {
			filter.Status = intPtr(parsed)
		}
	}
	if val, ok := options["type"]; ok {
		filter.Type = &val
	}

	monitors, err := SearchMonitorsServ(filter)
	if err != nil {
		return IMCommandResult{Reply: fmt.Sprintf("Error retrieving monitors: %v", err)}, nil
	}
	if len(monitors) == 0 {
		return IMCommandResult{Reply: "No monitors found."}, nil
	}
	var reply strings.Builder
	reply.WriteString(fmt.Sprintf("Monitors (%d):\n", len(monitors)))
	for i, monitor := range monitors {
		reply.WriteString(fmt.Sprintf("[%d] %s (Type: %d, Status: %d)\n", i+1, monitor.Name, monitor.Type, monitor.Status))
	}
	return IMCommandResult{Reply: reply.String()}, nil
}

func handleGroupsCommand(args []string, rawArgs string) (IMCommandResult, error) {
	options, _ := parseIMOptions(args)
	limit := 10
	if val, ok := options["limit"]; ok {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	filter := domain.GroupFilter{
		Query:  options["q"],
		Limit:  limit,
		Offset: 0,
	}
	if val, ok := options["status"]; ok {
		if parsed, err := strconv.Atoi(val); err == nil {
			filter.Status = intPtr(parsed)
		}
	}

	groups, err := SearchGroupsServ(filter)
	if err != nil {
		return IMCommandResult{Reply: fmt.Sprintf("Error retrieving groups: %v", err)}, nil
	}
	if len(groups) == 0 {
		return IMCommandResult{Reply: "No groups found."}, nil
	}
	var reply strings.Builder
	reply.WriteString(fmt.Sprintf("Groups (%d):\n", len(groups)))
	for i, group := range groups {
		reply.WriteString(fmt.Sprintf("[%d] %s (Status: %d)\n", i+1, group.Name, group.Status))
	}
	return IMCommandResult{Reply: reply.String()}, nil
}

func handleItemsCommand(args []string, rawArgs string) (IMCommandResult, error) {
	options, _ := parseIMOptions(args)
	limit := 10
	if val, ok := options["limit"]; ok {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	filter := domain.ItemFilter{
		Query:  options["q"],
		Limit:  limit,
		Offset: 0,
	}
	if val, ok := options["status"]; ok {
		if parsed, err := strconv.Atoi(val); err == nil {
			filter.Status = intPtr(parsed)
		}
	}
	if val, ok := options["host_id"]; ok {
		filter.HostID = &val
	}
	if val, ok := options["item_id"]; ok {
		filter.ItemID = &val
	}

	items, err := SearchItemsServ(filter)
	if err != nil {
		return IMCommandResult{Reply: fmt.Sprintf("Error retrieving items: %v", err)}, nil
	}
	if len(items) == 0 {
		return IMCommandResult{Reply: "No items found."}, nil
	}
	var reply strings.Builder
	reply.WriteString(fmt.Sprintf("Items (%d):\n", len(items)))
	for i, item := range items {
		reply.WriteString(fmt.Sprintf("[%d] %s (Status: %d, Value: %s)\n", i+1, item.Name, item.Status, item.Value))
	}
	return IMCommandResult{Reply: reply.String()}, nil
}

func handleLogsCommand(args []string, rawArgs string) (IMCommandResult, error) {
	options, _ := parseIMOptions(args)
	limit := 10
	if val, ok := options["limit"]; ok {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	filter := domain.LogFilter{
		Type:   options["type"],
		Query:  options["q"],
		Limit:  limit,
		Offset: 0,
	}
	if val, ok := options["severity"]; ok {
		if parsed, err := strconv.Atoi(val); err == nil {
			filter.Severity = intPtr(parsed)
		}
	}

	logs, err := SearchLogsServ(filter)
	if err != nil {
		return IMCommandResult{Reply: fmt.Sprintf("Error retrieving logs: %v", err)}, nil
	}
	if len(logs) == 0 {
		return IMCommandResult{Reply: "No logs found."}, nil
	}
	var reply strings.Builder
	reply.WriteString(fmt.Sprintf("Logs (%d):\n", len(logs)))
	for i, log := range logs {
		reply.WriteString(fmt.Sprintf("[%d] %s (Severity: %d)\n", i+1, log.Message, log.Severity))
	}
	return IMCommandResult{Reply: reply.String()}, nil
}

// handleChatCommand processes chat messages with LLM
func handleChatCommand(content string) (IMCommandResult, error) {
	// Get the default provider (first one)
	providers, err := GetAllProvidersServ()
	if err != nil || len(providers) == 0 {
		return IMCommandResult{Reply: "No LLM provider configured."}, nil
	}

	// Use the first available provider
	providerID := uint(providers[0].ID)

	chatReq := ChatReq{
		ProviderID: providerID,
		Content:    content,
		UseTools:   nil, // No tools for IM chat
		Privileges: 1,
	}

	// Get the model from provider
	if providers[0].DefaultModel != "" {
		chatReq.Model = providers[0].DefaultModel
	}

	resp, err := SendChatServ(chatReq)
	if err != nil {
		return IMCommandResult{Reply: fmt.Sprintf("Chat error: %v", err)}, nil
	}

	return IMCommandResult{Reply: resp.Content}, nil
}

func handleStatusCommand(args []string, rawArgs string) (IMCommandResult, error) {
	score, err := GetHealthScoreServ()
	if err != nil {
		return IMCommandResult{}, err
	}
	reply := fmt.Sprintf("Health Score: %d (monitors %d/%d, hosts %d/%d, items %d/%d)",
		score.Score,
		score.MonitorActive, score.MonitorTotal,
		score.HostActive, score.HostTotal,
		score.ItemActive, score.ItemTotal,
	)
	return IMCommandResult{Reply: reply}, nil
}

func handleChatWrapper(args []string, rawArgs string) (IMCommandResult, error) {
	content := strings.TrimSpace(rawArgs)
	if content == "" {
		return IMCommandResult{Reply: "Usage: /chat <message>"}, nil
	}
	return handleChatCommand(content)
}

func handleHelpCommand(args []string, rawArgs string) (IMCommandResult, error) {
	return IMCommandResult{Reply: buildHelpReply()}, nil
}

func parseIMCommand(message string) (cmd string, args []string, rawArgs string) {
	trimmed := strings.TrimSpace(message)
	if strings.HasPrefix(trimmed, "/") {
		trimmed = strings.TrimSpace(trimmed[1:])
	}
	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return "", nil, ""
	}
	cmd = strings.ToLower(fields[0])
	if len(fields) > 1 {
		args = fields[1:]
	}
	rawArgs = strings.TrimSpace(trimmed[len(fields[0]):])
	return cmd, args, rawArgs
}

func parseIMOptions(args []string) (map[string]string, []string) {
	options := make(map[string]string)
	flags := make([]string, 0, len(args))
	for _, arg := range args {
		value := strings.TrimSpace(arg)
		if value == "" {
			continue
		}
		if strings.Contains(value, "=") {
			parts := strings.SplitN(value, "=", 2)
			key := strings.ToLower(strings.TrimSpace(parts[0]))
			val := strings.TrimSpace(parts[1])
			options[key] = val
			continue
		}
		flags = append(flags, strings.ToLower(value))
	}
	return options, flags
}

func buildIMCommandMap() map[string]imCommand {
	commands := []imCommand{
		{
			Name:        "help",
			Aliases:     []string{"h", "?"},
			Usage:       "/help",
			Description: "Show available commands.",
			Handler:     handleHelpCommand,
		},
		{
			Name:        "status",
			Aliases:     []string{"health"},
			Usage:       "/status",
			Description: "Show system health summary.",
			Handler:     handleStatusCommand,
		},
		{
			Name:        "alerts",
			Aliases:     []string{"get_alert", "get_alerts"},
			Usage:       "/alerts [active|resolved|all] [severity=2] [limit=10] [q=keyword]",
			Description: "List alerts with optional filters.",
			Handler:     handleGetAlertsCommand,
		},
		{
			Name:        "hosts",
			Aliases:     []string{"get_hosts"},
			Usage:       "/hosts [q=keyword] [status=1] [monitor=1] [group=2] [ip=1.2.3.4] [limit=10]",
			Description: "List hosts with optional filters.",
			Handler:     handleHostsCommand,
		},
		{
			Name:        "monitors",
			Aliases:     []string{"get_monitors"},
			Usage:       "/monitors [q=keyword] [status=1] [type=zabbix] [limit=10]",
			Description: "List monitors with optional filters.",
			Handler:     handleMonitorsCommand,
		},
		{
			Name:        "groups",
			Aliases:     []string{"get_groups"},
			Usage:       "/groups [q=keyword] [status=1] [limit=10]",
			Description: "List groups with optional filters.",
			Handler:     handleGroupsCommand,
		},
		{
			Name:        "items",
			Aliases:     []string{"get_items"},
			Usage:       "/items [q=keyword] [status=1] [host_id=123] [item_id=456] [limit=10]",
			Description: "List items with optional filters.",
			Handler:     handleItemsCommand,
		},
		{
			Name:        "logs",
			Aliases:     []string{"get_logs"},
			Usage:       "/logs [type=system] [severity=2] [q=keyword] [limit=10]",
			Description: "List logs with optional filters.",
			Handler:     handleLogsCommand,
		},
		{
			Name:        "chat",
			Aliases:     []string{"ask"},
			Usage:       "/chat <message>",
			Description: "Chat with the configured AI provider.",
			Handler:     handleChatWrapper,
		},
	}

	commandMap := make(map[string]imCommand, len(commands)*2)
	for _, command := range commands {
		commandMap[command.Name] = command
		for _, alias := range command.Aliases {
			commandMap[alias] = command
		}
	}
	return commandMap
}

func buildHelpReply() string {
	commands := []string{
		"/help - Show available commands.",
		"/status - System health summary.",
		"/alerts [active|resolved|all] [severity=2] [limit=10] [q=keyword]",
		"/hosts [q=keyword] [status=1] [monitor=1] [group=2] [ip=1.2.3.4] [limit=10]",
		"/monitors [q=keyword] [status=1] [type=zabbix] [limit=10]",
		"/groups [q=keyword] [status=1] [limit=10]",
		"/items [q=keyword] [status=1] [host_id=123] [item_id=456] [limit=10]",
		"/logs [type=system] [severity=2] [q=keyword] [limit=10]",
		"/chat <message> - Chat with AI.",
	}
	return "Commands:\n" + strings.Join(commands, "\n")
}

func intPtr(value int) *int {
	return &value
}

// HandleIMCommandWithContext processes IM commands with media context
func HandleIMCommandWithContext(message string, ctx IMCommandContext) (IMCommandResult, error) {
	return HandleIMCommand(message)
}

func SendIMReply(mediaType, target, message string) error {
	if strings.TrimSpace(mediaType) == "" || strings.TrimSpace(target) == "" {
		return nil
	}
	return mediaSvc.GetService().SendMessage(context.Background(), strings.ToLower(mediaType), target, message)
}
