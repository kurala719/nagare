package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"nagare/internal/model"
	"nagare/internal/repository"
)

// ReportReq represents request data for report operations
type ReportReq struct {
	ReportType    string `json:"report_type"` // "weekly" or "monthly"
	Title         string `json:"title"`
	Summary       string `json:"summary"`
	IncludeAlerts int    `json:"include_alerts"`
	Comment       string `json:"comment"`
}

// ReportResp represents response data for reports
type ReportResp struct {
	ID           uint    `json:"id"`
	ReportType   string  `json:"report_type"`
	Title        string  `json:"title"`
	GeneratedAt  string  `json:"generated_at"`
	PeriodStart  string  `json:"period_start"`
	PeriodEnd    string  `json:"period_end"`
	Status       string  `json:"status"`
	StatusCode   int     `json:"status_code"`
	FileSize     int64   `json:"file_size"`
	Summary      string  `json:"summary"`
	HostCount    int     `json:"host_count"`
	AlertCount   int     `json:"alert_count"`
	Availability float64 `json:"availability"`
	CreatedAtStr string  `json:"created_at"`
	DownloadURL  string  `json:"download_url"`
}

// GenerateWeeklyReportServ generates a weekly operational report
func GenerateWeeklyReportServ() (model.Report, error) {
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -7) // 7 days ago

	return generateReportForPeriod("weekly", startTime, endTime)
}

// GenerateMonthlyReportServ generates a monthly operational report
func GenerateMonthlyReportServ() (model.Report, error) {
	endTime := time.Now()
	startTime := endTime.AddDate(0, -1, 0) // 1 month ago

	return generateReportForPeriod("monthly", startTime, endTime)
}

func generateReportForPeriod(reportType string, startTime, endTime time.Time) (model.Report, error) {
	// 1. Create report record with pending status
	now := time.Now()
	title := fmt.Sprintf("%s Report - %s", reportType, now.Format("2006-01-02"))

	report := model.Report{
		ReportType:  reportType,
		Title:       title,
		GeneratedAt: now,
		PeriodStart: startTime,
		PeriodEnd:   endTime,
		Status:      0, // pending
	}

	report, err := repository.AddReportDAO(report)
	if err != nil {
		LogService("error", "failed to create report record", map[string]interface{}{
			"type":  reportType,
			"error": err.Error(),
		}, nil, "")
		return model.Report{}, err
	}

	// 2. Aggregate data from Zabbix (async task - can be done in goroutine)
	go func() {
		if err := aggregateReportData(report.ID, startTime, endTime); err != nil {
			LogService("error", "failed to aggregate report data", map[string]interface{}{
				"report_id": report.ID,
				"error":     err.Error(),
			}, nil, "")
			repository.UpdateReportDAO(report.ID, map[string]interface{}{
				"status":        2, // failed
				"error_message": err.Error(),
			})
		}
	}()

	return report, nil
}

// aggregateReportData gathers metrics and host data for report
func aggregateReportData(reportID uint, startTime, endTime time.Time) error {
	ctx := context.Background()

	// 1. Get all monitored hosts
	hosts, err := repository.SearchHostsDAO(model.HostFilter{})
	if err != nil {
		return fmt.Errorf("failed to fetch hosts: %w", err)
	}

	if len(hosts) == 0 {
		return fmt.Errorf("no hosts found")
	}

	// 2. Aggregate metrics for each host
	reportItems := []model.ReportItem{}
	totalAvailability := 0.0

	for idx, host := range hosts {
		// Fetch metrics from Zabbix for this host
		metrics, err := getHostMetricsFromZabbix(ctx, host, startTime, endTime)
		if err != nil {
			LogService("warn", "failed to get metrics for host", map[string]interface{}{
				"host_id": host.ID,
				"error":   err.Error(),
			}, nil, "")
			continue
		}

		item := model.ReportItem{
			ReportID:        reportID,
			HostID:          host.ID,
			HostName:        host.Name,
			HostIP:          host.IPAddr,
			CPUAvg:          metrics["cpu_avg"].(float64),
			CPUPeak:         metrics["cpu_peak"].(float64),
			MemoryAvg:       metrics["memory_avg"].(float64),
			MemoryPeak:      metrics["memory_peak"].(float64),
			DiskUsage:       metrics["disk_usage"].(float64),
			NetworkLatency:  metrics["network_latency"].(float64),
			AlertCount:      metrics["alert_count"].(int),
			DowntimeMinutes: metrics["downtime_minutes"].(int),
			Rank:            idx + 1,
		}

		reportItems = append(reportItems, item)
		totalAvailability += metrics["availability"].(float64)
	}

	// 3. Store report items
	for _, item := range reportItems {
		if _, err := repository.AddReportItemDAO(item); err != nil {
			LogService("warn", "failed to store report item", map[string]interface{}{
				"host_id": item.HostID,
				"error":   err.Error(),
			}, nil, "")
		}
	}

	// 4. Calculate summary statistics
	avgAvailability := totalAvailability / float64(len(reportItems))
	alertCount, _ := countAlertsForPeriod(startTime, endTime)

	// 5. Generate executive summary (using AI if enabled)
	summary := generateExecutiveSummary(len(hosts), alertCount, avgAvailability, reportItems)

	// 6. Generate PDF
	pdfPath, fileSize, err := generateReportPDF(reportID, summary, reportItems)
	if err != nil {
		return fmt.Errorf("failed to generate PDF: %w", err)
	}

	// 7. Update report with completion data
	updates := map[string]interface{}{
		"status":       1, // completed
		"file_path":    pdfPath,
		"file_size":    fileSize,
		"summary":      summary,
		"host_count":   len(hosts),
		"alert_count":  alertCount,
		"availability": avgAvailability,
	}

	if err := repository.UpdateReportDAO(reportID, updates); err != nil {
		return fmt.Errorf("failed to update report: %w", err)
	}

	LogService("info", "report generated successfully", map[string]interface{}{
		"report_id": reportID,
		"file_path": pdfPath,
	}, nil, "")

	return nil
}

