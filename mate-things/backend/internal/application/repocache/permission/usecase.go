package applicationrepocachepermission

import (
	"context"
	"encoding/json"

	domaincontractscache "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/cache"
	domaincontractsrepository "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	domainusecasesrepocache "github.com/ABA-Developer/nusapala-things/backend/internal/domain/usecases/repocache"
	"github.com/google/uuid"
)

type usecase struct {
	repository          domaincontractsrepository.Permission
	cache               domaincontractscache.Permission
	roleCache           domaincontractscache.Role
	userCache           domaincontractscache.User
	rolePermissionCache domaincontractscache.RolePermission
}

func NewRepoCacheImpl(
	repository domaincontractsrepository.Permission,
	cache domaincontractscache.Permission,
	roleCache domaincontractscache.Role,
	userCache domaincontractscache.User,
	rolePermissionCache domaincontractscache.RolePermission,
) domainusecasesrepocache.Permission {
	return &usecase{
		repository:          repository,
		cache:               cache,
		roleCache:           roleCache,
		userCache:           userCache,
		rolePermissionCache: rolePermissionCache,
	}
}

func (u *usecase) Create(
	ctx context.Context,
	name string,
	description *string,
	createdBy *uuid.UUID,
) (uuid.UUID, error) {
	id, err := u.repository.Create(ctx, name, description, createdBy)
	if err != nil {
		return uuid.Nil, err
	}

	if err := u.invalidateAll(ctx); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (u *usecase) ReadById(ctx context.Context, id uuid.UUID) (*domainmodels.Permission, error) {
	if item, hit, err := u.cache.GetById(ctx, id); err != nil {
		return nil, err
	} else if hit {
		return item, nil
	}

	item, err := u.repository.ReadById(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := u.cache.SetById(ctx, id, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (u *usecase) ReadByName(ctx context.Context, name string) (*domainmodels.Permission, error) {
	if item, hit, err := u.cache.GetByName(ctx, name); err != nil {
		return nil, err
	} else if hit {
		return item, nil
	}

	item, err := u.repository.ReadByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if err := u.cache.SetByName(ctx, name, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (u *usecase) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	search *string,
) ([]domainmodels.Permission, int, error) {
	if pagination, hit, err := u.cache.GetPagination(ctx, page, limit, search); err != nil {
		return nil, 0, err
	} else if hit {
		return pagination.Items, pagination.Total, nil
	}

	items, total, err := u.repository.ReadByPagination(ctx, page, limit, search)
	if err != nil {
		return nil, 0, err
	}

	if err := u.cache.SetPagination(ctx, page, limit, search, domaincontractscache.Pagination[domainmodels.Permission]{
		Items: items,
		Total: total,
	}); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (u *usecase) UpdateById(
	ctx context.Context,
	id uuid.UUID,
	name *string,
	description *string,
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) error {
	if err := u.repository.UpdateById(ctx, id, name, description, preferences, updatedBy); err != nil {
		return err
	}

	return u.invalidateAll(ctx)
}

func (u *usecase) DeleteById(
	ctx context.Context,
	id uuid.UUID,
	deletedBy *uuid.UUID,
) error {
	if err := u.repository.DeleteById(ctx, id, deletedBy); err != nil {
		return err
	}

	return u.invalidateAll(ctx)
}

func (u *usecase) invalidateAll(ctx context.Context) error {
	if err := u.cache.InvalidateAll(ctx); err != nil {
		return err
	}
	if err := u.roleCache.InvalidatePermissions(ctx); err != nil {
		return err
	}
	if err := u.userCache.InvalidatePermissions(ctx); err != nil {
		return err
	}
	if err := u.rolePermissionCache.InvalidateAll(ctx); err != nil {
		return err
	}
	return nil
}
