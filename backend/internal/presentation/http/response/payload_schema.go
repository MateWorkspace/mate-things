package presentationhttpresponse

import (
	"encoding/json"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type PayloadSchemaResponse struct {
	Id          string          `json:"id" example:"b2f5d8a1-3c6e-4f9b-8d2a-7e1f4b6c9d03"`
	Name        string          `json:"name" example:"brew_command"`
	Version     int32           `json:"version" example:"1"`
	Definition  json.RawMessage `json:"definition" swaggertype:"object"`
	ValidFrom   time.Time       `json:"valid_from" example:"2026-07-01T00:00:00Z"`
	ValidTo     *time.Time      `json:"valid_to,omitempty" example:"2027-07-01T00:00:00Z"`
	Preferences json.RawMessage `json:"preferences" swaggertype:"object"`
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
