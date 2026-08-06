package domainmodels

import (
	"time"

	"github.com/google/uuid"
)

type ApiKey struct {
	Id          uuid.UUID  `db:"id" json:"id"`
	UserId      uuid.UUID  `db:"user_id" json:"user_id"`
	KeyHash     string     `db:"key_hash" json:"-"`
	KeyLastFour string     `db:"key_last_four" json:"key_last_four"`
	ExpiresAt   *time.Time `db:"expires_at" json:"expires_at,omitempty"`
	RevokedAt   *time.Time `db:"revoked_at" json:"revoked_at,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at" json:"updated_at,omitempty"`
	CreatedBy   *uuid.UUID `db:"created_by" json:"created_by,omitempty"`
	UpdatedBy   *uuid.UUID `db:"updated_by" json:"updated_by,omitempty"`
}

type ApiKeyWithUser struct {
	ApiKey
	UserName     string `json:"user_name"`
	UserUsername string `json:"user_username"`
}
