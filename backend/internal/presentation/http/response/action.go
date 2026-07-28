package presentationhttpresponse

import (
	"encoding/json"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type ActionResponse struct {
	Id                   string          `json:"id" example:"6d9e2f5a-8b1c-4d3e-9f6a-2c5d8e1f4b07"`
	NodeClassId          string          `json:"node_class_id" example:"3f1c9a2e-6d4b-4e7a-8c2f-1a9b3d5e7f01"`
	Name                 string          `json:"name" example:"brew_espresso"`
	Description          string          `json:"description" example:"Pulls a double shot at the requested temperature and duration."`
	PayloadSchemaName    string          `json:"payload_schema_name" example:"brew_command"`
	PayloadSchemaVersion int32           `json:"payload_schema_version" example:"1"`
	Preferences          json.RawMessage `json:"preferences" swaggertype:"object"`
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
