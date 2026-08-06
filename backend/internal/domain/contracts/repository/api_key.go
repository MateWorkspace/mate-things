package domaincontractsrepository

import (
	"context"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type ApiKey interface {
	Create(
		ctx context.Context,
		userId uuid.UUID,
		keyHash string,
		keyLastFour string,
		expiresAt *time.Time,
		createdBy *uuid.UUID,
	) (id uuid.UUID, err error)

	ReadByKeyHash(
		ctx context.Context,
		keyHash string,
	) (apiKey *domainmodels.ApiKey, err error)

	ReadByPagination(
		ctx context.Context,
		page int,
		limit int,
		search *string,
		status *string,
	) (apiKeys []domainmodels.ApiKeyWithUser, total int, err error)

	Regenerate(
		ctx context.Context,
		id uuid.UUID,
		keyHash string,
		keyLastFour string,
		expiresAt *time.Time,
		updatedBy *uuid.UUID,
	) (err error)

	Revoke(
		ctx context.Context,
		id uuid.UUID,
		updatedBy *uuid.UUID,
	) (err error)

	DeleteById(
		ctx context.Context,
		id uuid.UUID,
	) (err error)
}
