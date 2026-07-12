package virtualization

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/harvester/harvester-mcp-server/internal/middleware"
	"github.com/harvester/harvester-mcp-server/pkg/client"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const longhornProvisioner = "driver.longhorn.io"
const longhornSystemNamespace = "longhorn-system"

// inspectVolumeParams specifies the parameters needed to inspect a volume.
type inspectVolumeParams struct {
	Cluster   string `json:"cluster" jsonschema:"the name or ID of the Harvester cluster to query"`
	Namespace string `json:"namespace" jsonschema:"the namespace of the PersistentVolumeClaim"`
	Name      string `json:"name" jsonschema:"the name of the PersistentVolumeClaim"`
}

// inspectVolumeResult holds the aggregated details returned to the LLM.
type inspectVolumeResult struct {
	PVC            map[string]any  `json:"pvc"`
	StorageClass   map[string]any  `json:"storageClass,omitempty"`
	LonghornVolume *map[string]any `json:"longhornVolume,omitempty"`
}

// inspectVolume fetches a PVC, its StorageClass, and — when the provisioner is
// Longhorn — the corresponding longhorn.io/v1beta2 Volume resource.
func (t *Tools) inspectVolume(ctx context.Context, toolReq *mcp.CallToolRequest, params inspectVolumeParams) (*mcp.CallToolResult, any, error) {
	zap.L().Debug("inspectVolume called",
		zap.String("cluster", params.Cluster),
		zap.String("namespace", params.Namespace),
		zap.String("name", params.Name))

	token := middleware.Token(ctx)

	// 1. Fetch the PVC.
	pvcObj, err := t.client.GetResource(ctx, client.GetParams{
		Cluster:   params.Cluster,
		Kind:      "persistentvolumeclaim",
		Namespace: params.Namespace,
		Name:      params.Name,
		Token:     token,
	})
	if err != nil {
		zap.L().Error("failed to get PVC", zap.String("tool", "inspectVolume"), zap.Error(err))
		return nil, nil, fmt.Errorf("failed to get PVC %s/%s: %w", params.Namespace, params.Name, err)
	}

	unstructured.RemoveNestedField(pvcObj.Object, "metadata", "managedFields")

	result := inspectVolumeResult{
		PVC: pvcObj.Object,
	}

	// 2. Fetch the StorageClass referenced by the PVC.
	storageClassName, _, _ := unstructured.NestedString(pvcObj.Object, "spec", "storageClassName")
	if storageClassName != "" {
		scObj, err := t.client.GetResource(ctx, client.GetParams{
			Cluster:   params.Cluster,
			Kind:      "storageclass",
			Namespace: "",
			Name:      storageClassName,
			Token:     token,
		})
		if err != nil && !apierrors.IsNotFound(err) {
			zap.L().Warn("failed to get StorageClass", zap.String("storageClass", storageClassName), zap.Error(err))
		} else if err == nil {
			unstructured.RemoveNestedField(scObj.Object, "metadata", "managedFields")
			result.StorageClass = scObj.Object

			// 3. If provisioner is Longhorn, fetch the Longhorn Volume.
			provisioner, _, _ := unstructured.NestedString(scObj.Object, "provisioner")
			if strings.Contains(provisioner, longhornProvisioner) {
				volumeName, _, _ := unstructured.NestedString(pvcObj.Object, "spec", "volumeName")
				if volumeName != "" {
					lvObj, err := t.client.GetResource(ctx, client.GetParams{
						Cluster:   params.Cluster,
						Kind:      "longhornvolume",
						Namespace: longhornSystemNamespace,
						Name:      volumeName,
						Token:     token,
					})
					if err != nil && !apierrors.IsNotFound(err) {
						zap.L().Warn("failed to get Longhorn volume", zap.String("volume", volumeName), zap.Error(err))
					} else if err == nil {
						unstructured.RemoveNestedField(lvObj.Object, "metadata", "managedFields")
						result.LonghornVolume = &lvObj.Object
					}
				}
			}
		}
	}

	responseBytes, err := json.Marshal(result)
	if err != nil {
		zap.L().Error("failed to marshal response", zap.String("tool", "inspectVolume"), zap.Error(err))
		return nil, nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(responseBytes)}},
	}, nil, nil
}
