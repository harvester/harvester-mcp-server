package toolsets

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/harvester/harvester-mcp-server/pkg/client"
	"github.com/harvester/harvester-mcp-server/pkg/toolsets/virtualization"
)

// toolsAdder is an interface for types that can add tools to an MCP server.
type toolsAdder interface {
	AddTools(mcpServer *mcp.Server)
}

// AddAllTools adds all available tools to the MCP server.
func AddAllTools(client *client.Client, mcpServer *mcp.Server, readOnly bool) {
	for _, ta := range allToolSets(client, readOnly) {
		ta.AddTools(mcpServer)
	}
}

func allToolSets(client *client.Client, readOnly bool) []toolsAdder {
	return []toolsAdder{
		virtualization.NewTools(client, readOnly),
	}
}
