package infrastructurecacheapikey

import (
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

// cachedApiKey mirrors domainmodels.ApiKey but serializes key_hash.
//
// domainmodels.ApiKey tags KeyHash as json:"-" so it never leaks through
// HTTP responses. Caching that struct directly with encoding/json silently
// drops the hash from the cached blob. This type exists solely to
// round-trip the full row through Redis.
type cachedApiKey struct {
	Id          uuid.UUID  `json:"id"`
	UserId      uuid.UUID  `json:"user_id"`
	KeyHash     string     `json:"key_hash"`
	KeyLastFour string     `json:"key_last_four"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
	UpdatedBy   *uuid.UUID `json:"updated_by,omitempty"`
}

func newCachedApiKey(apiKey *domainmodels.ApiKey) *cachedApiKey {
	return &cachedApiKey{
		Id:          apiKey.Id,
		UserId:      apiKey.UserId,
		KeyHash:     apiKey.KeyHash,
		KeyLastFour: apiKey.KeyLastFour,
		ExpiresAt:   apiKey.ExpiresAt,
		RevokedAt:   apiKey.RevokedAt,
		CreatedAt:   apiKey.CreatedAt,
		UpdatedAt:   apiKey.UpdatedAt,
		CreatedBy:   apiKey.CreatedBy,
		UpdatedBy:   apiKey.UpdatedBy,
	}
}

func (c *cachedApiKey) toDomain() *domainmodels.ApiKey {
	return &domainmodels.ApiKey{
		Id:          c.Id,
		UserId:      c.UserId,
		KeyHash:     c.KeyHash,
		KeyLastFour: c.KeyLastFour,
		ExpiresAt:   c.ExpiresAt,
		RevokedAt:   c.RevokedAt,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
		CreatedBy:   c.CreatedBy,
		UpdatedBy:   c.UpdatedBy,
	}
}
