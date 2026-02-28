// package model contains the core business entities and interfaces for the Nagare system.
// This package follows clean architecture principles, keeping business logic
// independent of frameworks and external dependencies.
package model

import (
	"time"

	"gorm.io/gorm"
)

// Host represents a monitored host entity
type Host struct {
	gorm.Model
	Name              string `gorm:"type:varchar(255)" json:"name"`
	ExternalID        string `gorm:"column:external_id;type:varchar(64)" json:"external_id"` // External ID from monitoring system
	GroupID           uint   `gorm:"column:group_id;type:bigint unsigned" json:"group_id"`
	Group             Group  `gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	Description       string `gorm:"type:varchar(1024)" json:"description"`
	Enabled           int    `gorm:"type:tinyint;default:1" json:"enabled"`       // 0 = disabled, 1 = enabled
	Status            int    `gorm:"type:tinyint" json:"status"`                  // 0 = inactive, 1 = active, 2 = error, 3 = syncing
	StatusDescription string `gorm:"type:varchar(512)" json:"status_description"` // Reason for error status
	IPAddr            string `gorm:"column:ip_addr;type:varchar(39)" json:"ip_addr"`
	Comment           string `gorm:"type:text" json:"comment"`
	SSHUser           string `gorm:"column:ssh_user;type:varchar(100)" json:"ssh_user"`
	SSHPassword       string `gorm:"column:ssh_password;type:varchar(255)" json:"-"`
	SSHPort           int    `gorm:"column:ssh_port;default:22" json:"ssh_port"`
	// SNMP Configuration
	SNMPCommunity       string     `gorm:"column:snmp_community;type:varchar(100)" json:"snmp_community"`
	SNMPVersion         string     `gorm:"column:snmp_version;type:varchar(20)" json:"snmp_version"` // "v1", "v2c", "v3"
	SNMPPort            int        `gorm:"column:snmp_port;default:161" json:"snmp_port"`
	SNMPV3User          string     `gorm:"column:snmp_v3_user;type:varchar(100)" json:"snmp_v3_user"`
	SNMPV3AuthPass      string     `gorm:"column:snmp_v3_auth_pass;type:varchar(255)" json:"-"`
	SNMPV3PrivPass      string     `gorm:"column:snmp_v3_priv_pass;type:varchar(255)" json:"-"`
	SNMPV3AuthProtocol  string     `gorm:"column:snmp_v3_auth_protocol;type:varchar(50)" json:"snmp_v3_auth_protocol"`
	SNMPV3PrivProtocol  string     `gorm:"column:snmp_v3_priv_protocol;type:varchar(50)" json:"snmp_v3_priv_protocol"`
	SNMPV3SecurityLevel string     `gorm:"column:snmp_v3_security_level;type:varchar(50)" json:"snmp_v3_security_level"`
	LastSyncAt          *time.Time `json:"last_sync_at"`
	HealthScore         int        `gorm:"column:health_score;default:100" json:"health_score"`
}

// Group represents a logical group of hosts
type Group struct {
	gorm.Model
	Name              string     `gorm:"type:varchar(255)" json:"name"`
	Description       string     `gorm:"type:varchar(1024)" json:"description"`
	MonitorID         uint       `gorm:"column:monitor_id;type:bigint unsigned" json:"monitor_id"`
	Monitor           Monitor    `gorm:"foreignKey:MonitorID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	ExternalID        string     `gorm:"column:external_id;type:varchar(64)" json:"external_id"` // External ID from monitoring system (e.g., Zabbix groupid)
	Enabled           int        `gorm:"type:tinyint;default:1" json:"enabled"`                  // 0 = disabled, 1 = enabled
	Status            int        `gorm:"type:tinyint" json:"status"`                             // 0 = inactive, 1 = active, 2 = error, 3 = syncing
	StatusDescription string     `gorm:"type:varchar(512)" json:"status_description"`            // Reason for error status
	LastSyncAt        *time.Time `json:"last_sync_at"`
	HealthScore       int        `gorm:"column:health_score;default:100" json:"health_score"`
}

