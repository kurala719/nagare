package service

import (
	"testing"
)

func TestListTools(t *testing.T) {
	tools := ListTools()

	if len(tools) == 0 {
		t.Error("ListTools() should return a non-empty list of tools, but got an empty list")
	}

	for _, tool := range tools {
		t.Run(tool.Name, func(t *testing.T) {
			if tool.Name == "" {
				t.Error("ToolDefinition should have a name")
			}
			if tool.Description == "" {
				t.Error("ToolDefinition should have a description")
			}
			if tool.InputSchema == nil {
				t.Error("ToolDefinition should have an input schema")
			}
		})
	}
}
