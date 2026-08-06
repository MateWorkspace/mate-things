package infrastructurecacheapikey

import (
	"context"

	domaincontractscache "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/cache"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructurecacheshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/cache/shared"
	"github.com/redis/go-redis/v9"
)

type redisImpl struct {
	infrastructurecacheshared.BaseRedis
}

func NewRedisImpl(client redis.UniversalClient, namespace string, ttl infrastructurecacheshared.TtlConfig) domaincontractscache.ApiKey {
	return &redisImpl{BaseRedis: infrastructurecacheshared.NewRedis(client, namespace, "api_key", ttl)}
}

func (r *redisImpl) GetByKeyHash(ctx context.Context, keyHash string) (*domainmodels.ApiKey, bool, error) {
	key, err := r.ScopedKey(ctx, "identity", "key_hash", infrastructurecacheshared.StringPart(keyHash))
	if err != nil {
		return nil, false, err
	}
	var item cachedApiKey
	hit, err := r.Get(ctx, key, &item)
	return item.toDomain(), hit, err
}

func (r *redisImpl) SetByKeyHash(ctx context.Context, keyHash string, apiKey *domainmodels.ApiKey) error {
	key, err := r.ScopedKey(ctx, "identity", "key_hash", infrastructurecacheshared.StringPart(keyHash))
	if err != nil {
		return err
	}
	return r.Set(ctx, key, newCachedApiKey(apiKey), r.IdentityTtl())
}

func (r *redisImpl) GetPagination(ctx context.Context, page int, limit int, search *string, status *string) (domaincontractscache.Pagination[domainmodels.ApiKeyWithUser], bool, error) {
	key, err := r.paginationKey(ctx, page, limit, search, status)
	if err != nil {
		return domaincontractscache.Pagination[domainmodels.ApiKeyWithUser]{}, false, err
	}
	var pagination domaincontractscache.Pagination[domainmodels.ApiKeyWithUser]
	hit, err := r.Get(ctx, key, &pagination)
	return pagination, hit, err
}

func (r *redisImpl) SetPagination(ctx context.Context, page int, limit int, search *string, status *string, pagination domaincontractscache.Pagination[domainmodels.ApiKeyWithUser]) error {
	key, err := r.paginationKey(ctx, page, limit, search, status)
	if err != nil {
		return err
	}
	return r.Set(ctx, key, pagination, r.PaginationTtl())
}

func (r *redisImpl) InvalidatePagination(ctx context.Context) error {
	return r.Invalidate(ctx, "pagination")
}

func (r *redisImpl) InvalidateAll(ctx context.Context) error {
	return r.Invalidate(ctx, "identity", "pagination")
}

func (r *redisImpl) paginationKey(ctx context.Context, page int, limit int, search *string, status *string) (string, error) {
	hash, err := infrastructurecacheshared.HashPart(page, limit, search, status)
	if err != nil {
		return "", err
	}
	return r.ScopedKey(ctx, "pagination", hash)
}
