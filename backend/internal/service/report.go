package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nagare/internal/database"
	"nagare/internal/model"
	"nagare/internal/repository"
	"nagare/internal/repository/llm"
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
	"github.com/johnfercher/maroto/v2/pkg/core/entity"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

var translations = map[string]map[string]string{
	"en": {
		"page_pattern":          "Page {current} of {total}",
		"daily_title":           "Daily Operations Summary",
		"weekly_title":          "Weekly Operations Analytics",
		"monthly_title":         "Monthly Infrastructure Insight",
		"infra_health":          "Infrastructure Health & Trends",
		"chart_skipped":         "[Chart Generation Skipped due to data issues]",
		"critical_host":         "Critical Host Analytics",
		"top_resource":          "Top Resource Consumers (CPU Usage)",
		"stability_issues":      "Stability Issues (Frequency)",
		"asset_name":            "Asset Name",
		"ip_address":            "IP Address",
		"avg_usage":             "Avg Usage",
		"units":                 "Units",
		"status":                "Status",
		"summary":               "Summary",
		"alert_count":           "Alert Count",
		"infra_report_platform": "Infrastructure Intelligence Report | Nagare Platform",
		"executive_summary":     "Executive Summary",
		"total_alerts":          "Total Alerts",
		"avg_health":            "Avg Health",
		"critical_assets":       "Critical Assets",
		"status_distribution":   "Status Distribution",
		"alert_trend":           "Alert Trend in Period",
		"failure_frequency":     "Failure Frequency",
		"top_5_failures":        "Top 5 Most Frequent Failures (Times)",
		"ai_summary_disabled":   "AI Summary generation is disabled. Based on metrics, the system has %d alerts in the last period.",
		"ai_init_failed":        "Failed to initialize AI for summary: %v",
		"ai_summary_failed":     "Infrastructure remained operational. Total alerts: %d. (AI Summary failed: %v)",
		"ai_system_prompt":      "Generate a concise executive summary for an infrastructure report in English.",
		"ai_user_prompt":        "You are a senior infrastructure analyst. Summarize the following operational data into a professional executive summary (3-4 sentences) in English.\nData: %s",
		"detected_issues":       "Detected Issues",
		"alerts_count_suffix":   "%d alerts",
		"time":                  "Time",
		"count":                 "Count",
		"no_data":               "No Data",
		"na":                    "N/A",
		"hours_suffix":          "%s h",
		"times_suffix":          "%s times",
		"start":                 "Start",
		"end":                   "End",
		"mon":                   "M",
		"tue":                   "T",
		"wed":                   "W",
		"thu":                   "T",
		"fri":                   "F",
		"sat":                   "S",
		"sun":                   "S",
	},
	"zh": {
		"page_pattern":          "第 {current} 页，共 {total} 页",
		"daily_title":           "每日运维总结报告",
		"weekly_title":          "每周运维分析报告",
		"monthly_title":         "每月基础设施洞察报告",
		"infra_health":          "基础设施健康状况与趋势",
		"chart_skipped":         "[由于数据问题，跳过图表生成]",
		"critical_host":         "关键主机分析",
		"top_resource":          "资源消耗排行 (CPU 使用率)",
		"stability_issues":      "稳定性问题 (频率)",
		"asset_name":            "资产名称",
		"ip_address":            "IP 地址",
		"avg_usage":             "平均使用率",
		"units":                 "单位",
		"status":                "状态",
		"summary":               "摘要",
		"alert_count":           "告警次数",
		"infra_report_platform": "基础设施智能报告 | Nagare 平台",
		"executive_summary":     "执行摘要",
		"total_alerts":          "告警总数",
		"avg_health":            "平均健康度",
		"critical_assets":       "异常资产",
		"status_distribution":   "状态分布",
		"alert_trend":           "期间告警趋势",
		"failure_frequency":     "故障频率",
		"top_5_failures":        "故障最频繁的前 5 个主机 (次数)",
		"ai_summary_disabled":   "AI 摘要生成已禁用。根据指标，系统在上一周期共有 %d 条告警。",
		"ai_init_failed":        "初始化 AI 摘要失败: %v",
		"ai_summary_failed":     "基础设施运行正常。告警总数: %d。(AI 摘要失败: %v)",
		"ai_system_prompt":      "为基础设施报告生成一份简洁的中文执行摘要。",
		"ai_user_prompt":        "你是一位资深基础设施分析师。请将以下运营数据总结为一份专业的执行摘要（3-4 句话），请使用中文回复。\n数据: %s",
		"active":                "正常",
		"inactive":              "离线",
		"error":                 "错误",
		"syncing":               "同步中",
		"unknown":               "未知",
		"detected_issues":       "检测到问题",
		"alerts_count_suffix":   "%d 条告警",
		"time":                  "时间",
		"count":                 "数量",
		"no_data":               "无数据",
		"na":                    "无",
		"hours_suffix":          "%s 小时",
		"times_suffix":          "%s 次",
		"start":                 "开始",
		"end":                   "结束",
		"mon":                   "一",
		"tue":                   "二",
		"wed":                   "三",
		"thu":                   "四",
		"fri":                   "五",
		"sat":                   "六",
		"sun":                   "日",
	},
}

