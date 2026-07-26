package domaincontractscache

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type NodeClass interface {
	GetById(ctx context.Context, id uuid.UUID) (nodeClass *domainmodels.NodeClass, hit bool, err error)
	SetById(ctx context.Context, id uuid.UUID, nodeClass *domainmodels.NodeClass) error
	DeleteById(ctx context.Context, id uuid.UUID) error
	GetByName(ctx context.Context, name string) (nodeClass *domainmodels.NodeClass, hit bool, err error)
	SetByName(ctx context.Context, name string, nodeClass *domainmodels.NodeClass) error
	DeleteByName(ctx context.Context, name string) error
	GetPagination(ctx context.Context, page int, limit int, search *string) (pagination Pagination[domainmodels.NodeClass], hit bool, err error)
	SetPagination(ctx context.Context, page int, limit int, search *string, pagination Pagination[domainmodels.NodeClass]) error
	InvalidatePagination(ctx context.Context) error
	InvalidateAll(ctx context.Context) error
}
