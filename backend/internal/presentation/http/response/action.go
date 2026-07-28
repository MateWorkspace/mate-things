package presentationhttpresponse

import (
	"encoding/json"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type ActionResponse struct {
	Id                   string          `json:"id"`
	NodeClassId          string          `json:"node_class_id"`
	Name                 string          `json:"name"`
	Description          string          `json:"description"`
	PayloadSchemaName    string          `json:"payload_schema_name"`
	PayloadSchemaVersion int32           `json:"payload_schema_version"`
	Preferences          json.RawMessage `json:"preferences"`
	AuditResponse
}

func Action(action domainmodels.Action) ActionResponse {
	return ActionResponse{
		Id:                   UUIDString(action.Id),
		NodeClassId:          UUIDString(action.NodeClassId),
		Name:                 action.Name,
		Description:          action.Description,
		PayloadSchemaName:    action.PayloadSchemaName,
		PayloadSchemaVersion: action.PayloadSchemaVersion,
		Preferences:          NormalizeJSON(action.Preferences),
		AuditResponse: Audit(
			action.CreatedAt,
			action.UpdatedAt,
			action.DeletedAt,
			action.CreatedBy,
			action.UpdatedBy,
			action.DeletedBy,
		),
	}
}

func Actions(actions []domainmodels.Action) []ActionResponse {
	result := make([]ActionResponse, 0, len(actions))
	for _, action := range actions {
		result = append(result, Action(action))
	}
	return result
}