func T(lang, key string) string {
	if m, ok := translations[lang]; ok {
		if val, ok := m[key]; ok {
			return val
		}
	}
	// Fallback to English
	if val, ok := translations["en"][key]; ok {
		return val
	}
	return key
}

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

// GenerateDailyReportServ triggers a daily report generation
func GenerateDailyReportServ() (model.Report, error) {
	lang := "en"
	if cfg, err := repository.GetReportConfigDAO(); err == nil {
		lang = cfg.Language
	}
	title := fmt.Sprintf("%s - %s", T(lang, "daily_title"), time.Now().Format("2006-01-02"))
	return createAndProcessReport("daily", title, nil, nil)
}

// GenerateWeeklyReportServ triggers a weekly report generation
func GenerateWeeklyReportServ() (model.Report, error) {
	lang := "en"
	if cfg, err := repository.GetReportConfigDAO(); err == nil {
		lang = cfg.Language
	}
	title := fmt.Sprintf("%s - %s", T(lang, "weekly_title"), time.Now().Format("2006-01-02"))
	return createAndProcessReport("weekly", title, nil, nil)
}

// GenerateMonthlyReportServ triggers a monthly report generation
func GenerateMonthlyReportServ() (model.Report, error) {
	lang := "en"
	if cfg, err := repository.GetReportConfigDAO(); err == nil {
		lang = cfg.Language
	}
	title := fmt.Sprintf("%s - %s", T(lang, "monthly_title"), time.Now().Format("2006-01"))
	return createAndProcessReport("monthly", title, nil, nil)
}

// GenerateCustomReportServ triggers a custom report generation with specific timeframe
func GenerateCustomReportServ(title string, start, end time.Time) (model.Report, error) {
	return createAndProcessReport("custom", title, &start, &end)
}

func createAndProcessReport(rtype, title string, start, end *time.Time) (model.Report, error) {
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
	go processReport(report, start, end)

	return report, nil
}

