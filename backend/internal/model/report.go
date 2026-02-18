package model

import (
	"time"

	"gorm.io/gorm"
)

// Report represents a generated operational report
type Report struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	ReportType   string         `json:"report_type"`   // "weekly", "monthly"
	Title        string         `json:"title"`         // e.g., "Weekly Report 2025-02-17"
	GeneratedAt  time.Time      `json:"generated_at"`  // Report time period
	PeriodStart  time.Time      `json:"period_start"`  // Data collection start
	PeriodEnd    time.Time      `json:"period_end"`    // Data collection end
	FilePath     string         `json:"file_path"`     // Location of PDF file
	FileSize     int64          `json:"file_size"`     // File size in bytes
	Status       int            `json:"status"`        // 0=pending, 1=completed, 2=failed
	Summary      string         `json:"summary"`       // Executive summary (AI-generated)
	ErrorMessage string         `json:"error_message"` // If status=2
	HostCount    int            `json:"host_count"`    // Number of hosts in report
	AlertCount   int            `json:"alert_count"`   // Number of alerts during period
	Availability float64        `json:"availability"`  // Overall availability percentage
	Comment      string         `json:"comment"`       // Admin notes
}

// ReportItem represents a host's performance data in the report
type ReportItem struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ReportID        uint      `gorm:"index" json:"report_id"`
	HostID          uint      `json:"host_id"`
	HostName        string    `json:"host_name"`
	HostIP          string    `json:"host_ip"`
	CPUAvg          float64   `json:"cpu_avg"`          // Average CPU %
	CPUPeak         float64   `json:"cpu_peak"`         // Peak CPU %
	MemoryAvg       float64   `json:"memory_avg"`       // Average memory %
	MemoryPeak      float64   `json:"memory_peak"`      // Peak memory %
	DiskUsage       float64   `json:"disk_usage"`       // Disk usage %
	NetworkLatency  float64   `json:"network_latency"`  // Average latency ms
	AlertCount      int       `json:"alert_count"`      // Alerts for this host
	DowntimeMinutes int       `json:"downtime_minutes"` // Total downtime
	Rank            int       `json:"rank"`             // Ranking for Top N
	CreatedAt       time.Time `json:"created_at"`
}

// ReportConfig stores report generation settings
type ReportConfig struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	AutoGenerateWeekly  int       `json:"auto_generate_weekly"`  // 0=disabled, 1=enabled
	WeeklyGenerateDay   string    `json:"weekly_generate_day"`   // "Monday", "Sunday", etc
	WeeklyGenerateTime  string    `json:"weekly_generate_time"`  // "08:00:00"
	AutoGenerateMonthly int       `json:"auto_generate_monthly"` // 0=disabled, 1=enabled
	MonthlyGenerateDate int       `json:"monthly_generate_date"` // Day of month (1-28)
	MonthlyGenerateTime string    `json:"monthly_generate_time"` // "08:00:00"
	IncludeAlerts       int       `json:"include_alerts"`        // Include alerts section
	IncludeMetrics      int       `json:"include_metrics"`       // Include metrics section
	TopHostsCount       int       `json:"top_hosts_count"`       // How many top hosts to show (default 5)
	EnableLLMSummary    int       `json:"enable_llm_summary"`    // Use AI for executive summary
	EmailNotify         int       `json:"email_notify"`          // Send email after generation
	EmailRecipients     string    `json:"email_recipients"`      // Comma-separated emails
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// AlertTrend represents alert statistics for trend analysis
type AlertTrend struct {
	Date          time.Time `json:"date"`
	CriticalCount int       `json:"critical_count"`
	WarningCount  int       `json:"warning_count"`
	InfoCount     int       `json:"info_count"`
	ResolvedCount int       `json:"resolved_count"`
	TotalCount    int       `json:"total_count"`
}

// AvailabilityMetric represents availability data point
type AvailabilityMetric struct {
	Date             time.Time `json:"date"`
	AvailableHosts   int       `json:"available_hosts"`
	TotalHosts       int       `json:"total_hosts"`
	AvailabilityRate float64   `json:"availability_rate"`  // percentage
	ResponseTimeAvg  float64   `json:"response_time_avg"`  // ms
	ResponseTimePeak float64   `json:"response_time_peak"` // ms
}
