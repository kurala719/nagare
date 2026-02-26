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
	Name              string `json:"name"`
	Hostid            string `json:"hostid"` // External ID from monitoring system
	MonitorID         uint   `gorm:"column:m_id;type:bigint unsigned" json:"m_id"`
	Monitor           Monitor `gorm:"foreignKey:MonitorID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	GroupID           uint   `gorm:"column:group_id;type:bigint unsigned" json:"group_id"`
	Group             Group   `gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	Description       string `json:"description"`
	Enabled           int    `gorm:"default:1" json:"enabled"` // 0 = disabled, 1 = enabled
	Status            int    `json:"status"`                   // 0 = inactive, 1 = active, 2 = error, 3 = syncing
	StatusDescription string `json:"status_description"`       // Reason for error status
	ActiveAvailable   string `gorm:"column:active_available" json:"active_available"`
	IPAddr            string `gorm:"column:ip_addr" json:"ip_addr"`
	Comment           string `json:"comment"`
	SSHUser           string `gorm:"column:ssh_user" json:"ssh_user"`
	SSHPassword       string `gorm:"column:ssh_password" json:"-"`
	SSHPort           int    `gorm:"column:ssh_port;default:22" json:"ssh_port"`
	// SNMP Configuration
	SNMPCommunity       string     `gorm:"column:snmp_community" json:"snmp_community"`
	SNMPVersion         string     `gorm:"column:snmp_version" json:"snmp_version"` // "v1", "v2c", "v3"
	SNMPPort            int        `gorm:"column:snmp_port;default:161" json:"snmp_port"`
	SNMPV3User          string     `gorm:"column:snmp_v3_user" json:"snmp_v3_user"`
	SNMPV3AuthPass      string     `gorm:"column:snmp_v3_auth_pass" json:"-"`
	SNMPV3PrivPass      string     `gorm:"column:snmp_v3_priv_pass" json:"-"`
	SNMPV3AuthProtocol  string     `gorm:"column:snmp_v3_auth_protocol" json:"snmp_v3_auth_protocol"`
	SNMPV3PrivProtocol  string     `gorm:"column:snmp_v3_priv_protocol" json:"snmp_v3_priv_protocol"`
	SNMPV3SecurityLevel string     `gorm:"column:snmp_v3_security_level" json:"snmp_v3_security_level"`
	LastSyncAt          *time.Time `json:"last_sync_at"`
	ExternalSource      string     `gorm:"column:external_source" json:"external_source"`
	HealthScore         int        `gorm:"column:health_score;default:100" json:"health_score"`
	GroupName           string     `gorm:"->" json:"group_name"`
	MonitorName         string     `gorm:"->" json:"monitor_name"`
}

// Group represents a logical group of hosts
type Group struct {
	gorm.Model
	Name              string `json:"name"`
	Description       string `json:"description"`
	MonitorID         uint   `gorm:"column:m_id;type:bigint unsigned" json:"m_id"`
	Monitor           Monitor `gorm:"foreignKey:MonitorID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	ExternalID        string `gorm:"column:external_id" json:"external_id"` // External ID from monitoring system (e.g., Zabbix groupid)
	Enabled           int    `gorm:"default:1" json:"enabled"`              // 0 = disabled, 1 = enabled
	Status            int    `json:"status"`                                // 0 = inactive, 1 = active, 2 = error, 3 = syncing
	StatusDescription string `json:"status_description"`                    // Reason for error status
	ActiveAvailable   string `gorm:"column:active_available" json:"active_available"` // 0=unknown, 1=available, 2=not_available
	LastSyncAt        *time.Time `json:"last_sync_at"`
	ExternalSource    string `gorm:"column:external_source" json:"external_source"`
	HealthScore       int    `gorm:"column:health_score;default:100" json:"health_score"`
}

// Monitor represents a monitoring system (e.g., Zabbix)
type Monitor struct {
	gorm.Model
	Name              string `json:"name"`
	URL               string `json:"url"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	AuthToken         string `json:"auth_token"`
	EventToken        string `gorm:"size:64;uniqueIndex" json:"event_token"`
	Description       string `json:"description"`
	Type              int    `json:"type"` // 1 = snmp, 2 = zabbix, 3 = other
	Enabled           int    `gorm:"default:1" json:"enabled"` // 0 = disabled, 1 = enabled
	Status            int    `json:"status"` // 0 = inactive, 1 = active, 2 = error, 3 = syncing
	StatusDescription string `json:"status_description"` // Reason for error status (e.g., "connection timeout", "authentication failed")
	HealthScore       int    `gorm:"column:health_score;default:100" json:"health_score"`
}

