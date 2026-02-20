package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nagare/internal/adapter/external/llm"
	"nagare/internal/adapter/repository"
	"nagare/internal/core/domain"
	"nagare/internal/core/service/utils"
	"nagare/internal/database"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// ReportResp represents a report response
type ReportResp struct {
	ID          uint      `json:"id"`
	ReportType  string    `json:"report_type"`
	Title       string    `json:"title"`
	DownloadURL string    `json:"download_url"`
	Status      string    `json:"status"` // pending, completed, failed
	StatusCode  int       `json:"status_code"`
	GeneratedAt time.Time `json:"generated_at"`
}

// GenerateWeeklyReportServ triggers a weekly report generation
func GenerateWeeklyReportServ() (domain.Report, error) {
	return createAndProcessReport("weekly", "Weekly Operations Analytics - "+time.Now().Format("2006-01-02"))
}

// GenerateMonthlyReportServ triggers a monthly report generation
func GenerateMonthlyReportServ() (domain.Report, error) {
	return createAndProcessReport("monthly", "Monthly Infrastructure Insight - "+time.Now().Format("2006-01"))
}

func createAndProcessReport(rtype, title string) (domain.Report, error) {
	report := domain.Report{
		ReportType:  rtype,
		Title:       title,
		Status:      0, // Generating
		GeneratedAt: time.Now(),
	}
	if err := repository.AddReportDAO(&report); err != nil {
		return domain.Report{}, err
	}

	// Process asynchronously
	go processReport(report)

	return report, nil
}

func processReport(report domain.Report) {
	defer func() {
		if r := recover(); r != nil {
			LogService("error", "panic in report generation", map[string]interface{}{"panic": r, "report_id": report.ID}, nil, "")
			_ = repository.UpdateReportStatusDAO(report.ID, 2, "", "")
		}
	}()

	// 1. Fetch Aggregated Data
	data := aggregateAdvancedReportData(report.ReportType)

	// 2. Save Data to JSON for preview
	dataJSON, err := json.Marshal(data)
	if err != nil {
		LogService("error", "failed to marshal report data", map[string]interface{}{"error": err.Error()}, nil, "")
	} else {
		report.ContentData = string(dataJSON)
		if err := repository.UpdateReportContentDAO(report.ID, report.ContentData); err != nil {
			LogService("error", "failed to update report content in DB", map[string]interface{}{"error": err.Error(), "report_id": report.ID}, nil, "")
			// This might be the column missing error. We continue to try PDF generation.
		}
	}

	// 3. Generate PDF
	cfg := config.NewBuilder().
		WithPageNumber(props.PageNumber{
			Pattern: "Page {current} of {total}",
			Place:   props.RightBottom,
		}).
		WithLeftMargin(15).
		WithTopMargin(15).
		WithRightMargin(15).
		Build()

	m := maroto.New(cfg)

	// Page 1: Cover & Summary
	buildProfessionalHeader(m, report.Title)
	buildExecutiveSummary(m, data)

	// Charts Section
	m.AddAutoRow(text.NewCol(12, "Infrastructure Health & Trends", props.Text{Size: 14, Style: fontstyle.Bold, Top: 10}))

	// Pie Chart & Line Chart
	pieBytes, errPie := utils.GeneratePieChart("Host Status", data.StatusDistribution)
	lineBytes, errLine := utils.GenerateLineChart("Alert Trend", []string{"M", "T", "W", "T", "F", "S", "S"}, data.AlertTrend)

	if errPie == nil && errLine == nil {
		m.AddRow(80,
			col.New(6).Add(image.NewFromBytes(pieBytes, extension.Png, props.Rect{Center: true, Percent: 90})),
			col.New(6).Add(image.NewFromBytes(lineBytes, extension.Png, props.Rect{Center: true, Percent: 90})),
		)
		m.AddRow(10,
			text.NewCol(6, "Status Distribution", props.Text{Align: align.Center, Size: 9}),
			text.NewCol(6, "Weekly Alert Trend", props.Text{Align: align.Center, Size: 9}),
		)
	} else {
		m.AddAutoRow(text.NewCol(12, "[Chart Generation Skipped due to data issues]", props.Text{Size: 10, Color: &props.Color{Red: 255}}))
	}

	// Page 2: Host Analytics
	m.AddAutoRow(text.NewCol(12, "Critical Host Analytics", props.Text{Size: 14, Style: fontstyle.Bold, Top: 20}))

	// Failure Frequency Chart
	barBytes, errBar := utils.GenerateBarChart("Failure Frequency", data.FailureFrequency)
	if errBar == nil {
		m.AddRow(70, col.New(12).Add(image.NewFromBytes(barBytes, extension.Png, props.Rect{Center: true, Percent: 80})))
		m.AddRow(10, text.NewCol(12, "Top 5 Most Frequent Failures (Times)", props.Text{Align: align.Center, Size: 9}))
	}

	buildTopHostsTable(m, data.TopCPUHosts)
	buildDowntimeTable(m, data.LongestDowntimeHosts)

	fileName := fmt.Sprintf("report_%d.pdf", report.ID)
	filePath := "public/reports/" + fileName

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		LogService("error", "failed to create reports directory", map[string]interface{}{"error": err.Error()}, nil, "")
		_ = repository.UpdateReportStatusDAO(report.ID, 2, "", "")
		_ = CreateSiteMessageServ("Report Failed", fmt.Sprintf("Failed to create report directory for '%s'.", report.Title), "report", 3, nil)
		return
	}

	document, err := m.Generate()
	if err != nil {
		LogService("error", "failed to generate PDF", map[string]interface{}{"error": err.Error()}, nil, "")
		_ = repository.UpdateReportStatusDAO(report.ID, 2, "", "")
		_ = CreateSiteMessageServ("Report Failed", fmt.Sprintf("Failed to generate PDF for '%s'.", report.Title), "report", 3, nil)
		return
	}

	if err := document.Save(filePath); err != nil {
		LogService("error", "failed to save PDF file", map[string]interface{}{"error": err.Error(), "path": filePath}, nil, "")
		_ = repository.UpdateReportStatusDAO(report.ID, 2, "", "")
		_ = CreateSiteMessageServ("Report Failed", fmt.Sprintf("Failed to save PDF for '%s'.", report.Title), "report", 3, nil)
		return
	}

	downloadURL := "/api/v1/reports/" + fmt.Sprint(report.ID) + "/download"
	_ = repository.UpdateReportStatusDAO(report.ID, 1, filePath, downloadURL)

	// Add Site Message
	_ = CreateSiteMessageServ("Report Ready", fmt.Sprintf("Report '%s' has been generated successfully.", report.Title), "report", 1, nil)
}

