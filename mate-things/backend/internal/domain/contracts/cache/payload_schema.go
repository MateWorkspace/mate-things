package domaincontractscache

import (
	"context"
	"time"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type PayloadSchema interface {
	GetById(ctx context.Context, id uuid.UUID) (payloadSchema *domainmodels.PayloadSchema, hit bool, err error)
	SetById(ctx context.Context, id uuid.UUID, payloadSchema *domainmodels.PayloadSchema) error
	DeleteById(ctx context.Context, id uuid.UUID) error
	GetByNameAndVersion(ctx context.Context, name string, version int32) (payloadSchema *domainmodels.PayloadSchema, hit bool, err error)
	SetByNameAndVersion(ctx context.Context, name string, version int32, payloadSchema *domainmodels.PayloadSchema) error
	DeleteByNameAndVersion(ctx context.Context, name string, version int32) error
	GetLatestByName(ctx context.Context, name string) (payloadSchema *domainmodels.PayloadSchema, hit bool, err error)
	SetLatestByName(ctx context.Context, name string, payloadSchema *domainmodels.PayloadSchema) error
	DeleteLatestByName(ctx context.Context, name string) error
	GetPagination(ctx context.Context, page int, limit int, search *string, validAt *time.Time) (pagination Pagination[domainmodels.PayloadSchema], hit bool, err error)
	SetPagination(ctx context.Context, page int, limit int, search *string, validAt *time.Time, pagination Pagination[domainmodels.PayloadSchema]) error
	InvalidatePagination(ctx context.Context) error
	InvalidateAll(ctx context.Context) error
}
