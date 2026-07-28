package domaincontractsrepository

import (
	"context"
	"encoding/json"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type PayloadSchema interface {
	Create(
		ctx context.Context,
		name string,
		version int32,
		definition json.RawMessage,
		validFrom *time.Time,
		validTo *time.Time,
		createdBy *uuid.UUID,
	) (id uuid.UUID, err error)

	ReadById(
		ctx context.Context,
		id uuid.UUID,
	) (payloadSchema *domainmodels.PayloadSchema, err error)

	ReadByNameAndVersion(
		ctx context.Context,
		name string,
		version int32,
	) (payloadSchema *domainmodels.PayloadSchema, err error)

	ReadLatestByName(
		ctx context.Context,
		name string,
	) (payloadSchema *domainmodels.PayloadSchema, err error)

	ReadByPagination(
		ctx context.Context,
		page int,
		limit int,
		search *string,
		validAt *time.Time,
	) (payloadSchemas []domainmodels.PayloadSchema, total int, err error)

	UpdateById(
		ctx context.Context,
		id uuid.UUID,
		name *string,
		version *int32,
		definition *json.RawMessage,
		validFrom *time.Time,
		validTo *time.Time,
		preferences *json.RawMessage,
		updatedBy *uuid.UUID,
	) (err error)

	DeleteById(
		ctx context.Context,
		id uuid.UUID,
		deletedBy *uuid.UUID,
	) (err error)
}
