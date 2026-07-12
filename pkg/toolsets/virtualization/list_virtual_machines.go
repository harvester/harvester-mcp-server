package virtualization

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/harvester/harvester-mcp-server/internal/middleware"
	"github.com/harvester/harvester-mcp-server/pkg/client"
	"github.com/harvester/harvester-mcp-server/pkg/response"
	"go.uber.org/zap"
)

// listVirtualMachinesParams specifies the parameters needed to list VirtualMachine resources.
type listVirtualMachinesParams struct {
	Cluster   string `json:"cluster" jsonschema:"the name or ID of the Harvester cluster to query"`
	Namespace string `json:"namespace,omitempty" jsonschema:"the namespace to filter VMs; omit to list VMs across all namespaces"`
}

// listVirtualMachines retrieves all KubeVirt VirtualMachine resources from a Harvester /
// SUSE Virtualization cluster, optionally filtered by namespace.
func (t *Tools) listVirtualMachines(ctx context.Context, toolReq *mcp.CallToolRequest, params listVirtualMachinesParams) (*mcp.CallToolResult, any, error) {
	zap.L().Debug("listVirtualMachines called", zap.String("cluster", params.Cluster), zap.String("namespace", params.Namespace))

	vms, err := t.client.GetResources(ctx, client.ListParams{
		Cluster:   params.Cluster,
		Kind:      "virtualmachine",
		Namespace: params.Namespace,
		Token:     middleware.Token(ctx),
	})
	if err != nil {
		zap.L().Error("failed to list virtual machines", zap.String("tool", "listVirtualMachines"), zap.Error(err))
		return nil, nil, err
	}

	mcpResponse, err := response.CreateMcpResponse(vms, params.Cluster)
	if err != nil {
		zap.L().Error("failed to create mcp response", zap.String("tool", "listVirtualMachines"), zap.Error(err))
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: mcpResponse}},
	}, nil, nil
}