// Monitor represents a monitoring system (e.g., Zabbix)
type Monitor struct {
	gorm.Model
	Name              string `gorm:"type:varchar(255)" json:"name"`
	URL               string `gorm:"type:varchar(512)" json:"url"`
	Username          string `gorm:"type:varchar(100)" json:"username"`
	Password          string `gorm:"type:varchar(255)" json:"password"`
	AuthToken         string `gorm:"type:varchar(255)" json:"auth_token"`
	EventToken        string `gorm:"size:64;uniqueIndex" json:"event_token"`
	Description       string `gorm:"type:varchar(1024)" json:"description"`
	Type              int    `gorm:"type:tinyint" json:"type"`                    // 1 = snmp, 2 = zabbix, 3 = other
	Enabled           int    `gorm:"type:tinyint;default:1" json:"enabled"`       // 0 = disabled, 1 = enabled
	Status            int    `gorm:"type:tinyint" json:"status"`                  // 0 = inactive, 1 = active, 2 = error, 3 = syncing
	StatusDescription string `gorm:"type:varchar(512)" json:"status_description"` // Reason for error status (e.g., "connection timeout", "authentication failed")
	HealthScore       int    `gorm:"column:health_score;default:100" json:"health_score"`
}

// Alarm represents an external alert source (e.g., Zabbix)
type Alarm struct {
	gorm.Model
	Name              string `gorm:"type:varchar(255)" json:"name"`
	URL               string `gorm:"type:varchar(512)" json:"url"`
	Username          string `gorm:"type:varchar(100)" json:"username"`
	Password          string `gorm:"type:varchar(255)" json:"password"`
	AuthToken         string `gorm:"type:varchar(255)" json:"auth_token"`
	EventToken        string `gorm:"size:64;uniqueIndex" json:"event_token"`
	Description       string `gorm:"type:varchar(1024)" json:"description"`
	Type              int    `gorm:"type:tinyint" json:"type"`                    // 1 = snmp, 2 = zabbix, 3 = other
	Enabled           int    `gorm:"type:tinyint;default:1" json:"enabled"`       // 0 = disabled, 1 = enabled
	Status            int    `gorm:"type:tinyint" json:"status"`                  // 0 = inactive, 1 = active, 2 = error, 3 = syncing
	StatusDescription string `gorm:"type:varchar(512)" json:"status_description"` // Reason for error status (e.g., "connection timeout", "authentication failed")
}

// Item represents a monitoring item/metric
type Item struct {
	gorm.Model
	Name              string     `gorm:"type:varchar(255)" json:"name"`
	HostID            uint       `gorm:"index;type:bigint unsigned" json:"host_id"` // Internal host ID (foreign key to hosts table)
	Host              Host       `gorm:"foreignKey:HostID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	ExternalID        string     `gorm:"column:external_id;type:varchar(64)" json:"external_id"` // External ID from monitoring system
	ValueType         string     `gorm:"type:varchar(100)" json:"value_type"`
	LastValue         string     `gorm:"type:varchar(255)" json:"last_value"`
	Units             string     `gorm:"type:varchar(100)" json:"units"`
	Enabled           int        `gorm:"type:tinyint;default:1" json:"enabled"`       // 0 = disabled, 1 = enabled
	Status            int        `gorm:"type:tinyint" json:"status"`                  // 0 = inactive, 1 = active, 2 = error, 3 = syncing
	StatusDescription string     `gorm:"type:varchar(512)" json:"status_description"` // Reason for error status (e.g., "host is down", "pull failed")
	Comment           string     `gorm:"type:text" json:"comment"`
	LastSyncAt        *time.Time `json:"last_sync_at"`
}

// ItemHistory tracks item metric values over time.
type ItemHistory struct {
	gorm.Model
	ItemID    uint      `gorm:"index:idx_item_sampled,priority:1;type:bigint unsigned" json:"item_id"`
	Item      Item      `gorm:"foreignKey:ItemID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	HostID    uint      `gorm:"index;type:bigint unsigned" json:"host_id"`
	Host      Host      `gorm:"foreignKey:HostID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Value     string    `gorm:"type:varchar(128)" json:"value"`
	Units     string    `gorm:"type:varchar(100)" json:"units"`
	Status    int       `gorm:"type:tinyint" json:"status"`
	SampledAt time.Time `gorm:"index:idx_item_sampled,priority:2" json:"sampled_at"`
}

// HostHistory tracks host status over time.
type HostHistory struct {
	gorm.Model
	HostID            uint      `gorm:"index:idx_host_sampled,priority:1;type:bigint unsigned" json:"host_id"`
	Host              Host      `gorm:"foreignKey:HostID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Status            int       `gorm:"type:tinyint" json:"status"`
	StatusDescription string    `gorm:"type:varchar(512)" json:"status_description"`
	ItemTotal         int       `json:"item_total"`
	ItemActive        int       `json:"item_active"`
	IPAddr            string    `gorm:"type:varchar(39)" json:"ip_addr"`
	SampledAt         time.Time `gorm:"index:idx_host_sampled,priority:2" json:"sampled_at"`
}