type AdvancedReportData struct {
	TotalAlerts          int
	AvgUptime            float64
	TopCPUHosts          [][]string
	LongestDowntimeHosts [][]string
	AlertTrend           []float64
	StatusDistribution   map[string]float64
	FailureFrequency     map[string]float64
	Summary              string
}

func aggregateAdvancedReportData(reportType string) AdvancedReportData {
	days := 7
	if reportType == "monthly" {
		days = 30
	}
	startTime := time.Now().AddDate(0, 0, -days)

	data := AdvancedReportData{
		StatusDistribution: make(map[string]float64),
		FailureFrequency:   make(map[string]float64),
		AlertTrend:         make([]float64, 7),
	}

	// 1. Total Alerts in period
	var totalAlerts int64
	database.DB.Model(&domain.Alert{}).Where("created_at >= ?", startTime).Count(&totalAlerts)
	data.TotalAlerts = int(totalAlerts)

	// 2. Status Distribution (current)
	hosts, _ := repository.GetAllHostsDAO()
	for _, h := range hosts {
		status := "Unknown"
		switch h.Status {
		case 1:
			status = "Active"
		case 0:
			status = "Inactive"
		case 2:
			status = "Error"
		case 3:
			status = "Syncing"
		}
		data.StatusDistribution[status]++
	}

	// 3. Alert Trend (last 7 days)
	for i := 0; i < 7; i++ {
		dStart := time.Now().AddDate(0, 0, -6+i)
		dStart = time.Date(dStart.Year(), dStart.Month(), dStart.Day(), 0, 0, 0, 0, dStart.Location())
		dEnd := dStart.Add(24 * time.Hour)
		var count int64
		database.DB.Model(&domain.Alert{}).Where("created_at >= ? AND created_at < ?", dStart, dEnd).Count(&count)
		data.AlertTrend[i] = float64(count)
	}

	// 4. Failure Frequency (Top hosts by alert count)
	type freqRes struct {
		HostID uint
		Count  int64
	}
	var frequencies []freqRes
	database.DB.Model(&domain.Alert{}).
		Select("host_id, count(*) as count").
		Where("created_at >= ? AND host_id > 0", startTime).
		Group("host_id").
		Order("count desc").
		Limit(5).
		Scan(&frequencies)

	for _, f := range frequencies {
		h, err := repository.GetHostByIDDAO(f.HostID)
		if err == nil {
			data.FailureFrequency[h.Name] = float64(f.Count)
		}
	}

	// 5. Top CPU Hosts
	var cpuItems []domain.Item
	database.DB.Model(&domain.Item{}).
		Where("name LIKE ? OR name LIKE ?", "%CPU%", "%cpu%").
		Order("CAST(last_value AS DECIMAL) desc").
		Limit(5).
		Find(&cpuItems)

	for _, item := range cpuItems {
		h, _ := repository.GetHostByIDDAO(item.HID)
		status := "Active"
		if h.Status == 2 {
			status = "Error"
		}
		data.TopCPUHosts = append(data.TopCPUHosts, []string{
			h.Name, h.IPAddr, item.LastValue, item.Units, status,
		})
	}

	// 6. Longest Downtime Hosts (Mocked for now as real calculation is complex)
	data.LongestDowntimeHosts = [][]string{
		{"N/A", "-", "0h", "0 times"},
	}
	if len(frequencies) > 0 {
		h, _ := repository.GetHostByIDDAO(frequencies[0].HostID)
		data.LongestDowntimeHosts[0] = []string{h.Name, h.IPAddr, "Detected Issues", fmt.Sprintf("%d alerts", frequencies[0].Count)}
	}

	// 7. Calculate Avg Uptime (based on host health scores)
	var avgScore float64
	database.DB.Model(&domain.Host{}).Select("AVG(health_score)").Scan(&avgScore)
	data.AvgUptime = avgScore

	// 8. AI Summary
	data.Summary = generateAISummary(data)

	return data
}