// Alarm represents an external alert source (e.g., Zabbix)
type Alarm struct {
	gorm.Model
	Name              string `json:"name"`
	URL               string `json:"url"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	AuthToken         string `json:"auth_token"`
	EventToken        string `gorm:"size:64;uniqueIndex" json:"event_token"`
	Description       string `json:"description"`
	Type              int    `json:"type"` // 1 = snmp, 2 = zabbix, 3 = other
	Enabled           int    `gorm:"default:1" json:"enabled"` // 0 = disabled, 1 = enabled
	Status            int    `json:"status"` // 0 = inactive, 1 = active, 2 = error, 3 = syncing
	StatusDescription string `json:"status_description"` // Reason for error status (e.g., "connection timeout", "authentication failed")
}

// Item represents a monitoring item/metric
type Item struct {
	gorm.Model
	Name              string `json:"name"`
	HID               uint   `gorm:"column:hid;type:bigint unsigned" json:"hid"`       // Internal host ID (foreign key to hosts table)
	Host              Host   `gorm:"foreignKey:HID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	ItemID            string `gorm:"column:itemid" json:"itemid"` // External ID from monitoring system
	ExternalHostID    string `gorm:"column:hostid" json:"hostid"` // External host ID from monitoring system
	ValueType         string `json:"value_type"`
	LastValue         string `json:"last_value"`
	Units             string `json:"units"`
	Enabled           int    `gorm:"default:1" json:"enabled"` // 0 = disabled, 1 = enabled
	Status            int    `json:"status"`                   // 0 = inactive, 1 = active, 2 = error, 3 = syncing
	StatusDescription string `json:"status_description"`       // Reason for error status (e.g., "host is down", "pull failed")
	Comment           string `json:"comment"`
	LastSyncAt        *time.Time `json:"last_sync_at"`
	ExternalSource    string `gorm:"column:external_source" json:"external_source"`
	HostName          string `gorm:"->" json:"host_name"`
}

// ItemHistory tracks item metric values over time.
type ItemHistory struct {
	gorm.Model
	ItemID    uint `gorm:"index;type:bigint unsigned"`
	Item      Item `gorm:"foreignKey:ItemID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	HostID    uint `gorm:"index;type:bigint unsigned"`
	Host      Host `gorm:"foreignKey:HostID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Value     string
	Units     string
	Status    int
	SampledAt time.Time `gorm:"index"`
}

// HostHistory tracks host status over time.
type HostHistory struct {
	gorm.Model
	HostID            uint `gorm:"index;type:bigint unsigned"`
	Host              Host `gorm:"foreignKey:HostID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Status            int
	StatusDescription string
	IPAddr            string
	SampledAt         time.Time `gorm:"index"`
}

// NetworkStatusHistory tracks overall network health over time.
type NetworkStatusHistory struct {
	gorm.Model
	Score         int
	MonitorTotal  int
	MonitorActive int
	GroupTotal    int
	GroupActive   int
	GroupImpacted int
	HostTotal     int
	HostActive    int
	ItemTotal     int
	ItemActive    int
	SampledAt     time.Time `gorm:"index"`
}

// Alert represents an alert/notification
type Alert struct {
	gorm.Model
	Message   string `gorm:"size:512"`
	Severity  int    `gorm:"type:int"`
	Status    int    `gorm:"type:int"` // 0 = active, 1 = acknowledged, 2 = resolved
	AlarmID   *uint  `gorm:"column:alarm_id;type:bigint unsigned"`
	Alarm     *Alarm `gorm:"foreignKey:AlarmID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	TriggerID *uint  `gorm:"column:trigger_id;type:bigint unsigned"`
	Trigger   *Trigger `gorm:"foreignKey:TriggerID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	HostID    *uint  `gorm:"type:bigint unsigned"`
	Host      *Host  `gorm:"foreignKey:HostID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	ItemID    *uint  `gorm:"type:bigint unsigned"`
	// Item      *Item    `gorm:"foreignKey:ItemID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	Comment   string
	HostName  string `gorm:"->"`
	ItemName  string `gorm:"->"`
	AlarmName string `gorm:"->"`
}