// NetworkStatusHistory tracks overall network health over time.
type NetworkStatusHistory struct {
	gorm.Model
	Score         int       `json:"score"`
	MonitorTotal  int       `json:"monitor_total"`
	MonitorActive int       `json:"monitor_active"`
	GroupTotal    int       `json:"group_total"`
	GroupActive   int       `json:"group_active"`
	GroupImpacted int       `json:"group_impacted"`
	HostTotal     int       `json:"host_total"`
	HostActive    int       `json:"host_active"`
	ItemTotal     int       `json:"item_total"`
	ItemActive    int       `json:"item_active"`
	SampledAt     time.Time `gorm:"index" json:"sampled_at"`
}

// Alert represents an alert/notification
type Alert struct {
	gorm.Model
	Message   string   `gorm:"type:varchar(512)" json:"message"`
	Severity  int      `gorm:"type:tinyint" json:"severity"`
	Status    int      `gorm:"type:tinyint" json:"status"` // 0 = active, 1 = acknowledged, 2 = resolved
	AlarmID   *uint    `gorm:"column:alarm_id;type:bigint unsigned" json:"alarm_id"`
	Alarm     *Alarm   `gorm:"foreignKey:AlarmID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	TriggerID *uint    `gorm:"column:trigger_id;type:bigint unsigned" json:"trigger_id"`
	Trigger   *Trigger `gorm:"foreignKey:TriggerID;constraint:-;" json:"-"`
	HostID    *uint    `gorm:"type:bigint unsigned" json:"host_id"`
	Host      *Host    `gorm:"foreignKey:HostID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	ItemID    *uint    `gorm:"type:bigint unsigned" json:"item_id"`
	Comment   string   `gorm:"type:text" json:"comment"`
}

// Media represents a notification delivery target
type Media struct {
	gorm.Model
	Name        string            `gorm:"type:varchar(100)" json:"name"`
	Type        string            `gorm:"type:varchar(50)" json:"type"`    // "email", "other", "qq", etc.
	Target      string            `gorm:"type:varchar(255)" json:"target"` // address/endpoint/number
	Params      map[string]string `gorm:"type:json;serializer:json" json:"params"`
	Enabled     int               `gorm:"type:tinyint;default:1" json:"enabled"` // 0 = disabled, 1 = enabled
	Status      int               `gorm:"type:tinyint" json:"status"`            // 0 = inactive, 1 = active, 2 = error, 3 = syncing
	Description string            `gorm:"type:varchar(1024)" json:"description"`
}

// Action represents an action executed for alerts
type Action struct {
	gorm.Model
	Name        string `gorm:"type:varchar(255)" json:"name"`
	MediaID     uint   `gorm:"type:bigint unsigned" json:"media_id"`
	Media       Media  `gorm:"foreignKey:MediaID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Template    string `gorm:"type:text" json:"template"`             // message template
	Enabled     int    `gorm:"type:tinyint;default:1" json:"enabled"` // 0 = disabled, 1 = enabled
	Status      int    `gorm:"type:tinyint" json:"status"`            // 0 = inactive, 1 = active, 2 = error, 3 = syncing
	Description string `gorm:"type:varchar(1024)" json:"description"`
	SeverityMin *int   `gorm:"type:tinyint;default:0" json:"severity_min"` // Filter alerts with severity >= this
	HostID      *uint  `gorm:"index;type:bigint unsigned" json:"host_id"`  // Optional: specific host
	Host        *Host  `gorm:"foreignKey:HostID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	GroupID     *uint  `gorm:"index;type:bigint unsigned" json:"group_id"` // Optional: specific group
	Group       *Group `gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	AlertStatus *int   `gorm:"type:tinyint;default:0" json:"alert_status"` // Filter by alert status (0=active, 1=ack, 2=resolved)
	UserID      *uint  `gorm:"index;type:bigint unsigned" json:"user_id"`
	User        *User  `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
}

