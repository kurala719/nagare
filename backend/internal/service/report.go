package service

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"nagare/internal/model"
	"nagare/internal/repository"
	"nagare/internal/service/utils"

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
func GenerateWeeklyReportServ() (model.Report, error) {
	return createAndProcessReport("weekly", "Weekly Operations Analytics - "+time.Now().Format("2006-01-02"))
}

// GenerateMonthlyReportServ triggers a monthly report generation
func GenerateMonthlyReportServ() (model.Report, error) {
	return createAndProcessReport("monthly", "Monthly Infrastructure Insight - "+time.Now().Format("2006-01"))
}

func createAndProcessReport(rtype, title string) (model.Report, error) {
	report := model.Report{
		ReportType:  rtype,
		Title:       title,
		Status:      0, // Generating
		GeneratedAt: time.Now(),
	}
	if err := repository.AddReportDAO(&report); err != nil {
		return model.Report{}, err
	}

	// Process asynchronously
	go processReport(report)

	return report, nil
}

func processReport(report model.Report) {
	// 1. Fetch Aggregated Data
	data := aggregateAdvancedReportData()

	// 2. Generate PDF
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
	pieBytes, _ := utils.GeneratePieChart("Host Status", data.StatusDistribution)
	lineBytes, _ := utils.GenerateLineChart("Alert Trend", []string{"M", "T", "W", "T", "F", "S", "S"}, data.AlertTrend)
	
	m.AddRow(80,
		col.New(6).Add(image.NewFromBytes(pieBytes, extension.Png, props.Rect{Center: true, Percent: 90})),
		col.New(6).Add(image.NewFromBytes(lineBytes, extension.Png, props.Rect{Center: true, Percent: 90})),
	)
	m.AddRow(10,
		text.NewCol(6, "Status Distribution", props.Text{Align: align.Center, Size: 9}),
		text.NewCol(6, "Weekly Alert Trend", props.Text{Align: align.Center, Size: 9}),
	)

	// Page 2: Host Analytics
	m.AddAutoRow(text.NewCol(12, "Critical Host Analytics", props.Text{Size: 14, Style: fontstyle.Bold, Top: 20}))
	
	// Failure Frequency Chart
	barBytes, _ := utils.GenerateBarChart("Failure Frequency", data.FailureFrequency)
	m.AddRow(70, col.New(12).Add(image.NewFromBytes(barBytes, extension.Png, props.Rect{Center: true, Percent: 80})))
	m.AddRow(10, text.NewCol(12, "Top 5 Most Frequent Failures (Times)", props.Text{Align: align.Center, Size: 9}))

	buildTopHostsTable(m, data.TopCPUHosts)
	buildDowntimeTable(m, data.LongestDowntimeHosts)

	fileName := fmt.Sprintf("report_%d.pdf", report.ID)
	filePath := "public/reports/" + fileName
	
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		_ = repository.UpdateReportStatusDAO(report.ID, 2, "", "")
		_ = CreateSiteMessageServ("Report Failed", fmt.Sprintf("Failed to create report directory for '%s'.", report.Title), "report", 3, nil)
		return
	}
	
	document, err := m.Generate()
	if err != nil {
		_ = repository.UpdateReportStatusDAO(report.ID, 2, "", "")
		_ = CreateSiteMessageServ("Report Failed", fmt.Sprintf("Failed to generate PDF for '%s'.", report.Title), "report", 3, nil)
		return
	}

	if err := document.Save(filePath); err != nil {
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

func aggregateAdvancedReportData() AdvancedReportData {
	// Realistic mock data for visualization
	return AdvancedReportData{
		TotalAlerts: 156,
		AvgUptime:   99.82,
		StatusDistribution: map[string]float64{
			"Active":  85,
			"Warning": 10,
			"Error":   5,
		},
		AlertTrend: []float64{12, 15, 45, 20, 18, 10, 36},
		FailureFrequency: map[string]float64{
			"web-srv-01": 12,
			"db-master":  8,
			"app-node-2": 15,
			"gateway-01": 5,
			"cache-cli":  3,
		},
		TopCPUHosts: [][]string{
			{"app-node-2", "192.168.1.12", "92%", "Warning"},
			{"db-master", "192.168.1.20", "88%", "Active"},
			{"web-srv-01", "192.168.1.10", "75%", "Active"},
			{"worker-04", "192.168.1.44", "62%", "Active"},
			{"mq-broker", "192.168.1.30", "58%", "Active"},
		},
		LongestDowntimeHosts: [][]string{
			{"backup-srv", "192.168.1.99", "14h 20m", "5 times"},
			{"app-node-2", "192.168.1.12", "4h 15m", "15 times"},
			{"db-slave-2", "192.168.1.22", "2h 05m", "2 times"},
		},
		Summary: "Overall system stability remained at 99.82%. A major spike in alerts occurred on Wednesday due to network congestion in DC-East. 'app-node-2' remains the most unstable asset with 15 failure events this period. Immediate hardware health check is recommended for the backup server.",
	}
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
		text.NewCol(4, fmt.Sprintf("Avg Uptime: %.2f%%", data.AvgUptime), props.Text{Size: 11, Align: align.Center, Top: 5}),
		text.NewCol(4, fmt.Sprintf("Critical Issues: %d", int(data.StatusDistribution["Error"])), props.Text{Size: 11, Align: align.Center, Top: 5}),
	)

	m.AddAutoRow(text.NewCol(12, data.Summary, props.Text{Size: 10, Top: 5, Bottom: 10}))
}

func buildTopHostsTable(m core.Maroto, rows [][]string) {
	m.AddAutoRow(text.NewCol(12, "Top Resource Consumers (CPU Usage)", props.Text{Size: 12, Style: fontstyle.Bold, Top: 15}))
	
	header := []string{"Host Name", "IP Address", "Avg Usage", "Status"}
	
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

func buildDowntimeTable(m core.Maroto, rows [][]string) {
	m.AddAutoRow(text.NewCol(12, "Stability Issues (Longest Downtime)", props.Text{Size: 12, Style: fontstyle.Bold, Top: 15}))
	
	header := []string{"Host Name", "IP Address", "Total Downtime", "Frequency"}
	
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
