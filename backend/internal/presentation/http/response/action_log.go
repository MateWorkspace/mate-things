package presentationhttpresponse

import (
	"encoding/json"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type ActionLogResponse struct {
	Id            int64                     `json:"id"`
	ExecutionId   string                    `json:"execution_id"`
	ActionId      string                    `json:"action_id"`
	NodeId        *string                   `json:"node_id,omitempty"`
	ActionStatus  domainmodels.ActionStatus `json:"action_status"`
	ActionMessage *string                   `json:"action_message,omitempty"`
	Payload       json.RawMessage           `json:"payload"`
	ExecutedAt    time.Time                 `json:"executed_at"`
	CreatedAt     time.Time                 `json:"created_at"`
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