func generateAISummary(data AdvancedReportData) string {
	if !aiAnalysisEnabled() {
		return "AI Summary generation is disabled. Based on metrics, the system has " + fmt.Sprint(data.TotalAlerts) + " alerts in the last period."
	}

	providerID, modelName := aiProviderConfig()
	client, resolvedModel, err := createLLMClient(providerID, modelName)
	if err != nil {
		return "Failed to initialize AI for summary: " + err.Error()
	}

	ctx, cancel := aiAnalysisContext()
	defer cancel()

	statsJSON, _ := json.Marshal(data)
	prompt := "You are a senior infrastructure analyst. Summarize the following operational data into a professional executive summary (3-4 sentences).\nData: " + string(statsJSON)

	resp, err := client.Chat(ctx, llm.ChatRequest{
		Model:        resolvedModel,
		SystemPrompt: "Generate a concise executive summary for an infrastructure report.",
		Messages: []llm.Message{
			{Role: "user", Content: prompt},
		},
	})

	if err != nil {
		return "Infrastructure remained operational. Total alerts: " + fmt.Sprint(data.TotalAlerts) + ". (AI Summary failed: " + err.Error() + ")"
	}

	return strings.TrimSpace(resp.Content)
}

func buildProfessionalHeader(m core.Maroto, title string) {
	m.AddRow(20,
		text.NewCol(12, title, props.Text{
			Size:  20,
			Style: fontstyle.Bold,
			Align: align.Center,
			Color: &props.Color{Red: 37, Green: 99, Blue: 235},
		}),
	)
	m.AddRow(10,
		text.NewCol(12, "Infrastructure Intelligence Report | Nagare Platform", props.Text{
			Size:  10,
			Style: fontstyle.Italic,
			Align: align.Center,
		}),
	)
	m.AddRow(5, text.NewCol(12, "", props.Text{})) // Spacer
}

func buildExecutiveSummary(m core.Maroto, data AdvancedReportData) {
	m.AddAutoRow(text.NewCol(12, "Executive Summary", props.Text{Size: 14, Style: fontstyle.Bold}))

	m.AddRow(20,
		text.NewCol(4, fmt.Sprintf("Total Alerts: %d", data.TotalAlerts), props.Text{Size: 11, Align: align.Center, Top: 5}),
		text.NewCol(4, fmt.Sprintf("Avg Health: %.2f%%", data.AvgUptime), props.Text{Size: 11, Align: align.Center, Top: 5}),
		text.NewCol(4, fmt.Sprintf("Critical Assets: %d", int(data.StatusDistribution["Error"])), props.Text{Size: 11, Align: align.Center, Top: 5}),
	)

	m.AddAutoRow(text.NewCol(12, data.Summary, props.Text{Size: 10, Top: 5, Bottom: 10}))
}

