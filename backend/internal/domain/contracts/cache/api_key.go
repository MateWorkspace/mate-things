package domaincontractscache

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type ApiKey interface {
	GetByKeyHash(ctx context.Context, keyHash string) (apiKey *domainmodels.ApiKey, hit bool, err error)
	SetByKeyHash(ctx context.Context, keyHash string, apiKey *domainmodels.ApiKey) error
	GetPagination(ctx context.Context, page int, limit int, search *string, status *string) (pagination Pagination[domainmodels.ApiKeyWithUser], hit bool, err error)
	SetPagination(ctx context.Context, page int, limit int, search *string, status *string, pagination Pagination[domainmodels.ApiKeyWithUser]) error
	InvalidatePagination(ctx context.Context) error
	InvalidateAll(ctx context.Context) error
}
