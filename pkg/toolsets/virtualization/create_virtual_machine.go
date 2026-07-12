package virtualization

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/harvester/harvester-mcp-server/internal/middleware"
	"github.com/harvester/harvester-mcp-server/pkg/client"
	"github.com/harvester/harvester-mcp-server/pkg/converter"
	"github.com/harvester/harvester-mcp-server/pkg/response"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8srand "k8s.io/apimachinery/pkg/util/rand"
)

type createVirtualMachineParams struct {
	Cluster     string `json:"cluster" jsonschema:"the name or ID of the Harvester cluster"`
	Namespace   string `json:"namespace,omitempty" jsonschema:"namespace to create the VM in; defaults to 'default'"`
	Name        string `json:"name" jsonschema:"name of the virtual machine"`
	ImageID     string `json:"imageId" jsonschema:"VirtualMachineImage name (e.g. 'image-fnndf') or 'namespace/name' format (e.g. 'default/image-fnndf')"`
	NetworkName string `json:"networkName" jsonschema:"name of the NetworkAttachmentDefinition to attach (e.g. 'hvst-dev-vlan')"`
	CPUCores    int    `json:"cpuCores,omitempty" jsonschema:"number of CPU cores; defaults to 2"`
	Memory      string `json:"memory,omitempty" jsonschema:"memory amount (e.g. '1Gi', '2Gi'); defaults to '1Gi'"`
	DiskSize    string `json:"diskSize,omitempty" jsonschema:"boot disk size (e.g. '10Gi', '20Gi'); defaults to '10Gi'"`
}

// createVirtualMachineObj builds the VirtualMachine unstructured object from params.
// It fetches the VirtualMachineImage to derive the storage class and assembles the full manifest.
// It does not create the resource — callers are responsible for that.
func (t *Tools) createVirtualMachineObj(ctx context.Context, toolReq *mcp.CallToolRequest, params createVirtualMachineParams) (*unstructured.Unstructured, error) {
	ns := params.Namespace
	if ns == "" {
		ns = defaultNamespace
	}
	cpuCores := params.CPUCores
	if cpuCores == 0 {
		cpuCores = 2
	}
	memory := params.Memory
	if memory == "" {
		memory = "1Gi"
	}
	diskSize := params.DiskSize
	if diskSize == "" {
		diskSize = "10Gi"
	}

	imageNamespace, imageName := ns, params.ImageID
	if parts := strings.SplitN(params.ImageID, "/", 2); len(parts) == 2 {
		imageNamespace, imageName = parts[0], parts[1]
	}

	imageObj, err := t.client.GetResource(ctx, client.GetParams{
		Cluster:   params.Cluster,
		Kind:      "virtualmachineimage",
		Namespace: imageNamespace,
		Name:      imageName,
		Token:     middleware.Token(ctx),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get VirtualMachineImage %s/%s: %w", imageNamespace, imageName, err)
	}

	storageClassName, _, _ := unstructured.NestedString(imageObj.Object, "status", "storageClassName")
	if storageClassName == "" {
		return nil, fmt.Errorf("VirtualMachineImage %s/%s has no storageClassName in status; image may still be importing", imageNamespace, imageName)
	}

	pvcName := fmt.Sprintf("%s-disk-0-%s", params.Name, k8srand.String(5))
	imageRef := imageNamespace + "/" + imageName

	volumeClaimTemplates := []map[string]any{
		{
			"metadata": map[string]any{
				"name": pvcName,
				"annotations": map[string]any{
					"harvesterhci.io/imageId": imageRef,
				},
			},
			"spec": map[string]any{
				"accessModes": []string{"ReadWriteMany"},
				"resources": map[string]any{
					"requests": map[string]any{
						"storage": diskSize,
					},
				},
				"volumeMode":       "Block",
				"storageClassName": storageClassName,
			},
		},
	}

	volumeClaimTemplatesJSON, err := json.Marshal(volumeClaimTemplates)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal volumeClaimTemplates: %w", err)
	}

	networkRef := ns + "/" + params.NetworkName

	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "kubevirt.io/v1",
			"kind":       "VirtualMachine",
			"metadata": map[string]any{
				"name":      params.Name,
				"namespace": ns,
				"labels": map[string]any{
					"harvesterhci.io/creator": "harvester",
					"harvesterhci.io/os":      "linux",
				},
				"annotations": map[string]any{
					"harvesterhci.io/volumeClaimTemplates": string(volumeClaimTemplatesJSON),
					"network.harvesterhci.io/ips":          "[]",
				},
			},
			"spec": map[string]any{
				"runStrategy": "RerunOnFailure",
				"template": map[string]any{
					"metadata": map[string]any{
						"labels": map[string]any{
							"harvesterhci.io/vmName": params.Name,
						},
					},
					"spec": map[string]any{
						"domain": map[string]any{
							"machine": map[string]any{
								"type": "",
							},
							"cpu": map[string]any{
								"cores":   int64(cpuCores),
								"sockets": int64(1),
								"threads": int64(1),
							},
							"devices": map[string]any{
								"inputs": []any{
									map[string]any{
										"bus":  "usb",
										"name": "tablet",
										"type": "tablet",
									},
								},
								"interfaces": []any{
									map[string]any{
										"model":  "virtio",
										"name":   "nic-1",
										"bridge": map[string]any{},
									},
								},
								"disks": []any{
									map[string]any{
										"name": "disk-0",
										"disk": map[string]any{
											"bus": "virtio",
										},
										"bootOrder": int64(1),
									},
								},
							},
							"resources": map[string]any{
								"limits": map[string]any{
									"memory": memory,
									"cpu":    strconv.Itoa(cpuCores),
								},
							},
							"features": map[string]any{
								"acpi": map[string]any{
									"enabled": true,
								},
							},
						},
						"evictionStrategy":              "LiveMigrateIfPossible",
						"hostname":                      params.Name,
						"terminationGracePeriodSeconds": int64(120),
						"networks": []any{
							map[string]any{
								"name": "nic-1",
								"multus": map[string]any{
									"networkName": networkRef,
								},
							},
						},
						"volumes": []any{
							map[string]any{
								"name": "disk-0",
								"persistentVolumeClaim": map[string]any{
									"claimName": pvcName,
								},
							},
						},
						"affinity": map[string]any{},
					},
				},
			},
		},
	}, nil
}

