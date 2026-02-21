package service

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"nagare/internal/model"
	"nagare/internal/repository"
	"nagare/internal/repository/llm"
)

// AlertSeverity represents alert severity level
type AlertSeverity int

const (
	SeverityInfo     AlertSeverity = 0
	SeverityWarning  AlertSeverity = 1
	SeverityCritical AlertSeverity = 2
)

// AlertReq represents an alert request
type AlertReq struct {
	Message  string `json:"message" binding:"required"`
	Severity int    `json:"severity"`
	Status   int    `json:"status"`
	HostID   uint   `json:"host_id"`
	ItemID   uint   `json:"item_id"`
	AlarmID  uint   `json:"alarm_id"`
	Comment  string `json:"comment"`
}

// AlertRes represents an alert response
type AlertRes struct {
	ID       int    `json:"id"`
	Message  string `json:"message"`
	Severity int    `json:"severity"`
	Status   int    `json:"status"`
	HostID   uint   `json:"host_id"`
	ItemID   uint   `json:"item_id"`
	AlarmID  uint   `json:"alarm_id"`
	Comment  string `json:"comment"`
}

// GetAllAlertsServ retrieves all alerts
func GetAllAlertsServ() ([]AlertRes, error) {
	alerts, err := repository.GetAllAlertsDAO()
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}
	var alertResponses []AlertRes
	for _, alert := range alerts {
		alertResponses = append(alertResponses, AlertRes{
			ID:       int(alert.ID),
			Message:  alert.Message,
			Severity: alert.Severity,
			Status:   alert.Status,
			AlarmID:  alert.AlarmID,
			HostID:   alert.HostID,
			ItemID:   alert.ItemID,
			Comment:  alert.Comment,
		})
	}
	return alertResponses, nil
}

// SearchAlertsServ retrieves alerts by filter
func SearchAlertsServ(filter model.AlertFilter) ([]AlertRes, error) {
	alerts, err := repository.SearchAlertsDAO(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search alerts: %w", err)
	}
	responses := make([]AlertRes, 0, len(alerts))
	for _, alert := range alerts {
		responses = append(responses, AlertRes{
			ID:       int(alert.ID),
			Message:  alert.Message,
			Severity: alert.Severity,
			Status:   alert.Status,
			AlarmID:  alert.AlarmID,
			HostID:   alert.HostID,
			ItemID:   alert.ItemID,
			Comment:  alert.Comment,
		})
	}
	return responses, nil
}

// CountAlertsServ returns total count for alerts by filter
func CountAlertsServ(filter model.AlertFilter) (int64, error) {
	return repository.CountAlertsDAO(filter)
}

// GetAlertByIDServ retrieves an alert by ID
func GetAlertByIDServ(id int) (AlertRes, error) {
	alert, err := repository.GetAlertByIDDAO(id)
	if err != nil {
		return AlertRes{}, fmt.Errorf("failed to get alert by ID: %w", err)
	}
	return AlertRes{
		ID:       int(alert.ID),
		Message:  alert.Message,
		Severity: alert.Severity,
		Status:   alert.Status,
		AlarmID:  alert.AlarmID,
		HostID:   alert.HostID,
		ItemID:   alert.ItemID,
		Comment:  alert.Comment,
	}, nil
}

// AddAlertServ creates a new alert
func AddAlertServ(req AlertReq) error {
	alert := model.Alert{
		Message:  req.Message,
		Severity: req.Severity,
		Status:   0,
		AlarmID:  req.AlarmID,
		HostID:   req.HostID,
		ItemID:   req.ItemID,
		Comment:  req.Comment,
	}
	if err := repository.AddAlertDAO(&alert); err != nil {
		return err
	}

	// Trigger Site Message
	_ = CreateSiteMessageServ("New Alert Detected", alert.Message, "alert", alert.Severity, nil)

	go analyzeAndNotifyAlert(alert)
	return nil
}

func analyzeAndNotifyAlert(alert model.Alert) {
	if !aiAnalysisEnabled() {
		ExecuteActionsForAlert(alert)
		return
	}
	if alert.Severity < aiAnalysisMinSeverity() {
		LogService("info", "alert analysis skipped", map[string]interface{}{"alert_id": alert.ID, "severity": alert.Severity, "min_severity": aiAnalysisMinSeverity()}, nil, "")
		ExecuteActionsForAlert(alert)
		return
	}

	analysis, err := analyzeAlertWithAI(alert)
	if err != nil {
		LogService("warn", "alert analysis skipped", map[string]interface{}{"alert_id": alert.ID, "error": err.Error()}, nil, "")
		ExecuteActionsForAlert(alert)
		return
	}

	comment := mergeAlertComment(alert.Comment, analysis)
	if err := repository.UpdateAlertCommentDAO(int(alert.ID), comment); err != nil {
		LogService("warn", "alert analysis not saved", map[string]interface{}{"alert_id": alert.ID, "error": err.Error()}, nil, "")
	}
	alert.Comment = comment

	// AI Notification Guard check
	if aiNotificationGuardEnabled() {
		shouldNotify, reasoning := parseAIAlertDecision(analysis)
		if !shouldNotify {
			LogService("info", "alert notification suppressed by AI", map[string]interface{}{
				"alert_id":  alert.ID,
				"reasoning": reasoning,
			}, nil, "")
			return
		}
		LogService("info", "alert notification approved by AI", map[string]interface{}{
			"alert_id":  alert.ID,
			"reasoning": reasoning,
		}, nil, "")
	}

	ExecuteActionsForAlert(alert)
}

