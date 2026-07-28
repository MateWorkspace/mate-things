package infrastructurecacheuser

import (
	"encoding/json"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

// cachedUser mirrors domainmodels.User but serializes password_hash.
//
// domainmodels.User tags PasswordHash as json:"-" so it never leaks through
// HTTP responses. Caching that struct directly with encoding/json silently
// drops the hash from the cached blob, so any cache-hit read (including the
// one login performs) ends up comparing against an empty hash. This type
// exists solely to round-trip the full row through Redis.
type cachedUser struct {
	Id           uuid.UUID       `json:"id"`
	RoleId       uuid.UUID       `json:"role_id"`
	Name         string          `json:"name"`
	Bio          string          `json:"bio"`
	Username     string          `json:"username"`
	PasswordHash string          `json:"password_hash"`
	Preferences  json.RawMessage `json:"preferences"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    *time.Time      `json:"updated_at,omitempty"`
	DeletedAt    *time.Time      `json:"deleted_at,omitempty"`
	CreatedBy    *uuid.UUID      `json:"created_by,omitempty"`
	UpdatedBy    *uuid.UUID      `json:"updated_by,omitempty"`
	DeletedBy    *uuid.UUID      `json:"deleted_by,omitempty"`
}

func newCachedUser(user *domainmodels.User) *cachedUser {
	return &cachedUser{
		Id:           user.Id,
		RoleId:       user.RoleId,
		Name:         user.Name,
		Bio:          user.Bio,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Preferences:  user.Preferences,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		DeletedAt:    user.DeletedAt,
		CreatedBy:    user.CreatedBy,
		UpdatedBy:    user.UpdatedBy,
		DeletedBy:    user.DeletedBy,
	}
}

func (c *cachedUser) toDomain() *domainmodels.User {
	return &domainmodels.User{
		Id:           c.Id,
		RoleId:       c.RoleId,
		Name:         c.Name,
		Bio:          c.Bio,
		Username:     c.Username,
		PasswordHash: c.PasswordHash,
		Preferences:  c.Preferences,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
		DeletedAt:    c.DeletedAt,
		CreatedBy:    c.CreatedBy,
		UpdatedBy:    c.UpdatedBy,
		DeletedBy:    c.DeletedBy,
	}
}
