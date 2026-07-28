package presentationhttpresponse

import (
	"encoding/json"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type NodeClassResponse struct {
	Id          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Preferences json.RawMessage `json:"preferences"`
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
