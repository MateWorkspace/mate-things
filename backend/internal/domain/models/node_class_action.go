package domainmodels

import (
	"time"

	"github.com/google/uuid"
)

type NodeClassAction struct {
	Id          uuid.UUID  `db:"id" json:"id"`
	NodeClassId uuid.UUID  `db:"node_class_id" json:"node_class_id"`
	ActionId    uuid.UUID  `db:"action_id" json:"action_id"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	CreatedBy   *uuid.UUID `db:"created_by" json:"created_by,omitempty"`
}
