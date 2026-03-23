package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// MCPServerConfig holds configuration for external MCP servers
type MCPServerConfig struct {
	Name      string   `json:"name"`
	Command   string   `json:"command"`
	Args      []string `json:"args"`
	Enabled   bool              `json:"enabled"`
	Env       map[string]string `json:"env"`
}

// MCPConfigData represents the structure of mcp_config.json
type MCPConfigData struct {
	Servers []MCPServerConfig `json:"servers"`
}

var (
	mcpConfigPath  string = "configs/mcp_config.json"
	mcpConfigMutex sync.RWMutex
)

// SetMCPConfigPath overrides the default config path for testing or custom environments.
func SetMCPConfigPath(path string) {
	mcpConfigPath = path
}

// LoadMCPConfig reads the MCP configuration from disk.
func LoadMCPConfig() ([]MCPServerConfig, error) {
	mcpConfigMutex.RLock()
	defer mcpConfigMutex.RUnlock()

	data, err := os.ReadFile(mcpConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If file doesn't exist, return empty config.
			return []MCPServerConfig{}, nil
		}
		return nil, err
	}

	var configData MCPConfigData
	if err := json.Unmarshal(data, &configData); err != nil {
		return nil, err
	}
	if configData.Servers == nil {
		return []MCPServerConfig{}, nil
	}
	return configData.Servers, nil
}

// SaveMCPConfig writes the MCP configuration to disk.
func SaveMCPConfig(servers []MCPServerConfig) error {
	mcpConfigMutex.Lock()
	defer mcpConfigMutex.Unlock()

	// Ensure the directory exists
	dir := filepath.Dir(mcpConfigPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if servers == nil {
		servers = []MCPServerConfig{}
	}

	configData := MCPConfigData{Servers: servers}
	data, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(mcpConfigPath, data, 0644)
}
