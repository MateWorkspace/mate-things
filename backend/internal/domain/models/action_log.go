package domainmodels

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ActionStatus string

const (
	ActionStatusUnexecuted  ActionStatus = "UNEXECUTED"
	ActionStatusUnresponded ActionStatus = "UNRESPONDED"
	ActionStatusFailed      ActionStatus = "FAILED"
	ActionStatusSuccess     ActionStatus = "SUCCESS"
)

type ActionLog struct {
	Id            int64           `db:"id" json:"id"`
	ExecutionId   uuid.UUID       `db:"execution_id" json:"execution_id"`
	ActionId      uuid.UUID       `db:"action_id" json:"action_id"`
	NodeId        *uuid.UUID      `db:"node_id" json:"node_id,omitempty"`
	ActionStatus  ActionStatus    `db:"action_status" json:"action_status"`
	ActionMessage *string         `db:"action_message" json:"action_message,omitempty"`
	Payload       json.RawMessage `db:"payload" json:"payload"`
	ExecutedAt    time.Time       `db:"executed_at" json:"executed_at"`
	CreatedAt     time.Time       `db:"created_at" json:"created_at"`
}
