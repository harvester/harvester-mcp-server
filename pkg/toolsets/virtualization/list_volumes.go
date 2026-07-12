package virtualization

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/harvester/harvester-mcp-server/internal/middleware"
	"github.com/harvester/harvester-mcp-server/pkg/client"
	"github.com/harvester/harvester-mcp-server/pkg/response"
	"go.uber.org/zap"
)

// listVolumesParams specifies the parameters needed to list volume (PVC) resources.
type listVolumesParams struct {
	Cluster   string `json:"cluster" jsonschema:"the name or ID of the Harvester cluster to query"`
	Namespace string `json:"namespace,omitempty" jsonschema:"the namespace to filter volumes; omit to list volumes across all namespaces"`
}

// listVolumes retrieves all PersistentVolumeClaim resources from a Harvester / SUSE Virtualization
// cluster, optionally filtered by namespace.
func (t *Tools) listVolumes(ctx context.Context, toolReq *mcp.CallToolRequest, params listVolumesParams) (*mcp.CallToolResult, any, error) {
	zap.L().Debug("listVolumes called", zap.String("cluster", params.Cluster), zap.String("namespace", params.Namespace))

	volumes, err := t.client.GetResources(ctx, client.ListParams{
		Cluster:   params.Cluster,
		Kind:      "persistentvolumeclaim",
		Namespace: params.Namespace,
		Token:     middleware.Token(ctx),
	})
	if err != nil {
		zap.L().Error("failed to list volumes", zap.String("tool", "listVolumes"), zap.Error(err))
		return nil, nil, err
	}

	mcpResponse, err := response.CreateMcpResponse(volumes, params.Cluster)
	if err != nil {
		zap.L().Error("failed to create mcp response", zap.String("tool", "listVolumes"), zap.Error(err))
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: mcpResponse}},
	}, nil, nil
}