func (t *Tools) createVirtualMachine(ctx context.Context, toolReq *mcp.CallToolRequest, params createVirtualMachineParams) (*mcp.CallToolResult, any, error) {
	zap.L().Debug("createVirtualMachine called",
		zap.String("cluster", params.Cluster),
		zap.String("name", params.Name),
		zap.String("imageId", params.ImageID))

	obj, err := t.createVirtualMachineObj(ctx, toolReq, params)
	if err != nil {
		zap.L().Error("failed to build VM object", zap.String("tool", "createVirtualMachine"), zap.Error(err))
		return nil, nil, err
	}

	gvr := converter.K8sKindsToGVRs["virtualmachine"]
	resourceInterface, err := t.client.GetResourceInterface(ctx, middleware.Token(ctx), obj.GetNamespace(), params.Cluster, gvr)
	if err != nil {
		zap.L().Error("failed to get resource interface", zap.String("tool", "createVirtualMachine"), zap.Error(err))
		return nil, nil, fmt.Errorf("failed to get resource interface: %w", err)
	}

	created, err := resourceInterface.Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		zap.L().Error("failed to create VirtualMachine", zap.String("tool", "createVirtualMachine"), zap.Error(err))
		return nil, nil, fmt.Errorf("failed to create VirtualMachine: %w", err)
	}

	mcpResponse, err := response.CreateMcpResponse([]*unstructured.Unstructured{created}, params.Cluster)
	if err != nil {
		zap.L().Error("failed to create mcp response", zap.String("tool", "createVirtualMachine"), zap.Error(err))
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: mcpResponse}},
	}, nil, nil
}
