package domainmodels

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type PayloadSchemaDefinitionType string

const (
	PayloadSchemaDefinitionTypeString       PayloadSchemaDefinitionType = "string"
	PayloadSchemaDefinitionTypeFloat        PayloadSchemaDefinitionType = "float"
	PayloadSchemaDefinitionTypeInteger      PayloadSchemaDefinitionType = "integer"
	PayloadSchemaDefinitionTypeBoolean      PayloadSchemaDefinitionType = "boolean"
	PayloadSchemaDefinitionTypeEnum         PayloadSchemaDefinitionType = "enum"
	PayloadSchemaDefinitionTypeObject       PayloadSchemaDefinitionType = "object"
	PayloadSchemaDefinitionTypeStringArray  PayloadSchemaDefinitionType = "[]string"
	PayloadSchemaDefinitionTypeFloatArray   PayloadSchemaDefinitionType = "[]float"
	PayloadSchemaDefinitionTypeIntegerArray PayloadSchemaDefinitionType = "[]integer"
	PayloadSchemaDefinitionTypeBooleanArray PayloadSchemaDefinitionType = "[]boolean"
	PayloadSchemaDefinitionTypeEnumArray    PayloadSchemaDefinitionType = "[]enum"
	PayloadSchemaDefinitionTypeObjectArray  PayloadSchemaDefinitionType = "[]object"
)

type PayloadSchemaDefinition struct {
	Type          PayloadSchemaDefinitionType        `json:"type"`
	Required      []string                           `json:"required"`
	Options       []string                           `json:"options"`
	Unit          *string                            `json:"unit"`
	Minimum       *float64                           `json:"minimum"`
	Maximum       *float64                           `json:"maximum"`
	MinimumLength *int64                             `json:"minimum_length"`
	MaximumLength *int64                             `json:"maximum_length"`
	MinimumItem   *int64                             `json:"minimum_item"`
	MaximumItem   *int64                             `json:"maximum_item"`
	Properties    map[string]PayloadSchemaDefinition `json:"properties"`
	Items         *PayloadSchemaDefinition           `json:"items"`
}

type PayloadSchema struct {
	Id          uuid.UUID       `db:"id" json:"id"`
	Name        string          `db:"name" json:"name"`
	Version     int32           `db:"version" json:"version"`
	Definition  json.RawMessage `db:"definition" json:"definition"`
	ValidFrom   time.Time       `db:"valid_from" json:"valid_from"`
	ValidTo     *time.Time      `db:"valid_to" json:"valid_to,omitempty"`
	Preferences json.RawMessage `db:"preferences" json:"preferences"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   *time.Time      `db:"updated_at" json:"updated_at,omitempty"`
	DeletedAt   *time.Time      `db:"deleted_at" json:"deleted_at,omitempty"`
	CreatedBy   *uuid.UUID      `db:"created_by" json:"created_by,omitempty"`
	UpdatedBy   *uuid.UUID      `db:"updated_by" json:"updated_by,omitempty"`
	DeletedBy   *uuid.UUID      `db:"deleted_by" json:"deleted_by,omitempty"`
}
