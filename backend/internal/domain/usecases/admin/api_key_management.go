package domainusecasesadmin

import (
	"context"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type ApiKeyManagement interface {
	Create(ctx context.Context, request CreateApiKeyRequest) (key string, err error)
	Regenerate(ctx context.Context, request RegenerateApiKeyRequest) (key string, err error)
	Revoke(ctx context.Context, request RevokeApiKeyRequest) error
	ReadByPagination(ctx context.Context, request ReadApiKeysByPaginationRequest) ([]domainmodels.ApiKeyWithUser, int, error)
	DeleteById(ctx context.Context, request DeleteApiKeyRequest) error
}

type CreateApiKeyRequest struct {
	UserId    uuid.UUID
	ExpiresAt *time.Time
	CreatedBy *uuid.UUID
}

type RegenerateApiKeyRequest struct {
	Id        uuid.UUID
	ExpiresAt *time.Time
	UpdatedBy *uuid.UUID
}

type RevokeApiKeyRequest struct {
	Id        uuid.UUID
	UpdatedBy *uuid.UUID
}

type ReadApiKeysByPaginationRequest struct {
	Page   int
	Limit  int
	Search *string
	Status *string
}

type DeleteApiKeyRequest struct {
	Id uuid.UUID
}
