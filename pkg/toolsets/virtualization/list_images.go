package virtualization

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/harvester/harvester-mcp-server/internal/middleware"
	"github.com/harvester/harvester-mcp-server/pkg/client"
	"github.com/harvester/harvester-mcp-server/pkg/response"
	"go.uber.org/zap"
)

// listVirtualMachineImagesParams specifies the parameters needed to list VirtualMachineImage resources.
type listVirtualMachineImagesParams struct {
	Cluster   string `json:"cluster" jsonschema:"the name or ID of the Harvester cluster to query"`
	Namespace string `json:"namespace,omitempty" jsonschema:"the namespace to filter images; omit to list images across all namespaces"`
}

// listVirtualMachineImages retrieves all Harvester VirtualMachineImage resources from a SUSE Virtualization
// cluster, optionally filtered by namespace.
func (t *Tools) listVirtualMachineImages(ctx context.Context, toolReq *mcp.CallToolRequest, params listVirtualMachineImagesParams) (*mcp.CallToolResult, any, error) {
	zap.L().Debug("listVirtualMachineImages called", zap.String("cluster", params.Cluster), zap.String("namespace", params.Namespace))

	images, err := t.client.GetResources(ctx, client.ListParams{
		Cluster:   params.Cluster,
		Kind:      "virtualmachineimage",
		Namespace: params.Namespace,
		Token:     middleware.Token(ctx),
	})
	if err != nil {
		zap.L().Error("failed to list virtual machine images", zap.String("tool", "listVirtualMachineImages"), zap.Error(err))
		return nil, nil, err
	}

	mcpResponse, err := response.CreateMcpResponse(images, params.Cluster)
	if err != nil {
		zap.L().Error("failed to create mcp response", zap.String("tool", "listVirtualMachineImages"), zap.Error(err))
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: mcpResponse}},
	}, nil, nil
}