func processReport(report model.Report, customStart, customEnd *time.Time) {
	defer func() {
		if r := recover(); r != nil {
			LogService("error", "panic in report generation", map[string]interface{}{"panic": r, "report_id": report.ID}, nil, "")
			_ = repository.UpdateReportStatusDAO(report.ID, 2, "", "")
		}
	}()

	lang := "en"
	if cfg, err := repository.GetReportConfigDAO(); err == nil {
		lang = cfg.Language
	}

	// 1. Fetch Aggregated Data
	data := aggregateAdvancedReportData(report.ReportType, customStart, customEnd, lang)

	// 2. Save Data to JSON for preview
	dataJSON, err := json.Marshal(data)
	if err != nil {
		LogService("error", "failed to marshal report data", map[string]interface{}{"error": err.Error()}, nil, "")
	} else {
		report.ContentData = string(dataJSON)
		if err := repository.UpdateReportContentDAO(report.ID, report.ContentData); err != nil {
			LogService("error", "failed to update report content in DB", map[string]interface{}{"error": err.Error(), "report_id": report.ID}, nil, "")
		}
	}

	// 3. Generate PDF
	builder := config.NewBuilder().
		WithPageNumber(props.PageNumber{
			Pattern: T(lang, "page_pattern"),
			Place:   props.RightBottom,
		}).
		WithLeftMargin(15).
		WithTopMargin(15).
		WithRightMargin(15)

	if lang == "zh" {
		fontFile := "public/fonts/NotoSansSC-Regular.ttf"
		if _, err := os.Stat(fontFile); err == nil {
			customFonts := []*entity.CustomFont{
				{
					Family: "NotoSansSC",
					Style:  fontstyle.Normal,
					File:   fontFile,
				},
			}
			builder.WithCustomFonts(customFonts).
				WithDefaultFont(&props.Font{
					Family: "NotoSansSC",
				})
		} else {
			LogService("warn", "Chinese font not found, falling back to default", map[string]interface{}{"path": fontFile}, nil, "")
		}
	}

	cfg := builder.Build()
	m := maroto.New(cfg)

	// Page 1: Cover & Summary
	buildProfessionalHeader(m, report.Title, lang)
	buildExecutiveSummary(m, data, lang)

	// Charts Section
	m.AddAutoRow(text.NewCol(12, T(lang, "infra_health"), props.Text{Size: 14, Style: fontstyle.Bold, Top: 10}))

	// Pie Chart & Line Chart
	pieBytes, errPie := utils.GeneratePieChart(T(lang, "status_distribution"), data.StatusDistribution, T(lang, "no_data"))

	// Labels for trend
	labels := []string{T(lang, "mon"), T(lang, "tue"), T(lang, "wed"), T(lang, "thu"), T(lang, "fri"), T(lang, "sat"), T(lang, "sun")}
	if report.ReportType == "custom" && customStart != nil && customEnd != nil {
		labels = []string{T(lang, "start"), "...", T(lang, "end")}
	}
	lineBytes, errLine := utils.GenerateLineChart(T(lang, "alert_trend"), labels, data.AlertTrend, T(lang, "time"), T(lang, "count"))

	if errPie == nil && errLine == nil {
		m.AddRow(80,
			col.New(6).Add(image.NewFromBytes(pieBytes, extension.Png, props.Rect{Center: true, Percent: 90})),
			col.New(6).Add(image.NewFromBytes(lineBytes, extension.Png, props.Rect{Center: true, Percent: 90})),
		)
		m.AddRow(10,
			text.NewCol(6, T(lang, "status_distribution"), props.Text{Align: align.Center, Size: 9}),
			text.NewCol(6, T(lang, "alert_trend"), props.Text{Align: align.Center, Size: 9}),
		)
	} else {
		m.AddAutoRow(text.NewCol(12, T(lang, "chart_skipped"), props.Text{Size: 10, Color: &props.Color{Red: 255}}))
	}

	// Page 2: Host Analytics
	m.AddAutoRow(text.NewCol(12, T(lang, "critical_host"), props.Text{Size: 14, Style: fontstyle.Bold, Top: 20}))

	// Failure Frequency Chart
	barBytes, errBar := utils.GenerateBarChart(T(lang, "failure_frequency"), data.FailureFrequency, T(lang, "no_data"))
	if errBar == nil {
		m.AddRow(70, col.New(12).Add(image.NewFromBytes(barBytes, extension.Png, props.Rect{Center: true, Percent: 80})))
		m.AddRow(10, text.NewCol(12, T(lang, "top_5_failures"), props.Text{Align: align.Center, Size: 9}))
	}

	buildTopHostsTable(m, data.TopCPUHosts, lang)
	buildDowntimeTable(m, data.LongestDowntimeHosts, lang)

	fileName := fmt.Sprintf("report_%d.pdf", report.ID)
	filePath := "public/reports/" + fileName

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		LogService("error", fmt.Sprintf("Report Failed: Failed to create report directory for '%s'.", report.Title), map[string]interface{}{"error": err.Error()}, nil, "")
		_ = repository.UpdateReportStatusDAO(report.ID, 2, "", "")
		return
	}

	document, err := m.Generate()
	if err != nil {
		LogService("error", fmt.Sprintf("Report Failed: Failed to generate PDF for '%s'.", report.Title), map[string]interface{}{"error": err.Error()}, nil, "")
		_ = repository.UpdateReportStatusDAO(report.ID, 2, "", "")
		return
	}

	if err := document.Save(filePath); err != nil {
		LogService("error", fmt.Sprintf("Report Failed: Failed to save PDF for '%s'.", report.Title), map[string]interface{}{"error": err.Error(), "path": filePath}, nil, "")
		_ = repository.UpdateReportStatusDAO(report.ID, 2, "", "")
		return
	}

	downloadURL := "/api/v1/reports/" + fmt.Sprint(report.ID) + "/download"
	_ = repository.UpdateReportStatusDAO(report.ID, 1, filePath, downloadURL)

	LogService("info", fmt.Sprintf("Report Ready: Report '%s' has been generated successfully.", report.Title), map[string]interface{}{"report_id": report.ID}, nil, "")
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

