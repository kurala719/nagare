// Package monitors provides a unified interface for interacting with various monitoring systems.
package monitors

import (
	"context"
	"errors"
	"fmt"
)

// MonitorType represents the type of monitoring system
type MonitorType int

const (
	MonitorZabbix     MonitorType = iota + 1 // 1 = zabbix
	MonitorPrometheus                        // 2 = prometheus
	MonitorOther                             // 3 = other
)

// String returns the string representation of the monitor type
func (m MonitorType) String() string {
	switch m {
	case MonitorZabbix:
		return "zabbix"
	case MonitorPrometheus:
		return "prometheus"
	case MonitorOther:
		return "other"
	default:
		return "unknown"
	}
}

// ParseMonitorType parses an integer to MonitorType
func ParseMonitorType(s int) MonitorType {
	switch s {
	case 1:
		return MonitorZabbix
	case 2:
		return MonitorPrometheus
	case 3:
		return MonitorOther
	default:
		return MonitorZabbix
	}
}

// Host represents a monitored host
type Host struct {
	ID          string
	Name        string
	Description string
	Status      string // "up", "down", "unknown"
	IPAddress   string
	Metadata    map[string]string
}

// Item represents a monitoring item/metric
type Item struct {
	ID          string
	HostID      string
	Name        string
	Key         string
	Type        string
	Value       string
	Units       string
	ValueType   string
	Delay       string
	Status      string
	Timestamp   int64
	Description string
	InterfaceID string
	Metadata    map[string]string
}

// Alert represents an alert/problem
type Alert struct {
	ID          string
	HostID      string
	Name        string
	Severity    string // "disaster", "high", "average", "warning", "information"
	Status      string // "problem", "resolved"
	Description string
	Timestamp   int64
}

// Trigger represents an alert trigger/rule
type Trigger struct {
	ID          string
	Name        string
	Expression  string
	Priority    string
	Status      string
	Description string
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	URL      string
	Username string
	Password string
	Token    string // API token (alternative to username/password)
}

// Config holds the configuration for a monitoring provider
type Config struct {
	Name    string      // Provider name/identifier
	Type    MonitorType // Provider type
	Auth    AuthConfig  // Authentication configuration
	Timeout int         // Request timeout in seconds
}

// Provider defines the interface for monitoring providers
type Provider interface {
	// Authenticate authenticates with the monitoring system
	Authenticate(ctx context.Context) error
	// GetAuthToken returns the current auth token
	GetAuthToken() string
	// SetAuthToken sets the auth token (for pre-authenticated sessions)
	SetAuthToken(token string)

	// Host operations
	GetHosts(ctx context.Context) ([]Host, error)
	GetHostsByGroupID(ctx context.Context, groupID string) ([]Host, error)
	GetHostByName(ctx context.Context, name string) (*Host, error)
	GetHostByID(ctx context.Context, hostID string) (*Host, error)
	CreateHost(ctx context.Context, host Host) (Host, error)
	UpdateHost(ctx context.Context, host Host) (Host, error)
	DeleteHost(ctx context.Context, hostID string) error

	// Item/Metric operations
	GetItems(ctx context.Context, hostID string) ([]Item, error)
	GetItemByID(ctx context.Context, itemID string) (*Item, error)
	GetItemHistory(ctx context.Context, itemID string, from, to int64) ([]Item, error)
	CreateItem(ctx context.Context, item Item) (Item, error)
	UpdateItem(ctx context.Context, item Item) (Item, error)
	DeleteItem(ctx context.Context, itemID string) error

	// Alert operations
	GetAlerts(ctx context.Context) ([]Alert, error)
	GetAlertsByHost(ctx context.Context, hostID string) ([]Alert, error)

	// Trigger operations
	GetTriggers(ctx context.Context) ([]Trigger, error)
	GetTriggersByHost(ctx context.Context, hostID string) ([]Trigger, error)

	// Template operations
	GetTemplateidByName(ctx context.Context, name string) ([]string, error)

	// Host group operations could be added here if needed
	GetHostGroups(ctx context.Context) ([]string, error)
	GetHostGroupsDetails(ctx context.Context) ([]struct{ ID, Name string }, error)
	GetHostGroupByName(ctx context.Context, name string) (string, error)
	CreateHostGroup(ctx context.Context, name string) (string, error)
	UpdateHostGroup(ctx context.Context, id, name string) error
	DeleteHostGroup(ctx context.Context, id string) error

	// Provider info
	Name() string
	Type() MonitorType
}

// Client is the main monitoring client that wraps different providers
type Client struct {
	provider Provider
	config   Config
}