// Trigger represents a rule that filters alerts or logs to invoke an action
type Trigger struct {
	gorm.Model
	Name                  string   `gorm:"type:varchar(255)" json:"name"`
	Severity              int      `gorm:"type:tinyint" json:"severity"`
	AlertID               *uint    `gorm:"column:alert_id;type:bigint unsigned" json:"alert_id"`
	Alert                 *Alert   `gorm:"foreignKey:AlertID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	ItemStatus            *int     `gorm:"column:item_status;type:tinyint" json:"item_status"`
	ItemValueThreshold    *float64 `gorm:"column:item_value_threshold" json:"item_value_threshold"`
	ItemValueThresholdMax *float64 `gorm:"column:item_value_threshold_max" json:"item_value_threshold_max"`
	ItemValueOperator     string   `gorm:"column:item_value_operator;type:varchar(50)" json:"item_value_operator"`
	Enabled               int      `gorm:"type:tinyint;default:1" json:"enabled"` // 0 = disabled, 1 = enabled
	Status                int      `gorm:"type:tinyint" json:"status"`            // 0 = inactive, 1 = active, 2 = error, 3 = syncing
}

// Provider represents an AI provider (e.g., Google Gemini)
type Provider struct {
	gorm.Model
	Name         string   `gorm:"type:varchar(100)" json:"name"`
	URL          string   `gorm:"type:varchar(512)" json:"url"`
	APIKey       string   `gorm:"type:varchar(255)" json:"api_key"`
	DefaultModel string   `gorm:"type:varchar(100)" json:"default_model"`
	Models       []string `gorm:"type:json;serializer:json" json:"models"` // List of available models
	Type         int      `gorm:"type:tinyint" json:"type"`                // Provider type: 1 = Gemini, 2 = OpenAI
	Description  string   `gorm:"type:varchar(1024)" json:"description"`
	Enabled      int      `gorm:"type:tinyint;default:1" json:"enabled"` // 0 = disabled, 1 = enabled
	Status       int      `gorm:"type:tinyint" json:"status"`            // 0 = inactive, 1 = active, 2 = error, 3 = syncing
}

// Chat represents a chat message
type Chat struct {
	gorm.Model
	UserID     uint     `json:"user_id"`
	User       User     `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	ProviderID uint     `json:"provider_id"`
	Provider   Provider `gorm:"foreignKey:ProviderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	LLMModel   string   `gorm:"column:model;type:varchar(100)" json:"model"`
	Role       string   `gorm:"type:varchar(50)" json:"role"` // "user" or "assistant"
	Content    string   `gorm:"type:text" json:"content"`
}

// ChatMessage is used for AI interactions
type ChatMessage struct {
	gorm.Model
	Role    string `gorm:"type:varchar(50)" json:"role"` // "user" or "assistant"
	Content string `gorm:"type:text" json:"content"`
}

// LogEntry represents a system log entry
type LogEntry struct {
	gorm.Model
	Severity int    `gorm:"column:level;type:tinyint" json:"severity"` // 0=info, 1=warn, 2=error
	Message  string `gorm:"type:varchar(512)" json:"message"`
	Context  string `gorm:"type:text" json:"context"`
	UserID   *uint  `gorm:"type:bigint unsigned" json:"user_id"`
	User     *User  `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	IP       string `gorm:"type:varchar(39)" json:"ip"`
}

