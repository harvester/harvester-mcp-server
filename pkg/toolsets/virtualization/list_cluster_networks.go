package virtualization

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/harvester/harvester-mcp-server/internal/middleware"
	"github.com/harvester/harvester-mcp-server/pkg/client"
	"github.com/harvester/harvester-mcp-server/pkg/response"
	"go.uber.org/zap"
)

// listClusterNetworksParams specifies the parameters needed to list ClusterNetwork resources.
type listClusterNetworksParams struct {
	Cluster string `json:"cluster" jsonschema:"the name or ID of the Harvester cluster to query"`
}

// listClusterNetworks retrieves all ClusterNetwork resources from a Harvester / SUSE Virtualization
// cluster. ClusterNetworks are cluster-scoped (not namespaced).
func (t *Tools) listClusterNetworks(ctx context.Context, toolReq *mcp.CallToolRequest, params listClusterNetworksParams) (*mcp.CallToolResult, any, error) {
	zap.L().Debug("listClusterNetworks called", zap.String("cluster", params.Cluster))

	networks, err := t.client.GetResources(ctx, client.ListParams{
		Cluster: params.Cluster,
		Kind:    "clusternetwork",
		Token:   middleware.Token(ctx),
	})
	if err != nil {
		zap.L().Error("failed to list cluster networks", zap.String("tool", "listClusterNetworks"), zap.Error(err))
		return nil, nil, err
	}

	mcpResponse, err := response.CreateMcpResponse(networks, params.Cluster)
	if err != nil {
		zap.L().Error("failed to create mcp response", zap.String("tool", "listClusterNetworks"), zap.Error(err))
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: mcpResponse}},
	}, nil, nil
}
