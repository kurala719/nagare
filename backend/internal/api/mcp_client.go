package api

import (
	"net/http"

	"nagare/internal/mcp"
	"nagare/internal/repository"

	"github.com/gin-gonic/gin"
)

// GetMCPClientStatusCtrl returns runtime status for configured MCP clients.
func GetMCPClientStatusCtrl(c *gin.Context) {
	statuses := mcp.GetClientStatuses()
	connected := 0
	for _, st := range statuses {
		if st.Connected {
			connected++
		}
	}

	respondSuccess(c, http.StatusOK, gin.H{
		"items":           statuses,
		"total":           len(statuses),
		"connected_total": connected,
	})
}

// TestMCPClientCtrl tests a single MCP server definition without persisting config.
func TestMCPClientCtrl(c *gin.Context) {
	var req repository.MCPServerConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	res := mcp.TestServerConfig(req)
	if !res.Connected {
		respondSuccess(c, http.StatusOK, gin.H{
			"ok":      false,
			"message": res.Error,
			"result":  res,
		})
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{
		"ok":      true,
		"message": "connection successful",
		"result":  res,
	})
}