func parseAIAlertDecision(analysis string) (bool, string) {
	lines := strings.Split(analysis, "\n")
	decisionLine := ""
	for _, line := range lines {
		if strings.Contains(strings.ToUpper(line), "DECISION:") {
			decisionLine = line
			break
		}
	}

	if decisionLine == "" {
		// Default to true if decision not found
		return true, "no clear decision found in analysis"
	}

	upperDecision := strings.ToUpper(decisionLine)
	if strings.Contains(upperDecision, "[SUPPRESS]") || strings.Contains(upperDecision, "SUPPRESS") {
		return false, decisionLine
	}

	return true, decisionLine
}

func analyzeAlertWithAI(alert model.Alert) (string, error) {
	providerID, model := aiProviderConfig()
	client, resolvedModel, err := createLLMClient(providerID, model)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(providerID, 2)
		return "", err
	}

	hostName, hostIP, itemName := "", "", ""
	if alert.HostID > 0 {
		if host, err := repository.GetHostByIDDAO(alert.HostID); err == nil {
			hostName = host.Name
			hostIP = host.IPAddr
		}
	}
	if alert.ItemID > 0 {
		if item, err := repository.GetItemByIDDAO(alert.ItemID); err == nil {
			itemName = item.Name
		}
	}

	ctx, cancel := aiAnalysisContext()
	defer cancel()
	start := time.Now()

	// Use RAG to fetch context from local knowledge base
	localContext := RetrieveContext(alert.Message)

	alertData := fmt.Sprintf(
		"Alert ID: %d\nAlarm ID: %d\nHost ID: %d\nHost Name: %s\nHost IP: %s\nItem ID: %d\nItem Name: %s\nSeverity: %d\nStatus: %d\nCreated At: %s\nMessage: %s\nComment: %s%s",
		alert.ID,
		alert.AlarmID,
		alert.HostID,
		sanitizeSensitiveText(hostName),
		sanitizeSensitiveText(hostIP),
		alert.ItemID,
		sanitizeSensitiveText(itemName),
		alert.Severity,
		alert.Status,
		alert.CreatedAt.Format(time.RFC3339),
		sanitizeSensitiveText(alert.Message),
		sanitizeSensitiveText(alert.Comment),
		localContext,
	)

	resp, err := client.Chat(ctx, llm.ChatRequest{
		Model:        resolvedModel,
		SystemPrompt: alertAnalysisPrompt(),
		Messages: []llm.Message{
			{Role: "user", Content: alertData},
		},
	})
	logLLMRequest("alert_analysis", providerID, resolvedModel, time.Since(start), err)
	if err != nil {
		_ = repository.UpdateProviderStatusDAO(providerID, 2)
		return "", err
	}
	_ = repository.UpdateProviderStatusDAO(providerID, 1)
	return strings.TrimSpace(resp.Content), nil
}

func mergeAlertComment(existing, analysis string) string {
	trimmed := strings.TrimSpace(existing)
	if trimmed == "" {
		return analysis
	}
	return trimmed + "\n\nAI Analysis:\n" + analysis
}

func alertAnalysisPrompt() string {
	return "You are an expert system administrator and DevOps engineer.\n" +
		"Analyze the alert data and produce a concise, actionable assessment.\n\n" +
		"Rules:\n" +
		"- Use only the provided data; do not invent metrics or events.\n" +
		"- If data is missing, state what is missing and how it affects confidence.\n" +
		"- Severity mapping: severity 0-1=Normal, 2=Warning, 3+=Critical.\n\n" +
		"Decision Requirement:\n" +
		"- You must conclude with a clear decision on whether to notify a human user.\n" +
		"- Decisions: [NOTIFY] or [SUPPRESS].\n" +
		"- Suppress if the alert is a known false positive, duplicate, or trivial noise.\n" +
		"- Notify if the alert requires immediate or near-term human attention.\n\n" +
		"Output format (use headings):\n" +
		"Summary:\n" +
		"- What the alert means in plain language.\n\n" +
		"Likely Causes:\n" +
		"- Bullet list of the most probable causes.\n\n" +
		"Recommended Actions:\n" +
		"- Immediate steps first, then follow-ups.\n\n" +
		"Decision:\n" +
		"- [NOTIFY] or [SUPPRESS] followed by a one-sentence justification.\n\n" +
		"Severity:\n" +
		"- Critical/Warning/Normal with brief justification.\n\n" +
		"Assumptions:\n" +
		"- Any assumptions or unknowns."
}

