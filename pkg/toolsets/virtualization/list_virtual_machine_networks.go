package virtualization

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/harvester/harvester-mcp-server/internal/middleware"
	"github.com/harvester/harvester-mcp-server/pkg/client"
	"github.com/harvester/harvester-mcp-server/pkg/response"
	"go.uber.org/zap"
)

// listVirtualMachineNetworksParams specifies the parameters needed to list NetworkAttachmentDefinition resources.
type listVirtualMachineNetworksParams struct {
	Cluster   string `json:"cluster" jsonschema:"the name or ID of the Harvester cluster to query"`
	Namespace string `json:"namespace,omitempty" jsonschema:"the namespace to filter networks; omit to list networks across all namespaces"`
}

// listVirtualMachineNetworks retrieves all NetworkAttachmentDefinition resources from a Harvester /
// SUSE Virtualization cluster, optionally filtered by namespace.
func (t *Tools) listVirtualMachineNetworks(ctx context.Context, toolReq *mcp.CallToolRequest, params listVirtualMachineNetworksParams) (*mcp.CallToolResult, any, error) {
	zap.L().Debug("listVirtualMachineNetworks called", zap.String("cluster", params.Cluster), zap.String("namespace", params.Namespace))

	networks, err := t.client.GetResources(ctx, client.ListParams{
		Cluster:   params.Cluster,
		Kind:      "networkattachmentdefinition",
		Namespace: params.Namespace,
		Token:     middleware.Token(ctx),
	})
	if err != nil {
		zap.L().Error("failed to list virtual machine networks", zap.String("tool", "listVirtualMachineNetworks"), zap.Error(err))
		return nil, nil, err
	}

	mcpResponse, err := response.CreateMcpResponse(networks, params.Cluster)
	if err != nil {
		zap.L().Error("failed to create mcp response", zap.String("tool", "listVirtualMachineNetworks"), zap.Error(err))
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: mcpResponse}},
	}, nil, nil
}
