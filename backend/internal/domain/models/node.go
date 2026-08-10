package domainmodels

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Node struct {
	Id          uuid.UUID       `db:"id" json:"id"`
	NodeClassId uuid.UUID       `db:"node_class_id" json:"node_class_id"`
	DeviceId    string          `db:"device_id" json:"device_id"`
	DeviceInfo  string          `db:"device_info" json:"device_info"`
	Name        string          `db:"name" json:"name"`
	FirmwareId  *uuid.UUID      `db:"firmware_id" json:"firmware_id,omitempty"`
	Description string          `db:"description" json:"description"`
	IsConnected bool            `db:"is_connected" json:"is_connected"`
	Preferences json.RawMessage `db:"preferences" json:"preferences"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   *time.Time      `db:"updated_at" json:"updated_at,omitempty"`
	DeletedAt   *time.Time      `db:"deleted_at" json:"deleted_at,omitempty"`
	CreatedBy   *uuid.UUID      `db:"created_by" json:"created_by,omitempty"`
	UpdatedBy   *uuid.UUID      `db:"updated_by" json:"updated_by,omitempty"`
	DeletedBy   *uuid.UUID      `db:"deleted_by" json:"deleted_by,omitempty"`
}