func buildTopHostsTable(m core.Maroto, rows [][]string) {
	m.AddAutoRow(text.NewCol(12, "Top Resource Consumers (CPU Usage)", props.Text{Size: 12, Style: fontstyle.Bold, Top: 15}))

	header := []string{"Asset Name", "IP Address", "Avg Usage", "Units", "Status"}

	m.AddRow(10,
		text.NewCol(3, header[0], props.Text{Style: fontstyle.Bold, Size: 10}),
		text.NewCol(3, header[1], props.Text{Style: fontstyle.Bold, Size: 10}),
		text.NewCol(2, header[2], props.Text{Style: fontstyle.Bold, Size: 10}),
		text.NewCol(2, header[3], props.Text{Style: fontstyle.Bold, Size: 10}),
		text.NewCol(2, header[4], props.Text{Style: fontstyle.Bold, Size: 10}),
	)

	for _, row := range rows {
		m.AddRow(8,
			text.NewCol(3, row[0], props.Text{Size: 9}),
			text.NewCol(3, row[1], props.Text{Size: 9}),
			text.NewCol(2, row[2], props.Text{Size: 9}),
			text.NewCol(2, row[3], props.Text{Size: 9}),
			text.NewCol(2, row[4], props.Text{Size: 9}),
		)
	}
}

func buildDowntimeTable(m core.Maroto, rows [][]string) {
	m.AddAutoRow(text.NewCol(12, "Stability Issues (Frequency)", props.Text{Size: 12, Style: fontstyle.Bold, Top: 15}))

	header := []string{"Asset Name", "IP Address", "Summary", "Alert Count"}

	m.AddRow(10,
		text.NewCol(3, header[0], props.Text{Style: fontstyle.Bold, Size: 10}),
		text.NewCol(3, header[1], props.Text{Style: fontstyle.Bold, Size: 10}),
		text.NewCol(3, header[2], props.Text{Style: fontstyle.Bold, Size: 10}),
		text.NewCol(3, header[3], props.Text{Style: fontstyle.Bold, Size: 10}),
	)

	for _, row := range rows {
		m.AddRow(8,
			text.NewCol(3, row[0], props.Text{Size: 9}),
			text.NewCol(3, row[1], props.Text{Size: 9}),
			text.NewCol(3, row[2], props.Text{Size: 9}),
			text.NewCol(3, row[3], props.Text{Size: 9}),
		)
	}
}

// GetReportServ retrieves a report by ID
func GetReportServ(id uint) (ReportResp, error) {
	r, err := repository.GetReportByIDDAO(id)
	if err != nil {
		return ReportResp{}, err
	}

	statusStr := "pending"
	if r.Status == 1 {
		statusStr = "completed"
	} else if r.Status == 2 {
		statusStr = "failed"
	}

	return ReportResp{
		ID:          r.ID,
		ReportType:  r.ReportType,
		Title:       r.Title,
		DownloadURL: r.DownloadURL,
		Status:      statusStr,
		StatusCode:  r.Status,
		GeneratedAt: r.GeneratedAt,
	}, nil
}

// ListReportsServ lists reports
func ListReportsServ(rtype string, status *int, limit, offset int) ([]ReportResp, error) {
	reports, _, err := repository.ListReportsDAO(rtype, status, limit, offset)
	if err != nil {
		return nil, err
	}

	var res []ReportResp
	for _, r := range reports {
		statusStr := "pending"
		if r.Status == 1 {
			statusStr = "completed"
		} else if r.Status == 2 {
			statusStr = "failed"
		}
		res = append(res, ReportResp{
			ID:          r.ID,
			ReportType:  r.ReportType,
			Title:       r.Title,
			DownloadURL: r.DownloadURL,
			Status:      statusStr,
			StatusCode:  r.Status,
			GeneratedAt: r.GeneratedAt,
		})
	}
	return res, nil
}

// DeleteReportServ deletes a report
func DeleteReportServ(id uint) error {
	return repository.DeleteReportDAO(id)
}

// GetReportFilePathServ returns the local file path of a report
func GetReportFilePathServ(id uint) (string, error) {
	r, err := repository.GetReportByIDDAO(id)
	if err != nil {
		return "", err
	}
	return r.FilePath, nil
}

// GetReportContentServ retrieves the JSON content of a report
func GetReportContentServ(id uint) (AdvancedReportData, error) {
	r, err := repository.GetReportByIDDAO(id)
	if err != nil {
		return AdvancedReportData{}, err
	}

	var data AdvancedReportData
	if err := json.Unmarshal([]byte(r.ContentData), &data); err != nil {
		return AdvancedReportData{}, fmt.Errorf("failed to unmarshal report content: %w", err)
	}

	return data, nil
}