// Media represents a notification delivery target
type Media struct {
	gorm.Model
	Name        string
	Type        string            // "email", "other", "qq", etc.
	Target      string            // address/endpoint/number
	Params      map[string]string `gorm:"type:json;serializer:json"`
	Enabled     int               `gorm:"default:1"` // 0 = disabled, 1 = enabled
	Status      int               // 0 = inactive, 1 = active, 2 = error, 3 = syncing
	Description string
}

// Action represents an action executed for alerts
type Action struct {
	gorm.Model
	Name        string
	MediaID     uint     `gorm:"type:bigint unsigned"`
	Media       Media    `gorm:"foreignKey:MediaID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Template    string // message template
	Enabled     int    `gorm:"default:1;type:int"` // 0 = disabled, 1 = enabled
	Status      int    `gorm:"type:int"`           // 0 = inactive, 1 = active, 2 = error, 3 = syncing
	Description string
	// Filter conditions for executing this action
	SeverityMin *int  `gorm:"default:0;type:int"` // Filter alerts with severity >= this
	TriggerID   *uint `gorm:"index;type:bigint unsigned"`     // Optional: specific trigger
	Trigger     *Trigger `gorm:"foreignKey:TriggerID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	HostID      *uint `gorm:"index;type:bigint unsigned"`     // Optional: specific host
	Host        *Host    `gorm:"foreignKey:HostID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	GroupID     *uint `gorm:"index;type:bigint unsigned"`     // Optional: specific group
	Group       *Group   `gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	AlertStatus *int  `gorm:"default:0;type:int"` // Filter by alert status (0=active, 1=ack, 2=resolved)
	Users       []User `gorm:"many2many:action_users;"`
}

// Trigger represents a rule that filters alerts or logs to invoke an action
type Trigger struct {
	gorm.Model
	Name                  string   `json:"name"`
	Entity                string   `json:"entity"` // "alert" or "log"
	Severity              int      `json:"severity;type:int"`
	AlertID               *uint    `gorm:"column:alert_id;type:bigint unsigned" json:"alert_id"`
	Alert                 *Alert   `gorm:"foreignKey:AlertID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	AlertStatus           *int     `gorm:"column:alert_status;type:int" json:"alert_status"`
	AlertGroupID          *uint    `gorm:"column:alert_group_id;type:bigint unsigned" json:"alert_group_id"`
	AlertGroup            *Group   `gorm:"foreignKey:AlertGroupID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	AlertMonitorID        *uint    `gorm:"column:alert_monitor_id;type:bigint unsigned" json:"alert_monitor_id"`
	AlertMonitor          *Monitor `gorm:"foreignKey:AlertMonitorID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	AlertHostID           *uint    `gorm:"column:alert_host_id;type:bigint unsigned" json:"alert_host_id"`
	AlertHost             *Host    `gorm:"foreignKey:AlertHostID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	AlertItemID           *uint    `gorm:"column:alert_item_id;type:bigint unsigned" json:"alert_item_id"`
	AlertItem             *Item    `gorm:"foreignKey:AlertItemID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	AlertQuery            string   `json:"alert_query"`
	LogType               string   `gorm:"column:log_type" json:"log_type"`
	LogSeverity           *int     `gorm:"column:log_level;type:int" json:"log_level"`
	LogQuery              string   `gorm:"column:log_query" json:"log_query"`
	ItemStatus            *int     `gorm:"column:item_status;type:int" json:"item_status"`
	ItemValueThreshold    *float64 `gorm:"column:item_value_threshold" json:"item_value_threshold"`
	ItemValueThresholdMax *float64 `gorm:"column:item_value_threshold_max" json:"item_value_threshold_max"`
	ItemValueOperator     string   `gorm:"column:item_value_operator" json:"item_value_operator"`
	Enabled               int      `gorm:"default:1;type:int" json:"enabled"` // 0 = disabled, 1 = enabled
	Status                int      `json:"status;type:int"`                   // 0 = inactive, 1 = active, 2 = error, 3 = syncing
}

// Provider represents an AI provider (e.g., Google Gemini)
type Provider struct {
	gorm.Model
	Name         string
	URL          string
	APIKey       string
	DefaultModel string
	Models       []string `gorm:"type:json;serializer:json"` // List of available models
	Type         int      // Provider type: 1 = Gemini, 2 = OpenAI
	Description  string
	Enabled      int `gorm:"default:1"` // 0 = disabled, 1 = enabled
	Status       int // 0 = inactive, 1 = active, 2 = error, 3 = syncing
}

// Chat represents a chat message
type Chat struct {
	gorm.Model
	UserID     uint
	User       User     `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	ProviderID uint
	Provider   Provider `gorm:"foreignKey:ProviderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	LLMModel   string `gorm:"column:model"`
	Role       string // "user" or "assistant"
	Content    string
}

// ChatMessage is used for AI interactions
type ChatMessage struct {
	gorm.Model
	Role    string // "user" or "assistant"
	Content string
}

// LogEntry represents a system or service log entry
type LogEntry struct {
	gorm.Model
	Type     string // "system" or "service"
	Severity int    `gorm:"column:level;type:int"` // 0=info, 1=warn, 2=error
	Message  string `gorm:"size:512"`
	Context  string `gorm:"type:text"`
	UserID   *uint  `gorm:"type:bigint unsigned"`
	User     *User  `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	IP       string
}

