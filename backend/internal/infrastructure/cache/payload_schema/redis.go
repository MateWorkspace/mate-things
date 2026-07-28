package infrastructurecachepayloadschema

import (
	"context"
	"time"

	domaincontractscache "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/cache"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructurecacheshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/cache/shared"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type redisImpl struct {
	infrastructurecacheshared.BaseRedis
}

func NewRedisImpl(client redis.UniversalClient, namespace string, ttl infrastructurecacheshared.TtlConfig) domaincontractscache.PayloadSchema {
	return &redisImpl{BaseRedis: infrastructurecacheshared.NewRedis(client, namespace, "payload_schema", ttl)}
}

func (r *redisImpl) GetById(ctx context.Context, id uuid.UUID) (*domainmodels.PayloadSchema, bool, error) {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return nil, false, err
	}
	var item domainmodels.PayloadSchema
	hit, err := r.Get(ctx, key, &item)
	return &item, hit, err
}

func (r *redisImpl) SetById(ctx context.Context, id uuid.UUID, payloadSchema *domainmodels.PayloadSchema) error {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return err
	}
	return r.Set(ctx, key, payloadSchema, r.IdentityTtl())
}

func (r *redisImpl) DeleteById(ctx context.Context, id uuid.UUID) error {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return err
	}
	return r.Delete(ctx, key)
}

func (r *redisImpl) GetByNameAndVersion(ctx context.Context, name string, version int32) (*domainmodels.PayloadSchema, bool, error) {
	key, err := r.ScopedKey(ctx, "identity", "name_version", infrastructurecacheshared.StringPart(name), infrastructurecacheshared.Int32Part(version))
	if err != nil {
		return nil, false, err
	}
	var item domainmodels.PayloadSchema
	hit, err := r.Get(ctx, key, &item)
	return &item, hit, err
}

func (r *redisImpl) SetByNameAndVersion(ctx context.Context, name string, version int32, payloadSchema *domainmodels.PayloadSchema) error {
	key, err := r.ScopedKey(ctx, "identity", "name_version", infrastructurecacheshared.StringPart(name), infrastructurecacheshared.Int32Part(version))
	if err != nil {
		return err
	}
	return r.Set(ctx, key, payloadSchema, r.IdentityTtl())
}

func (r *redisImpl) DeleteByNameAndVersion(ctx context.Context, name string, version int32) error {
	key, err := r.ScopedKey(ctx, "identity", "name_version", infrastructurecacheshared.StringPart(name), infrastructurecacheshared.Int32Part(version))
	if err != nil {
		return err
	}
	return r.Delete(ctx, key)
}

func (r *redisImpl) GetLatestByName(ctx context.Context, name string) (*domainmodels.PayloadSchema, bool, error) {
	key, err := r.ScopedKey(ctx, "identity", "latest", infrastructurecacheshared.StringPart(name))
	if err != nil {
		return nil, false, err
	}
	var item domainmodels.PayloadSchema
	hit, err := r.Get(ctx, key, &item)
	return &item, hit, err
}

func (r *redisImpl) SetLatestByName(ctx context.Context, name string, payloadSchema *domainmodels.PayloadSchema) error {
	key, err := r.ScopedKey(ctx, "identity", "latest", infrastructurecacheshared.StringPart(name))
	if err != nil {
		return err
	}
	return r.Set(ctx, key, payloadSchema, r.IdentityTtl())
}

func (r *redisImpl) DeleteLatestByName(ctx context.Context, name string) error {
	key, err := r.ScopedKey(ctx, "identity", "latest", infrastructurecacheshared.StringPart(name))
	if err != nil {
		return err
	}
	return r.Delete(ctx, key)
}

func (r *redisImpl) GetPagination(ctx context.Context, page int, limit int, search *string, validAt *time.Time) (domaincontractscache.Pagination[domainmodels.PayloadSchema], bool, error) {
	key, err := r.paginationKey(ctx, page, limit, search, validAt)
	if err != nil {
		return domaincontractscache.Pagination[domainmodels.PayloadSchema]{}, false, err
	}
	var pagination domaincontractscache.Pagination[domainmodels.PayloadSchema]
	hit, err := r.Get(ctx, key, &pagination)
	return pagination, hit, err
}

func (r *redisImpl) SetPagination(ctx context.Context, page int, limit int, search *string, validAt *time.Time, pagination domaincontractscache.Pagination[domainmodels.PayloadSchema]) error {
	key, err := r.paginationKey(ctx, page, limit, search, validAt)
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

func (r *redisImpl) paginationKey(ctx context.Context, page int, limit int, search *string, validAt *time.Time) (string, error) {
	hash, err := infrastructurecacheshared.HashPart(page, limit, search, validAt)
	if err != nil {
		return "", err
	}
	return r.ScopedKey(ctx, "pagination", hash)
}
