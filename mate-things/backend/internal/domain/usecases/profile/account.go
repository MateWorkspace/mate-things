package domainusecasesprofile

import (
	"context"

	"github.com/google/uuid"
)

type Account interface {
	UpdateProfile(ctx context.Context, request UpdateProfileRequest) error
}

type UpdateProfileRequest struct {
	UserId    uuid.UUID
	Name      *string
	Bio       *string
	Username  *string
	UpdatedBy *uuid.UUID
}
