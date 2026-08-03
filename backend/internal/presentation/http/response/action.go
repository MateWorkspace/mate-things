package presentationhttpresponse

import (
	"encoding/json"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type ActionResponse struct {
	Id                       string          `json:"id" example:"6d9e2f5a-8b1c-4d3e-9f6a-2c5d8e1f4b07"`
	Name                     string          `json:"name" example:"pull_espresso_shot"`
	Description              string          `json:"description" example:"Runs a timed espresso extraction on the target node."`
	PayloadSchemaName        string          `json:"payload_schema_name" example:"pull_espresso_shot"`
	PayloadSchemaVersion     int32           `json:"payload_schema_version" example:"1"`
	Preferences              json.RawMessage `json:"preferences" swaggertype:"object"`
	CompatibleNodeClassCount int             `json:"compatible_node_class_count,omitempty" example:"3"`
	AuditResponse
}

func Action(action domainmodels.Action) ActionResponse {
	return ActionResponse{
		Id:                   UUIDString(action.Id),
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

// ActionListItem maps the list read-model (base Action fields plus the
// compatible-class count) - mirrors ActionLogListItem() in action_log.go.
// The single-item Action() mapper above is unchanged and leaves
// CompatibleNodeClassCount at its zero value.
func ActionListItem(item domainmodels.ActionListItem) ActionResponse {
	resp := Action(item.Action)
	resp.CompatibleNodeClassCount = item.CompatibleNodeClassCount
	return resp
}

func ActionListItems(items []domainmodels.ActionListItem) []ActionResponse {
	result := make([]ActionResponse, 0, len(items))
	for _, item := range items {
		result = append(result, ActionListItem(item))
	}
	return result
}
