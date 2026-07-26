package infrastructurecachefirmware

import (
	"context"

	domaincontractscache "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/cache"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructurecacheshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/cache/shared"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type redisImpl struct {
	infrastructurecacheshared.BaseRedis
}

func NewRedisImpl(client redis.UniversalClient, namespace string, ttl infrastructurecacheshared.TtlConfig) domaincontractscache.Firmware {
	return &redisImpl{BaseRedis: infrastructurecacheshared.NewRedis(client, namespace, "firmware", ttl)}
}

func (r *redisImpl) GetById(ctx context.Context, id uuid.UUID) (*domainmodels.Firmware, bool, error) {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return nil, false, err
	}
	var item domainmodels.Firmware
	hit, err := r.Get(ctx, key, &item)
	return &item, hit, err
}

func (r *redisImpl) SetById(ctx context.Context, id uuid.UUID, firmware *domainmodels.Firmware) error {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return err
	}
	return r.Set(ctx, key, firmware, r.IdentityTtl())
}

func (r *redisImpl) DeleteById(ctx context.Context, id uuid.UUID) error {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return err
	}
	return r.Delete(ctx, key)
}

func (r *redisImpl) GetByName(ctx context.Context, name string) (*domainmodels.Firmware, bool, error) {
	key, err := r.ScopedKey(ctx, "identity", "name", infrastructurecacheshared.StringPart(name))
	if err != nil {
		return nil, false, err
	}
	var item domainmodels.Firmware
	hit, err := r.Get(ctx, key, &item)
	return &item, hit, err
}

func (r *redisImpl) SetByName(ctx context.Context, name string, firmware *domainmodels.Firmware) error {
	key, err := r.ScopedKey(ctx, "identity", "name", infrastructurecacheshared.StringPart(name))
	if err != nil {
		return err
	}
	return r.Set(ctx, key, firmware, r.IdentityTtl())
}

func (r *redisImpl) DeleteByName(ctx context.Context, name string) error {
	key, err := r.ScopedKey(ctx, "identity", "name", infrastructurecacheshared.StringPart(name))
	if err != nil {
		return err
	}
	return r.Delete(ctx, key)
}

func (r *redisImpl) GetPagination(ctx context.Context, page int, limit int, search *string, nodeClassId *uuid.UUID) (domaincontractscache.Pagination[domainmodels.Firmware], bool, error) {
	key, err := r.paginationKey(ctx, page, limit, search, nodeClassId)
	if err != nil {
		return domaincontractscache.Pagination[domainmodels.Firmware]{}, false, err
	}
	var pagination domaincontractscache.Pagination[domainmodels.Firmware]
	hit, err := r.Get(ctx, key, &pagination)
	return pagination, hit, err
}

func (r *redisImpl) SetPagination(ctx context.Context, page int, limit int, search *string, nodeClassId *uuid.UUID, pagination domaincontractscache.Pagination[domainmodels.Firmware]) error {
	key, err := r.paginationKey(ctx, page, limit, search, nodeClassId)
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

func (r *redisImpl) paginationKey(ctx context.Context, page int, limit int, search *string, nodeClassId *uuid.UUID) (string, error) {
	hash, err := infrastructurecacheshared.HashPart(page, limit, search, nodeClassId)
	if err != nil {
		return "", err
	}
	return r.ScopedKey(ctx, "pagination", hash)
}