// AuditLog represents a record of a user's operational action for security and compliance.
type AuditLog struct {
	gorm.Model
	UserID    *uint  `gorm:"index" json:"user_id"`
	User      *User  `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	Username  string `gorm:"type:varchar(100)" json:"username"`
	Action    string `gorm:"type:varchar(255)" json:"action"` // e.g., "Create Host", "Delete Monitor"
	Method    string `gorm:"type:varchar(10)" json:"method"`  // GET, POST, PUT, DELETE
	Path      string `gorm:"type:varchar(255)" json:"path"`   // API path
	IP        string `gorm:"type:varchar(39)" json:"ip"`
	Status    int    `gorm:"type:smallint" json:"status"` // HTTP status code
	Latency   int64  `json:"latency"`                     // Latency in microseconds
	UserAgent string `gorm:"type:varchar(255)" json:"user_agent"`
}

// User represents the unified authentication and profile information
type User struct {
	// Fields managed manually in migration.go to avoid GORM schema issues
	ID           uint       `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"-"`
	Username     string     `gorm:"size:100;uniqueIndex:idx_username" json:"username"`
	Password     string     `json:"-"`                           // Hashed password, excluded from JSON by default
	Privileges   int        `gorm:"default:1" json:"privileges"` // 0=unauthorized, 1=user, 2=admin, 3=superadmin
	Status       int        `gorm:"default:1" json:"status"`     // 0=inactive, 1=active
	Email        string     `gorm:"size:255" json:"email"`
	Phone        string     `gorm:"size:20" json:"phone"`
	Avatar       string     `gorm:"size:255" json:"avatar"`
	Address      string     `gorm:"size:255" json:"address"`
	Introduction string     `gorm:"type:text" json:"introduction"`
	Nickname     string     `gorm:"size:100" json:"nickname"`
	QQ           string     `gorm:"size:20" json:"qq"`
}

// RegisterApplication represents a pending registration request from an unregistered user
type RegisterApplication struct {
	gorm.Model
	Username   string `json:"username"`
	Password   string `json:"-"`
	Email      string `gorm:"size:255" json:"email"`
	Status     int    `json:"status"` // 0 = pending, 1 = approved, 2 = rejected
	Reason     string `json:"reason"` // rejection or approval note
	ApprovedBy *uint  `gorm:"column:approved_by" json:"approved_by"`
	Approver   *User  `gorm:"foreignKey:ApprovedBy;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
}

// PasswordResetApplication represents a request to reset a user password
type PasswordResetApplication struct {
	gorm.Model
	UserID      uint   `json:"user_id"`
	User        User   `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Username    string `json:"username"`
	NewPassword string `json:"-"`      // Hashed new password to be applied upon approval
	Status      int    `json:"status"` // 0 = pending, 1 = approved, 2 = rejected
	Reason      string `json:"reason"`
	ApprovedBy  *uint  `gorm:"column:approved_by" json:"approved_by"`
	Approver    *User  `gorm:"foreignKey:ApprovedBy;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
}

// QQWhitelist represents an allowed QQ user or group for commands and alerts
type QQWhitelist struct {
	gorm.Model
	QQIdentifier string `gorm:"index;uniqueIndex:idx_qq_type" json:"qq_identifier"` // QQ user ID or group ID
	Type         int    `gorm:"uniqueIndex:idx_qq_type" json:"type"`                // 0 = user, 1 = group
	Nickname     string `json:"nickname"`                                           // Nickname or group name
	CanCommand   int    `gorm:"default:1" json:"can_command"`                       // 0 = no, 1 = yes
	CanReceive   int    `gorm:"default:1" json:"can_receive"`                       // 0 = no, 1 = yes (receive alerts)
	Enabled      int    `gorm:"default:1" json:"enabled"`                           // 0 = disabled, 1 = enabled
	Comment      string `json:"comment"`
}

// KnowledgeBase represents a local knowledge entry for RAG
type KnowledgeBase struct {
	gorm.Model
	Topic    string `gorm:"type:varchar(255);index" json:"topic"`
	Content  string `gorm:"type:text" json:"content"`
	Keywords string `gorm:"type:varchar(255);index" json:"keywords"` // Comma-separated keywords
	Category string `gorm:"type:varchar(50);index" json:"category"`
}

