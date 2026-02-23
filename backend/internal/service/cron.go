package service

import (
	"fmt"

	"nagare/internal/model"
	"nagare/internal/repository"

	"github.com/robfig/cron/v3"
)

var cronScheduler *cron.Cron

// InitCronScheduler initializes the cron job scheduler
func InitCronScheduler() error {
	cronScheduler = cron.New()

	// Add report generation jobs based on configuration
	cfgData, err := GetReportConfigServ()
	if err == nil {
		// Extract values from map with safe assertions
		autoDaily, _ := cfgData["auto_generate_daily"].(int)
		dailyTime, _ := cfgData["daily_generate_time"].(string)
		autoWeekly, _ := cfgData["auto_generate_weekly"].(int)
		weeklyDay, _ := cfgData["weekly_generate_day"].(string)
		weeklyTime, _ := cfgData["weekly_generate_time"].(string)
		autoMonthly, _ := cfgData["auto_generate_monthly"].(int)
		monthlyDate, _ := cfgData["monthly_generate_date"].(int)
		monthlyTime, _ := cfgData["monthly_generate_time"].(string)
		
		if autoDaily == 1 && dailyTime != "" {
			hour, minute := parseTime(dailyTime)
			cronExpr := fmt.Sprintf("%d %d * * *", minute, hour)
			if _, err := cronScheduler.AddFunc(cronExpr, func() {
				if _, err := GenerateDailyReportServ(); err != nil {
					LogService("error", "daily report generation failed", map[string]interface{}{
						"error": err.Error(),
					}, nil, "")
				}
			}); err != nil {
				LogService("warn", "failed to schedule daily report", map[string]interface{}{
					"error": err.Error(),
					"expr":  cronExpr,
				}, nil, "")
			}
		}

		if autoWeekly == 1 && weeklyDay != "" {
			cronExpr := buildWeeklyCronExpression(weeklyDay, weeklyTime)
			if _, err := cronScheduler.AddFunc(cronExpr, func() {
				if _, err := GenerateWeeklyReportServ(); err != nil {
					LogService("error", "weekly report generation failed", map[string]interface{}{
						"error": err.Error(),
					}, nil, "")
				}
			}); err != nil {
				LogService("warn", "failed to schedule weekly report", map[string]interface{}{
					"error": err.Error(),
					"expr":  cronExpr,
				}, nil, "")
			}
		}

		if autoMonthly == 1 && monthlyDate > 0 {
			cronExpr := buildMonthlyCronExpression(monthlyDate, monthlyTime)
			if _, err := cronScheduler.AddFunc(cronExpr, func() {
				if _, err := GenerateMonthlyReportServ(); err != nil {
					LogService("error", "monthly report generation failed", map[string]interface{}{
						"error": err.Error(),
					}, nil, "")
				}
			}); err != nil {
				LogService("warn", "failed to schedule monthly report", map[string]interface{}{
					"error": err.Error(),
					"expr":  cronExpr,
				}, nil, "")
			}
		}
	}

	// Add daily data retention cleanup job (2:00 AM)
	if _, err := cronScheduler.AddFunc("0 2 * * *", func() {
		PerformDataRetentionCleanupServ()
	}); err != nil {
		LogService("warn", "failed to schedule data retention cleanup job", map[string]interface{}{
			"error": err.Error(),
		}, nil, "")
	}

	cronScheduler.Start()
	LogService("info", "cron scheduler started", map[string]interface{}{
		"jobs": len(cronScheduler.Entries()),
	}, nil, "")

	return nil
}

// StopCronScheduler stops the cron scheduler
func StopCronScheduler() {
	if cronScheduler != nil {
		cronScheduler.Stop()
		LogService("info", "cron scheduler stopped", nil, nil, "")
	}
}

// buildWeeklyCronExpression builds a cron expression for weekly reports
// dayName should be one of: Sunday, Monday, Tuesday, Wednesday, Thursday, Friday, Saturday
// timeStr should be in format "HH:MM:SS" or "HH:MM"
func buildWeeklyCronExpression(dayName string, timeStr string) string {
	// Map day names to cron day-of-week values (0=Sunday, 1=Monday, etc.)
	dayMap := map[string]int{
		"Sunday":    0,
		"Monday":    1,
		"Tuesday":   2,
		"Wednesday": 3,
		"Thursday":  4,
		"Friday":    5,
		"Saturday":  6,
	}

	dayOfWeek := dayMap[dayName]
	hour, minute := parseTime(timeStr)

	// Cron format: minute hour * * dayOfWeek
	return fmt.Sprintf("%d %d * * %d", minute, hour, dayOfWeek)
}

