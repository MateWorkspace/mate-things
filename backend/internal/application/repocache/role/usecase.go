package applicationrepocacherole

import (
	"context"
	"encoding/json"

	domaincontractscache "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/cache"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
	"github.com/google/uuid"
)

type usecase struct {
	repository          domaincontractsrepository.Role
	cache               domaincontractscache.Role
	rolePermissionCache domaincontractscache.RolePermission
}

func NewRepoCacheImpl(
	repository domaincontractsrepository.Role,
	cache domaincontractscache.Role,
	rolePermissionCache domaincontractscache.RolePermission,
) domainusecasesrepocache.Role {
	return &usecase{
		repository:          repository,
		cache:               cache,
		rolePermissionCache: rolePermissionCache,
	}
}

func (u *usecase) Create(
	ctx context.Context,
	name string,
	description *string,
	isDefault *bool,
	createdBy *uuid.UUID,
) (uuid.UUID, error) {
	id, err := u.repository.Create(ctx, name, description, isDefault, createdBy)
	if err != nil {
		return uuid.Nil, err
	}

	if err := u.invalidateAll(ctx); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (u *usecase) ReadById(ctx context.Context, id uuid.UUID) (*domainmodels.Role, error) {
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

func (u *usecase) ReadByName(ctx context.Context, name string) (*domainmodels.Role, error) {
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

func (u *usecase) ReadDefault(ctx context.Context) (*domainmodels.Role, error) {
	if item, hit, err := u.cache.GetDefault(ctx); err != nil {
		return nil, err
	} else if hit {
		return item, nil
	}

	item, err := u.repository.ReadDefault(ctx)
	if err != nil {
		return nil, err
	}

	if err := u.cache.SetDefault(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (u *usecase) ReadPermissions(ctx context.Context, roleId uuid.UUID) ([]domainmodels.Permission, error) {
	if items, hit, err := u.cache.GetPermissions(ctx, roleId); err != nil {
		return nil, err
	} else if hit {
		return items, nil
	}

	items, err := u.repository.ReadPermissions(ctx, roleId)
	if err != nil {
		return nil, err
	}

	if err := u.cache.SetPermissions(ctx, roleId, items); err != nil {
		return nil, err
	}

	return items, nil
}

func (u *usecase) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	search *string,
) ([]domainmodels.Role, int, error) {
	if pagination, hit, err := u.cache.GetPagination(ctx, page, limit, search); err != nil {
		return nil, 0, err
	} else if hit {
		return pagination.Items, pagination.Total, nil
	}

	items, total, err := u.repository.ReadByPagination(ctx, page, limit, search)
	if err != nil {
		return nil, 0, err
	}

	if err := u.cache.SetPagination(ctx, page, limit, search, domaincontractscache.Pagination[domainmodels.Role]{
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
	isDefault *bool,
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) error {
	if err := u.repository.UpdateById(ctx, id, name, description, isDefault, preferences, updatedBy); err != nil {
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
	if err := u.rolePermissionCache.InvalidateAll(ctx); err != nil {
		return err
	}
	return nil
}