func aggregateAdvancedReportData(reportType string, customStart, customEnd *time.Time, lang string) AdvancedReportData {
	var startTime time.Time
	if customStart != nil {
		startTime = *customStart
	} else {
		days := 7
		if reportType == "monthly" {
			days = 30
		} else if reportType == "daily" {
			days = 1
		}
		startTime = time.Now().AddDate(0, 0, -days)
	}

	var endTime time.Time
	if customEnd != nil {
		endTime = *customEnd
	} else {
		endTime = time.Now()
	}

	data := AdvancedReportData{
		StatusDistribution: make(map[string]float64),
		FailureFrequency:   make(map[string]float64),
		AlertTrend:         make([]float64, 7),
	}

	// 1. Total Alerts in period
	var totalAlerts int64
	database.DB.Model(&model.Alert{}).Where("created_at >= ? AND created_at <= ?", startTime, endTime).Count(&totalAlerts)
	data.TotalAlerts = int(totalAlerts)

	// 2. Status Distribution (current)
	hosts, _ := repository.GetAllHostsDAO()
	for _, h := range hosts {
		status := T(lang, "unknown")
		switch h.Status {
		case 1:
			status = T(lang, "active")
		case 0:
			status = T(lang, "inactive")
		case 2:
			status = T(lang, "error")
		case 3:
			status = T(lang, "syncing")
		}
		data.StatusDistribution[status]++
	}

	// 3. Alert Trend (split into 7 points for the period)
	duration := endTime.Sub(startTime)
	interval := duration / 7
	for i := 0; i < 7; i++ {
		pStart := startTime.Add(interval * time.Duration(i))
		pEnd := pStart.Add(interval)
		var count int64
		database.DB.Model(&model.Alert{}).Where("created_at >= ? AND created_at < ?", pStart, pEnd).Count(&count)
		data.AlertTrend[i] = float64(count)
	}

	// 4. Failure Frequency (Top hosts by alert count)
	type freqRes struct {
		HostID uint
		Count  int64
	}
	var frequencies []freqRes
	database.DB.Model(&model.Alert{}).
		Select("host_id, count(*) as count").
		Where("created_at >= ? AND created_at <= ? AND host_id > 0", startTime, endTime).
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
	type cpuResult struct {
		model.Item
		HostName   string
		HostIP     string
		HostStatus int
	}
	var cpuResults []cpuResult
	database.DB.Table("items").
		Select("items.*, hosts.name as host_name, hosts.ip_addr as host_ip, hosts.status as host_status").
		Joins("left join hosts on hosts.id = items.hid").
		Where("(items.name LIKE ? OR items.name LIKE ? OR items.name LIKE ?) AND items.last_value != ''", "%CPU%", "%cpu%", "%处理器%").
		Order("(items.last_value + 0) desc").
		Limit(5).
		Scan(&cpuResults)

	for _, res := range cpuResults {
		status := T(lang, "active")
		if res.HostStatus == 2 {
			status = T(lang, "error")
		}
		data.TopCPUHosts = append(data.TopCPUHosts, []string{
			res.HostName, res.HostIP, res.LastValue, res.Units, status,
		})
	}

	// 6. Longest Downtime Hosts (Mocked for now as real calculation is complex)
	data.LongestDowntimeHosts = [][]string{
		{T(lang, "na"), "-", fmt.Sprintf(T(lang, "hours_suffix"), "0"), fmt.Sprintf(T(lang, "times_suffix"), "0")},
	}
	if len(frequencies) > 0 {
		h, _ := repository.GetHostByIDDAO(frequencies[0].HostID)
		data.LongestDowntimeHosts[0] = []string{h.Name, h.IPAddr, T(lang, "detected_issues"), fmt.Sprintf(T(lang, "alerts_count_suffix"), frequencies[0].Count)}
	}

	// 7. Calculate Avg Uptime (based on host health scores)
	var avgScore float64
	database.DB.Model(&model.Host{}).Select("AVG(health_score)").Scan(&avgScore)
	data.AvgUptime = avgScore

	// 8. AI Summary
	data.Summary = generateAISummary(data, lang)

	return data
}

