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
	Name              string
	Hostid            string // External ID from monitoring system
	MonitorID         uint   `gorm:"column:m_id"`
	GroupID           uint   `gorm:"column:group_id"`
	Description       string
	Enabled           int    `gorm:"default:1"` // 0 = disabled, 1 = enabled
	Status            int    // 0 = inactive, 1 = active, 2 = error, 3 = syncing
	StatusDescription string // Reason for error status (e.g., "monitor is down", "connection failed")
	ActiveAvailable   string `gorm:"column:active_available"` // 0=unknown, 1=available, 2=not_available
	IPAddr            string `gorm:"column:ip_addr"`
	Comment           string
}

// Group represents a logical group of hosts
type Group struct {
	gorm.Model
	Name        string
	Description string
	MonitorID   uint   `gorm:"column:m_id"`
	ExternalID  string `gorm:"column:external_id"` // External ID from monitoring system (e.g., Zabbix groupid)
	Enabled     int    `gorm:"default:1"`          // 0 = disabled, 1 = enabled
	Status      int    // 0 = inactive, 1 = active, 2 = error, 3 = syncing
}

// Monitor represents a monitoring system (e.g., Zabbix)
type Monitor struct {
	gorm.Model
	Name              string
	URL               string
	Username          string
	Password          string
	AuthToken         string
	EventToken        string `gorm:"size:64;uniqueIndex"`
	Description       string
	Type              int    // 1 = zabbix, 2 = prometheus, 3 = other
	Enabled           int    `gorm:"default:1"` // 0 = disabled, 1 = enabled
	Status            int    // 0 = inactive, 1 = active, 2 = error, 3 = syncing
	StatusDescription string // Reason for error status (e.g., "connection timeout", "authentication failed")
}

// Alarm represents an external alert source (e.g., Zabbix)
type Alarm struct {
	gorm.Model
	Name              string
	URL               string
	Username          string
	Password          string
	AuthToken         string
	EventToken        string `gorm:"size:64;uniqueIndex"`
	Description       string
	Type              int    // 1 = zabbix, 2 = prometheus, 3 = other
	Enabled           int    `gorm:"default:1"` // 0 = disabled, 1 = enabled
	Status            int    // 0 = inactive, 1 = active, 2 = error, 3 = syncing
	StatusDescription string // Reason for error status (e.g., "connection timeout", "authentication failed")
}

// Item represents a monitoring item/metric
type Item struct {
	gorm.Model
	Name              string
	HID               uint   `gorm:"column:hid"`    // Internal host ID (foreign key to hosts table)
	ItemID            string `gorm:"column:itemid"` // External ID from monitoring system
	ExternalHostID    string `gorm:"column:hostid"` // External host ID from monitoring system
	ValueType         string
	LastValue         string
	Units             string
	Enabled           int    `gorm:"default:1"` // 0 = disabled, 1 = enabled
	Status            int    // 0 = inactive, 1 = active, 2 = error, 3 = syncing
	StatusDescription string // Reason for error status (e.g., "host is down", "pull failed")
	Comment           string
}

// ItemHistory tracks item metric values over time.
type ItemHistory struct {
	gorm.Model
	ItemID    uint `gorm:"index"`
	HostID    uint `gorm:"index"`
	Value     string
	Units     string
	Status    int
	SampledAt time.Time `gorm:"index"`
}

// HostHistory tracks host status over time.
type HostHistory struct {
	gorm.Model
	HostID            uint `gorm:"index"`
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
	Message  string
	Severity int
	Status   int  // 0 = active, 1 = acknowledged, 2 = resolved
	AlarmID  uint `gorm:"column:alarm_id"`
	HostID   uint
	ItemID   uint
	Comment  string
}

// Media represents a notification delivery target
type Media struct {
	gorm.Model
	Name        string
	Type        string            // "email", "webhook", "sms", etc. (cached from media type)
	MediaTypeID uint              `gorm:"column:media_type_id"`
	Target      string            // address/endpoint/number
	Params      map[string]string `gorm:"type:json;serializer:json"`
	Enabled     int               `gorm:"default:1"` // 0 = disabled, 1 = enabled
	Status      int               // 0 = inactive, 1 = active, 2 = error, 3 = syncing
	Description string
}