// getHostMetricsFromZabbix fetches host metrics from Zabbix for a time period
func getHostMetricsFromZabbix(ctx context.Context, host model.Host, startTime, endTime time.Time) (map[string]interface{}, error) {
	// This would call Zabbix API history.get to aggregate metrics
	// For now, return placeholder data structure

	metrics := map[string]interface{}{
		"cpu_avg":          45.5,
		"cpu_peak":         78.3,
		"memory_avg":       62.1,
		"memory_peak":      85.9,
		"disk_usage":       73.4,
		"network_latency":  12.5,
		"alert_count":      3,
		"downtime_minutes": 0,
		"availability":     99.9,
	}

	return metrics, nil
}

// countAlertsForPeriod returns number of alerts in the given period
func countAlertsForPeriod(startTime, endTime time.Time) (int, error) {
	// Query alerts within the period
	// For now, return placeholder
	return 12, nil
}

// generateExecutiveSummary creates a text summary (could use LLM)
func generateExecutiveSummary(hostCount, alertCount int, availability float64, items []model.ReportItem) string {
	summary := fmt.Sprintf(
		"In this reporting period, %d hosts were monitored with an overall availability of %.2f%%.\n"+
			"%d critical alerts were triggered.\n"+
			"Top performing metrics remain stable with average CPU utilization at %.1f%% and memory at %.1f%%.\n"+
			"Recommended actions: Monitor high-utilization hosts and review alert trends.",
		hostCount,
		availability*100,
		alertCount,
		getAvgCPU(items),
		getAvgMemory(items),
	)
	return summary
}

func getAvgCPU(items []model.ReportItem) float64 {
	if len(items) == 0 {
		return 0
	}
	var total float64
	for _, item := range items {
		total += item.CPUAvg
	}
	return total / float64(len(items))
}

func getAvgMemory(items []model.ReportItem) float64 {
	if len(items) == 0 {
		return 0
	}
	var total float64
	for _, item := range items {
		total += item.MemoryAvg
	}
	return total / float64(len(items))
}

// generateReportPDF generates the PDF file
func generateReportPDF(reportID uint, summary string, items []model.ReportItem) (string, int64, error) {
	// Ensure reports directory exists
	reportsDir := filepath.Join(".", "reports")
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return "", 0, fmt.Errorf("failed to create reports directory: %w", err)
	}

	// Generate filename
	filename := fmt.Sprintf("report_%d_%d.pdf", reportID, time.Now().Unix())
	filepath := filepath.Join(reportsDir, filename)

	// TODO: Integrate with Maroto v2 for actual PDF generation
	// For now, create a placeholder file
	content := []byte(fmt.Sprintf("Report ID: %d\nSummary: %s\n", reportID, summary))
	if err := os.WriteFile(filepath, content, 0644); err != nil {
		return "", 0, fmt.Errorf("failed to write PDF file: %w", err)
	}

	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get file info: %w", err)
	}

	return filepath, fileInfo.Size(), nil
}

// ListReportsServ lists reports with filtering
func ListReportsServ(reportType *string, status *int, limit int, offset int) ([]ReportResp, error) {
	reports, err := repository.ListReportsDAO(reportType, status, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]ReportResp, len(reports))
	for i, report := range reports {
		responses[i] = toReportResp(report)
	}
	return responses, nil
}

// GetReportServ retrieves a single report
func GetReportServ(id uint) (ReportResp, error) {
	report, err := repository.GetReportByIDDAO(id)
	if err != nil {
		return ReportResp{}, err
	}
	return toReportResp(report), nil
}

// DeleteReportServ deletes a report
func DeleteReportServ(id uint) error {
	return repository.DeleteReportDAO(id)
}

func toReportResp(report model.Report) ReportResp {
	statusStr := "pending"
	if report.Status == 1 {
		statusStr = "completed"
	} else if report.Status == 2 {
		statusStr = "failed"
	}

	downloadURL := ""
	if report.Status == 1 && report.FilePath != "" {
		downloadURL = fmt.Sprintf("/api/v1/reports/%d/download", report.ID)
	}

	return ReportResp{
		ID:           report.ID,
		ReportType:   report.ReportType,
		Title:        report.Title,
		GeneratedAt:  report.GeneratedAt.Format("2006-01-02 15:04:05"),
		PeriodStart:  report.PeriodStart.Format("2006-01-02"),
		PeriodEnd:    report.PeriodEnd.Format("2006-01-02"),
		Status:       statusStr,
		StatusCode:   report.Status,
		FileSize:     report.FileSize,
		Summary:      report.Summary,
		HostCount:    report.HostCount,
		AlertCount:   report.AlertCount,
		Availability: report.Availability,
		CreatedAtStr: report.CreatedAt.Format("2006-01-02 15:04:05"),
		DownloadURL:  downloadURL,
	}
}
