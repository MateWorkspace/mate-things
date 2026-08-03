package infrastructurecachenodeclass

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

func NewRedisImpl(client redis.UniversalClient, namespace string, ttl infrastructurecacheshared.TtlConfig) domaincontractscache.NodeClass {
	return &redisImpl{BaseRedis: infrastructurecacheshared.NewRedis(client, namespace, "node_class", ttl)}
}

func (r *redisImpl) GetById(ctx context.Context, id uuid.UUID) (*domainmodels.NodeClass, bool, error) {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return nil, false, err
	}
	var item domainmodels.NodeClass
	hit, err := r.Get(ctx, key, &item)
	return &item, hit, err
}

func (r *redisImpl) SetById(ctx context.Context, id uuid.UUID, nodeClass *domainmodels.NodeClass) error {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return err
	}
	return r.Set(ctx, key, nodeClass, r.IdentityTtl())
}

func (r *redisImpl) DeleteById(ctx context.Context, id uuid.UUID) error {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return err
	}
	return r.Delete(ctx, key)
}

func (r *redisImpl) GetByName(ctx context.Context, name string) (*domainmodels.NodeClass, bool, error) {
	key, err := r.ScopedKey(ctx, "identity", "name", infrastructurecacheshared.StringPart(name))
	if err != nil {
		return nil, false, err
	}
	var item domainmodels.NodeClass
	hit, err := r.Get(ctx, key, &item)
	return &item, hit, err
}

func (r *redisImpl) SetByName(ctx context.Context, name string, nodeClass *domainmodels.NodeClass) error {
	key, err := r.ScopedKey(ctx, "identity", "name", infrastructurecacheshared.StringPart(name))
	if err != nil {
		return err
	}
	return r.Set(ctx, key, nodeClass, r.IdentityTtl())
}

func (r *redisImpl) DeleteByName(ctx context.Context, name string) error {
	key, err := r.ScopedKey(ctx, "identity", "name", infrastructurecacheshared.StringPart(name))
	if err != nil {
		return err
	}
	return r.Delete(ctx, key)
}

func (r *redisImpl) GetPagination(ctx context.Context, page int, limit int, search *string) (domaincontractscache.Pagination[domainmodels.NodeClass], bool, error) {
	key, err := r.paginationKey(ctx, page, limit, search)
	if err != nil {
		return domaincontractscache.Pagination[domainmodels.NodeClass]{}, false, err
	}
	var pagination domaincontractscache.Pagination[domainmodels.NodeClass]
	hit, err := r.Get(ctx, key, &pagination)
	return pagination, hit, err
}

func (r *redisImpl) SetPagination(ctx context.Context, page int, limit int, search *string, pagination domaincontractscache.Pagination[domainmodels.NodeClass]) error {
	key, err := r.paginationKey(ctx, page, limit, search)
	if err != nil {
		return err
	}
	return r.Set(ctx, key, pagination, r.PaginationTtl())
}

func (r *redisImpl) GetActions(ctx context.Context, nodeClassId uuid.UUID) ([]domainmodels.Action, bool, error) {
	key, err := r.ScopedKey(ctx, "actions", "node_class", nodeClassId.String())
	if err != nil {
		return nil, false, err
	}
	var items []domainmodels.Action
	hit, err := r.Get(ctx, key, &items)
	return items, hit, err
}

func (r *redisImpl) SetActions(ctx context.Context, nodeClassId uuid.UUID, actions []domainmodels.Action) error {
	key, err := r.ScopedKey(ctx, "actions", "node_class", nodeClassId.String())
	if err != nil {
		return err
	}
	return r.Set(ctx, key, actions, r.RelationTtl())
}

func (r *redisImpl) DeleteActions(ctx context.Context, nodeClassId uuid.UUID) error {
	key, err := r.ScopedKey(ctx, "actions", "node_class", nodeClassId.String())
	if err != nil {
		return err
	}
	return r.Delete(ctx, key)
}

func (r *redisImpl) InvalidateActions(ctx context.Context) error {
	return r.Invalidate(ctx, "actions")
}

func (r *redisImpl) InvalidatePagination(ctx context.Context) error {
	return r.Invalidate(ctx, "pagination")
}

func (r *redisImpl) InvalidateAll(ctx context.Context) error {
	return r.Invalidate(ctx, "identity", "actions", "pagination")
}

func (r *redisImpl) paginationKey(ctx context.Context, page int, limit int, search *string) (string, error) {
	hash, err := infrastructurecacheshared.HashPart(page, limit, search)
	if err != nil {
		return "", err
	}
	return r.ScopedKey(ctx, "pagination", hash)
}
