package presentationhttpresponse

import (
	"encoding/json"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type NodeClassResponse struct {
	Id          string          `json:"id" example:"3f1c9a2e-6d4b-4e7a-8c2f-1a9b3d5e7f01"`
	Name        string          `json:"name" example:"Espresso Machine"`
	Description string          `json:"description" example:"Dual-boiler espresso machines with ESP32-controlled brew heads."`
	Preferences json.RawMessage `json:"preferences" swaggertype:"object"`
	AuditResponse
}

func NodeClass(nodeClass domainmodels.NodeClass) NodeClassResponse {
	return NodeClassResponse{
		Id:          UUIDString(nodeClass.Id),
		Name:        nodeClass.Name,
		Description: nodeClass.Description,
		Preferences: NormalizeJSON(nodeClass.Preferences),
		AuditResponse: Audit(
			nodeClass.CreatedAt,
			nodeClass.UpdatedAt,
			nodeClass.DeletedAt,
			nodeClass.CreatedBy,
			nodeClass.UpdatedBy,
			nodeClass.DeletedBy,
		),
	}
}

func NodeClasses(nodeClasses []domainmodels.NodeClass) []NodeClassResponse {
	result := make([]NodeClassResponse, 0, len(nodeClasses))
	for _, nodeClass := range nodeClasses {
		result = append(result, NodeClass(nodeClass))
	}
	return result
}