// AuditLog represents a record of a user's operational action for security and compliance.
type AuditLog struct {
	gorm.Model
	UserID    *uint  `gorm:"index"`
	User      *User   `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	Username  string `gorm:"size:100"`
	Action    string `gorm:"size:255"` // e.g., "Create Host", "Delete Monitor"
	Method    string `gorm:"size:10"`  // GET, POST, PUT, DELETE
	Path      string `gorm:"size:255"` // API path
	IP        string `gorm:"size:45"`
	Status    int    // HTTP status code
	Latency   int64  // Latency in microseconds
	UserAgent string `gorm:"size:255"`
}

// User represents the unified authentication and profile information
type User struct {
	// Fields managed manually in migration.go to avoid GORM schema issues
	ID           uint `gorm:"primaryKey"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time `gorm:"index"`
	Username     string     `gorm:"size:100;uniqueIndex:idx_username"`
	Password     string     `json:"-"`         // Hashed password, excluded from JSON by default
	Privileges   int        `gorm:"default:1"` // 0=unauthorized, 1=user, 2=admin, 3=superadmin
	Status       int        `gorm:"default:1"` // 0=inactive, 1=active
	Email        string     `gorm:"size:255"`
	Phone        string     `gorm:"size:20"`
	Avatar       string     `gorm:"size:255"`
	Address      string     `gorm:"size:255"`
	Introduction string     `gorm:"type:text"`
	Nickname     string     `gorm:"size:100"`
	QQ           string     `gorm:"size:20"`
}

// RegisterApplication represents a pending registration request from an unregistered user
type RegisterApplication struct {
	gorm.Model
	Username   string
	Password   string
	Email      string `gorm:"size:255"`
	Status     int    // 0 = pending, 1 = approved, 2 = rejected
	Reason     string // rejection or approval note
	ApprovedBy *uint  `gorm:"column:approved_by"`
	Approver   *User  `gorm:"foreignKey:ApprovedBy;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
}

// EmailVerification stores temporary verification codes sent to users
type EmailVerification struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"index;size:255"`
	Code      string    `gorm:"size:10"`
	ExpiresAt time.Time `gorm:"index"`
}

// PasswordResetApplication represents a request to reset a user password
type PasswordResetApplication struct {
	gorm.Model
	UserID      uint
	User        User   `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Username    string
	NewPassword string // Hashed new password to be applied upon approval
	Status      int    // 0 = pending, 1 = approved, 2 = rejected
	Reason      string
	ApprovedBy  *uint `gorm:"column:approved_by"`
	Approver    *User `gorm:"foreignKey:ApprovedBy;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
}

// QQWhitelist represents an allowed QQ user or group for commands and alerts
type QQWhitelist struct {
	gorm.Model
	QQIdentifier string `gorm:"index;uniqueIndex:idx_qq_type"` // QQ user ID or group ID
	Type         int    `gorm:"uniqueIndex:idx_qq_type"`       // 0 = user, 1 = group
	Nickname     string // Nickname or group name
	CanCommand   int    `gorm:"default:1"` // 0 = no, 1 = yes
	CanReceive   int    `gorm:"default:1"` // 0 = no, 1 = yes (receive alerts)
	Enabled      int    `gorm:"default:1"` // 0 = disabled, 1 = enabled
	Comment      string
}

// KnowledgeBase represents a local knowledge entry for RAG
type KnowledgeBase struct {
	gorm.Model
	Topic    string `gorm:"size:255;index"`
	Content  string `gorm:"type:text"`
	Keywords string `gorm:"size:255;index"` // Comma-separated keywords
	Category string `gorm:"size:50;index"`
}

