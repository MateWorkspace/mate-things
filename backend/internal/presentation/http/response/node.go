package presentationhttpresponse

import (
	"encoding/json"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type NodeResponse struct {
	Id          string          `json:"id" example:"5e8a1c3f-2b7d-4f6a-9c1e-3a8b6d2f4e09"`
	NodeClassId string          `json:"node_class_id" example:"3f1c9a2e-6d4b-4e7a-8c2f-1a9b3d5e7f01"`
	DeviceId    string          `json:"device_id" example:"ESP32-BARISTA-07"`
	DeviceInfo  string          `json:"device_info" example:"ESP32-WROOM-32E, rev 3, 240MHz dual-core"`
	Name        string          `json:"name" example:"Kitchen Espresso Machine"`
	FirmwareId  *string         `json:"firmware_id,omitempty" example:"9c4e2b7a-1f3d-4a6c-8b5e-2d7f9a1c3e08"`
	Description string          `json:"description" example:"The espresso machine behind the office kitchen counter."`
	IsConnected bool            `json:"is_connected" example:"true"`
	Preferences json.RawMessage `json:"preferences" swaggertype:"object"`
	AuditResponse
}

func Node(node domainmodels.Node) NodeResponse {
	return NodeResponse{
		Id:          UUIDString(node.Id),
		NodeClassId: UUIDString(node.NodeClassId),
		DeviceId:    node.DeviceId,
		DeviceInfo:  node.DeviceInfo,
		Name:        node.Name,
		FirmwareId:  UUIDPtrString(node.FirmwareId),
		Description: node.Description,
		IsConnected: node.IsConnected,
		Preferences: NormalizeJSON(node.Preferences),
		AuditResponse: Audit(
			node.CreatedAt,
			node.UpdatedAt,
			node.DeletedAt,
			node.CreatedBy,
			node.UpdatedBy,
			node.DeletedBy,
		),
	}
}

func Nodes(nodes []domainmodels.Node) []NodeResponse {
	result := make([]NodeResponse, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, Node(node))
	}
	return result
}