// MediaParamField defines a parameter expected by a media type template
type MediaParamField struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Required bool   `json:"required"`
	Default  string `json:"default"`
	Pattern  string `json:"pattern"`
}

// MediaType represents a supported media delivery type
type MediaType struct {
	gorm.Model
	Name        string
	Key         string // "email", "webhook", "sms", etc.
	Enabled     int    `gorm:"default:1"` // 0 = disabled, 1 = enabled
	Status      int    // 0 = inactive, 1 = active, 2 = error, 3 = syncing
	Description string
	Template    string            `gorm:"type:text"`
	Fields      []MediaParamField `gorm:"type:json;serializer:json"`
}

// Action represents an action executed for alerts
type Action struct {
	gorm.Model
	Name        string
	MediaID     uint
	Template    string // message template
	Enabled     int    `gorm:"default:1"` // 0 = disabled, 1 = enabled
	Status      int    // 0 = inactive, 1 = active, 2 = error, 3 = syncing
	Description string
}

// Trigger represents a rule that filters alerts or logs to invoke an action
type Trigger struct {
	gorm.Model
	Name           string
	Entity         string // "alert" or "log"
	SeverityMin    int
	ActionID       uint
	AlertID        *uint `gorm:"column:alert_id"`
	AlertStatus    *int  `gorm:"column:alert_status"`
	AlertGroupID   *uint `gorm:"column:alert_group_id"`
	AlertMonitorID *uint `gorm:"column:alert_monitor_id"`
	AlertHostID    *uint `gorm:"column:alert_host_id"`
	AlertItemID    *uint `gorm:"column:alert_item_id"`
	AlertQuery     string
	LogType        string `gorm:"column:log_type"`
	LogSeverity    *int   `gorm:"column:log_level"`
	LogQuery       string `gorm:"column:log_query"`
	Enabled        int    `gorm:"default:1"` // 0 = disabled, 1 = enabled
	Status         int    // 0 = inactive, 1 = active, 2 = error, 3 = syncing
}

// Provider represents an AI provider (e.g., Google Gemini)
type Provider struct {
	gorm.Model
	Name         string
	URL          string
	APIKey       string
	DefaultModel string
	Models       []string `gorm:"type:json;serializer:json"` // List of available models
	Type         int      // Provider type: 1 = Gemini, 2 = OpenAI, 3 = Ollama
	Description  string
	Enabled      int `gorm:"default:1"` // 0 = disabled, 1 = enabled
	Status       int // 0 = inactive, 1 = active, 2 = error, 3 = syncing
}

// Chat represents a chat message
type Chat struct {
	gorm.Model
	UserID     uint
	ProviderID uint
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
	Severity int    `gorm:"column:level"` // 0=info, 1=warn, 2=error
	Message  string
	Context  string `gorm:"type:text"`
	UserID   *uint
	IP       string
}

// User represents the authentication and authorization information
type User struct {
	gorm.Model
	Username   string
	Password   string // Hashed password
	Privileges int    // 0 = unauthorized, 1 = user, 2 = admin, 3 = superadmin
	Status     int    // 0 = inactive, 1 = active
}

// UserInformation represents the user profile and personal information
type UserInformation struct {
	gorm.Model
	UserID       uint `gorm:"column:user_id"` // Foreign key to User
	Email        string
	Phone        string
	Avatar       string
	Address      string
	Introduction string
	Nickname     string
}

// RegisterApplication represents a pending registration request from an unregistered user
type RegisterApplication struct {
	gorm.Model
	Username   string
	Password   string
	Status     int    // 0 = pending, 1 = approved, 2 = rejected
	Reason     string // rejection or approval note
	ApprovedBy *uint  `gorm:"column:approved_by"`
}

// UserWithInfo combines User and UserInformation for convenient querying
type UserWithInfo struct {
	User
	UserInformation UserInformation `gorm:"foreignKey:UserID"`
}
