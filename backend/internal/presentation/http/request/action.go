package presentationhttprequest

import (
	"encoding/json"
	"time"
)

type ActionPostRequest struct {
	NodeClassId          string  `json:"node_class_id"`
	Name                 string  `json:"name"`
	Description          *string `json:"description"`
	PayloadSchemaName    string  `json:"payload_schema_name"`
	PayloadSchemaVersion int32   `json:"payload_schema_version"`
}

type ActionPatchRequest struct {
	NodeClassId          *string `json:"node_class_id"`
	Name                 *string `json:"name"`
	Description          *string `json:"description"`
	PayloadSchemaName    *string `json:"payload_schema_name"`
	PayloadSchemaVersion *int32  `json:"payload_schema_version"`
}

type ActionDispatchRequest struct {
	NodeId     string          `json:"node_id"`
	Payload    json.RawMessage `json:"payload"`
	ExecutedAt *time.Time      `json:"executed_at"`
}