// buildMonthlyCronExpression builds a cron expression for monthly reports
// dateOfMonth should be 1-28
func buildMonthlyCronExpression(dateOfMonth int, timeStr string) string {
	if dateOfMonth < 1 || dateOfMonth > 28 {
		dateOfMonth = 1 // fallback to first day
	}

	hour, minute := parseTime(timeStr)

	// Cron format: minute hour dayOfMonth * *
	return fmt.Sprintf("%d %d %d * *", minute, hour, dateOfMonth)
}

func parseTime(timeStr string) (int, int) {
	var hour, minute int
	_, err := fmt.Sscanf(timeStr, "%d:%d", &hour, &minute)
	if err != nil {
		// Default to 8:00
		hour = 8
		minute = 0
	}
	if hour < 0 || hour > 23 {
		hour = 8
	}
	if minute < 0 || minute > 59 {
		minute = 0
	}
	return hour, minute
}

// GetReportConfigServ retrieves the report configuration
func GetReportConfigServ() (map[string]interface{}, error) {
	config, err := repository.GetReportConfigDAO()
	if err != nil {
		// Return default config if not found
		return map[string]interface{}{
			"auto_generate_daily":   0,
			"daily_generate_time":   "09:00",
			"auto_generate_weekly":  0,
			"weekly_generate_day":   "Monday",
			"weekly_generate_time":  "09:00",
			"auto_generate_monthly": 0,
			"monthly_generate_date": 1,
			"monthly_generate_time": "09:00",
			"include_alerts":        1,
			"include_metrics":       1,
			"top_hosts_count":       5,
			"enable_llm_summary":    1,
			"email_notify":          0,
			"email_recipients":      "",
			"language":              "en",
		}, nil
	}

	return map[string]interface{}{
		"id":                    config.ID,
		"auto_generate_daily":   config.AutoGenerateDaily,
		"daily_generate_time":   config.DailyGenerateTime,
		"auto_generate_weekly":  config.AutoGenerateWeekly,
		"weekly_generate_day":   config.WeeklyGenerateDay,
		"weekly_generate_time":  config.WeeklyGenerateTime,
		"auto_generate_monthly": config.AutoGenerateMonthly,
		"monthly_generate_date": config.MonthlyGenerateDate,
		"monthly_generate_time": config.MonthlyGenerateTime,
		"include_alerts":        config.IncludeAlerts,
		"include_metrics":       config.IncludeMetrics,
		"top_hosts_count":       config.TopHostsCount,
		"enable_llm_summary":    config.EnableLLMSummary,
		"email_notify":          config.EmailNotify,
		"email_recipients":      config.EmailRecipients,
		"language":              config.Language,
	}, nil
}

// UpdateReportConfigServ updates the report configuration
func UpdateReportConfigServ(updates map[string]interface{}) error {
	config, err := repository.GetReportConfigDAO()
	if err != nil {
		// Create new config
		config = model.ReportConfig{
			AutoGenerateDaily:   0,
			AutoGenerateWeekly:  0,
			AutoGenerateMonthly: 0,
			TopHostsCount:       5,
			EnableLLMSummary:    1,
			IncludeAlerts:       1,
			IncludeMetrics:      1,
		}
	}

	// Apply updates
	if autoDaily, ok := updates["auto_generate_daily"].(int); ok {
		config.AutoGenerateDaily = autoDaily
	}
	if dailyTime, ok := updates["daily_generate_time"].(string); ok && dailyTime != "" {
		config.DailyGenerateTime = dailyTime
	}
	if autoWeek, ok := updates["auto_generate_weekly"].(int); ok {
		config.AutoGenerateWeekly = autoWeek
	}
	if weekDay, ok := updates["weekly_generate_day"].(string); ok && weekDay != "" {
		config.WeeklyGenerateDay = weekDay
	}
	if weekTime, ok := updates["weekly_generate_time"].(string); ok && weekTime != "" {
		config.WeeklyGenerateTime = weekTime
	}
	if autoMonth, ok := updates["auto_generate_monthly"].(int); ok {
		config.AutoGenerateMonthly = autoMonth
	}
	if monthDate, ok := updates["monthly_generate_date"].(int); ok {
		config.MonthlyGenerateDate = monthDate
	}
	if monthTime, ok := updates["monthly_generate_time"].(string); ok && monthTime != "" {
		config.MonthlyGenerateTime = monthTime
	}
	if topHosts, ok := updates["top_hosts_count"].(int); ok && topHosts > 0 {
		config.TopHostsCount = topHosts
	}
	if lang, ok := updates["language"].(string); ok && lang != "" {
		config.Language = lang
	}
	if enableLLM, ok := updates["enable_llm_summary"].(int); ok {
		config.EnableLLMSummary = enableLLM
	}

	if err := repository.UpdateReportConfigDAO(config); err != nil {
		return err
	}

	// Restart scheduler with new config
	StopCronScheduler()
	InitCronScheduler()

	return nil
}
