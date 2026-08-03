package domaincontractscache

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type Action interface {
	GetById(ctx context.Context, id uuid.UUID) (action *domainmodels.Action, hit bool, err error)
	SetById(ctx context.Context, id uuid.UUID, action *domainmodels.Action) error
	DeleteById(ctx context.Context, id uuid.UUID) error
	GetByName(ctx context.Context, name string) (action *domainmodels.Action, hit bool, err error)
	SetByName(ctx context.Context, name string, action *domainmodels.Action) error
	DeleteByName(ctx context.Context, name string) error
	GetPagination(ctx context.Context, page int, limit int, search *string, nodeClassId *uuid.UUID, payloadSchemaName *string, payloadSchemaVersion *int32) (pagination Pagination[domainmodels.ActionListItem], hit bool, err error)
	SetPagination(ctx context.Context, page int, limit int, search *string, nodeClassId *uuid.UUID, payloadSchemaName *string, payloadSchemaVersion *int32, pagination Pagination[domainmodels.ActionListItem]) error
	InvalidatePagination(ctx context.Context) error
	InvalidateAll(ctx context.Context) error
}
