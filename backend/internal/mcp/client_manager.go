package mcp

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"nagare/internal/repository"
	"nagare/internal/service"
)

var (
	clients                    map[string]*Client
	clientsMu                  sync.RWMutex
	clientStatuses             []ClientStatus
	registerConfigObserverOnce sync.Once
)

// ClientStatus provides runtime visibility for configured MCP clients.
type ClientStatus struct {
	Name      string   `json:"name"`
	Enabled   bool     `json:"enabled"`
	Command   string   `json:"command"`
	Args      []string `json:"args"`
	Connected bool     `json:"connected"`
	ToolCount int      `json:"tool_count"`
	LastError string   `json:"last_error,omitempty"`
	UpdatedAt string   `json:"updated_at"`
}

// ClientTestResult contains connectivity test output for an MCP server config.
type ClientTestResult struct {
	Connected bool     `json:"connected"`
	ToolCount int      `json:"tool_count"`
	ToolNames []string `json:"tool_names"`
	Error     string   `json:"error,omitempty"`
}

func init() {
	clients = make(map[string]*Client)
	clientStatuses = make([]ClientStatus, 0)
}

// InitClients reads configuration and connects to external MCP servers
func InitClients() {
	registerConfigObserverOnce.Do(func() {
		repository.RegisterConfigObserver(func() {
			go InitClients()
		})
	})

	config, err := repository.LoadMCPConfig()
	if err != nil {
		log.Printf("Failed to load generic config for MCP clients: %v", err)
		return
	}

	clientsMu.Lock()
	defer clientsMu.Unlock()

	// Close existing clients
	for _, c := range clients {
		c.Close()
	}
	clients = make(map[string]*Client)
	clientStatuses = make([]ClientStatus, 0, len(config))

	for _, srv := range config {
		status := ClientStatus{
			Name:      srv.Name,
			Enabled:   srv.Enabled,
			Command:   srv.Command,
			Args:      append([]string(nil), srv.Args...),
			Connected: false,
			ToolCount: 0,
			UpdatedAt: time.Now().Format(time.RFC3339),
		}

		if !srv.Enabled {
			status.LastError = "disabled"
			clientStatuses = append(clientStatuses, status)
			continue
		}

		if srv.Command == "" {
			status.LastError = "missing command"
			clientStatuses = append(clientStatuses, status)
			continue
		}

		client, err := NewClient(srv.Name, srv.Command, srv.Args, srv.Env)
		if err != nil {
			log.Printf("Failed to start MCP server %s: %v", srv.Name, err)
			status.LastError = err.Error()
			clientStatuses = append(clientStatuses, status)
			continue
		}

		if err := client.Initialize(); err != nil {
			log.Printf("Failed to initialize MCP server %s: %v", srv.Name, err)
			client.Close()
			status.LastError = err.Error()
			clientStatuses = append(clientStatuses, status)
			continue
		}

		if err := client.LoadTools(); err != nil {
			log.Printf("Failed to load tools from MCP server %s: %v", srv.Name, err)
			client.Close()
			status.LastError = err.Error()
			clientStatuses = append(clientStatuses, status)
			continue
		}

		clients[srv.Name] = client
		status.Connected = true
		status.ToolCount = len(client.GetTools())
		status.LastError = ""
		status.UpdatedAt = time.Now().Format(time.RFC3339)
		clientStatuses = append(clientStatuses, status)
		log.Printf("Successfully connected to external MCP server: %s (%d tools loaded)", srv.Name, len(client.GetTools()))
	}

	// Inject tool provider into `service` layer
	service.ExternalToolsProvider = func() []service.ToolDefinition {
		extTools := GetAllExternalTools()
		var res []service.ToolDefinition
		for _, t := range extTools {
			res = append(res, service.ToolDefinition{
				Name:        t.Name,
				Description: t.Description,
				InputSchema: t.InputSchema,
			})
		}
		return res
	}

	service.ExternalToolCaller = CallExternalTool
}

// GetClientStatuses returns a snapshot of configured MCP client statuses.
func GetClientStatuses() []ClientStatus {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	res := make([]ClientStatus, len(clientStatuses))
	copy(res, clientStatuses)
	return res
}

// TestServerConfig validates a single MCP server config by opening a temporary connection.
func TestServerConfig(cfg repository.MCPServerConfig) ClientTestResult {
	if cfg.Command == "" {
		return ClientTestResult{Connected: false, Error: "command is required"}
	}

	name := cfg.Name
	if name == "" {
		name = "test"
	}

	client, err := NewClient(name, cfg.Command, cfg.Args, cfg.Env)
	if err != nil {
		return ClientTestResult{Connected: false, Error: err.Error()}
	}
	defer client.Close()

	if err := client.Initialize(); err != nil {
		return ClientTestResult{Connected: false, Error: err.Error()}
	}

	if err := client.LoadTools(); err != nil {
		return ClientTestResult{Connected: false, Error: err.Error()}
	}

	tools := client.GetTools()
	toolNames := make([]string, 0, len(tools))
	for _, tool := range tools {
		toolNames = append(toolNames, tool.Name)
	}

	return ClientTestResult{
		Connected: true,
		ToolCount: len(tools),
		ToolNames: toolNames,
	}
}

// GetAllExternalTools returns all tools gathered from all external clients
func GetAllExternalTools() []ExternalTool {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	var allTools []ExternalTool
	for _, c := range clients {
		allTools = append(allTools, c.GetTools()...)
	}
	return allTools
}

// CallExternalTool routes a tool call to the appropriate external client
func CallExternalTool(prefixedName string, args json.RawMessage) (interface{}, error) {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	var targetClient *Client
	for name, c := range clients {
		// The tool name was prefixed with `Name_`
		if len(prefixedName) > len(name) && prefixedName[:len(name)+1] == name+"_" {
			targetClient = c
			break
		}
	}

	if targetClient == nil {
		return nil, fmt.Errorf("external tool %s not found", prefixedName)
	}

	var parsedArgs map[string]interface{}
	if err := json.Unmarshal(args, &parsedArgs); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	return targetClient.CallTool(prefixedName, parsedArgs)
}