// ReportConfig stores configuration for automated report generation
type ReportConfig struct {
	gorm.Model
	AutoGenerateDaily   int    `gorm:"type:tinyint;default:0" json:"auto_generate_daily"`   // 0=disabled, 1=enabled
	DailyGenerateTime   string `gorm:"type:varchar(20)" json:"daily_generate_time"`         // "09:00"
	AutoGenerateWeekly  int    `gorm:"type:tinyint;default:0" json:"auto_generate_weekly"`  // 0=disabled, 1=enabled
	WeeklyGenerateDay   string `gorm:"type:varchar(20)" json:"weekly_generate_day"`         // "Monday", "Friday", etc.
	WeeklyGenerateTime  string `gorm:"type:varchar(20)" json:"weekly_generate_time"`        // "09:00"
	AutoGenerateMonthly int    `gorm:"type:tinyint;default:0" json:"auto_generate_monthly"` // 0=disabled, 1=enabled
	MonthlyGenerateDate int    `gorm:"type:tinyint" json:"monthly_generate_date"`           // 1-28
	MonthlyGenerateTime string `gorm:"type:varchar(20)" json:"monthly_generate_time"`       // "09:00"
	IncludeAlerts       int    `gorm:"type:tinyint;default:1" json:"include_alerts"`
	IncludeMetrics      int    `gorm:"type:tinyint;default:1" json:"include_metrics"`
	TopHostsCount       int    `gorm:"type:int;default:5" json:"top_hosts_count"`
	EnableLLMSummary    int    `gorm:"type:tinyint;default:1" json:"enable_llm_summary"`
	EmailNotify         int    `gorm:"type:tinyint;default:0" json:"email_notify"`
	EmailRecipients     string `gorm:"type:text" json:"email_recipients"`             // Comma-separated emails
	Language            string `gorm:"type:varchar(10);default:'en'" json:"language"` // "en", "zh"
}

// Report represents a generated PDF report
type Report struct {
	gorm.Model
	ReportType  string    `gorm:"type:varchar(50)" json:"report_type"` // "weekly", "monthly", "manual"
	Title       string    `gorm:"type:varchar(255)" json:"title"`
	FilePath    string    `gorm:"type:varchar(512)" json:"file_path"`
	DownloadURL string    `gorm:"type:varchar(512)" json:"download_url"`
	Status      int       `gorm:"type:tinyint" json:"status"` // 0=generating, 1=completed, 2=failed
	GeneratedAt time.Time `json:"generated_at"`
	ContentData string    `gorm:"type:longtext" json:"content_data"` // JSON content for preview
}

// SiteMessage represents an internal system notification for users
type SiteMessage struct {
	gorm.Model
	Title    string `gorm:"type:varchar(255)" json:"title"`
	Content  string `gorm:"type:text" json:"content"`
	Type     string `gorm:"type:varchar(50)" json:"type"`              // "alert", "sync", "system", "report"
	Severity int    `gorm:"type:tinyint" json:"severity"`              // 0=info, 1=success, 2=warn, 3=error
	IsRead   int    `gorm:"type:tinyint;default:0" json:"is_read"`     // 0=unread, 1=read
	UserID   *uint  `gorm:"index;type:bigint unsigned" json:"user_id"` // Optional: target specific user, null for all
	User     *User  `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
}

type UserWithInfo struct {
	User
}

// RetentionPolicy defines how long data for a specific part of the system should be kept.
type RetentionPolicy struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	DataType      string         `gorm:"type:varchar(50);uniqueIndex" json:"data_type"` // e.g., 'logs', 'alerts', 'audit_logs', 'item_history', 'host_history'
	RetentionDays int            `gorm:"type:int;default:30" json:"retention_days"`     // 0 means keep forever
	Enabled       *int           `gorm:"type:tinyint;default:1" json:"enabled"`         // 0 = disabled, 1 = enabled
	Description   string         `gorm:"type:varchar(255)" json:"description"`
}

// PacketAnalysis represents a packet capture analysis request
type PacketAnalysis struct {
	gorm.Model
	Name       string    `gorm:"type:varchar(255)" json:"name"`
	FilePath   string    `gorm:"type:varchar(500)" json:"file_path"`
	RawContent string    `gorm:"type:longtext" json:"raw_content"` // Hex or text snippet
	ProviderID *uint     `json:"provider_id"`
	Provider   *Provider `gorm:"foreignKey:ProviderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	AIModel    string    `gorm:"column:model;type:varchar(100)" json:"model"`
	Status     int       `gorm:"type:tinyint;default:0" json:"status"` // 0=pending, 1=analyzing, 2=completed, 3=failed
	RiskLevel  string    `gorm:"type:varchar(50)" json:"risk_level"`   // "clean", "notable", "malicious"
	Analysis   string    `gorm:"type:longtext" json:"analysis"`
	UserID     *uint     `json:"user_id"`
	User       *User     `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
}
