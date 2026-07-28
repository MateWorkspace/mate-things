package presentationhttpresponse

import (
	"encoding/json"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type ActionLogResponse struct {
	Id            int64                     `json:"id" example:"4821"`
	ExecutionId   string                    `json:"execution_id" example:"a4d7f1c9-3e6b-4a8d-9c2f-5b1e7d4a6c02"`
	ActionId      string                    `json:"action_id" example:"6d9e2f5a-8b1c-4d3e-9f6a-2c5d8e1f4b07"`
	NodeId        *string                   `json:"node_id,omitempty" example:"5e8a1c3f-2b7d-4f6a-9c1e-3a8b6d2f4e09"`
	ActionStatus  domainmodels.ActionStatus `json:"action_status" example:"SUCCESS"`
	ActionMessage *string                   `json:"action_message,omitempty" example:"Shot pulled: 93.5C for 28s. Crema looked great."`
	Payload       json.RawMessage           `json:"payload" swaggertype:"object"`
	ExecutedAt    time.Time                 `json:"executed_at" example:"2026-07-28T08:15:00Z"`
	CreatedAt     time.Time                 `json:"created_at" example:"2026-07-28T08:15:02Z"`
}

func ActionLog(actionLog domainmodels.ActionLog) ActionLogResponse {
	return ActionLogResponse{
		Id:            actionLog.Id,
		ExecutionId:   UUIDString(actionLog.ExecutionId),
		ActionId:      UUIDString(actionLog.ActionId),
		NodeId:        UUIDPtrString(actionLog.NodeId),
		ActionStatus:  actionLog.ActionStatus,
		ActionMessage: actionLog.ActionMessage,
		Payload:       NormalizeJSON(actionLog.Payload),
		ExecutedAt:    actionLog.ExecutedAt,
		CreatedAt:     actionLog.CreatedAt,
	}
}

func ActionLogs(actionLogs []domainmodels.ActionLog) []ActionLogResponse {
	result := make([]ActionLogResponse, 0, len(actionLogs))
	for _, actionLog := range actionLogs {
		result = append(result, ActionLog(actionLog))
	}
	return result
}
