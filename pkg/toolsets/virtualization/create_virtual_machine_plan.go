package virtualization

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/harvester/harvester-mcp-server/pkg/response"
	"go.uber.org/zap"
)

// createVirtualMachinePlan plans the creation of a VirtualMachine.
// It returns the full JSON manifest that would be submitted without actually creating the resource.
func (t *Tools) createVirtualMachinePlan(ctx context.Context, toolReq *mcp.CallToolRequest, params createVirtualMachineParams) (*mcp.CallToolResult, any, error) {
	zap.L().Debug("createVirtualMachine_plan called",
		zap.String("cluster", params.Cluster),
		zap.String("name", params.Name))

	obj, err := t.createVirtualMachineObj(ctx, toolReq, params)
	if err != nil {
		zap.L().Error("failed to build VM object", zap.String("tool", "createVirtualMachine_plan"), zap.Error(err))
		return nil, nil, fmt.Errorf("failed to build VM object: %w", err)
	}

	createResource := response.NewCreateResourceInput(obj, params.Cluster)
	mcpResponse, err := response.CreatePlanResponse([]response.PlanResource{createResource})
	if err != nil {
		zap.L().Error("failed to create plan response", zap.String("tool", "createVirtualMachine_plan"), zap.Error(err))
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: mcpResponse}},
	}, nil, nil
}
