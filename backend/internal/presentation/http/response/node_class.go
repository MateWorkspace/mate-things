package presentationhttpresponse

import (
	"encoding/json"
	"time"

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

type NodeClassActionResponse struct {
	Id          string    `json:"id" example:"a4d7f1c9-3e6b-4a8d-9c2f-5b1e7d4a6c02"`
	NodeClassId string    `json:"node_class_id" example:"3f1c9a2e-6d4b-4e7a-8c2f-1a9b3d5e7f01"`
	ActionId    string    `json:"action_id" example:"6d9e2f5a-8b1c-4d3e-9f6a-2c5d8e1f4b07"`
	CreatedAt   time.Time `json:"created_at" example:"2026-08-03T09:00:00Z"`
	CreatedBy   *string   `json:"created_by,omitempty" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
}

type NodeClassActionDetailResponse struct {
	NodeClassAction NodeClassActionResponse `json:"node_class_action"`
	NodeClass       NodeClassResponse       `json:"node_class"`
	Action          ActionResponse          `json:"action"`
}

func NodeClassAction(nodeClassAction domainmodels.NodeClassAction) NodeClassActionResponse {
	return NodeClassActionResponse{
		Id:          UUIDString(nodeClassAction.Id),
		NodeClassId: UUIDString(nodeClassAction.NodeClassId),
		ActionId:    UUIDString(nodeClassAction.ActionId),
		CreatedAt:   nodeClassAction.CreatedAt,
		CreatedBy:   UUIDPtrString(nodeClassAction.CreatedBy),
	}
}

func NodeClassActionDetail(nodeClassAction domainmodels.NodeClassAction, nodeClass domainmodels.NodeClass, action domainmodels.Action) NodeClassActionDetailResponse {
	return NodeClassActionDetailResponse{
		NodeClassAction: NodeClassAction(nodeClassAction),
		NodeClass:       NodeClass(nodeClass),
		Action:          Action(action),
	}
}

func NodeClassActionDetails(
	nodeClassActions []domainmodels.NodeClassAction,
	nodeClasses []domainmodels.NodeClass,
	actions []domainmodels.Action,
) []NodeClassActionDetailResponse {
	result := make([]NodeClassActionDetailResponse, 0, len(nodeClassActions))
	for i, nodeClassAction := range nodeClassActions {
		var nodeClass domainmodels.NodeClass
		if i < len(nodeClasses) {
			nodeClass = nodeClasses[i]
		}

		var action domainmodels.Action
		if i < len(actions) {
			action = actions[i]
		}

		result = append(result, NodeClassActionDetail(nodeClassAction, nodeClass, action))
	}
	return result
}