func generateAISummary(data AdvancedReportData, lang string) string {
	if !aiAnalysisEnabled() {
		return fmt.Sprintf(T(lang, "ai_summary_disabled"), data.TotalAlerts)
	}

	providerID, modelName := aiProviderConfig()
	client, resolvedModel, err := createLLMClient(providerID, modelName)
	if err != nil {
		return fmt.Sprintf(T(lang, "ai_init_failed"), err)
	}

	ctx, cancel := aiAnalysisContext()
	defer cancel()

	statsJSON, _ := json.Marshal(data)
	prompt := fmt.Sprintf(T(lang, "ai_user_prompt"), string(statsJSON))

	resp, err := client.Chat(ctx, llm.ChatRequest{
		Model:        resolvedModel,
		SystemPrompt: T(lang, "ai_system_prompt"),
		Messages: []llm.Message{
			{Role: "user", Content: prompt},
		},
	})

	if err != nil {
		return fmt.Sprintf(T(lang, "ai_summary_failed"), data.TotalAlerts, err)
	}

	return strings.TrimSpace(resp.Content)
}

func buildProfessionalHeader(m core.Maroto, title string, lang string) {
	m.AddRow(20,
		text.NewCol(12, title, props.Text{
			Size:  20,
			Style: fontstyle.Bold,
			Align: align.Center,
			Color: &props.Color{Red: 37, Green: 99, Blue: 235},
		}),
	)
	m.AddRow(10,
		text.NewCol(12, T(lang, "infra_report_platform"), props.Text{
			Size:  10,
			Style: fontstyle.Italic,
			Align: align.Center,
		}),
	)
	m.AddRow(5, text.NewCol(12, "", props.Text{})) // Spacer
}

func buildExecutiveSummary(m core.Maroto, data AdvancedReportData, lang string) {
	m.AddAutoRow(text.NewCol(12, T(lang, "executive_summary"), props.Text{Size: 14, Style: fontstyle.Bold}))

	m.AddRow(20,
		text.NewCol(4, fmt.Sprintf("%s: %d", T(lang, "total_alerts"), data.TotalAlerts), props.Text{Size: 11, Align: align.Center, Top: 5}),
		text.NewCol(4, fmt.Sprintf("%s: %.2f%%", T(lang, "avg_health"), data.AvgUptime), props.Text{Size: 11, Align: align.Center, Top: 5}),
		text.NewCol(4, fmt.Sprintf("%s: %d", T(lang, "critical_assets"), int(data.StatusDistribution[T(lang, "error")])), props.Text{Size: 11, Align: align.Center, Top: 5}),
	)

	m.AddAutoRow(text.NewCol(12, data.Summary, props.Text{Size: 10, Top: 5, Bottom: 10}))
}

func buildTopHostsTable(m core.Maroto, rows [][]string, lang string) {
	m.AddAutoRow(text.NewCol(12, T(lang, "top_resource"), props.Text{Size: 12, Style: fontstyle.Bold, Top: 15}))

	header := []string{T(lang, "asset_name"), T(lang, "ip_address"), T(lang, "avg_usage"), T(lang, "units"), T(lang, "status")}

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

func buildDowntimeTable(m core.Maroto, rows [][]string, lang string) {
	m.AddAutoRow(text.NewCol(12, T(lang, "stability_issues"), props.Text{Size: 12, Style: fontstyle.Bold, Top: 15}))

	header := []string{T(lang, "asset_name"), T(lang, "ip_address"), T(lang, "summary"), T(lang, "alert_count")}

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

	status := "pending"
	switch r.Status {
	case 1:
		status = "completed"
	case 2:
		status = "failed"
	}

	return ReportResp{
		ID:          r.ID,
		ReportType:  r.ReportType,
		Title:       r.Title,
		DownloadURL: r.DownloadURL,
		Status:      status,
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
		status := "pending"
		switch r.Status {
		case 1:
			status = "completed"
		case 2:
			status = "failed"
		}
		res = append(res, ReportResp{
			ID:          r.ID,
			ReportType:  r.ReportType,
			Title:       r.Title,
			DownloadURL: r.DownloadURL,
			Status:      status,
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
