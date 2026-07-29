package presentationhttpresponse

import (
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type FirmwareConfigParameterResponse struct {
	Key       string `json:"key" example:"mqtt_host"`
	ValueType string `json:"value_type" example:"string"`
}

func FirmwareConfigParameters(params []domainmodels.FirmwareConfigParameter) []FirmwareConfigParameterResponse {
	result := make([]FirmwareConfigParameterResponse, 0, len(params))
	for _, param := range params {
		result = append(result, FirmwareConfigParameterResponse{
			Key:       param.Key,
			ValueType: param.ValueType,
		})
	}
	return result
}

// NodeConfigValueResponse echoes config values verbatim, including secret-ish
// keys such as `mqtt_pass`, which are stored as plaintext. Accepted tradeoff:
// the endpoint is gated behind `node_config:get` (super/admin only, not user).
type NodeConfigValueResponse struct {
	Key       string `json:"key" example:"mqtt_host"`
	Value     string `json:"value" example:"broker.example.com"`
	UpdatedAt string `json:"updated_at,omitempty" example:"2026-07-29T14:05:00Z"`
}

func NodeConfigValues(values []domainmodels.NodeConfigValue) []NodeConfigValueResponse {
	result := make([]NodeConfigValueResponse, 0, len(values))
	for _, value := range values {
		updatedAt := ""
		if value.UpdatedAt != nil {
			updatedAt = value.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		result = append(result, NodeConfigValueResponse{
			Key:       value.Key,
			Value:     value.Value,
			UpdatedAt: updatedAt,
		})
	}
	return result
}
