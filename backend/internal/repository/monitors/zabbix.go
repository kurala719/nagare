package monitors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"
)

// ZabbixProvider implements the Provider interface for Zabbix
type ZabbixProvider struct {
	url       string
	username  string
	password  string
	authToken string
	client    *http.Client
	reqID     int
}

// Zabbix API request/response structures
type zabbixRequest struct {
	Jsonrpc string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
	Auth    string                 `json:"auth,omitempty"`
	ID      int                    `json:"id"`
}

type zabbixResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *zabbixError    `json:"error,omitempty"`
	ID      int             `json:"id"`
}

type zabbixError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

type zabbixHost struct {
	HostID        string      `json:"hostid"`
	Host          string      `json:"host"`
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	Status        interface{} `json:"status"`
	Available     interface{} `json:"available"`
	Error         string      `json:"error"`
	SnmpAvailable interface{} `json:"snmp_available"`
	SnmpError     string      `json:"snmp_error"`
	IpmiAvailable interface{} `json:"ipmi_available"`
	IpmiError     string      `json:"ipmi_error"`
	JmxAvailable  interface{} `json:"jmx_available"`
	JmxError      string      `json:"jmx_error"`
	Interfaces    []struct {
		InterfaceID interface{} `json:"interfaceid"`
		IP          string      `json:"ip"`
		DNS         string      `json:"dns"`
		Port        interface{} `json:"port"`
		Type        interface{} `json:"type"`
		Main        interface{} `json:"main"`
		UseIP       interface{} `json:"useip"`
		Available   interface{} `json:"available"`
		Error       string      `json:"error"`
	} `json:"interfaces"`
}

type zabbixHostWithGroups struct {
	zabbixHost
	Groups []struct {
		GroupID string `json:"groupid"`
		Name    string `json:"name"`
	} `json:"groups"`
	HostGroups []struct {
		GroupID string `json:"groupid"`
		Name    string `json:"name"`
	} `json:"hostgroups"`
}

type zabbixItem struct {
	ItemID    string      `json:"itemid"`
	HostID    string      `json:"hostid"`
	Name      string      `json:"name"`
	Key       string      `json:"key_"`
	LastValue interface{} `json:"lastvalue"`
	Units     string      `json:"units"`
	ValueType interface{} `json:"value_type"`
	Type      interface{} `json:"type"`
	Delay     string      `json:"delay"`
	Desc      string      `json:"description"`
	Status    interface{} `json:"status"`
	LastClock string      `json:"lastclock"`
}

type zabbixProblem struct {
	EventID      string      `json:"eventid"`
	ObjectID     string      `json:"objectid"`
	Name         string      `json:"name"`
	Severity     interface{} `json:"severity"`
	Acknowledged interface{} `json:"acknowledged"`
	Clock        string      `json:"clock"`
}

type zabbixTrigger struct {
	TriggerID   string      `json:"triggerid"`
	Description string      `json:"description"`
	Expression  string      `json:"expression"`
	Priority    interface{} `json:"priority"`
	Status      interface{} `json:"status"`
	Value       interface{} `json:"value"`
}

type ZabbixWebhookSetupConfig struct {
	WebhookURL       string
	EventToken       string
	ActionName       string
	UserLookup       string
	ZabbixUserID     string // Optional: if set, explicitly use this user ID instead of looking up by username
	MediaTypeName    string
	UserMediaSendTo  string
	ActionEscalation string
}

type ZabbixWebhookSetupResult struct {
	WebhookURL  string
	MediaTypeID string
	ActionID    string
	ActionName  string
	UserID      string
	Username    string
}

// NewZabbixProvider creates a new Zabbix provider
func NewZabbixProvider(cfg Config) (*ZabbixProvider, error) {
	if cfg.Auth.URL == "" {
		return nil, fmt.Errorf("URL is required for Zabbix provider")
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30
	}

	url := cfg.Auth.URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}
	url = strings.TrimSuffix(url, "/")

	// Create HTTP client with cookie jar for session management (Zabbix 6.0+)
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	return &ZabbixProvider{
		url:       url,
		username:  cfg.Auth.Username,
		password:  cfg.Auth.Password,
		authToken: cfg.Auth.Token,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
			Jar:     jar,
		},
	}, nil
}

