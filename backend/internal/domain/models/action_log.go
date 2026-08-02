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

// ActionLogListItem is the read model ReadByFilter returns - it embeds the
// table-mapped ActionLog (used as-is by Create and the single-dispatch
// response) plus the human-readable names joined in from actions/nodes,
// so the frontend's Action History table doesn't have to resolve UUIDs
// to names itself. NodeDeviceId/NodeName are pointers because node_id
// itself is nullable (an action log can predate/lack a resolved node);
// ActionName is a plain string because action_id is NOT NULL and actions
// are soft-deleted (never hard-deleted), so the join always matches.
type ActionLogListItem struct {
	ActionLog
	ActionName   string
	NodeDeviceId *string
	NodeName     *string
}
