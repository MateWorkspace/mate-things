package presentationhttpresponse

import (
	"encoding/json"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type PayloadSchemaResponse struct {
	Id          string          `json:"id"`
	Name        string          `json:"name"`
	Version     int32           `json:"version"`
	Definition  json.RawMessage `json:"definition"`
	ValidFrom   time.Time       `json:"valid_from"`
	ValidTo     *time.Time      `json:"valid_to,omitempty"`
	Preferences json.RawMessage `json:"preferences"`
	AuditResponse
}

func PayloadSchema(payloadSchema domainmodels.PayloadSchema) PayloadSchemaResponse {
	return PayloadSchemaResponse{
		Id:          UUIDString(payloadSchema.Id),
		Name:        payloadSchema.Name,
		Version:     payloadSchema.Version,
		Definition:  NormalizeJSON(payloadSchema.Definition),
		ValidFrom:   payloadSchema.ValidFrom,
		ValidTo:     payloadSchema.ValidTo,
		Preferences: NormalizeJSON(payloadSchema.Preferences),
		AuditResponse: Audit(
			payloadSchema.CreatedAt,
			payloadSchema.UpdatedAt,
			payloadSchema.DeletedAt,
			payloadSchema.CreatedBy,
			payloadSchema.UpdatedBy,
			payloadSchema.DeletedBy,
		),
	}
}

func PayloadSchemas(payloadSchemas []domainmodels.PayloadSchema) []PayloadSchemaResponse {
	result := make([]PayloadSchemaResponse, 0, len(payloadSchemas))
	for _, payloadSchema := range payloadSchemas {
		result = append(result, PayloadSchema(payloadSchema))
	}
	return result
}
