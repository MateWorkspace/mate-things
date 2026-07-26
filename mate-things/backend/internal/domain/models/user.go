package domainmodels

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id           uuid.UUID       `db:"id" json:"id"`
	RoleId       uuid.UUID       `db:"role_id" json:"role_id"`
	Name         string          `db:"name" json:"name"`
	Bio          string          `db:"bio" json:"bio"`
	Username     string          `db:"username" json:"username"`
	PasswordHash string          `db:"password_hash" json:"-"`
	Preferences  json.RawMessage `db:"preferences" json:"preferences"`
	CreatedAt    time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt    *time.Time      `db:"updated_at" json:"updated_at,omitempty"`
	DeletedAt    *time.Time      `db:"deleted_at" json:"deleted_at,omitempty"`
	CreatedBy    *uuid.UUID      `db:"created_by" json:"created_by,omitempty"`
	UpdatedBy    *uuid.UUID      `db:"updated_by" json:"updated_by,omitempty"`
	DeletedBy    *uuid.UUID      `db:"deleted_by" json:"deleted_by,omitempty"`
}
