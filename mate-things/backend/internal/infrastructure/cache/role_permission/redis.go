package infrastructurecacherolepermission

import (
	"context"

	domaincontractscache "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/cache"
	infrastructurecacheshared "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/cache/shared"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type redisImpl struct {
	infrastructurecacheshared.BaseRedis
}

func NewRedisImpl(client redis.UniversalClient, namespace string, ttl infrastructurecacheshared.TtlConfig) domaincontractscache.RolePermission {
	return &redisImpl{BaseRedis: infrastructurecacheshared.NewRedis(client, namespace, "role_permission", ttl)}
}

func (r *redisImpl) GetById(ctx context.Context, id uuid.UUID) (*domaincontractscache.RolePermissionItem, bool, error) {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return nil, false, err
	}
	var item domaincontractscache.RolePermissionItem
	hit, err := r.Get(ctx, key, &item)
	return &item, hit, err
}

func (r *redisImpl) SetById(ctx context.Context, id uuid.UUID, item *domaincontractscache.RolePermissionItem) error {
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

func (r *redisImpl) GetByRoleIdAndPermissionId(ctx context.Context, roleId uuid.UUID, permissionId uuid.UUID) (*domaincontractscache.RolePermissionItem, bool, error) {
	key, err := r.ScopedKey(ctx, "identity", "role_permission", roleId.String(), permissionId.String())
	if err != nil {
		return nil, false, err
	}
	var item domaincontractscache.RolePermissionItem
	hit, err := r.Get(ctx, key, &item)
	return &item, hit, err
}

func (r *redisImpl) SetByRoleIdAndPermissionId(ctx context.Context, roleId uuid.UUID, permissionId uuid.UUID, item *domaincontractscache.RolePermissionItem) error {
	key, err := r.ScopedKey(ctx, "identity", "role_permission", roleId.String(), permissionId.String())
	if err != nil {
		return err
	}
	return r.Set(ctx, key, item, r.IdentityTtl())
}

func (r *redisImpl) DeleteByRoleIdAndPermissionId(ctx context.Context, roleId uuid.UUID, permissionId uuid.UUID) error {
	key, err := r.ScopedKey(ctx, "identity", "role_permission", roleId.String(), permissionId.String())
	if err != nil {
		return err
	}
	return r.Delete(ctx, key)
}

func (r *redisImpl) GetPagination(ctx context.Context, page int, limit int, roleId *uuid.UUID, permissionId *uuid.UUID) (domaincontractscache.Pagination[domaincontractscache.RolePermissionItem], bool, error) {
	key, err := r.paginationKey(ctx, page, limit, roleId, permissionId)
	if err != nil {
		return domaincontractscache.Pagination[domaincontractscache.RolePermissionItem]{}, false, err
	}
	var pagination domaincontractscache.Pagination[domaincontractscache.RolePermissionItem]
	hit, err := r.Get(ctx, key, &pagination)
	return pagination, hit, err
}

func (r *redisImpl) SetPagination(ctx context.Context, page int, limit int, roleId *uuid.UUID, permissionId *uuid.UUID, pagination domaincontractscache.Pagination[domaincontractscache.RolePermissionItem]) error {
	key, err := r.paginationKey(ctx, page, limit, roleId, permissionId)
	if err != nil {
		return err
	}
	return r.Set(ctx, key, pagination, r.PaginationTtl())
}

func (r *redisImpl) InvalidatePagination(ctx context.Context) error {
	return r.Invalidate(ctx, "pagination")
}

func (r *redisImpl) InvalidateByRoleId(ctx context.Context, roleId uuid.UUID) error {
	return r.InvalidateAll(ctx)
}

func (r *redisImpl) InvalidateByPermissionId(ctx context.Context, permissionId uuid.UUID) error {
	return r.InvalidateAll(ctx)
}

func (r *redisImpl) InvalidateAll(ctx context.Context) error {
	return r.Invalidate(ctx, "identity", "pagination")
}

func (r *redisImpl) paginationKey(ctx context.Context, page int, limit int, roleId *uuid.UUID, permissionId *uuid.UUID) (string, error) {
	hash, err := infrastructurecacheshared.HashPart(page, limit, roleId, permissionId)
	if err != nil {
		return "", err
	}
	return r.ScopedKey(ctx, "pagination", hash)
}
