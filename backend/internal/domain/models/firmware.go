package domainmodels

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Firmware struct {
	Id          uuid.UUID       `db:"id" json:"id"`
	NodeClassId uuid.UUID       `db:"node_class_id" json:"node_class_id"`
	Name        string          `db:"name" json:"name"`
	Size        int32           `db:"size" json:"size"`
	Checksum    string          `db:"checksum" json:"checksum"`
	BinaryPath  string          `db:"binary_path" json:"binary_path"`
	Preferences json.RawMessage `db:"preferences" json:"preferences"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   *time.Time      `db:"updated_at" json:"updated_at,omitempty"`
	DeletedAt   *time.Time      `db:"deleted_at" json:"deleted_at,omitempty"`
	CreatedBy   *uuid.UUID      `db:"created_by" json:"created_by,omitempty"`
	UpdatedBy   *uuid.UUID      `db:"updated_by" json:"updated_by,omitempty"`
	DeletedBy   *uuid.UUID      `db:"deleted_by" json:"deleted_by,omitempty"`
}
