package infrastructurecacheuser

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

func NewRedisImpl(client redis.UniversalClient, namespace string, ttl infrastructurecacheshared.TtlConfig) domaincontractscache.User {
	return &redisImpl{BaseRedis: infrastructurecacheshared.NewRedis(client, namespace, "user", ttl)}
}

func (r *redisImpl) GetById(ctx context.Context, id uuid.UUID) (*domainmodels.User, bool, error) {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return nil, false, err
	}
	var item domainmodels.User
	hit, err := r.Get(ctx, key, &item)
	return &item, hit, err
}

func (r *redisImpl) SetById(ctx context.Context, id uuid.UUID, user *domainmodels.User) error {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return err
	}
	return r.Set(ctx, key, user, r.IdentityTtl())
}

func (r *redisImpl) DeleteById(ctx context.Context, id uuid.UUID) error {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return err
	}
	return r.Delete(ctx, key)
}

func (r *redisImpl) GetByUsername(ctx context.Context, username string) (*domainmodels.User, bool, error) {
	key, err := r.ScopedKey(ctx, "identity", "username", infrastructurecacheshared.StringPart(username))
	if err != nil {
		return nil, false, err
	}
	var item domainmodels.User
	hit, err := r.Get(ctx, key, &item)
	return &item, hit, err
}

func (r *redisImpl) SetByUsername(ctx context.Context, username string, user *domainmodels.User) error {
	key, err := r.ScopedKey(ctx, "identity", "username", infrastructurecacheshared.StringPart(username))
	if err != nil {
		return err
	}
	return r.Set(ctx, key, user, r.IdentityTtl())
}

func (r *redisImpl) DeleteByUsername(ctx context.Context, username string) error {
	key, err := r.ScopedKey(ctx, "identity", "username", infrastructurecacheshared.StringPart(username))
	if err != nil {
		return err
	}
	return r.Delete(ctx, key)
}

func (r *redisImpl) GetPermissions(ctx context.Context, userId uuid.UUID) ([]domainmodels.Permission, bool, error) {
	key, err := r.ScopedKey(ctx, "permissions", "user", userId.String())
	if err != nil {
		return nil, false, err
	}
	var items []domainmodels.Permission
	hit, err := r.Get(ctx, key, &items)
	return items, hit, err
}

func (r *redisImpl) SetPermissions(ctx context.Context, userId uuid.UUID, permissions []domainmodels.Permission) error {
	key, err := r.ScopedKey(ctx, "permissions", "user", userId.String())
	if err != nil {
		return err
	}
	return r.Set(ctx, key, permissions, r.RelationTtl())
}

func (r *redisImpl) DeletePermissions(ctx context.Context, userId uuid.UUID) error {
	key, err := r.ScopedKey(ctx, "permissions", "user", userId.String())
	if err != nil {
		return err
	}
	return r.Delete(ctx, key)
}

func (r *redisImpl) InvalidatePermissions(ctx context.Context) error {
	return r.Invalidate(ctx, "permissions")
}

func (r *redisImpl) GetPagination(ctx context.Context, page int, limit int, search *string, roleId *uuid.UUID) (domaincontractscache.Pagination[domainmodels.User], bool, error) {
	key, err := r.paginationKey(ctx, page, limit, search, roleId)
	if err != nil {
		return domaincontractscache.Pagination[domainmodels.User]{}, false, err
	}
	var pagination domaincontractscache.Pagination[domainmodels.User]
	hit, err := r.Get(ctx, key, &pagination)
	return pagination, hit, err
}

func (r *redisImpl) SetPagination(ctx context.Context, page int, limit int, search *string, roleId *uuid.UUID, pagination domaincontractscache.Pagination[domainmodels.User]) error {
	key, err := r.paginationKey(ctx, page, limit, search, roleId)
	if err != nil {
		return err
	}
	return r.Set(ctx, key, pagination, r.PaginationTtl())
}

func (r *redisImpl) InvalidatePagination(ctx context.Context) error {
	return r.Invalidate(ctx, "pagination")
}

func (r *redisImpl) InvalidateAll(ctx context.Context) error {
	return r.Invalidate(ctx, "identity", "permissions", "pagination")
}

func (r *redisImpl) paginationKey(ctx context.Context, page int, limit int, search *string, roleId *uuid.UUID) (string, error) {
	hash, err := infrastructurecacheshared.HashPart(page, limit, search, roleId)
	if err != nil {
		return "", err
	}
	return r.ScopedKey(ctx, "pagination", hash)
}
