package presentationhttprequest

import (
	"encoding/json"
	"time"
)

type ActionPostRequest struct {
	NodeClassId          string  `json:"node_class_id" example:"3f1c9a2e-6d4b-4e7a-8c2f-1a9b3d5e7f01"`
	Name                 string  `json:"name" example:"brew_espresso"`
	Description          *string `json:"description" example:"Pulls a double shot at the requested temperature and duration."`
	PayloadSchemaName    string  `json:"payload_schema_name" example:"brew_command"`
	PayloadSchemaVersion int32   `json:"payload_schema_version" example:"1"`
}

type ActionPatchRequest struct {
	NodeClassId          *string `json:"node_class_id" example:"3f1c9a2e-6d4b-4e7a-8c2f-1a9b3d5e7f01"`
	Name                 *string `json:"name" example:"brew_espresso"`
	Description          *string `json:"description" example:"Pulls a double shot at the requested temperature and duration."`
	PayloadSchemaName    *string `json:"payload_schema_name" example:"brew_command"`
	PayloadSchemaVersion *int32  `json:"payload_schema_version" example:"1"`
}

type ActionDispatchRequest struct {
	NodeId     string          `json:"node_id" example:"5e8a1c3f-2b7d-4f6a-9c1e-3a8b6d2f4e09"`
	Payload    json.RawMessage `json:"payload" swaggertype:"object"`
	ExecutedAt *time.Time      `json:"executed_at" example:"2026-07-28T08:15:00Z"`
}