// AnsiblePlaybook stores YAML content for Ansible operations
type AnsiblePlaybook struct {
	gorm.Model
	Name        string `gorm:"size:255"`
	Description string `gorm:"type:text"`
	Content     string `gorm:"type:text"` // YAML content
	Tags        string `gorm:"size:255"`  // Comma-separated tags
}

// AnsibleJob tracks execution of playbooks
type AnsibleJob struct {
	gorm.Model
	PlaybookID  uint
	Playbook    AnsiblePlaybook `gorm:"foreignKey:PlaybookID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Status      string          `gorm:"size:50"` // "pending", "running", "success", "failed"
	Output      string          `gorm:"type:longtext"`
	TriggeredBy *uint           `gorm:"index"`
	User        *User           `gorm:"foreignKey:TriggeredBy;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	HostFilter  string          `gorm:"size:255"` // Specific host or group filter
}

// ReportConfig stores configuration for automated report generation
type ReportConfig struct {
	gorm.Model
	AutoGenerateDaily   int    `gorm:"default:0"` // 0=disabled, 1=enabled
	DailyGenerateTime   string // "09:00"
	AutoGenerateWeekly  int    `gorm:"default:0"` // 0=disabled, 1=enabled
	WeeklyGenerateDay   string // "Monday", "Friday", etc.
	WeeklyGenerateTime  string // "09:00"
	AutoGenerateMonthly int    `gorm:"default:0"` // 0=disabled, 1=enabled
	MonthlyGenerateDate int    // 1-28
	MonthlyGenerateTime string // "09:00"
	IncludeAlerts       int    `gorm:"default:1"`
	IncludeMetrics      int    `gorm:"default:1"`
	TopHostsCount       int    `gorm:"default:5"`
	EnableLLMSummary    int    `gorm:"default:1"`
	EmailNotify         int    `gorm:"default:0"`
	EmailRecipients     string // Comma-separated emails
	Language            string `gorm:"size:10;default:en"` // "en", "zh"
}

// Report represents a generated PDF report
type Report struct {
	gorm.Model
	ReportType  string // "weekly", "monthly", "manual"
	Title       string
	FilePath    string
	DownloadURL string
	Status      int // 0=generating, 1=completed, 2=failed
	GeneratedAt time.Time
	ContentData string `gorm:"type:longtext"` // JSON content for preview
}

// SiteMessage represents an internal system notification for users
type SiteMessage struct {
	gorm.Model
	Title    string `gorm:"size:255"`
	Content  string `gorm:"type:text"`
	Type     string `gorm:"size:50"` // "alert", "sync", "system", "report"
	Severity int    `gorm:"type:int"` // 0=info, 1=success, 2=warn, 3=error
	IsRead   int    `gorm:"default:0;type:int"` // 0=unread, 1=read
	UserID   *uint  `gorm:"index;type:bigint unsigned"`     // Optional: target specific user, null for all
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
	DataType      string         `gorm:"size:50;uniqueIndex" json:"data_type"` // e.g., 'logs', 'alerts', 'audit_logs', 'item_history', 'host_history'
	RetentionDays int            `gorm:"default:30" json:"retention_days"`     // 0 means keep forever
	Enabled       *int           `gorm:"default:1" json:"enabled"`             // 0 = disabled, 1 = enabled
	Description   string         `gorm:"size:255" json:"description"`
}

// PacketAnalysis represents a packet capture analysis request
type PacketAnalysis struct {
	gorm.Model
	Name        string `gorm:"size:255" json:"name"`
	FilePath    string `gorm:"size:500" json:"file_path"`
	RawContent  string `gorm:"type:longtext" json:"raw_content"` // Hex or text snippet
	ProviderID  *uint  `json:"provider_id"`
	Provider    *Provider `gorm:"foreignKey:ProviderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	AIModel     string `gorm:"column:model;size:100" json:"model"`
	Status      int    `gorm:"default:0" json:"status"` // 0=pending, 1=analyzing, 2=completed, 3=failed
	RiskLevel   string `gorm:"size:50" json:"risk_level"` // "clean", "notable", "malicious"
	Analysis    string `gorm:"type:longtext" json:"analysis"`
	UserID      *uint  `json:"user_id"`
	User        *User   `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
}