// DeleteAlertServ deletes an alert by ID
func DeleteAlertServ(id int) error {
	return repository.DeleteAlertByIDDAO(id)
}

// UpdateAlertServ updates an existing alert
func UpdateAlertServ(id int, req AlertReq) error {
	alert, err := repository.GetAlertByIDDAO(id)
	if err != nil {
		return err
	}

	// Use provided status or keep existing status if not provided
	status := req.Status
	if status == 0 && alert.Status != 0 {
		// Only use default status if the alert previously had a different status
		// This allows explicit status updates
		status = alert.Status
	}

	return repository.UpdateAlertDAO(id, model.Alert{
		Message:  req.Message,
		Severity: req.Severity,
		Status:   status,
		AlarmID:  req.AlarmID,
		HostID:   req.HostID,
		ItemID:   req.ItemID,
		Comment:  req.Comment,
	})
}

// GenerateTestAlerts generates simulated alerts for testing
func GenerateTestAlerts(count int) error {
	if count <= 0 {
		count = 5
	}

	rand.Seed(time.Now().UnixNano())

	// Get random hosts and items
	hosts, err := repository.GetAllHostsDAO()
	if err != nil || len(hosts) == 0 {
		return fmt.Errorf("no hosts available for alert generation")
	}

	items, err := repository.GetAllItemsDAO()
	if err != nil {
		items = []model.Item{}
	}

	messages := []string{
		"High CPU usage detected",
		"Memory threshold exceeded",
		"Disk space running low",
		"Network latency spike",
		"Service health degraded",
		"Connection timeout",
		"Database query slow",
		"Certificate expiration warning",
		"Backup failed",
		"Configuration mismatch",
		"Performance degradation",
		"Unusual traffic pattern",
		"Failed authentication attempts",
		"System resource exhausted",
		"Response time exceeded threshold",
	}

	for i := 0; i < count; i++ {
		hostIdx := rand.Intn(len(hosts))
		host := hosts[hostIdx]

		var itemID uint = 0
		if len(items) > 0 {
			// Filter items belonging to this host
			for j := 0; j < len(items); j++ {
				if items[j].HID == host.ID {
					itemID = items[j].ID
					break
				}
			}
		}

		severity := AlertSeverity(rand.Intn(3))
		messageIdx := rand.Intn(len(messages))

		alert := model.Alert{
			Message:  messages[messageIdx],
			Severity: int(severity),
			Status:   0, // 0 = active
			HostID:   host.ID,
			ItemID:   itemID,
			Comment:  fmt.Sprintf("Auto-generated test alert #%d at %s", i+1, time.Now().Format(time.RFC3339)),
		}

		if err := repository.AddAlertDAO(&alert); err != nil {
			LogService("warn", "failed to add test alert", map[string]interface{}{"error": err.Error()}, nil, "")
			continue
		}

		// Simulate host/item status update based on severity
		if severity == SeverityCritical {
			// Mark host as error
			_ = repository.UpdateHostStatusAndDescriptionDAO(host.ID, 2, fmt.Sprintf("Critical alert: %s", alert.Message))
			// Mark related items as error
			if itemID > 0 {
				_ = repository.UpdateItemStatusAndDescriptionDAO(itemID, 2, fmt.Sprintf("Critical alert: %s", alert.Message))
			}
		}
	}

	LogService("info", "test alerts generated", map[string]interface{}{"count": count}, nil, "")
	return nil
}

// CalculateAlertScore calculates a composite alert score (0-100)
func CalculateAlertScore() (int, error) {
	alerts, err := repository.GetAllAlertsDAO()
	if err != nil {
		return 0, err
	}

	if len(alerts) == 0 {
		return 100, nil
	}

	totalScore := 0.0
	weights := map[AlertSeverity]float64{
		SeverityInfo:     1.0,
		SeverityWarning:  5.0,
		SeverityCritical: 20.0,
	}

	for _, alert := range alerts {
		severity := AlertSeverity(alert.Severity)
		weight := weights[severity]
		if alert.Status != 0 { // Resolved alerts count less
			weight *= 0.5
		}
		totalScore += weight
	}

	// Normalize to 0-100 scale using logarithmic decay
	score := 100 - int(math.Min(100, 50*math.Log(1+totalScore/10)))
	return score, nil
}
