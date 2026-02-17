package service

import (
	"nagare/internal/model"
	"nagare/internal/repository"
)

// PublicStatusSummary represents the data for the public status page
type PublicStatusSummary struct {
	OverallStatus   string           `json:"overall_status"`   // "operational", "degraded", "outage"
	OverallMessage  string           `json:"overall_message"`  // "All systems operational", etc.
	ActiveIncidents []PublicIncident `json:"active_incidents"` // List of active critical alerts
	Groups          []PublicGroup    `json:"groups"`           // List of groups/services
}

type PublicIncident struct {
	ID        uint   `json:"id"`
	Message   string `json:"message"`
	Severity  string `json:"severity"`
	CreatedAt string `json:"created_at"`
}

type PublicGroup struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"` // "operational", "degraded", "outage"
	Uptime string `json:"uptime"` // e.g., "99.9%" (placeholder for now)
}

// GetPublicStatusSummaryServ generates the status page data
func GetPublicStatusSummaryServ() (PublicStatusSummary, error) {
	// 1. Fetch Groups
	groups, err := repository.GetAllGroupsDAO()
	if err != nil {
		return PublicStatusSummary{}, err
	}

	var publicGroups []PublicGroup
	groupsDown := 0
	groupsDegraded := 0

	for _, group := range groups {
		if group.Enabled != 1 {
			continue
		}

		statusStr := "operational"
		switch group.Status {
		case 2: // Error
			statusStr = "outage"
			groupsDown++
		case 3: // Syncing/Unknown often treated as degraded or operational depending on policy
			statusStr = "operational" // Assume syncing is fine for public display, or check actual logic
		}

		publicGroups = append(publicGroups, PublicGroup{
			ID:     group.ID,
			Name:   group.Name,
			Status: statusStr,
			Uptime: "100.0%", // Placeholder: Implement real calculation later
		})
	}

	// 2. Fetch Incidents (Active Critical/Warning Alerts)
	// We'll consider severity 2 (Critical) as incidents for the status page.
	statusActive := 0
	alertFilter := model.AlertFilter{
		Status: &statusActive,
	}
	alerts, err := repository.SearchAlertsDAO(alertFilter)
	if err != nil {
		// Log error but continue with empty alerts? Better to return error.
		return PublicStatusSummary{}, err
	}

	var publicIncidents []PublicIncident
	for _, alert := range alerts {
		if alert.Severity >= 2 { // Critical
			publicIncidents = append(publicIncidents, PublicIncident{
				ID:        alert.ID,
				Message:   alert.Message,
				Severity:  "critical",
				CreatedAt: alert.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}

	// 3. Calculate Overall Status
	overallStatus := "operational"
	overallMessage := "All systems operational"

	if len(publicIncidents) > 0 || groupsDown > 0 {
		overallStatus = "outage"
		overallMessage = "Service Outage"
	} else if groupsDegraded > 0 {
		overallStatus = "degraded"
		overallMessage = "Partial System Degradation"
	}

	return PublicStatusSummary{
		OverallStatus:   overallStatus,
		OverallMessage:  overallMessage,
		ActiveIncidents: publicIncidents,
		Groups:          publicGroups,
	}, nil
}
