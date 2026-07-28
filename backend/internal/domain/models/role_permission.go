package domainmodels

import (
	"time"

	"github.com/google/uuid"
)

type RolePermission struct {
	Id           uuid.UUID  `db:"id" json:"id"`
	RoleId       uuid.UUID  `db:"role_id" json:"role_id"`
	PermissionId uuid.UUID  `db:"permission_id" json:"permission_id"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	CreatedBy    *uuid.UUID `db:"created_by" json:"created_by,omitempty"`
}
