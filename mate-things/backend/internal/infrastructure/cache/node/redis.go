package infrastructurecachenode

import (
	"context"

	domaincontractscache "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/cache"
	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	infrastructurecacheshared "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/cache/shared"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type redisImpl struct {
	infrastructurecacheshared.BaseRedis
}

func NewRedisImpl(client redis.UniversalClient, namespace string, ttl infrastructurecacheshared.TtlConfig) domaincontractscache.Node {
	return &redisImpl{BaseRedis: infrastructurecacheshared.NewRedis(client, namespace, "node", ttl)}
}

func (r *redisImpl) GetById(ctx context.Context, id uuid.UUID) (*domainmodels.Node, bool, error) {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return nil, false, err
	}
	var item domainmodels.Node
	hit, err := r.Get(ctx, key, &item)
	return &item, hit, err
}

func (r *redisImpl) SetById(ctx context.Context, id uuid.UUID, node *domainmodels.Node) error {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return err
	}
	return r.Set(ctx, key, node, r.IdentityTtl())
}

func (r *redisImpl) DeleteById(ctx context.Context, id uuid.UUID) error {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return err
	}
	return r.Delete(ctx, key)
}

func (r *redisImpl) GetByDeviceId(ctx context.Context, deviceId string) (*domainmodels.Node, bool, error) {
	key, err := r.ScopedKey(ctx, "identity", "device_id", infrastructurecacheshared.StringPart(deviceId))
	if err != nil {
		return nil, false, err
	}
	var item domainmodels.Node
	hit, err := r.Get(ctx, key, &item)
	return &item, hit, err
}

func (r *redisImpl) SetByDeviceId(ctx context.Context, deviceId string, node *domainmodels.Node) error {
	key, err := r.ScopedKey(ctx, "identity", "device_id", infrastructurecacheshared.StringPart(deviceId))
	if err != nil {
		return err
	}
	return r.Set(ctx, key, node, r.IdentityTtl())
}

func (r *redisImpl) DeleteByDeviceId(ctx context.Context, deviceId string) error {
	key, err := r.ScopedKey(ctx, "identity", "device_id", infrastructurecacheshared.StringPart(deviceId))
	if err != nil {
		return err
	}
	return r.Delete(ctx, key)
}

func (r *redisImpl) GetPagination(ctx context.Context, page int, limit int, search *string, nodeClassId *uuid.UUID, firmwareName *string) (domaincontractscache.Pagination[domainmodels.Node], bool, error) {
	key, err := r.paginationKey(ctx, page, limit, search, nodeClassId, firmwareName)
	if err != nil {
		return domaincontractscache.Pagination[domainmodels.Node]{}, false, err
	}
	var pagination domaincontractscache.Pagination[domainmodels.Node]
	hit, err := r.Get(ctx, key, &pagination)
	return pagination, hit, err
}

func (r *redisImpl) SetPagination(ctx context.Context, page int, limit int, search *string, nodeClassId *uuid.UUID, firmwareName *string, pagination domaincontractscache.Pagination[domainmodels.Node]) error {
	key, err := r.paginationKey(ctx, page, limit, search, nodeClassId, firmwareName)
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

func (r *redisImpl) paginationKey(ctx context.Context, page int, limit int, search *string, nodeClassId *uuid.UUID, firmwareName *string) (string, error) {
	hash, err := infrastructurecacheshared.HashPart(page, limit, search, nodeClassId, firmwareName)
	if err != nil {
		return "", err
	}
	return r.ScopedKey(ctx, "pagination", hash)
}
