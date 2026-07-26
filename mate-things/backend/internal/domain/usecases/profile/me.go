package domainusecasesprofile

import (
	"context"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type Me interface {
	GetProfile(ctx context.Context, request GetProfileRequest) (*domainmodels.User, error)
	GetPermissions(ctx context.Context, request GetProfilePermissionsRequest) ([]domainmodels.Permission, error)
}

type GetProfileRequest struct {
	UserId uuid.UUID
}

type GetProfilePermissionsRequest struct {
	UserId uuid.UUID
}
