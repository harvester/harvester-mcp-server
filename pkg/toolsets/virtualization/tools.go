package virtualization

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/harvester/harvester-mcp-server/pkg/client"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

const (
	toolsSet    = "virtualization"
	toolsSetAnn = "toolset"
)

type toolsClient interface {
	GetResource(ctx context.Context, params client.GetParams) (*unstructured.Unstructured, error)
	GetResourceInterface(ctx context.Context, token string, namespace string, cluster string, gvr schema.GroupVersionResource) (dynamic.ResourceInterface, error)
	GetResources(ctx context.Context, params client.ListParams) ([]*unstructured.Unstructured, error)
	CreateClientSet(ctx context.Context, token string, cluster string) (kubernetes.Interface, error)
	GetClusterID(ctx context.Context, token string, clusterNameOrID string) (string, error)
}

// Tools contains tools for accessing provisioning information.
type Tools struct {
	client   toolsClient
	ReadOnly bool
}

// NewTools creates and returns a new Tools instance.
func NewTools(client toolsClient, readOnly bool) *Tools {
	return &Tools{
		client:   client,
		ReadOnly: readOnly,
	}
}

func (t *Tools) AddTools(mcpServer *mcp.Server) {
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "listVirtualMachines",
		Meta: map[string]any{
			toolsSetAnn: toolsSet,
		},
		Description: `Lists VirtualMachine resources on a Harvester / SUSE Virtualization cluster.
Returns the name, namespace, running state, CPU, memory, and status conditions for each VM.
Use this when the user wants an overview of their virtual machines.`},
		t.listVirtualMachines)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "listVirtualMachineImages",
		Meta: map[string]any{
			toolsSetAnn: toolsSet,
		},
		Description: `Lists VirtualMachineImage resources on a Harvester / SUSE Virtualization cluster.
Returns the display name, source type, download URL, import status, progress, size, and virtual size for each image.
Use this when the user wants to see available VM images or check the status of an image import.`},
		t.listVirtualMachineImages)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "listVolumes",
		Meta: map[string]any{
			toolsSetAnn: toolsSet,
		},
		Description: `Lists storage volumes (PersistentVolumeClaims) on a Harvester / SUSE Virtualization cluster.
Returns the name, namespace, capacity, access modes, storage class, and binding status for each volume.
Use this when the user wants to see available storage volumes or check their status.`},
		t.listVolumes)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "listVirtualMachineNetworks",
		Meta: map[string]any{
			toolsSetAnn: toolsSet,
		},
		Description: `Lists NetworkAttachmentDefinition resources on a Harvester / SUSE Virtualization cluster.
Returns the name, namespace, VLAN ID, cluster network, network type, and connectivity status for each network.
Use this when the user wants to see available VM networks or check their configuration.`},
		t.listVirtualMachineNetworks)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "listClusterNetworks",
		Meta: map[string]any{
			toolsSetAnn: toolsSet,
		},
		Description: `Lists ClusterNetwork resources on a Harvester / SUSE Virtualization cluster.
Returns the name, readiness conditions, MTU, and uplink configuration for each cluster-level network.
Use this when the user wants to see the underlying cluster networks that VM networks are built on.`},
		t.listClusterNetworks)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "listNodes",
		Meta: map[string]any{
			toolsSetAnn: toolsSet,
		},
		Description: `Lists nodes (hosts) on a Harvester / SUSE Virtualization cluster.
Returns each node's name, roles, status conditions, CPU, memory, and resource utilization metrics when available.
In a Harvester cluster, nodes and hosts refer to the same physical or virtual machines that form the cluster.
Use this when the user wants an overview of the cluster hosts or needs to check node health and capacity.`},
		t.listNodes)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "inspectVolume",
		Meta: map[string]any{
			toolsSetAnn: toolsSet,
		},
		Description: `Inspects a storage volume (PersistentVolumeClaim) on a Harvester / SUSE Virtualization cluster.
Returns the PVC details along with its StorageClass. If the volume is backed by Longhorn (provisioner contains "driver.longhorn.io"),
also returns the corresponding Longhorn Volume resource with replica scheduling, robustness, and condition details.
Use this when the user wants detailed information about a specific volume or needs to troubleshoot storage issues.`},
		t.inspectVolume)

	if !t.ReadOnly {
		mcp.AddTool(mcpServer, &mcp.Tool{
			Name: "createVirtualMachineImage",
			Meta: map[string]any{
				toolsSetAnn: toolsSet,
			},
			Description: `Creates a VirtualMachineImage resource on a Harvester / SUSE Virtualization cluster by downloading from a URL.
The image will be imported into the harvester-longhorn storage class by default.
Use this when the user wants to add a new VM image by providing a download URL.`},
			t.createVirtualMachineImage)

		mcp.AddTool(mcpServer, &mcp.Tool{
			Name: "createVirtualMachine",
			Meta: map[string]any{
				toolsSetAnn: toolsSet,
			},
			Description: `Creates a VirtualMachine on a Harvester / SUSE Virtualization cluster.
Fetches the specified VirtualMachineImage to derive the storage class, then provisions a VM with the given CPU, memory, disk size, and network attachment.
Use this when the user wants to create a new virtual machine from an existing image.`},
			t.createVirtualMachine)

		mcp.AddTool(mcpServer, &mcp.Tool{
			Name: "createVirtualMachinePlan",
			Meta: map[string]any{
				toolsSetAnn: toolsSet,
			},
			Description: `Plans to create a VirtualMachine on a Harvester / SUSE Virtualization cluster.
Returns the full JSON manifest that would be submitted without actually creating the resource.
Only used for displaying the resource when using human validation.`},
			t.createVirtualMachinePlan)

		mcp.AddTool(mcpServer, &mcp.Tool{
			Name: "deleteVirtualMachine",
			Meta: map[string]any{
				toolsSetAnn: toolsSet,
			},
			Description: `Deletes a VirtualMachine on a Harvester / SUSE Virtualization cluster.
Use this when the user wants to permanently remove a virtual machine.`},
			t.deleteVirtualMachine)

		mcp.AddTool(mcpServer, &mcp.Tool{
			Name: "deleteVirtualMachinePlan",
			Meta: map[string]any{
				toolsSetAnn: toolsSet,
			},
			Description: `Plans to delete a VirtualMachine on a Harvester / SUSE Virtualization cluster.
Returns the resource that would be deleted without actually removing it.
Only used for displaying the target when using human validation.`},
			t.deleteVirtualMachinePlan)
	}
}
