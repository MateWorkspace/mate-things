package presentationhttpresponse

import (
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type ApiKeyResponse struct {
	Id           string     `json:"id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	UserId       string     `json:"user_id" example:"7a3c2e10-4b1a-4b8e-9f3a-2b6e7c9d1a04"`
	UserName     string     `json:"user_name" example:"Grace Hopper"`
	UserUsername string     `json:"user_username" example:"grace.hopper"`
	KeyLastFour  string     `json:"key_last_four" example:"a91f"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty" example:"2027-01-01T00:00:00Z"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty" example:"2026-07-20T11:00:00Z"`
	CreatedAt    time.Time  `json:"created_at" example:"2026-06-15T09:30:00Z"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty" example:"2026-07-01T14:05:00Z"`
}

func ApiKey(apiKey domainmodels.ApiKeyWithUser) ApiKeyResponse {
	return ApiKeyResponse{
		Id:           UUIDString(apiKey.Id),
		UserId:       UUIDString(apiKey.UserId),
		UserName:     apiKey.UserName,
		UserUsername: apiKey.UserUsername,
		KeyLastFour:  apiKey.KeyLastFour,
		ExpiresAt:    apiKey.ExpiresAt,
		RevokedAt:    apiKey.RevokedAt,
		CreatedAt:    apiKey.CreatedAt,
		UpdatedAt:    apiKey.UpdatedAt,
	}
}

func ApiKeys(apiKeys []domainmodels.ApiKeyWithUser) []ApiKeyResponse {
	result := make([]ApiKeyResponse, 0, len(apiKeys))
	for _, apiKey := range apiKeys {
		result = append(result, ApiKey(apiKey))
	}
	return result
}

type ApiKeySecretResponse struct {
	Key string `json:"key" example:"mate_9f2ac1b4e7d3..."`
}