// NewClient creates a new monitoring client based on the provider type
func NewClient(cfg Config) (*Client, error) {
	var provider Provider
	var err error

	switch cfg.Type {
	case MonitorZabbix:
		provider, err = NewZabbixProvider(cfg)
	case MonitorPrometheus:
		provider, err = NewPrometheusProvider(cfg)
	default:
		return nil, errors.New("unsupported monitor type")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return &Client{
		provider: provider,
		config:   cfg,
	}, nil
}

// Authenticate authenticates with the monitoring system
func (c *Client) Authenticate(ctx context.Context) error {
	return c.provider.Authenticate(ctx)
}

// GetAuthToken returns the current auth token
func (c *Client) GetAuthToken() string {
	return c.provider.GetAuthToken()
}

// SetAuthToken sets the auth token
func (c *Client) SetAuthToken(token string) {
	c.provider.SetAuthToken(token)
}

// GetHosts retrieves all hosts from the monitoring system
func (c *Client) GetHosts(ctx context.Context) ([]Host, error) {
	return c.provider.GetHosts(ctx)
}

// GetHostsByGroupID retrieves all hosts for a specific host group
func (c *Client) GetHostsByGroupID(ctx context.Context, groupID string) ([]Host, error) {
	return c.provider.GetHostsByGroupID(ctx, groupID)
}

// GetHostByName retrieves a specific host by name
func (c *Client) GetHostByName(ctx context.Context, name string) (*Host, error) {
	return c.provider.GetHostByName(ctx, name)
}

// GetHostByID retrieves a specific host by ID
func (c *Client) GetHostByID(ctx context.Context, hostID string) (*Host, error) {
	return c.provider.GetHostByID(ctx, hostID)
}

// CreateHost creates a host in the monitoring system
func (c *Client) CreateHost(ctx context.Context, host Host) (Host, error) {
	return c.provider.CreateHost(ctx, host)
}

// UpdateHost updates a host in the monitoring system
func (c *Client) UpdateHost(ctx context.Context, host Host) (Host, error) {
	return c.provider.UpdateHost(ctx, host)
}

// DeleteHost deletes a host in the monitoring system
func (c *Client) DeleteHost(ctx context.Context, hostID string) error {
	return c.provider.DeleteHost(ctx, hostID)
}

// GetItems retrieves all items for a host
func (c *Client) GetItems(ctx context.Context, hostID string) ([]Item, error) {
	return c.provider.GetItems(ctx, hostID)
}

// GetItemByID retrieves a specific item by ID
func (c *Client) GetItemByID(ctx context.Context, itemID string) (*Item, error) {
	return c.provider.GetItemByID(ctx, itemID)
}

// GetItemHistory retrieves historical data for an item
func (c *Client) GetItemHistory(ctx context.Context, itemID string, from, to int64) ([]Item, error) {
	return c.provider.GetItemHistory(ctx, itemID, from, to)
}

// CreateItem creates an item in the monitoring system
func (c *Client) CreateItem(ctx context.Context, item Item) (Item, error) {
	return c.provider.CreateItem(ctx, item)
}

// UpdateItem updates an item in the monitoring system
func (c *Client) UpdateItem(ctx context.Context, item Item) (Item, error) {
	return c.provider.UpdateItem(ctx, item)
}

// DeleteItem deletes an item in the monitoring system
func (c *Client) DeleteItem(ctx context.Context, itemID string) error {
	return c.provider.DeleteItem(ctx, itemID)
}

// GetAlerts retrieves all active alerts
func (c *Client) GetAlerts(ctx context.Context) ([]Alert, error) {
	return c.provider.GetAlerts(ctx)
}

// GetAlertsByHost retrieves alerts for a specific host
func (c *Client) GetAlertsByHost(ctx context.Context, hostID string) ([]Alert, error) {
	return c.provider.GetAlertsByHost(ctx, hostID)
}

// GetTriggers retrieves all triggers
func (c *Client) GetTriggers(ctx context.Context) ([]Trigger, error) {
	return c.provider.GetTriggers(ctx)
}

// GetTriggersByHost retrieves triggers for a specific host
func (c *Client) GetTriggersByHost(ctx context.Context, hostID string) ([]Trigger, error) {
	return c.provider.GetTriggersByHost(ctx, hostID)
}

// ProviderName returns the name of the current provider
func (c *Client) ProviderName() string {
	return c.provider.Name()
}

// ProviderType returns the type of the current provider
func (c *Client) ProviderType() MonitorType {
	return c.provider.Type()
}

// GetTemplateidByName retrieves template IDs by name from the monitoring system
func (c *Client) GetTemplateidByName(ctx context.Context, name string) ([]string, error) {
	return c.provider.GetTemplateidByName(ctx, name)
}

// GetHostGroups retrieves all host groups from the monitoring system
func (c *Client) GetHostGroups(ctx context.Context) ([]string, error) {
	return c.provider.GetHostGroups(ctx)
}

// GetHostGroupsDetails retrieves all host groups with details from the monitoring system
func (c *Client) GetHostGroupsDetails(ctx context.Context) ([]struct{ ID, Name string }, error) {
	return c.provider.GetHostGroupsDetails(ctx)
}

// GetHostGroupByName retrieves a host group ID by name
func (c *Client) GetHostGroupByName(ctx context.Context, name string) (string, error) {
	return c.provider.GetHostGroupByName(ctx, name)
}

// CreateHostGroup creates a host group in the monitoring system
func (c *Client) CreateHostGroup(ctx context.Context, name string) (string, error) {
	return c.provider.CreateHostGroup(ctx, name)
}

// UpdateHostGroup updates a host group in the monitoring system
func (c *Client) UpdateHostGroup(ctx context.Context, id, name string) error {
	return c.provider.UpdateHostGroup(ctx, id, name)
}

// DeleteHostGroup deletes a host group in the monitoring system
func (c *Client) DeleteHostGroup(ctx context.Context, id string) error {
	return c.provider.DeleteHostGroup(ctx, id)
}
