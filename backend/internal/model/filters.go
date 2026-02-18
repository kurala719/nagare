package model

// AlertFilter represents search and filter options for alerts
// Query matches alert message (LIKE)
type AlertFilter struct {
	Query     string
	Severity  *int
	Status    *int
	AlarmID   *int
	HostID    *int
	ItemID    *int
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// MediaFilter represents search and filter options for media
// Query matches name/type/target/description (LIKE)
type MediaFilter struct {
	Query     string
	Status    *int
	TypeID    *uint
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// MediaTypeFilter represents search and filter options for media types
// Query matches name/key/description (LIKE)
type MediaTypeFilter struct {
	Query     string
	Status    *int
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// ActionFilter represents search and filter options for actions
// Query matches name/description/template (LIKE)
type ActionFilter struct {
	Query     string
	Status    *int
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// TriggerFilter represents search and filter options for triggers
// Query matches name (LIKE)
type TriggerFilter struct {
	Query          string
	Status         *int
	SeverityMin    *int
	Entity         *string
	ActionID       *uint
	AlertID        *uint
	AlertMonitorID *uint
	AlertGroupID   *uint
	AlertHostID    *uint
	AlertItemID    *uint
	Limit          int
	Offset         int
	SortBy         string
	SortOrder      string
}

// LogFilter represents search and filter options for logs
// Query matches message/context (LIKE)
type LogFilter struct {
	Type      string
	Severity  *int
	Query     string
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// MonitorFilter represents search and filter options for monitors
// Query matches name/url/description (LIKE)
type MonitorFilter struct {
	Query     string
	Type      *string
	Status    *int
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// AlarmFilter represents search and filter options for alarms
// Query matches name/url/description (LIKE)
type AlarmFilter struct {
	Query     string
	Type      *int
	Status    *int
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// HostFilter represents search and filter options for hosts
// Query matches name/hostid/ip_addr/description (LIKE)
type HostFilter struct {
	Query     string
	MID       *uint
	GroupID   *uint
	Status    *int
	IPAddr    *string
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// GroupFilter represents search and filter options for groups
// Query matches name/description (LIKE)
type GroupFilter struct {
	Query     string
	Status    *int
	MonitorID *uint
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// ItemFilter represents search and filter options for items
// Query matches name/itemid/hostid (LIKE)
type ItemFilter struct {
	Query     string
	HID       *uint
	ValueType *string
	Status    *int
	HostID    *string
	ItemID    *string
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// ProviderFilter represents search and filter options for providers
// Query matches name/url/description/default_model (LIKE)
type ProviderFilter struct {
	Query     string
	Type      *int
	Status    *int
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// UserFilter represents search and filter options for users
// Query matches username (LIKE)
type UserFilter struct {
	Query               string
	Privileges          *int
	Status              *int
	Limit               int
	Offset              int
	SortBy              string
	SortOrder           string
	RequesterPrivileges int
}

// RegisterApplicationFilter represents search and filter options for registration applications
// Query matches username (LIKE)
type RegisterApplicationFilter struct {
	Query     string
	Status    *int
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// ChatFilter represents search and filter options for chats
// Query matches content (LIKE)
type ChatFilter struct {
	Query      string
	Role       *string
	ProviderID *int
	UserID     *int
	Model      *string
	Limit      int
	Offset     int
	SortBy     string
	SortOrder  string
}
