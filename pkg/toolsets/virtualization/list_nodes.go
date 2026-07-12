package virtualization

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/harvester/harvester-mcp-server/internal/middleware"
	"github.com/harvester/harvester-mcp-server/pkg/client"
	"github.com/harvester/harvester-mcp-server/pkg/response"
	"go.uber.org/zap"
)

// listNodesParams specifies the parameters needed to list nodes/hosts.
type listNodesParams struct {
	Cluster string `json:"cluster" jsonschema:"the name or ID of the Harvester cluster to query"`
}

// listNodes retrieves all nodes (hosts) in a Harvester / SUSE Virtualization cluster,
// along with their resource utilization metrics when available.
func (t *Tools) listNodes(ctx context.Context, toolReq *mcp.CallToolRequest, params listNodesParams) (*mcp.CallToolResult, any, error) {
	zap.L().Debug("listNodes called", zap.String("cluster", params.Cluster))

	nodes, err := t.client.GetResources(ctx, client.ListParams{
		Cluster: params.Cluster,
		Kind:    "node",
		Token:   middleware.Token(ctx),
	})
	if err != nil {
		zap.L().Error("failed to list nodes", zap.String("tool", "listNodes"), zap.Error(err))
		return nil, nil, err
	}

	// ignore error as Metrics Server might not be installed in the cluster
	nodeMetrics, _ := t.client.GetResources(ctx, client.ListParams{
		Cluster: params.Cluster,
		Kind:    "node.metrics.k8s.io",
		Token:   middleware.Token(ctx),
	})

	mcpResponse, err := response.CreateMcpResponse(append(nodes, nodeMetrics...), params.Cluster)
	if err != nil {
		zap.L().Error("failed to create mcp response", zap.String("tool", "listNodes"), zap.Error(err))
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: mcpResponse}},
	}, nil, nil
}
