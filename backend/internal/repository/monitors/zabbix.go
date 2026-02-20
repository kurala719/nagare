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
	HostID          string `json:"hostid"`
	Host            string `json:"host"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Status          string `json:"status"`
	ActiveAvailable string `json:"active_available"`
	Interfaces      []struct {
		InterfaceID string `json:"interfaceid"`
		IP          string `json:"ip"`
		DNS         string `json:"dns"`
		Port        string `json:"port"`
		Type        string `json:"type"`
		Main        string `json:"main"`
		UseIP       string `json:"useip"`
	} `json:"interfaces"`
}

type zabbixItem struct {
	ItemID    string `json:"itemid"`
	HostID    string `json:"hostid"`
	Name      string `json:"name"`
	Key       string `json:"key_"`
	LastValue string `json:"lastvalue"`
	Units     string `json:"units"`
	ValueType string `json:"value_type"`
	Type      string `json:"type"`
	Delay     string `json:"delay"`
	Desc      string `json:"description"`
	Status    string `json:"status"`
	LastClock string `json:"lastclock"`
}

type zabbixProblem struct {
	EventID      string `json:"eventid"`
	ObjectID     string `json:"objectid"`
	Name         string `json:"name"`
	Severity     string `json:"severity"`
	Acknowledged string `json:"acknowledged"`
	Clock        string `json:"clock"`
}

type zabbixTrigger struct {
	TriggerID   string `json:"triggerid"`
	Description string `json:"description"`
	Expression  string `json:"expression"`
	Priority    string `json:"priority"`
	Status      string `json:"status"`
	Value       string `json:"value"`
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
		return fmt.Errorf("failed to parse auth token: %w", err)
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
		"output":           []string{"hostid", "host", "name", "description", "status", "active_available"},
		"selectInterfaces": []string{"interfaceid", "ip", "dns", "port", "type", "main", "useip"},
		"selectHostGroups": "extend",
		"selectGroups":     "extend",
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
		"output":           []string{"hostid", "host", "name", "description", "status", "active_available"},
		"selectInterfaces": []string{"interfaceid", "ip", "dns", "port", "type", "main", "useip"},
		"selectHostGroups": "extend",
		"selectGroups":     "extend",
		"groupids":         groupID,
	}

	resp, err := p.sendRequest(ctx, "host.get", params)
	if err != nil {
		return nil, err
	}

	return p.parseZabbixHosts(resp.Result)
}

// parseZabbixHosts parses Zabbix host.get result into common Host slice
func (p *ZabbixProvider) parseZabbixHosts(result json.RawMessage) ([]Host, error) {
	// Define a custom struct to handle groups since standard zabbixHost struct doesn't have it
	var zabbixHosts []struct {
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
	if err := json.Unmarshal(result, &zabbixHosts); err != nil {
		return nil, fmt.Errorf("failed to parse hosts: %w", err)
	}

	hosts := make([]Host, 0, len(zabbixHosts))
	for _, zh := range zabbixHosts {
		ip := ""
		if len(zh.Interfaces) > 0 {
			ip = zh.Interfaces[0].IP
		}

		status := "unknown"
		switch zh.Status {
		case "0":
			status = "up"
		case "1":
			status = "down"
		}

		metadata := map[string]string{
			"host":             zh.Host,
			"active_available": zh.ActiveAvailable,
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

		hosts = append(hosts, Host{
			ID:          zh.HostID,
			Name:        zh.Name,
			Description: zh.Description,
			Status:      status,
			IPAddress:   ip,
			Metadata:    metadata,
		})
	}

	return hosts, nil
}

// GetHostByName implements the Provider interface
func (p *ZabbixProvider) GetHostByName(ctx context.Context, name string) (*Host, error) {
	params := map[string]interface{}{
		"output":           []string{"hostid", "host", "name", "description", "status", "active_available"},
		"selectInterfaces": []string{"interfaceid", "ip", "dns", "port", "type", "main", "useip"},
		"selectHostGroups": "extend",
		"selectGroups":     "extend",
		"filter": map[string]interface{}{
			"host": []string{name},
		},
	}

	resp, err := p.sendRequest(ctx, "host.get", params)
	if err != nil {
		return nil, err
	}

	var zabbixHosts []struct {
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
	if err := json.Unmarshal(resp.Result, &zabbixHosts); err != nil {
		return nil, fmt.Errorf("failed to parse host: %w", err)
	}

	if len(zabbixHosts) == 0 {
		return nil, fmt.Errorf("host not found: %s", name)
	}

	zh := zabbixHosts[0]
	ip := ""
	if len(zh.Interfaces) > 0 {
		ip = zh.Interfaces[0].IP
	}

	status := "unknown"
	switch zh.Status {
	case "0":
		status = "up"
	case "1":
		status = "down"
	}

	metadata := map[string]string{
		"host":             zh.Host,
		"active_available": zh.ActiveAvailable,
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

	return &Host{
		ID:          zh.HostID,
		Name:        zh.Name,
		Description: zh.Description,
		Status:      status,
		IPAddress:   ip,
		Metadata:    metadata,
	}, nil
}

// GetHostByID implements the Provider interface
func (p *ZabbixProvider) GetHostByID(ctx context.Context, hostID string) (*Host, error) {
	params := map[string]interface{}{
		"output":           []string{"hostid", "host", "name", "description", "status", "active_available"},
		"selectInterfaces": []string{"interfaceid", "ip", "dns", "port", "type", "main", "useip"},
		"selectHostGroups": "extend",
		"selectGroups":     "extend",
		"hostids":          hostID,
	}

	resp, err := p.sendRequest(ctx, "host.get", params)
	if err != nil {
		return nil, err
	}

	var zabbixHosts []struct {
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
	if err := json.Unmarshal(resp.Result, &zabbixHosts); err != nil {
		return nil, fmt.Errorf("failed to parse host: %w", err)
	}

	if len(zabbixHosts) == 0 {
		return nil, fmt.Errorf("host not found: %s", hostID)
	}

	zh := zabbixHosts[0]
	ip := ""
	if len(zh.Interfaces) > 0 {
		ip = zh.Interfaces[0].IP
	}

	status := "unknown"
	switch zh.Status {
	case "0":
		status = "up"
	case "1":
		status = "down"
	}

	metadata := map[string]string{
		"host":             zh.Host,
		"active_available": zh.ActiveAvailable,
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

	return &Host{
		ID:          zh.HostID,
		Name:        zh.Name,
		Description: zh.Description,
		Status:      status,
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
		items = append(items, Item{
			ID:          zi.ItemID,
			HostID:      zi.HostID,
			Name:        zi.Name,
			Key:         zi.Key,
			Type:        zi.Type,
			Value:       zi.LastValue,
			Units:       zi.Units,
			ValueType:   zi.ValueType,
			Delay:       zi.Delay,
			Status:      zi.Status,
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
	return &Item{
		ID:          zi.ItemID,
		HostID:      zi.HostID,
		Name:        zi.Name,
		Key:         zi.Key,
		Type:        zi.Type,
		Value:       zi.LastValue,
		Units:       zi.Units,
		ValueType:   zi.ValueType,
		Delay:       zi.Delay,
		Status:      zi.Status,
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
		items = append(items, Item{
			ID:        h.ItemID,
			Value:     h.Value,
			HostID:    item.HostID,
			Name:      item.Name,
			Units:     item.Units,
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
		if iface.Main == "1" {
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
		InterfaceID: selected.InterfaceID,
		IP:          selected.IP,
		DNS:         selected.DNS,
		Port:        selected.Port,
		Type:        selected.Type,
		Main:        selected.Main,
		UseIP:       selected.UseIP,
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
		EventID      string `json:"eventid"`
		ObjectID     string `json:"objectid"`
		Name         string `json:"name"`
		Severity     string `json:"severity"`
		Acknowledged string `json:"acknowledged"`
		Clock        string `json:"clock"`
		Hosts        []struct {
			HostID string `json:"hostid"`
		} `json:"hosts"`
	}
	if err := json.Unmarshal(resp.Result, &problems); err != nil {
		return nil, fmt.Errorf("failed to parse problems: %w", err)
	}

	alerts := make([]Alert, 0, len(problems))
	for _, prob := range problems {
		hostID := ""
		if len(prob.Hosts) > 0 {
			hostID = prob.Hosts[0].HostID
		}

		severity := "information"
		switch prob.Severity {
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
		switch prob.Severity {
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
		switch zt.Priority {
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
		if zt.Value == "1" {
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
		switch zt.Priority {
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
		if zt.Value == "1" {
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
	if host.Metadata != nil {
		if gid, ok := host.Metadata["groupid"]; ok && gid != "" {
			groupID = gid
		}
	}

	// Default to Zabbix Agent interface
	interfaces := []map[string]interface{}{
		{
			"type":  1, // Agent
			"main":  1,
			"useip": 1,
			"ip":    host.IPAddress,
			"dns":   "",
			"port":  "10050",
		},
	}

	// Prepare templates
	var templates []map[string]string

	// Try to use template from metadata
	if host.Metadata != nil {
		if tid, ok := host.Metadata["templateid"]; ok && tid != "" {
			templates = append(templates, map[string]string{"templateid": tid})
		}
	}

	// If no template specified, try to find a default one (optional)
	// For now we don't force a template if not provided to avoid errors

	params := map[string]interface{}{
		"host":       host.Name,
		"interfaces": interfaces,
		"groups":     []map[string]string{{"groupid": groupID}},
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
	params := map[string]interface{}{
		"hostid":      host.ID,
		"host":        host.Name,
		"name":        host.Name,
		"description": host.Description,
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
			if iface.Type == "2" {
				ifaceParams["details"] = map[string]interface{}{
					"version":   2,
					"community": "{$SNMP_COMMUNITY}",
				}
			}
			params["interfaces"] = []map[string]interface{}{ifaceParams}
		} else {
			// If no interface found (unlikely for existing host), create one
			// or just ignore. Logic below handles update.
			// Ideally we should try to create one if missing, but let's stick to update logic.
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
