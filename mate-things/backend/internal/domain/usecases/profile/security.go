package domainusecasesprofile

import (
	"context"

	"github.com/google/uuid"
)

type Security interface {
	ChangePassword(ctx context.Context, request ChangePasswordRequest) error
}

type ChangePasswordRequest struct {
	UserId          uuid.UUID
	CurrentPassword string
	NewPassword     string
	UpdatedBy       *uuid.UUID
}
