package domaincontractscache

import (
	"context"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type Permission interface {
	GetById(ctx context.Context, id uuid.UUID) (permission *domainmodels.Permission, hit bool, err error)
	SetById(ctx context.Context, id uuid.UUID, permission *domainmodels.Permission) error
	DeleteById(ctx context.Context, id uuid.UUID) error
	GetByName(ctx context.Context, name string) (permission *domainmodels.Permission, hit bool, err error)
	SetByName(ctx context.Context, name string, permission *domainmodels.Permission) error
	DeleteByName(ctx context.Context, name string) error
	GetPagination(ctx context.Context, page int, limit int, search *string) (pagination Pagination[domainmodels.Permission], hit bool, err error)
	SetPagination(ctx context.Context, page int, limit int, search *string, pagination Pagination[domainmodels.Permission]) error
	InvalidatePagination(ctx context.Context) error
	InvalidateAll(ctx context.Context) error
}
