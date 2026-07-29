package domainmodels

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type NodeConfigValue struct {
	Id          uuid.UUID       `db:"id" json:"id"`
	NodeId      uuid.UUID       `db:"node_id" json:"node_id"`
	FirmwareId  uuid.UUID       `db:"firmware_id" json:"firmware_id"`
	Key         string          `db:"key" json:"key"`
	Value       string          `db:"value" json:"value"`
	Preferences json.RawMessage `db:"preferences" json:"preferences"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   *time.Time      `db:"updated_at" json:"updated_at,omitempty"`
	DeletedAt   *time.Time      `db:"deleted_at" json:"deleted_at,omitempty"`
	CreatedBy   *uuid.UUID      `db:"created_by" json:"created_by,omitempty"`
	UpdatedBy   *uuid.UUID      `db:"updated_by" json:"updated_by,omitempty"`
	DeletedBy   *uuid.UUID      `db:"deleted_by" json:"deleted_by,omitempty"`
}
