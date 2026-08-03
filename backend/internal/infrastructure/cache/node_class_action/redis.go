// backend/internal/infrastructure/cache/node_class_action/redis.go
package infrastructurecachenodeclassaction

import (
	"context"

	domaincontractscache "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/cache"
	infrastructurecacheshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/cache/shared"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type redisImpl struct {
	infrastructurecacheshared.BaseRedis
}

func NewRedisImpl(client redis.UniversalClient, namespace string, ttl infrastructurecacheshared.TtlConfig) domaincontractscache.NodeClassAction {
	return &redisImpl{BaseRedis: infrastructurecacheshared.NewRedis(client, namespace, "node_class_action", ttl)}
}

func (r *redisImpl) GetById(ctx context.Context, id uuid.UUID) (*domaincontractscache.NodeClassActionItem, bool, error) {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return nil, false, err
	}
	var item domaincontractscache.NodeClassActionItem
	hit, err := r.Get(ctx, key, &item)
	return &item, hit, err
}

func (r *redisImpl) SetById(ctx context.Context, id uuid.UUID, item *domaincontractscache.NodeClassActionItem) error {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return err
	}
	return r.Set(ctx, key, item, r.IdentityTtl())
}

func (r *redisImpl) DeleteById(ctx context.Context, id uuid.UUID) error {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return err
	}
	return r.Delete(ctx, key)
}

func (r *redisImpl) GetByNodeClassIdAndActionId(ctx context.Context, nodeClassId uuid.UUID, actionId uuid.UUID) (*domaincontractscache.NodeClassActionItem, bool, error) {
	key, err := r.ScopedKey(ctx, "identity", "node_class_action", nodeClassId.String(), actionId.String())
	if err != nil {
		return nil, false, err
	}
	var item domaincontractscache.NodeClassActionItem
	hit, err := r.Get(ctx, key, &item)
	return &item, hit, err
}

func (r *redisImpl) SetByNodeClassIdAndActionId(ctx context.Context, nodeClassId uuid.UUID, actionId uuid.UUID, item *domaincontractscache.NodeClassActionItem) error {
	key, err := r.ScopedKey(ctx, "identity", "node_class_action", nodeClassId.String(), actionId.String())
	if err != nil {
		return err
	}
	return r.Set(ctx, key, item, r.IdentityTtl())
}

func (r *redisImpl) DeleteByNodeClassIdAndActionId(ctx context.Context, nodeClassId uuid.UUID, actionId uuid.UUID) error {
	key, err := r.ScopedKey(ctx, "identity", "node_class_action", nodeClassId.String(), actionId.String())
	if err != nil {
		return err
	}
	return r.Delete(ctx, key)
}

func (r *redisImpl) GetPagination(ctx context.Context, page int, limit int, nodeClassId *uuid.UUID, actionId *uuid.UUID) (domaincontractscache.Pagination[domaincontractscache.NodeClassActionItem], bool, error) {
	key, err := r.paginationKey(ctx, page, limit, nodeClassId, actionId)
	if err != nil {
		return domaincontractscache.Pagination[domaincontractscache.NodeClassActionItem]{}, false, err
	}
	var pagination domaincontractscache.Pagination[domaincontractscache.NodeClassActionItem]
	hit, err := r.Get(ctx, key, &pagination)
	return pagination, hit, err
}

func (r *redisImpl) SetPagination(ctx context.Context, page int, limit int, nodeClassId *uuid.UUID, actionId *uuid.UUID, pagination domaincontractscache.Pagination[domaincontractscache.NodeClassActionItem]) error {
	key, err := r.paginationKey(ctx, page, limit, nodeClassId, actionId)
	if err != nil {
		return err
	}
	return r.Set(ctx, key, pagination, r.PaginationTtl())
}

func (r *redisImpl) InvalidatePagination(ctx context.Context) error {
	return r.Invalidate(ctx, "pagination")
}

func (r *redisImpl) InvalidateByNodeClassId(ctx context.Context, nodeClassId uuid.UUID) error {
	return r.InvalidateAll(ctx)
}

func (r *redisImpl) InvalidateByActionId(ctx context.Context, actionId uuid.UUID) error {
	return r.InvalidateAll(ctx)
}

func (r *redisImpl) InvalidateAll(ctx context.Context) error {
	return r.Invalidate(ctx, "identity", "pagination")
}

func (r *redisImpl) paginationKey(ctx context.Context, page int, limit int, nodeClassId *uuid.UUID, actionId *uuid.UUID) (string, error) {
	hash, err := infrastructurecacheshared.HashPart(page, limit, nodeClassId, actionId)
	if err != nil {
		return "", err
	}
	return r.ScopedKey(ctx, "pagination", hash)
}
