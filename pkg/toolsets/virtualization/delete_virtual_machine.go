package virtualization

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/harvester/harvester-mcp-server/internal/middleware"
	"github.com/harvester/harvester-mcp-server/pkg/client"
	"github.com/harvester/harvester-mcp-server/pkg/converter"
	"github.com/harvester/harvester-mcp-server/pkg/response"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deleteVirtualMachineParams struct {
	Cluster   string `json:"cluster" jsonschema:"the name or ID of the Harvester cluster"`
	Namespace string `json:"namespace,omitempty" jsonschema:"namespace of the virtual machine; defaults to 'default'"`
	Name      string `json:"name" jsonschema:"name of the virtual machine to delete"`
}

func (t *Tools) deleteVirtualMachine(ctx context.Context, toolReq *mcp.CallToolRequest, params deleteVirtualMachineParams) (*mcp.CallToolResult, any, error) {
	zap.L().Debug("deleteVirtualMachine called",
		zap.String("cluster", params.Cluster),
		zap.String("namespace", params.Namespace),
		zap.String("name", params.Name))

	ns := params.Namespace
	if ns == "" {
		ns = defaultNamespace
	}

	token := middleware.Token(ctx)

	gvr := converter.K8sKindsToGVRs["virtualmachine"]
	resourceInterface, err := t.client.GetResourceInterface(ctx, token, ns, params.Cluster, gvr)
	if err != nil {
		zap.L().Error("failed to get resource interface", zap.String("tool", "deleteVirtualMachine"), zap.Error(err))
		return nil, nil, fmt.Errorf("failed to get resource interface: %w", err)
	}

	if err := resourceInterface.Delete(ctx, params.Name, metav1.DeleteOptions{}); err != nil {
		zap.L().Error("failed to delete VirtualMachine", zap.String("tool", "deleteVirtualMachine"), zap.Error(err))
		return nil, nil, fmt.Errorf("failed to delete VirtualMachine %s/%s: %w", ns, params.Name, err)
	}

	mcpResponse, err := response.CreateMcpResponseAny(fmt.Sprintf("VirtualMachine %s/%s deleted successfully", ns, params.Name))
	if err != nil {
		zap.L().Error("failed to create mcp response", zap.String("tool", "deleteVirtualMachine"), zap.Error(err))
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: mcpResponse}},
	}, nil, nil
}

func (t *Tools) deleteVirtualMachinePlan(ctx context.Context, toolReq *mcp.CallToolRequest, params deleteVirtualMachineParams) (*mcp.CallToolResult, any, error) {
	zap.L().Debug("deleteVirtualMachine_plan called",
		zap.String("cluster", params.Cluster),
		zap.String("namespace", params.Namespace),
		zap.String("name", params.Name))

	ns := params.Namespace
	if ns == "" {
		ns = defaultNamespace
	}

	vmObj, err := t.client.GetResource(ctx, client.GetParams{
		Cluster:   params.Cluster,
		Kind:      "virtualmachine",
		Namespace: ns,
		Name:      params.Name,
		Token:     middleware.Token(ctx),
	})
	if err != nil {
		zap.L().Error("failed to get VirtualMachine", zap.String("tool", "deleteVirtualMachine_plan"), zap.Error(err))
		return nil, nil, fmt.Errorf("failed to get VirtualMachine %s/%s: %w", ns, params.Name, err)
	}

	planResource := response.PlanResource{
		Type: response.OperationDelete,
		Resource: response.Resource{
			Name:      vmObj.GetName(),
			Kind:      vmObj.GetKind(),
			Cluster:   params.Cluster,
			Namespace: vmObj.GetNamespace(),
		},
		Payload: vmObj,
	}

	mcpResponse, err := response.CreatePlanResponse([]response.PlanResource{planResource})
	if err != nil {
		zap.L().Error("failed to create plan response", zap.String("tool", "deleteVirtualMachine_plan"), zap.Error(err))
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: mcpResponse}},
	}, nil, nil
}
