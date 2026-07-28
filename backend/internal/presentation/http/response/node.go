package presentationhttpresponse

import (
	"encoding/json"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type NodeResponse struct {
	Id          string          `json:"id"`
	NodeClassId string          `json:"node_class_id"`
	DeviceId    string          `json:"device_id"`
	DeviceInfo  string          `json:"device_info"`
	Name        string          `json:"name"`
	FirmwareId  string          `json:"firmware_id"`
	Description string          `json:"description"`
	IsConnected bool            `json:"is_connected"`
	Preferences json.RawMessage `json:"preferences"`
	AuditResponse
}

func Node(node domainmodels.Node) NodeResponse {
	return NodeResponse{
		Id:          UUIDString(node.Id),
		NodeClassId: UUIDString(node.NodeClassId),
		DeviceId:    node.DeviceId,
		DeviceInfo:  node.DeviceInfo,
		Name:        node.Name,
		FirmwareId:  UUIDString(node.FirmwareId),
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
