package domainusecasesadmin

import (
	"context"
	"encoding/json"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type SchemaRegistry interface {
	Create(ctx context.Context, request CreatePayloadSchemaRequest) (uuid.UUID, error)
	ReadById(ctx context.Context, request ReadPayloadSchemaByIdRequest) (*domainmodels.PayloadSchema, error)
	ReadByNameAndVersion(ctx context.Context, request ReadPayloadSchemaByNameAndVersionRequest) (*domainmodels.PayloadSchema, error)
	ReadLatestByName(ctx context.Context, request ReadLatestPayloadSchemaByNameRequest) (*domainmodels.PayloadSchema, error)
	ReadByPagination(ctx context.Context, request ReadPayloadSchemasByPaginationRequest) ([]domainmodels.PayloadSchema, int, error)
	UpdateById(ctx context.Context, request UpdatePayloadSchemaRequest) error
	DeleteById(ctx context.Context, request DeletePayloadSchemaRequest) error
}

type CreatePayloadSchemaRequest struct {
	Name       string
	Version    int32
	Definition json.RawMessage
	ValidFrom  *time.Time
	ValidTo    *time.Time
	CreatedBy  *uuid.UUID
}

type ReadPayloadSchemaByIdRequest struct {
	Id uuid.UUID
}

type ReadPayloadSchemaByNameAndVersionRequest struct {
	Name    string
	Version int32
}

type ReadLatestPayloadSchemaByNameRequest struct {
	Name string
}

type ReadPayloadSchemasByPaginationRequest struct {
	Page    int
	Limit   int
	Search  *string
	ValidAt *time.Time
}

type UpdatePayloadSchemaRequest struct {
	Id         uuid.UUID
	Name       *string
	Version    *int32
	Definition *json.RawMessage
	ValidFrom  *time.Time
	ValidTo    *time.Time
	UpdatedBy  *uuid.UUID
}

type DeletePayloadSchemaRequest struct {
	Id        uuid.UUID
	DeletedBy *uuid.UUID
}
