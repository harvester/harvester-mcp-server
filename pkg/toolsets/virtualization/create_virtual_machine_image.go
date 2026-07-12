package virtualization

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/harvester/harvester-mcp-server/internal/middleware"
	"github.com/harvester/harvester-mcp-server/pkg/converter"
	"github.com/harvester/harvester-mcp-server/pkg/response"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	defaultStorageClass = "harvester-longhorn"
	defaultNamespace    = "default"
)

// createVirtualMachineImageParams defines the parameters for creating a VirtualMachineImage.
type createVirtualMachineImageParams struct {
	Cluster     string `json:"cluster" jsonschema:"the name or ID of the Harvester cluster"`
	Namespace   string `json:"namespace,omitempty" jsonschema:"the namespace to create the image in; defaults to 'default'"`
	DisplayName string `json:"displayName" jsonschema:"human-readable name for the image"`
	URL         string `json:"url" jsonschema:"the download URL of the image (qcow2 or raw format)"`
	Description string `json:"description,omitempty" jsonschema:"optional description of the image"`
	Checksum    string `json:"checksum,omitempty" jsonschema:"optional SHA512 checksum to verify the downloaded image"`
}

// createVirtualMachineImage creates a new VirtualMachineImage resource on a Harvester cluster.
func (t *Tools) createVirtualMachineImage(ctx context.Context, toolReq *mcp.CallToolRequest, params createVirtualMachineImageParams) (*mcp.CallToolResult, any, error) {
	zap.L().Debug("createVirtualMachineImage called",
		zap.String("cluster", params.Cluster),
		zap.String("displayName", params.DisplayName))

	ns := params.Namespace
	if ns == "" {
		ns = defaultNamespace
	}

	obj := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "harvesterhci.io/v1beta1",
			"kind":       "VirtualMachineImage",
			"metadata": map[string]any{
				"generateName": "image-",
				"namespace":    ns,
				"annotations": map[string]any{
					"harvesterhci.io/storageClassName": defaultStorageClass,
				},
				"labels": map[string]any{
					"harvesterhci.io/os-type":    "",
					"harvesterhci.io/image-type": "raw_qcow2",
				},
			},
			"spec": map[string]any{
				"backend":                "backingimage",
				"displayName":            params.DisplayName,
				"sourceType":             "download",
				"url":                    params.URL,
				"targetStorageClassName": defaultStorageClass,
				"checksum":               params.Checksum,
				"description":            params.Description,
			},
		},
	}

	gvr := converter.K8sKindsToGVRs["virtualmachineimage"]
	resourceInterface, err := t.client.GetResourceInterface(
		ctx, middleware.Token(ctx),
		ns, params.Cluster, gvr)
	if err != nil {
		zap.L().Error("failed to get resource interface", zap.String("tool", "createVirtualMachineImage"), zap.Error(err))
		return nil, nil, fmt.Errorf("failed to get resource interface: %w", err)
	}

	created, err := resourceInterface.Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		zap.L().Error("failed to create VirtualMachineImage", zap.String("tool", "createVirtualMachineImage"), zap.Error(err))
		return nil, nil, fmt.Errorf("failed to create VirtualMachineImage: %w", err)
	}

	mcpResponse, err := response.CreateMcpResponse([]*unstructured.Unstructured{created}, params.Cluster)
	if err != nil {
		zap.L().Error("failed to create mcp response", zap.String("tool", "createVirtualMachineImage"), zap.Error(err))
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: mcpResponse}},
	}, nil, nil
}
