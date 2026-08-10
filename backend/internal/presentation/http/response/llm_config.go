package presentationhttpresponse

import (
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

// LlmConfigResponse never carries the API key in any form — not full, not
// partial. Only whether one is currently set.
type LlmConfigResponse struct {
	Provider  string     `json:"provider" example:"CLAUDE"`
	Model     string     `json:"model" example:"claude-opus-5"`
	BaseURL   *string    `json:"base_url,omitempty" example:"https://api.anthropic.com"`
	ApiKeySet bool       `json:"api_key_set" example:"true"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	UpdatedBy *string    `json:"updated_by,omitempty" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
}

func LlmConfig(model *domainmodels.LlmConfig) LlmConfigResponse {
	if model == nil {
		return LlmConfigResponse{ApiKeySet: false}
	}
	updatedAt := model.UpdatedAt
	return LlmConfigResponse{
		Provider:  string(model.Provider),
		Model:     model.Model,
		BaseURL:   model.BaseURL,
		ApiKeySet: len(model.ApiKeyEncrypted) > 0,
		UpdatedAt: &updatedAt,
		UpdatedBy: UUIDPtrString(model.UpdatedBy),
	}
}

type LlmConnectionStatusResponse struct {
	Status string `json:"status" example:"CONNECTED"`
}
