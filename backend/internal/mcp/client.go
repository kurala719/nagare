package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// ExternalTool represents a tool provided by an external MCP server
type ExternalTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// Client represents a connection to an external MCP server via stdio
type Client struct {
	Name      string
	mcpClient *client.Client
	tools     []ExternalTool
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewClient creates a new MCP client using the official SDK
func NewClient(name, command string, args []string, customEnv map[string]string) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())
	envArray := buildMCPEnv(customEnv)

	// Create stdio transport
	stdioTransport := transport.NewStdio(command, envArray, args...)

	// Create the client
	mcpClient := client.NewClient(stdioTransport)

	// Start it
	if err := mcpClient.Start(ctx); err != nil {
		cancel()
		return nil, err
	}

	return &Client{
		Name:      name,
		mcpClient: mcpClient,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

func buildMCPEnv(customEnv map[string]string) []string {
	var envList []string

	// Process custom env vars first
	for k, v := range customEnv {
		if strings.TrimSpace(k) != "" {
			envList = append(envList, fmt.Sprintf("%s=%s", strings.TrimSpace(k), v))
		}
	}

	if len(envList) == 0 {
		return nil
	}
	return envList
}

// Close terminates the client
func (c *Client) Close() {
	if c.mcpClient != nil {
		c.mcpClient.Close()
	}
	c.cancel()
}

// Initialize performs the MCP handshake
func (c *Client) Initialize() error {
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "nagare",
		Version: "1.0",
	}

	_, err := c.mcpClient.Initialize(c.ctx, initRequest)
	return err
}

// LoadTools fetches tools from the server and caches them with a prefix
func (c *Client) LoadTools() error {
	req := mcp.ListToolsRequest{}
	res, err := c.mcpClient.ListTools(c.ctx, req)
	if err != nil {
		return err
	}

	c.tools = make([]ExternalTool, 0, len(res.Tools))
	for _, t := range res.Tools {
		var schema map[string]interface{}

		// Try standard input schema first
		schemaBytes, err := json.Marshal(t.InputSchema)
		if err == nil {
			json.Unmarshal(schemaBytes, &schema)
		} else if len(t.RawInputSchema) > 0 {
			// Fallback to raw schema
			json.Unmarshal(t.RawInputSchema, &schema)
		}

		prefixedName := c.Name + "_" + t.Name

		c.tools = append(c.tools, ExternalTool{
			Name:        prefixedName,
			Description: fmt.Sprintf("[%s] %s", c.Name, t.Description),
			InputSchema: schema,
		})
	}

	return nil
}

// CallTool calls a prefixed tool on the remote server
func (c *Client) CallTool(prefixedName string, args map[string]interface{}) (interface{}, error) {
	// Strip the prefix to get the original name
	originalName := prefixedName[len(c.Name)+1:]

	req := mcp.CallToolRequest{}
	req.Params.Name = originalName
	req.Params.Arguments = args

	resp, err := c.mcpClient.CallTool(c.ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.IsError {
		// Log or extract the error message from the response
		if len(resp.Content) > 0 {
			if textContent, ok := resp.Content[0].(mcp.TextContent); ok {
				return nil, fmt.Errorf("tool error: %s", textContent.Text)
			}
		}
		return nil, fmt.Errorf("tool execution failed on target server")
	}

	if len(resp.Content) > 0 {
		if textContent, ok := resp.Content[0].(mcp.TextContent); ok {
			return textContent.Text, nil
		}
		// If it's an image or other representation, we fallback to json encoding to pass it to our LLM layer.
		b, _ := json.Marshal(resp.Content[0])
		return string(b), nil
	}

	return "No content returned", nil
}

// GetTools returns the cached tools
func (c *Client) GetTools() []ExternalTool {
	return c.tools
}