// sendRequest sends a request to the Zabbix API
func (p *ZabbixProvider) sendRequest(ctx context.Context, method string, params map[string]interface{}) (*zabbixResponse, error) {
	p.reqID++

	req := zabbixRequest{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		ID:      p.reqID,
	}

	// Try with auth at root level first (Zabbix < 6.0)
	if p.authToken != "" && method != "user.login" {
		req.Auth = p.authToken
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.url+"/api_jsonrpc.php", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var zabbixResp zabbixResponse
	if err := json.Unmarshal(respBody, &zabbixResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if zabbixResp.Error != nil {
		// If we get "unexpected parameter auth" or "Not authorized" error, retry with Bearer token (Zabbix 6.0+)
		if (strings.Contains(zabbixResp.Error.Data, "unexpected parameter \"auth\"") ||
			strings.Contains(zabbixResp.Error.Data, "Not authorized")) &&
			p.authToken != "" && method != "user.login" {
			// Retry with Bearer token in Authorization header instead
			p.reqID++
			req2 := zabbixRequest{
				Jsonrpc: "2.0",
				Method:  method,
				Params:  params,
				ID:      p.reqID,
			}
			// Don't include Auth at root level this time

			body2, err := json.Marshal(req2)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal retry request: %w", err)
			}

			httpReq2, err := http.NewRequestWithContext(ctx, "POST", p.url+"/api_jsonrpc.php", bytes.NewReader(body2))
			if err != nil {
				return nil, fmt.Errorf("failed to create retry request: %w", err)
			}

			httpReq2.Header.Set("Content-Type", "application/json")
			// Use Bearer token in Authorization header for Zabbix 6.0+
			httpReq2.Header.Set("Authorization", "Bearer "+p.authToken)

			resp2, err := p.client.Do(httpReq2)
			if err != nil {
				return nil, fmt.Errorf("failed to send retry request: %w", err)
			}
			defer resp2.Body.Close()

			respBody2, err := io.ReadAll(resp2.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read retry response: %w", err)
			}

			var zabbixResp2 zabbixResponse
			if err := json.Unmarshal(respBody2, &zabbixResp2); err != nil {
				return nil, fmt.Errorf("failed to unmarshal retry response: %w", err)
			}

			if zabbixResp2.Error != nil {
				return nil, fmt.Errorf("Zabbix API error: %s - %s", zabbixResp2.Error.Message, zabbixResp2.Error.Data)
			}

			return &zabbixResp2, nil
		}

		return nil, fmt.Errorf("Zabbix API error: %s - %s", zabbixResp.Error.Message, zabbixResp.Error.Data)
	}

	return &zabbixResp, nil
}

// Authenticate implements the Provider interface
func (p *ZabbixProvider) Authenticate(ctx context.Context) error {
	// If we already have a token, try to verify it first to avoid unnecessary logins
	if p.authToken != "" {
		_, err := p.sendRequest(ctx, "user.get", map[string]interface{}{
			"output": []string{"userid"},
			"limit":  1,
		})
		if err == nil {
			// Token is still valid, no need to re-authenticate
			return nil
		}

		// If token verification fails and we have no credentials, we can't do anything
		if p.username == "" || p.password == "" {
			return fmt.Errorf("provided token is invalid and no credentials supplied for login")
		}
	}

	// Otherwise, proceed with normal user.login
	params := map[string]interface{}{
		"username": p.username,
		"password": p.password,
	}

	resp, err := p.sendRequest(ctx, "user.login", params)
	if err != nil {
		if strings.Contains(err.Error(), "unexpected parameter \"username\"") {
			params = map[string]interface{}{
				"user":     p.username,
				"password": p.password,
			}
			resp, err = p.sendRequest(ctx, "user.login", params)
		} else if strings.Contains(err.Error(), "unexpected parameter \"user\"") {
			params = map[string]interface{}{
				"username": p.username,
				"password": p.password,
			}
			resp, err = p.sendRequest(ctx, "user.login", params)
		}
	}
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	var token string
	if err := json.Unmarshal(resp.Result, &token); err != nil {
		// Some Zabbix versions might return the token wrapped in a different way or
		// if it's already an API token used as password, we might need different handling.
		// For now, try to catch empty result.
		if string(resp.Result) == "null" || string(resp.Result) == "" {
			return fmt.Errorf("authentication succeeded but Zabbix returned an empty token")
		}
		return fmt.Errorf("failed to parse auth token: %w (result: %s)", err, string(resp.Result))
	}

	if token == "" {
		return fmt.Errorf("authentication succeeded but received empty token string")
	}

	p.authToken = token
	return nil
}

// GetAuthToken returns the current auth token
func (p *ZabbixProvider) GetAuthToken() string {
	return p.authToken
}

// SetAuthToken sets the auth token
func (p *ZabbixProvider) SetAuthToken(token string) {
	p.authToken = token
}

// GetHosts implements the Provider interface
func (p *ZabbixProvider) GetHosts(ctx context.Context) ([]Host, error) {
	params := map[string]interface{}{
		"output": []string{
			"hostid", "host", "name", "description", "status",
			"available", "error", "snmp_available", "snmp_error",
			"ipmi_available", "ipmi_error", "jmx_available", "jmx_error",
		},
		"selectInterfaces": "extend",
		"selectGroups":     "extend",
		"selectHostGroups": "extend",
	}

	resp, err := p.sendRequest(ctx, "host.get", params)
	if err != nil {
		return nil, err
	}

	return p.parseZabbixHosts(resp.Result)
}

// GetHostsByGroupID implements the Provider interface
func (p *ZabbixProvider) GetHostsByGroupID(ctx context.Context, groupID string) ([]Host, error) {
	params := map[string]interface{}{
		"output": []string{
			"hostid", "host", "name", "description", "status",
			"available", "error", "snmp_available", "snmp_error",
			"ipmi_available", "ipmi_error", "jmx_available", "jmx_error",
		},
		"selectInterfaces": "extend",
		"selectGroups":     "extend",
		"selectHostGroups": "extend",
		"groupids":         groupID,
	}

	resp, err := p.sendRequest(ctx, "host.get", params)
	if err != nil {
		return nil, err
	}

	return p.parseZabbixHosts(resp.Result)
}

func determineZabbixAvailability(zh zabbixHostWithGroups) (string, string, string) {
	status := toString(zh.Status)
	available := toString(zh.Available)
	snmpAvailable := toString(zh.SnmpAvailable)
	ipmiAvailable := toString(zh.IpmiAvailable)
	jmxAvailable := toString(zh.JmxAvailable)

	// status == "1" means Disabled in Zabbix
	if status == "1" {
		return "unknown", "0", "" // Mapping to Inactive (0)
	}

	// Priority 1: Check for any explicit error messages at host level
	if zh.Error != "" {
		return "down", "2", zh.Error
	}
	if zh.SnmpError != "" {
		return "down", "2", zh.SnmpError
	}
	if zh.IpmiError != "" {
		return "down", "2", zh.IpmiError
	}
	if zh.JmxError != "" {
		return "down", "2", zh.JmxError
	}

	// Priority 2: Deep check - Check interface-level availability and errors
	// Zabbix 6.0+ often keeps the error details here
	for _, iface := range zh.Interfaces {
		ifaceAvailable := toString(iface.Available)
		if ifaceAvailable == "2" || iface.Error != "" {
			return "down", "2", iface.Error
		}
	}

	// Priority 3: Check for explicit "Not Available" (2) numeric status at host level
	if available == "2" || snmpAvailable == "2" || ipmiAvailable == "2" || jmxAvailable == "2" {
		return "down", "2", ""
	}

	// Priority 4: Check for any explicit "Available" (1) status (Host or Interface)
	if available == "1" || snmpAvailable == "1" || ipmiAvailable == "1" || jmxAvailable == "1" {
		return "up", "1", ""
	}

	for _, iface := range zh.Interfaces {
		if toString(iface.Available) == "1" {
			return "up", "1", ""
		}
	}

	// Default: Enabled but Zabbix hasn't determined availability (0).
	// We return "0" (unknown) to let Nagare's sync process try to poll it.
	return "unknown", "0", ""
}

func zabbixInterfaceTypeLabel(raw string) string {
	switch raw {
	case "1":
		return "agent"
	case "2":
		return "snmp"
	case "3":
		return "ipmi"
	case "4":
		return "jmx"
	default:
		return "unknown"
	}
}

func mapZabbixIfTypeLabel(raw string) string {
	switch raw {
	case "1":
		return "other"
	case "6":
		return "ethernetCsmacd"
	case "24":
		return "softwareLoopback"
	case "53":
		return "propVirtual"
	case "54":
		return "propMultiplexor"
	case "117":
		return "gigabitEthernet"
	case "131":
		return "tunnel"
	case "135":
		return "l2vlan"
	case "136":
		return "l3ipvlan"
	case "161":
		return "ieee8023adLag"
	default:
		return raw
	}
}

func mapZabbixFanStatus(raw string) string {
	switch raw {
	case "1":
		return "Normal"
	case "2":
		return "Abnormal"
	case "3":
		return "Not Supported"
	case "4":
		return "Unknown"
	default:
		return raw
	}
}

func normalizeZabbixItemValue(name, key, value, units string) (string, string) {
	nameLower := strings.ToLower(strings.TrimSpace(name))
	keyLower := strings.ToLower(strings.TrimSpace(key))
	unitsLower := strings.ToLower(strings.TrimSpace(units))
	value = strings.TrimSpace(value)

	if strings.Contains(nameLower, "fan") && strings.Contains(nameLower, "status") {
		return mapZabbixFanStatus(value), units
	}
	if strings.Contains(nameLower, "interface type") || strings.Contains(keyLower, "iftype") {
		return mapZabbixIfTypeLabel(value), units
	}

	if (unitsLower == "bps" || unitsLower == "b/s" || unitsLower == "bit/s") &&
		(strings.Contains(nameLower, "speed") || strings.Contains(keyLower, "ifspeed") || strings.Contains(keyLower, "ifhighspeed")) {
		if numeric, err := strconv.ParseFloat(value, 64); err == nil {
			scaledValue, scaledUnits := scaleBitsPerSecond(numeric)
			return scaledValue, scaledUnits
		}
	}

	return value, units
}

func scaleBitsPerSecond(value float64) (string, string) {
	switch {
	case value >= 1e9:
		return fmt.Sprintf("%.2f", value/1e9), "Gbps"
	case value >= 1e6:
		return fmt.Sprintf("%.2f", value/1e6), "Mbps"
	case value >= 1e3:
		return fmt.Sprintf("%.2f", value/1e3), "Kbps"
	default:
		return fmt.Sprintf("%.0f", value), "bps"
	}
}

func zabbixPrimaryInterfaceInfo(host zabbixHost) (string, string, string) {
	if len(host.Interfaces) == 0 {
		return "", "", ""
	}
	selected := host.Interfaces[0]
	for _, iface := range host.Interfaces {
		if toString(iface.Main) == "1" {
			selected = iface
			break
		}
	}
	interfaceTypeID := toString(selected.Type)
	return selected.IP, interfaceTypeID, zabbixInterfaceTypeLabel(interfaceTypeID)
}

// parseZabbixHosts parses Zabbix host.get result into common Host slice
func (p *ZabbixProvider) parseZabbixHosts(result json.RawMessage) ([]Host, error) {
	var zabbixHosts []zabbixHostWithGroups
	if err := json.Unmarshal(result, &zabbixHosts); err != nil {
		return nil, fmt.Errorf("failed to parse hosts: %w", err)
	}

	hosts := make([]Host, 0, len(zabbixHosts))
	for _, zh := range zabbixHosts {
		ip, interfaceTypeID, interfaceTypeLabel := zabbixPrimaryInterfaceInfo(zh.zabbixHost)

		status, activeAvailable, statusDesc := determineZabbixAvailability(zh)

		metadata := map[string]string{
			"host":               zh.Host,
			"active_available":   activeAvailable,
			"status_description": statusDesc,
		}
		if interfaceTypeID != "" {
			metadata["interface_type_id"] = interfaceTypeID
			metadata["interface_type"] = interfaceTypeLabel
		}
		selectedGroups := zh.HostGroups
		if len(selectedGroups) == 0 {
			selectedGroups = zh.Groups
		}

		// Add groupid to metadata if available
		if len(selectedGroups) > 0 {
			metadata["groupid"] = selectedGroups[0].GroupID
			metadata["groupname"] = selectedGroups[0].Name
			var gids []string
			var gnames []string
			for _, g := range selectedGroups {
				gids = append(gids, g.GroupID)
				gnames = append(gnames, g.Name)
			}
			metadata["groupids"] = strings.Join(gids, ",")
			metadata["groupnames"] = strings.Join(gnames, ",")
		}

		// Zabbix status: 0 = monitored, 1 = unmonitored
		enabled := 1
		if toString(zh.Status) == "1" {
			enabled = 0
		}

		hosts = append(hosts, Host{
			ID:          zh.HostID,
			Name:        zh.Name,
			Description: zh.Description,
			Status:      status,
			Enabled:     enabled,
			IPAddress:   ip,
			Metadata:    metadata,
		})
	}

	return hosts, nil
}

// GetHostByName implements the Provider interface
func (p *ZabbixProvider) GetHostByName(ctx context.Context, name string) (*Host, error) {
	params := map[string]interface{}{
		"output":           "extend",
		"selectInterfaces": "extend",
		"selectGroups":     "extend",
		"filter": map[string]interface{}{
			"host": []string{name},
		},
	}

	resp, err := p.sendRequest(ctx, "host.get", params)
	if err != nil {
		return nil, err
	}

	var zabbixHosts []zabbixHostWithGroups
	if err := json.Unmarshal(resp.Result, &zabbixHosts); err != nil {
		return nil, fmt.Errorf("failed to parse host: %w", err)
	}

	if len(zabbixHosts) == 0 {
		return nil, fmt.Errorf("host not found: %s", name)
	}

	zh := zabbixHosts[0]
	ip, interfaceTypeID, interfaceTypeLabel := zabbixPrimaryInterfaceInfo(zh.zabbixHost)

	status, activeAvailable, statusDesc := determineZabbixAvailability(zh)

	metadata := map[string]string{
		"host":               zh.Host,
		"active_available":   activeAvailable,
		"status_description": statusDesc,
	}
	if interfaceTypeID != "" {
		metadata["interface_type_id"] = interfaceTypeID
		metadata["interface_type"] = interfaceTypeLabel
	}
	selectedGroups := zh.HostGroups
	if len(selectedGroups) == 0 {
		selectedGroups = zh.Groups
	}

	if len(selectedGroups) > 0 {
		metadata["groupid"] = selectedGroups[0].GroupID
		metadata["groupname"] = selectedGroups[0].Name
		var gids []string
		var gnames []string
		for _, g := range selectedGroups {
			gids = append(gids, g.GroupID)
			gnames = append(gnames, g.Name)
		}
		metadata["groupids"] = strings.Join(gids, ",")
		metadata["groupnames"] = strings.Join(gnames, ",")
	}

	enabled := 1
	if toString(zh.Status) == "1" {
		enabled = 0
	}

	return &Host{
		ID:          zh.HostID,
		Name:        zh.Name,
		Description: zh.Description,
		Status:      status,
		Enabled:     enabled,
		IPAddress:   ip,
		Metadata:    metadata,
	}, nil
}

// GetHostByID implements the Provider interface
func (p *ZabbixProvider) GetHostByID(ctx context.Context, hostID string) (*Host, error) {
	params := map[string]interface{}{
		"output":           []string{"hostid", "host", "name", "description", "status", "available", "error", "snmp_available", "snmp_error", "ipmi_available", "ipmi_error", "jmx_available", "jmx_error"},
		"selectInterfaces": []string{"interfaceid", "ip", "dns", "port", "type", "main", "useip"},
		"selectHostGroups": "extend",
		"selectGroups":     "extend",
		"hostids":          hostID,
	}

	resp, err := p.sendRequest(ctx, "host.get", params)
	if err != nil {
		return nil, err
	}

	var zabbixHosts []zabbixHostWithGroups
	if err := json.Unmarshal(resp.Result, &zabbixHosts); err != nil {
		return nil, fmt.Errorf("failed to parse host: %w", err)
	}

	if len(zabbixHosts) == 0 {
		return nil, fmt.Errorf("host not found: %s", hostID)
	}

	zh := zabbixHosts[0]
	ip, interfaceTypeID, interfaceTypeLabel := zabbixPrimaryInterfaceInfo(zh.zabbixHost)

	status, activeAvailable, statusDesc := determineZabbixAvailability(zh)

	metadata := map[string]string{
		"host":               zh.Host,
		"active_available":   activeAvailable,
		"status_description": statusDesc,
	}
	if interfaceTypeID != "" {
		metadata["interface_type_id"] = interfaceTypeID
		metadata["interface_type"] = interfaceTypeLabel
	}
	selectedGroups := zh.HostGroups
	if len(selectedGroups) == 0 {
		selectedGroups = zh.Groups
	}

	if len(selectedGroups) > 0 {
		metadata["groupid"] = selectedGroups[0].GroupID
		metadata["groupname"] = selectedGroups[0].Name
		var gids []string
		var gnames []string
		for _, g := range selectedGroups {
			gids = append(gids, g.GroupID)
			gnames = append(gnames, g.Name)
		}
		metadata["groupids"] = strings.Join(gids, ",")
		metadata["groupnames"] = strings.Join(gnames, ",")
	}

	enabled := 1
	if toString(zh.Status) == "1" {
		enabled = 0
	}

	return &Host{
		ID:          zh.HostID,
		Name:        zh.Name,
		Description: zh.Description,
		Status:      status,
		Enabled:     enabled,
		IPAddress:   ip,
		Metadata:    metadata,
	}, nil
}

// GetItems implements the Provider interface
func (p *ZabbixProvider) GetItems(ctx context.Context, hostID string) ([]Item, error) {
	params := map[string]interface{}{
		"output":  []string{"itemid", "hostid", "name", "key_", "lastvalue", "units", "value_type", "status", "lastclock", "type", "delay", "description"},
		"hostids": hostID,
	}

	resp, err := p.sendRequest(ctx, "item.get", params)
	if err != nil {
		return nil, err
	}

	var zabbixItems []zabbixItem
	if err := json.Unmarshal(resp.Result, &zabbixItems); err != nil {
		return nil, fmt.Errorf("failed to parse items: %w", err)
	}

	items := make([]Item, 0, len(zabbixItems))
	for _, zi := range zabbixItems {
		timestamp := parseUnixClock(zi.LastClock)
		value := toString(zi.LastValue)
		units := zi.Units
		value, units = normalizeZabbixItemValue(zi.Name, zi.Key, value, units)
		items = append(items, Item{
			ID:          zi.ItemID,
			HostID:      zi.HostID,
			Name:        zi.Name,
			Key:         zi.Key,
			Type:        toString(zi.Type),
			Value:       value,
			Units:       units,
			ValueType:   toString(zi.ValueType),
			Delay:       zi.Delay,
			Status:      toString(zi.Status),
			Timestamp:   timestamp,
			Description: zi.Desc,
		})
	}

	return items, nil
}

// GetItemByID implements the Provider interface
func (p *ZabbixProvider) GetItemByID(ctx context.Context, itemID string) (*Item, error) {
	params := map[string]interface{}{
		"output":  []string{"itemid", "hostid", "name", "key_", "lastvalue", "units", "value_type", "status", "lastclock", "type", "delay", "description"},
		"itemids": itemID,
	}

	resp, err := p.sendRequest(ctx, "item.get", params)
	if err != nil {
		return nil, err
	}

	var zabbixItems []zabbixItem
	if err := json.Unmarshal(resp.Result, &zabbixItems); err != nil {
		return nil, fmt.Errorf("failed to parse item: %w", err)
	}

	if len(zabbixItems) == 0 {
		return nil, fmt.Errorf("item not found: %s", itemID)
	}

	zi := zabbixItems[0]
	value := toString(zi.LastValue)
	units := zi.Units
	value, units = normalizeZabbixItemValue(zi.Name, zi.Key, value, units)
	return &Item{
		ID:          zi.ItemID,
		HostID:      zi.HostID,
		Name:        zi.Name,
		Key:         zi.Key,
		Type:        toString(zi.Type),
		Value:       value,
		Units:       units,
		ValueType:   toString(zi.ValueType),
		Delay:       zi.Delay,
		Status:      toString(zi.Status),
		Timestamp:   parseUnixClock(zi.LastClock),
		Description: zi.Desc,
	}, nil
}

// GetItemHistory implements the Provider interface
func (p *ZabbixProvider) GetItemHistory(ctx context.Context, itemID string, from, to int64) ([]Item, error) {
	// First get the item to determine value type
	item, err := p.GetItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	params := map[string]interface{}{
		"output":    []string{"itemid", "clock", "value"},
		"itemids":   itemID,
		"time_from": from,
		"time_till": to,
		"history":   item.ValueType,
		"sortfield": "clock",
		"sortorder": "DESC",
		"limit":     1000,
	}

	resp, err := p.sendRequest(ctx, "history.get", params)
	if err != nil {
		return nil, err
	}

	var history []struct {
		ItemID string `json:"itemid"`
		Clock  string `json:"clock"`
		Value  string `json:"value"`
	}
	if err := json.Unmarshal(resp.Result, &history); err != nil {
		return nil, fmt.Errorf("failed to parse history: %w", err)
	}

	items := make([]Item, 0, len(history))
	for _, h := range history {
		timestamp := parseUnixClock(h.Clock)
		value := h.Value
		units := item.Units
		value, units = normalizeZabbixItemValue(item.Name, item.Key, value, units)
		items = append(items, Item{
			ID:        h.ItemID,
			Value:     value,
			HostID:    item.HostID,
			Name:      item.Name,
			Units:     units,
			Timestamp: timestamp,
		})
	}

	return items, nil
}

func parseUnixClock(value string) int64 {
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

// CreateItem creates an item in Zabbix
func (p *ZabbixProvider) CreateItem(ctx context.Context, item Item) (Item, error) {
	if item.HostID == "" {
		return Item{}, fmt.Errorf("host ID is required")
	}
	iface, err := p.getPrimaryInterface(ctx, item.HostID)
	if err != nil || iface.InterfaceID == "" {
		return Item{}, fmt.Errorf("failed to get host interface: %w", err)
	}
	key := item.Key
	if key == "" && item.Metadata != nil {
		key = item.Metadata["key"]
	}
	if key == "" {
		return Item{}, fmt.Errorf("item key is required")
	}
	itemType := item.Type
	if itemType == "" && item.Metadata != nil {
		itemType = item.Metadata["type"]
	}
	if itemType == "" {
		itemType = defaultItemTypeForInterface(iface.Type)
	}
	valueType := item.ValueType
	if valueType == "" && item.Metadata != nil {
		valueType = item.Metadata["value_type"]
	}
	if valueType == "" {
		valueType = "0"
	}
	status := "0"
	if item.Status == "1" || strings.ToLower(item.Status) == "disabled" {
		status = "1"
	}
	delay := item.Delay
	if delay == "" && item.Metadata != nil {
		delay = item.Metadata["delay"]
	}
	if delay == "" {
		delay = "30s"
	}
	description := item.Description
	if description == "" && item.Metadata != nil {
		description = item.Metadata["description"]
	}
	params := map[string]interface{}{
		"name":        item.Name,
		"key_":        key,
		"hostid":      item.HostID,
		"type":        itemType,
		"value_type":  valueType,
		"interfaceid": iface.InterfaceID,
		"units":       item.Units,
		"status":      status,
		"delay":       delay,
		"description": description,
	}
	resp, err := p.sendRequest(ctx, "item.create", params)
	if err != nil {
		return Item{}, fmt.Errorf("failed to create item: %w", err)
	}
	var result struct {
		ItemIDs []string `json:"itemids"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return Item{}, fmt.Errorf("failed to parse create item response: %w", err)
	}
	if len(result.ItemIDs) == 0 {
		return Item{}, fmt.Errorf("no item ID returned after creation")
	}
	item.ID = result.ItemIDs[0]
	return item, nil
}

// UpdateItem updates an item in Zabbix
func (p *ZabbixProvider) UpdateItem(ctx context.Context, item Item) (Item, error) {
	if item.ID == "" {
		return Item{}, fmt.Errorf("item ID is required")
	}
	params := map[string]interface{}{
		"itemid": item.ID,
	}
	if strings.TrimSpace(item.Name) != "" {
		params["name"] = item.Name
	}
	if strings.TrimSpace(item.Key) != "" {
		params["key_"] = item.Key
	}
	if strings.TrimSpace(item.Units) != "" {
		params["units"] = item.Units
	}
	if strings.TrimSpace(item.ValueType) != "" {
		params["value_type"] = item.ValueType
	}
	if strings.TrimSpace(item.Type) != "" {
		params["type"] = item.Type
	}
	if strings.TrimSpace(item.Delay) != "" {
		params["delay"] = item.Delay
	}
	if strings.TrimSpace(item.Description) != "" {
		params["description"] = item.Description
	}
	if strings.TrimSpace(item.InterfaceID) != "" {
		params["interfaceid"] = item.InterfaceID
	}
	if item.Status != "" {
		status := "0"
		if item.Status == "1" || strings.ToLower(item.Status) == "disabled" {
			status = "1"
		}
		params["status"] = status
	}
	if _, err := p.sendRequest(ctx, "item.update", params); err != nil {
		return Item{}, fmt.Errorf("failed to update item: %w", err)
	}
	return item, nil
}

func defaultItemTypeForInterface(ifaceType string) string {
	switch ifaceType {
	case "2":
		return "20" // SNMP agent
	case "3":
		return "12" // IPMI agent
	case "4":
		return "16" // JMX agent
	default:
		return "0" // Zabbix agent
	}
}

// DeleteItem deletes an item in Zabbix
func (p *ZabbixProvider) DeleteItem(ctx context.Context, itemID string) error {
	if itemID == "" {
		return fmt.Errorf("item ID is required")
	}
	params := map[string]interface{}{
		"itemids": []string{itemID},
	}
	if _, err := p.sendRequest(ctx, "item.delete", params); err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	return nil
}

func (p *ZabbixProvider) getPrimaryInterface(ctx context.Context, hostID string) (struct {
	InterfaceID string
	IP          string
	DNS         string
	Port        string
	Type        string
	Main        string
	UseIP       string
}, error) {
	params := map[string]interface{}{
		"output":           []string{"hostid"},
		"selectInterfaces": []string{"interfaceid", "ip", "dns", "port", "type", "main", "useip"},
		"hostids":          hostID,
	}
	resp, err := p.sendRequest(ctx, "host.get", params)
	if err != nil {
		return struct {
			InterfaceID string
			IP          string
			DNS         string
			Port        string
			Type        string
			Main        string
			UseIP       string
		}{}, err
	}
	var zabbixHosts []zabbixHost
	if err := json.Unmarshal(resp.Result, &zabbixHosts); err != nil {
		return struct {
			InterfaceID string
			IP          string
			DNS         string
			Port        string
			Type        string
			Main        string
			UseIP       string
		}{}, err
	}
	if len(zabbixHosts) == 0 || len(zabbixHosts[0].Interfaces) == 0 {
		return struct {
			InterfaceID string
			IP          string
			DNS         string
			Port        string
			Type        string
			Main        string
			UseIP       string
		}{}, fmt.Errorf("no interface found")
	}
	selected := zabbixHosts[0].Interfaces[0]
	for _, iface := range zabbixHosts[0].Interfaces {
		if toString(iface.Main) == "1" {
			selected = iface
			break
		}
	}
	return struct {
		InterfaceID string
		IP          string
		DNS         string
		Port        string
		Type        string
		Main        string
		UseIP       string
	}{
		InterfaceID: toString(selected.InterfaceID),
		IP:          selected.IP,
		DNS:         selected.DNS,
		Port:        toString(selected.Port),
		Type:        toString(selected.Type),
		Main:        toString(selected.Main),
		UseIP:       toString(selected.UseIP),
	}, nil
}

// GetAlerts implements the Provider interface
func (p *ZabbixProvider) GetAlerts(ctx context.Context) ([]Alert, error) {
	params := map[string]interface{}{
		"output":      []string{"eventid", "objectid", "name", "severity", "acknowledged", "clock"},
		"recent":      true,
		"sortfield":   []string{"eventid"},
		"sortorder":   "DESC",
		"selectHosts": []string{"hostid"},
	}

	resp, err := p.sendRequest(ctx, "problem.get", params)
	if err != nil {
		return nil, err
	}

	var problems []struct {
		EventID      string      `json:"eventid"`
		ObjectID     string      `json:"objectid"`
		Name         string      `json:"name"`
		Severity     interface{} `json:"severity"`
		Acknowledged interface{} `json:"acknowledged"`
		Clock        interface{} `json:"clock"`
		Hosts        []struct {
			HostID interface{} `json:"hostid"`
		} `json:"hosts"`
	}
	if err := json.Unmarshal(resp.Result, &problems); err != nil {
		return nil, fmt.Errorf("failed to parse problems: %w", err)
	}

	alerts := make([]Alert, 0, len(problems))
	for _, prob := range problems {
		hostID := ""
		if len(prob.Hosts) > 0 {
			hostID = toString(prob.Hosts[0].HostID)
		}

		severity := "information"
		switch toString(prob.Severity) {
		case "5":
			severity = "disaster"
		case "4":
			severity = "high"
		case "3":
			severity = "average"
		case "2":
			severity = "warning"
		case "1":
			severity = "information"
		}

		alerts = append(alerts, Alert{
			ID:       prob.EventID,
			HostID:   hostID,
			Name:     prob.Name,
			Severity: severity,
			Status:   "problem",
		})
	}

	return alerts, nil
}

// GetAlertsByHost implements the Provider interface
func (p *ZabbixProvider) GetAlertsByHost(ctx context.Context, hostID string) ([]Alert, error) {
	params := map[string]interface{}{
		"output":    []string{"eventid", "objectid", "name", "severity", "acknowledged", "clock"},
		"hostids":   hostID,
		"recent":    true,
		"sortfield": []string{"eventid"},
		"sortorder": "DESC",
	}

	resp, err := p.sendRequest(ctx, "problem.get", params)
	if err != nil {
		return nil, err
	}

	var problems []zabbixProblem
	if err := json.Unmarshal(resp.Result, &problems); err != nil {
		return nil, fmt.Errorf("failed to parse problems: %w", err)
	}

	alerts := make([]Alert, 0, len(problems))
	for _, prob := range problems {
		severity := "information"
		switch toString(prob.Severity) {
		case "5":
			severity = "disaster"
		case "4":
			severity = "high"
		case "3":
			severity = "average"
		case "2":
			severity = "warning"
		case "1":
			severity = "information"
		}

		alerts = append(alerts, Alert{
			ID:       prob.EventID,
			HostID:   hostID,
			Name:     prob.Name,
			Severity: severity,
			Status:   "problem",
		})
	}

	return alerts, nil
}

// GetTriggers implements the Provider interface
func (p *ZabbixProvider) GetTriggers(ctx context.Context) ([]Trigger, error) {
	params := map[string]interface{}{
		"output":            []string{"triggerid", "description", "expression", "priority", "status", "value"},
		"selectHosts":       []string{"hostid"},
		"only_true":         true,
		"active":            true,
		"expandDescription": true,
	}

	resp, err := p.sendRequest(ctx, "trigger.get", params)
	if err != nil {
		return nil, err
	}

	var zabbixTriggers []zabbixTrigger
	if err := json.Unmarshal(resp.Result, &zabbixTriggers); err != nil {
		return nil, fmt.Errorf("failed to parse triggers: %w", err)
	}

	triggers := make([]Trigger, 0, len(zabbixTriggers))
	for _, zt := range zabbixTriggers {
		priority := "information"
		switch toString(zt.Priority) {
		case "5":
			priority = "disaster"
		case "4":
			priority = "high"
		case "3":
			priority = "average"
		case "2":
			priority = "warning"
		case "1":
			priority = "information"
		}

		status := "ok"
		if toString(zt.Value) == "1" {
			status = "problem"
		}

		triggers = append(triggers, Trigger{
			ID:          zt.TriggerID,
			Name:        zt.Description,
			Expression:  zt.Expression,
			Priority:    priority,
			Status:      status,
			Description: zt.Description,
		})
	}

	return triggers, nil
}

// GetTriggersByHost implements the Provider interface
func (p *ZabbixProvider) GetTriggersByHost(ctx context.Context, hostID string) ([]Trigger, error) {
	params := map[string]interface{}{
		"output":            []string{"triggerid", "description", "expression", "priority", "status", "value"},
		"hostids":           hostID,
		"only_true":         true,
		"active":            true,
		"expandDescription": true,
	}

	resp, err := p.sendRequest(ctx, "trigger.get", params)
	if err != nil {
		return nil, err
	}

	var zabbixTriggers []zabbixTrigger
	if err := json.Unmarshal(resp.Result, &zabbixTriggers); err != nil {
		return nil, fmt.Errorf("failed to parse triggers: %w", err)
	}

	triggers := make([]Trigger, 0, len(zabbixTriggers))
	for _, zt := range zabbixTriggers {
		priority := "information"
		switch toString(zt.Priority) {
		case "5":
			priority = "disaster"
		case "4":
			priority = "high"
		case "3":
			priority = "average"
		case "2":
			priority = "warning"
		case "1":
			priority = "information"
		}

		status := "ok"
		if toString(zt.Value) == "1" {
			status = "problem"
		}

		triggers = append(triggers, Trigger{
			ID:          zt.TriggerID,
			Name:        zt.Description,
			Expression:  zt.Expression,
			Priority:    priority,
			Status:      status,
			Description: zt.Description,
		})
	}

	return triggers, nil
}

func (p *ZabbixProvider) SetupWebhookMediaActionAndUser(ctx context.Context, cfg ZabbixWebhookSetupConfig) (ZabbixWebhookSetupResult, error) {
	if strings.TrimSpace(cfg.WebhookURL) == "" {
		return ZabbixWebhookSetupResult{}, fmt.Errorf("webhook URL is required")
	}
	if strings.TrimSpace(cfg.EventToken) == "" {
		return ZabbixWebhookSetupResult{}, fmt.Errorf("event token is required")
	}
	if strings.TrimSpace(cfg.UserLookup) == "" && strings.TrimSpace(cfg.ZabbixUserID) == "" {
		return ZabbixWebhookSetupResult{}, fmt.Errorf("user lookup or ZabbixUserID is required")
	}
	if strings.TrimSpace(cfg.MediaTypeName) == "" {
		cfg.MediaTypeName = "Nagare Webhook"
	}
	if strings.TrimSpace(cfg.ActionName) == "" {
		cfg.ActionName = "Nagare Alert Push"
	}
	if strings.TrimSpace(cfg.UserMediaSendTo) == "" {
		cfg.UserMediaSendTo = "nagare-webhook"
	}
	if strings.TrimSpace(cfg.ActionEscalation) == "" {
		cfg.ActionEscalation = "1m"
	}

	if strings.TrimSpace(p.GetAuthToken()) == "" {
		if err := p.Authenticate(ctx); err != nil {
			return ZabbixWebhookSetupResult{}, fmt.Errorf("failed to authenticate with zabbix: %w", err)
		}
	}

	mediaTypeID, err := p.ensureWebhookMediaType(ctx, cfg)
	if err != nil {
		return ZabbixWebhookSetupResult{}, err
	}

	var userID, username string
	if strings.TrimSpace(cfg.ZabbixUserID) != "" {
		userID, username, err = p.getUserByID(ctx, cfg.ZabbixUserID)
	} else {
		userID, username, err = p.findUserByLogin(ctx, cfg.UserLookup)
	}
	if err != nil {
		return ZabbixWebhookSetupResult{}, err
	}

	if err := p.ensureUserMedia(ctx, userID, mediaTypeID, cfg.UserMediaSendTo); err != nil {
		return ZabbixWebhookSetupResult{}, err
	}

	actionID, actionName, err := p.ensureActionBound(ctx, cfg.ActionName, mediaTypeID, userID, cfg.ActionEscalation)
	if err != nil {
		return ZabbixWebhookSetupResult{}, err
	}

	return ZabbixWebhookSetupResult{
		WebhookURL:  cfg.WebhookURL,
		MediaTypeID: mediaTypeID,
		ActionID:    actionID,
		ActionName:  actionName,
		UserID:      userID,
		Username:    username,
	}, nil
}

func (p *ZabbixProvider) ensureWebhookMediaType(ctx context.Context, cfg ZabbixWebhookSetupConfig) (string, error) {
	getParams := map[string]interface{}{
		"output": []string{"mediatypeid", "name"},
		"filter": map[string]interface{}{"name": []string{cfg.MediaTypeName}},
	}

	getResp, err := p.sendRequest(ctx, "mediatype.get", getParams)
	if err != nil {
		return "", fmt.Errorf("failed to query zabbix media type: %w", err)
	}

	var existing []struct {
		MediaTypeID string `json:"mediatypeid"`
		Name        string `json:"name"`
	}
	if err := json.Unmarshal(getResp.Result, &existing); err != nil {
		return "", fmt.Errorf("failed to parse zabbix media type list: %w", err)
	}

	script := "var req = new HttpRequest();\n" +
		"req.addHeader('Content-Type: application/json');\n" +
		"var params = JSON.parse(value);\n" +
		"var payload = {\n" +
		"  message: params.Subject,\n" +
		"  details: params.Message,\n" +
		"  severity: params.Severity,\n" +
		"  host_id: params.HostID,\n" +
		"  item_id: params.ItemID,\n" +
		"  host_name: params.HostName,\n" +
		"  item_name: params.ItemName,\n" +
		"  event_id: params.EventID,\n" +
		"  status: params.EventValue == '1' ? 'open' : 'resolved',\n" +
		"  token: params.EventToken\n" +
		"};\n" +
		"var response = req.post(params.URL, JSON.stringify(payload));\n" +
		"return response;"

	parameters := []map[string]interface{}{
		{"name": "URL", "value": cfg.WebhookURL},
		{"name": "EventToken", "value": cfg.EventToken},
		{"name": "Subject", "value": "{EVENT.NAME}"},
		{"name": "Message", "value": "Host: {HOST.NAME} ({HOST.IP}), Item: {ITEM.NAME}, Value: {ITEM.VALUE}, Severity: {EVENT.SEVERITY}, Time: {EVENT.DATE} {EVENT.TIME}"},
		{"name": "Severity", "value": "{EVENT.NSEVERITY}"},
		{"name": "HostID", "value": "{HOST.ID}"},
		{"name": "HostName", "value": "{HOST.NAME}"},
		{"name": "ItemID", "value": "{ITEM.ID}"},
		{"name": "ItemName", "value": "{ITEM.NAME}"},
		{"name": "EventID", "value": "{EVENT.ID}"},
		{"name": "EventValue", "value": "{EVENT.VALUE}"},
	}

	messageTemplates := []map[string]interface{}{
		{
			"eventsource": 0, // Triggers
			"recovery":    0, // Problem
			"subject":     "Nagare Alert: {EVENT.NAME}",
			"message":     "Host: {HOST.NAME} ({HOST.IP})\nItem: {ITEM.NAME}\nValue: {ITEM.VALUE}\nSeverity: {EVENT.SEVERITY}\nTime: {EVENT.DATE} {EVENT.TIME}\n\n{EVENT.OPDATA}",
		},
		{
			"eventsource": 0, // Triggers
			"recovery":    1, // Recovery
			"subject":     "Resolved: {EVENT.NAME}",
			"message":     "Host: {HOST.NAME} ({HOST.IP})\nItem: {ITEM.NAME}\nStatus: Resolved\nTime: {EVENT.RECOVERY.DATE} {EVENT.RECOVERY.TIME}",
		},
		{
			"eventsource": 0, // Triggers
			"recovery":    2, // Update
			"subject":     "Updated: {EVENT.NAME}",
			"message":     "Host: {HOST.NAME} ({HOST.IP})\nUser: {USER.FULLNAME}\nAction: {EVENT.UPDATE.ACTION}\nTime: {EVENT.UPDATE.DATE} {EVENT.UPDATE.TIME}",
		},
	}

	if len(existing) > 0 {
		id := existing[0].MediaTypeID
		updateParams := map[string]interface{}{
			"mediatypeid":       id,
			"parameters":        parameters,
			"message_templates": messageTemplates,
			"script":            script,
		}
		if _, err := p.sendRequest(ctx, "mediatype.update", updateParams); err != nil {
			return id, fmt.Errorf("failed to update zabbix media type: %w", err)
		}
		return id, nil
	}

	createParams := map[string]interface{}{
		"name":              cfg.MediaTypeName,
		"type":              4,
		"status":            0,
		"parameters":        parameters,
		"message_templates": messageTemplates,
		"script":            script,
	}

	createResp, err := p.sendRequest(ctx, "mediatype.create", createParams)
	if err != nil {
		return "", fmt.Errorf("failed to create zabbix media type: %w", err)
	}

	var created struct {
		MediaTypeIDs []string `json:"mediatypeids"`
	}
	if err := json.Unmarshal(createResp.Result, &created); err != nil {
		return "", fmt.Errorf("failed to parse zabbix media type create result: %w", err)
	}
	if len(created.MediaTypeIDs) == 0 {
		return "", fmt.Errorf("zabbix media type created but no mediatypeid returned")
	}

	return created.MediaTypeIDs[0], nil
}

func (p *ZabbixProvider) findUserByLogin(ctx context.Context, lookup string) (string, string, error) {
	type zabbixUser struct {
		UserID   string `json:"userid"`
		Username string `json:"username"`
		Alias    string `json:"alias"`
	}

	// Try filters one by one to avoid errors if a field doesn't exist in the current Zabbix version
	fieldNames := []string{"username", "alias"}
	for _, field := range fieldNames {
		resp, err := p.sendRequest(ctx, "user.get", map[string]interface{}{
			"output": "extend",
			"filter": map[string]interface{}{
				field: []string{lookup},
			},
		})
		if err != nil {
			// Continue if this specific field filter is not supported
			continue
		}

		var users []zabbixUser
		if err := json.Unmarshal(resp.Result, &users); err != nil {
			continue
		}
		if len(users) > 0 {
			name := users[0].Username
			if strings.TrimSpace(name) == "" {
				name = users[0].Alias
			}
			return users[0].UserID, name, nil
		}
	}

	// If exact match not found, try to get all users using 'extend' to be safe across versions
	availableUsers, err := p.getAvailableUsers(ctx)
	if err == nil && len(availableUsers) > 0 {
		// Try to pick a sensible default - look for Admin first (case-insensitive)
		for _, user := range availableUsers {
			uName := user.Username
			if uName == "" {
				uName = user.Alias
			}
			if strings.EqualFold(uName, "admin") || strings.EqualFold(uName, "Admin") {
				return user.UserID, uName, nil
			}
		}

		// If no Admin, use first available user
		name := availableUsers[0].Username
		if strings.TrimSpace(name) == "" {
			name = availableUsers[0].Alias
		}

		return availableUsers[0].UserID, name, nil
	}

	// Last resort fallback: query all users without any filter and look for matches in the code
	resp, err := p.sendRequest(ctx, "user.get", map[string]interface{}{
		"output": "extend",
	})
	if err == nil {
		var users []zabbixUser
		if err := json.Unmarshal(resp.Result, &users); err == nil && len(users) > 0 {
			for _, user := range users {
				uName := user.Username
				if uName == "" {
					uName = user.Alias
				}
				if strings.EqualFold(uName, lookup) || strings.EqualFold(uName, "admin") {
					return user.UserID, uName, nil
				}
			}
			// Just return the first one if we found anything
			uName := users[0].Username
			if uName == "" {
				uName = users[0].Alias
			}
			return users[0].UserID, uName, nil
		}
	}

	// If all else fails, return detailed error
	return "", "", fmt.Errorf("zabbix user '%s' not found: unable to lookup user or retrieve user list from zabbix", lookup)
}

func (p *ZabbixProvider) getUserByID(ctx context.Context, userID string) (string, string, error) {
	if strings.TrimSpace(userID) == "" {
		return "", "", fmt.Errorf("user ID is required")
	}

	resp, err := p.sendRequest(ctx, "user.get", map[string]interface{}{
		"output":  "extend",
		"userids": []string{userID},
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to query zabbix user by ID: %w", err)
	}

	type zabbixUser struct {
		UserID   string `json:"userid"`
		Username string `json:"username"`
		Alias    string `json:"alias"`
	}

	var users []zabbixUser
	if err := json.Unmarshal(resp.Result, &users); err != nil {
		return "", "", fmt.Errorf("failed to parse zabbix user: %w", err)
	}

	if len(users) == 0 {
		return "", "", fmt.Errorf("zabbix user with ID '%s' not found", userID)
	}

	username := users[0].Username
	if strings.TrimSpace(username) == "" {
		username = users[0].Alias
	}

	return users[0].UserID, username, nil
}

func (p *ZabbixProvider) getAvailableUsers(ctx context.Context) ([]struct {
	UserID   string
	Username string
	Alias    string
}, error) {
	resp, err := p.sendRequest(ctx, "user.get", map[string]interface{}{
		"output": "extend",
		"limit":  100,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query zabbix users: %w", err)
	}

	type zabbixUser struct {
		UserID   string `json:"userid"`
		Username string `json:"username"`
		Alias    string `json:"alias"`
	}

	var users []zabbixUser
	if err := json.Unmarshal(resp.Result, &users); err != nil {
		return nil, fmt.Errorf("failed to parse zabbix users: %w", err)
	}

	// Build result with same struct type as return type
	result := make([]struct {
		UserID   string
		Username string
		Alias    string
	}, len(users))

	for i, user := range users {
		result[i].UserID = user.UserID
		// Map correctly based on which field is populated
		result[i].Username = user.Username
		if result[i].Username == "" {
			result[i].Username = user.Alias
		}
		result[i].Alias = user.Alias
	}

	return result, nil
}

func (p *ZabbixProvider) ensureUserMedia(ctx context.Context, userID, mediaTypeID, sendTo string) error {
	resp, err := p.sendRequest(ctx, "user.get", map[string]interface{}{
		"output":       []string{"userid"},
		"userids":      []string{userID},
		"selectMedias": "extend",
	})
	if err != nil {
		return fmt.Errorf("failed to load zabbix user medias: %w", err)
	}

	var users []struct {
		UserID string                   `json:"userid"`
		Medias []map[string]interface{} `json:"medias"`
	}
	if err := json.Unmarshal(resp.Result, &users); err != nil {
		return fmt.Errorf("failed to parse zabbix user medias: %w", err)
	}
	if len(users) == 0 {
		return fmt.Errorf("zabbix user not found for userid: %s", userID)
	}

	medias := users[0].Medias
	for _, media := range medias {
		if toString(media["mediatypeid"]) == mediaTypeID {
			return nil
		}
	}

	// Clean up existing medias to include only supported fields for update
	// This prevents errors like "unexpected parameter 'provisioned'"
	cleanedMedias := make([]map[string]interface{}, 0, len(medias)+1)
	for _, m := range medias {
		cleanedMedia := map[string]interface{}{
			"mediatypeid": m["mediatypeid"],
			"sendto":      m["sendto"],
			"active":      m["active"],
			"severity":    m["severity"],
			"period":      m["period"],
		}
		cleanedMedias = append(cleanedMedias, cleanedMedia)
	}

	// Add new Nagare media
	cleanedMedias = append(cleanedMedias, map[string]interface{}{
		"mediatypeid": mediaTypeID,
		"sendto":      sendTo,
		"active":      0,
		"severity":    63,
		"period":      "1-7,00:00-24:00",
	})

	if _, err := p.sendRequest(ctx, "user.update", map[string]interface{}{
		"userid": userID,
		"medias": cleanedMedias,
	}); err != nil {
		return fmt.Errorf("failed to bind media type to zabbix user: %w", err)
	}

	return nil
}

func (p *ZabbixProvider) ensureActionBound(ctx context.Context, actionName, mediaTypeID, userID, escPeriod string) (string, string, error) {
	existingID, _, err := p.findActionByName(ctx, actionName)
	actionExists := err == nil && strings.TrimSpace(existingID) != ""

	baseFilter, err := p.findBaseTriggerActionFilter(ctx)
	if err != nil {
		return "", "", err
	}

	operations := []map[string]interface{}{
		{
			"operationtype": 0, // Send message
			"opmessage": map[string]interface{}{
				"default_msg": 1,
				"mediatypeid": mediaTypeID,
			},
			"opmessage_usr": []map[string]interface{}{{"userid": userID}},
		},
	}
	recoveryOperations := []map[string]interface{}{
		{
			"operationtype": 11, // Recovery message
			"opmessage": map[string]interface{}{
				"default_msg": 1,
			},
		},
	}
	updateOperations := []map[string]interface{}{
		{
			"operationtype": 12, // Update message
			"opmessage": map[string]interface{}{
				"default_msg": 1,
			},
		},
	}

	if actionExists {
		updateParams := map[string]interface{}{
			"actionid":            existingID,
			"operations":          operations,
			"recovery_operations": recoveryOperations,
			"update_operations":   updateOperations,
		}
		if _, err := p.sendRequest(ctx, "action.update", updateParams); err != nil {
			return existingID, actionName, fmt.Errorf("failed to update zabbix action: %w", err)
		}
		return existingID, actionName, nil
	}

	createParams := map[string]interface{}{
		"name":                actionName,
		"eventsource":         0,
		"status":              0,
		"esc_period":          escPeriod,
		"filter":              baseFilter,
		"operations":          operations,
		"recovery_operations": recoveryOperations,
		"update_operations":   updateOperations,
	}

	createResp, err := p.sendRequest(ctx, "action.create", createParams)
	if err != nil {
		return "", "", fmt.Errorf("failed to create zabbix action: %w", err)
	}

	var created struct {
		ActionIDs []string `json:"actionids"`
	}
	if err := json.Unmarshal(createResp.Result, &created); err != nil {
		return "", "", fmt.Errorf("failed to parse zabbix action create result: %w", err)
	}
	if len(created.ActionIDs) == 0 {
		return "", "", fmt.Errorf("zabbix action created but no actionid returned")
	}

	return created.ActionIDs[0], actionName, nil
}

func (p *ZabbixProvider) findActionByName(ctx context.Context, actionName string) (string, string, error) {
	resp, err := p.sendRequest(ctx, "action.get", map[string]interface{}{
		"output": []string{"actionid", "name"},
		"filter": map[string]interface{}{"name": []string{actionName}},
		"limit":  1,
	})
	if err != nil {
		return "", "", err
	}

	var actions []struct {
		ActionID string `json:"actionid"`
		Name     string `json:"name"`
	}
	if err := json.Unmarshal(resp.Result, &actions); err != nil {
		return "", "", err
	}
	if len(actions) == 0 {
		return "", "", fmt.Errorf("action not found")
	}

	return actions[0].ActionID, actions[0].Name, nil
}

func (p *ZabbixProvider) findBaseTriggerActionFilter(ctx context.Context) (map[string]interface{}, error) {
	resp, err := p.sendRequest(ctx, "action.get", map[string]interface{}{
		"output":       []string{"actionid", "name", "eventsource", "status"},
		"eventsource":  0,
		"selectFilter": "extend",
		"sortfield":    "actionid",
		"sortorder":    "ASC",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query zabbix base action: %w", err)
	}

	var actions []struct {
		Status interface{}            `json:"status"`
		Filter map[string]interface{} `json:"filter"`
	}
	if err := json.Unmarshal(resp.Result, &actions); err != nil {
		return nil, fmt.Errorf("failed to parse zabbix base action: %w", err)
	}

	for _, action := range actions {
		if toString(action.Status) == "0" && len(action.Filter) > 0 {
			// Clean up read-only fields from filter to prevent action.create failure
			delete(action.Filter, "eval_formula")
			delete(action.Filter, "formula")
			if conds, ok := action.Filter["conditions"].([]interface{}); ok {
				for i, c := range conds {
					if condMap, ok := c.(map[string]interface{}); ok {
						delete(condMap, "conditionid")
						delete(condMap, "actionid")
						conds[i] = condMap
					}
				}
				action.Filter["conditions"] = conds
			}
			return action.Filter, nil
		}
	}

	// Default fallback: Trigger severity >= Not classified (conditiontype 4, operator 5, value 0)
	// This ensures setup works even if no actions exist yet, and is compatible with Zabbix 6.0+.
	return map[string]interface{}{
		"evaltype": 0, // AND/OR
		"conditions": []map[string]interface{}{
			{
				"conditiontype": 4,
				"operator":      5,
				"value":         "0",
			},
		},
	}, nil
}

// Name returns the provider name
func (p *ZabbixProvider) Name() string {
	return "zabbix"
}

// Type returns the provider type
func (p *ZabbixProvider) Type() MonitorType {
	return MonitorZabbix
}

func (p *ZabbixProvider) CreateHost(ctx context.Context, host Host) (Host, error) {
	groupID := "7"
	isSNMP := false
	snmpVersion := 2 // default v2c
	snmpCommunity := "public"
	snmpPort := "161"

	if host.Metadata != nil {
		if gid, ok := host.Metadata["groupid"]; ok && gid != "" {
			groupID = gid
		}
		if mType, ok := host.Metadata["monitor_type"]; ok && mType == "2" {
			isSNMP = true
		}
		if ver, ok := host.Metadata["snmp_version"]; ok {
			if strings.Contains(ver, "1") {
				snmpVersion = 1
			} else if strings.Contains(ver, "3") {
				snmpVersion = 3
			}
		}
		if comm, ok := host.Metadata["snmp_community"]; ok && comm != "" {
			snmpCommunity = comm
		}
		if port, ok := host.Metadata["snmp_port"]; ok && port != "" && port != "0" {
			snmpPort = port
		}
	}

	var interfaces []map[string]interface{}
	if isSNMP {
		iface := map[string]interface{}{
			"type":  2, // SNMP
			"main":  1,
			"useip": 1,
			"ip":    host.IPAddress,
			"dns":   "",
			"port":  snmpPort,
			"details": map[string]interface{}{
				"version":   snmpVersion,
				"bulk":      1,
				"community": snmpCommunity,
			},
		}
		// Zabbix 6.0+ requires community at top level of details or as a macro
		interfaces = append(interfaces, iface)
	} else {
		// Default to Zabbix Agent interface
		interfaces = append(interfaces, map[string]interface{}{
			"type":  1, // Agent
			"main":  1,
			"useip": 1,
			"ip":    host.IPAddress,
			"dns":   "",
			"port":  "10050",
		})
	}

	// Prepare templates
	var templates []map[string]string

	// Try to use template from metadata
	if host.Metadata != nil {
		if tid, ok := host.Metadata["templateid"]; ok && tid != "" {
			templates = append(templates, map[string]string{"templateid": tid})
		}
	}

	// If no template provided and it's SNMP, use "Huawei VRP by SNMP" as default
	if len(templates) == 0 && isSNMP {
		tids, err := p.GetTemplateidByName(ctx, "Huawei VRP by SNMP")
		if err == nil && len(tids) > 0 {
			templates = append(templates, map[string]string{"templateid": tids[0]})
		} else {
			// Fallback: search by case-insensitive name if exact match fails
			params := map[string]interface{}{
				"output": []string{"templateid", "name"},
				"search": map[string]interface{}{"name": "Huawei VRP"},
			}
			resp, err := p.sendRequest(ctx, "template.get", params)
			if err == nil {
				var searchTmpls []struct {
					TemplateID string `json:"templateid"`
					Name       string `json:"name"`
				}
				if json.Unmarshal(resp.Result, &searchTmpls) == nil && len(searchTmpls) > 0 {
					templates = append(templates, map[string]string{"templateid": searchTmpls[0].TemplateID})
				}
			}
		}
	}

	// Zabbix status: 0 = monitored, 1 = unmonitored
	zabbixStatus := 0
	if host.Enabled == 0 {
		zabbixStatus = 1
	}

	params := map[string]interface{}{
		"host":       host.Name,
		"interfaces": interfaces,
		"groups":     []map[string]string{{"groupid": groupID}},
		"status":     zabbixStatus,
	}

	if len(templates) > 0 {
		params["templates"] = templates
	}

	resp, err := p.sendRequest(ctx, "host.create", params)
	if err != nil {
		return Host{}, fmt.Errorf("failed to create host: %w", err)
	}
	var result struct {
		HostIDs []string `json:"hostids"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return Host{}, fmt.Errorf("failed to parse create host response: %w", err)
	}
	if len(result.HostIDs) == 0 {
		return Host{}, fmt.Errorf("no host ID returned after creation")
	}
	host.ID = result.HostIDs[0]
	return host, nil
}

func (p *ZabbixProvider) UpdateHost(ctx context.Context, host Host) (Host, error) {
	if host.ID == "" {
		return Host{}, fmt.Errorf("host ID is required")
	}
	// Zabbix status: 0 = monitored, 1 = unmonitored
	zabbixStatus := 0
	if host.Enabled == 0 {
		zabbixStatus = 1
	}

	params := map[string]interface{}{
		"hostid":      host.ID,
		"host":        host.Name,
		"name":        host.Name,
		"description": host.Description,
		"status":      zabbixStatus,
	}
	if host.Metadata != nil {
		if gid, ok := host.Metadata["groupid"]; ok && gid != "" {
			params["groups"] = []map[string]string{{"groupid": gid}}
		}
	}
	if host.IPAddress != "" {
		iface, err := p.getPrimaryInterface(ctx, host.ID)
		if err == nil && iface.InterfaceID != "" {
			ifaceParams := map[string]interface{}{
				"interfaceid": iface.InterfaceID,
				"type":        iface.Type,
				"main":        iface.Main,
				"useip":       iface.UseIP,
				"ip":          host.IPAddress,
				"dns":         iface.DNS,
				"port":        iface.Port,
			}
			
			// Update SNMP details if provided in metadata
			if host.Metadata != nil && host.Metadata["monitor_type"] == "2" {
				snmpVersion := 2
				if ver, ok := host.Metadata["snmp_version"]; ok {
					if strings.Contains(ver, "1") {
						snmpVersion = 1
					} else if strings.Contains(ver, "3") {
						snmpVersion = 3
					}
				}
				comm := host.Metadata["snmp_community"]
				if comm == "" {
					comm = "public"
				}
				
				ifaceParams["details"] = map[string]interface{}{
					"version":   snmpVersion,
					"bulk":      1,
					"community": comm,
				}
				ifaceParams["type"] = 2 // Ensure type is SNMP
			}
			
			params["interfaces"] = []map[string]interface{}{ifaceParams}
		}
	}
	if _, err := p.sendRequest(ctx, "host.update", params); err != nil {
		return Host{}, fmt.Errorf("failed to update host: %w", err)
	}
	return host, nil
}

func (p *ZabbixProvider) DeleteHost(ctx context.Context, hostID string) error {
	if hostID == "" {
		return fmt.Errorf("host ID is required")
	}
	params := map[string]interface{}{
		"hostids": []string{hostID},
	}
	if _, err := p.sendRequest(ctx, "host.delete", params); err != nil {
		return fmt.Errorf("failed to delete host: %w", err)
	}
	return nil
}
func (p *ZabbixProvider) GetTemplateidByName(ctx context.Context, name string) ([]string, error) {
	params := map[string]interface{}{
		"output": []string{"templateid", "name"},
		"filter": map[string]interface{}{"name": []string{name}},
	}
	resp, err := p.sendRequest(ctx, "template.get", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get templates by name: %w", err)
	}
	var templates []struct {
		TemplateID string `json:"templateid"`
		Name       string `json:"name"`
	}
	if err := json.Unmarshal(resp.Result, &templates); err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}
	templateIDs := make([]string, 0, len(templates))
	for _, t := range templates {
		templateIDs = append(templateIDs, t.TemplateID)
	}
	return templateIDs, nil
}

func (p *ZabbixProvider) GetHostGroups(ctx context.Context) ([]string, error) {
	params := map[string]interface{}{
		"output": []string{"groupid", "name"},
	}
	resp, err := p.sendRequest(ctx, "hostgroup.get", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get host groups: %w", err)
	}
	var groups []struct {
		GroupID string `json:"groupid"`
		Name    string `json:"name"`
	}
	if err := json.Unmarshal(resp.Result, &groups); err != nil {
		return nil, fmt.Errorf("failed to parse host groups: %w", err)
	}
	groupNames := make([]string, 0, len(groups))
	for _, g := range groups {
		groupNames = append(groupNames, g.Name)
	}
	return groupNames, nil
}

func (p *ZabbixProvider) GetHostGroupsDetails(ctx context.Context) ([]struct{ ID, Name string }, error) {
	params := map[string]interface{}{
		"output": "extend",
	}
	resp, err := p.sendRequest(ctx, "hostgroup.get", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get host groups: %w", err)
	}

	// Debug logging
	// fmt.Printf("Zabbix HostGroups Response: %s\n", string(resp.Result))

	var groups []struct {
		GroupID string `json:"groupid"`
		Name    string `json:"name"`
	}
	if err := json.Unmarshal(resp.Result, &groups); err != nil {
		return nil, fmt.Errorf("failed to parse host groups: %w", err)
	}
	result := make([]struct{ ID, Name string }, 0, len(groups))
	for _, g := range groups {
		result = append(result, struct{ ID, Name string }{ID: g.GroupID, Name: g.Name})
	}
	return result, nil
}

func (p *ZabbixProvider) UpdateHostGroup(ctx context.Context, id, name string) error {
	params := map[string]interface{}{
		"groupid": id,
		"name":    name,
	}
	if _, err := p.sendRequest(ctx, "hostgroup.update", params); err != nil {
		return fmt.Errorf("failed to update host group: %w", err)
	}
	return nil
}

func (p *ZabbixProvider) DeleteHostGroup(ctx context.Context, id string) error {
	params := map[string]interface{}{
		"groupids": []string{id},
	}
	if _, err := p.sendRequest(ctx, "hostgroup.delete", params); err != nil {
		return fmt.Errorf("failed to delete host group: %w", err)
	}
	return nil
}

func (p *ZabbixProvider) GetHostGroupByName(ctx context.Context, name string) (string, error) {
	params := map[string]interface{}{
		"output": "extend",
		"filter": map[string]interface{}{"name": []string{name}},
	}
	resp, err := p.sendRequest(ctx, "hostgroup.get", params)
	if err != nil {
		return "", fmt.Errorf("failed to get host group by name: %w", err)
	}
	var groups []struct {
		GroupID string `json:"groupid"`
		Name    string `json:"name"`
	}
	if err := json.Unmarshal(resp.Result, &groups); err != nil {
		return "", fmt.Errorf("failed to parse host groups: %w", err)
	}
	if len(groups) == 0 {
		return "", fmt.Errorf("host group not found: %s", name)
	}
	return groups[0].GroupID, nil
}

// CreateHostGroup ensures a host group exists and returns its ID
func (p *ZabbixProvider) CreateHostGroup(ctx context.Context, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("host group name is required")
	}
	getParams := map[string]interface{}{
		"output": "extend",
		"filter": map[string]interface{}{"name": []string{name}},
	}
	resp, err := p.sendRequest(ctx, "hostgroup.get", getParams)
	if err == nil {
		var result []struct {
			GroupID string `json:"groupid"`
			Name    string `json:"name"`
		}
		if err := json.Unmarshal(resp.Result, &result); err == nil && len(result) > 0 {
			return result[0].GroupID, nil
		}
	}
	createParams := map[string]interface{}{
		"name": name,
	}
	createResp, err := p.sendRequest(ctx, "hostgroup.create", createParams)
	if err != nil {
		return "", fmt.Errorf("failed to create host group: %w", err)
	}
	var createResult struct {
		GroupIDs []string `json:"groupids"`
	}
	if err := json.Unmarshal(createResp.Result, &createResult); err != nil {
		return "", fmt.Errorf("failed to parse create host group response: %w", err)
	}
	if len(createResult.GroupIDs) == 0 {
		return "", fmt.Errorf("no host group ID returned after creation")
	}
	return createResult.GroupIDs[0], nil
}

func toString(value interface{}) string {
	if value == nil {
		return "0"
	}
	switch v := value.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	case float64:
		return strconv.FormatInt(int64(v), 10)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func toInt(value interface{}, fallback int) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case json.Number:
		if parsed, err := strconv.Atoi(v.String()); err == nil {
			return parsed
		}
	case string:
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return fallback
}
