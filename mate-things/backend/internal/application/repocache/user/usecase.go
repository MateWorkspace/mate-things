package applicationrepocacheuser

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
	repository domaincontractsrepository.User
	cache      domaincontractscache.User
}

func NewRepoCacheImpl(
	repository domaincontractsrepository.User,
	cache domaincontractscache.User,
) domainusecasesrepocache.User {
	return &usecase{
		repository: repository,
		cache:      cache,
	}
}

func (u *usecase) Create(
	ctx context.Context,
	roleId uuid.UUID,
	name string,
	bio *string,
	username string,
	passwordHash string,
	createdBy *uuid.UUID,
) (uuid.UUID, error) {
	id, err := u.repository.Create(ctx, roleId, name, bio, username, passwordHash, createdBy)
	if err != nil {
		return uuid.Nil, err
	}

	if err := u.cache.InvalidateAll(ctx); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (u *usecase) ReadById(ctx context.Context, id uuid.UUID) (*domainmodels.User, error) {
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

func (u *usecase) ReadByUsername(ctx context.Context, username string) (*domainmodels.User, error) {
	if item, hit, err := u.cache.GetByUsername(ctx, username); err != nil {
		return nil, err
	} else if hit {
		return item, nil
	}

	item, err := u.repository.ReadByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if err := u.cache.SetByUsername(ctx, username, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (u *usecase) ReadPermissions(ctx context.Context, userId uuid.UUID) ([]domainmodels.Permission, error) {
	if items, hit, err := u.cache.GetPermissions(ctx, userId); err != nil {
		return nil, err
	} else if hit {
		return items, nil
	}

	items, err := u.repository.ReadPermissions(ctx, userId)
	if err != nil {
		return nil, err
	}

	if err := u.cache.SetPermissions(ctx, userId, items); err != nil {
		return nil, err
	}

	return items, nil
}

func (u *usecase) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	search *string,
	roleId *uuid.UUID,
) ([]domainmodels.User, int, error) {
	if pagination, hit, err := u.cache.GetPagination(ctx, page, limit, search, roleId); err != nil {
		return nil, 0, err
	} else if hit {
		return pagination.Items, pagination.Total, nil
	}

	items, total, err := u.repository.ReadByPagination(ctx, page, limit, search, roleId)
	if err != nil {
		return nil, 0, err
	}

	if err := u.cache.SetPagination(ctx, page, limit, search, roleId, domaincontractscache.Pagination[domainmodels.User]{
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
	roleId *uuid.UUID,
	name *string,
	bio *string,
	username *string,
	passwordHash *string,
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) error {
	if err := u.repository.UpdateById(ctx, id, roleId, name, bio, username, passwordHash, preferences, updatedBy); err != nil {
		return err
	}

	return u.cache.InvalidateAll(ctx)
}

func (u *usecase) DeleteById(
	ctx context.Context,
	id uuid.UUID,
	deletedBy *uuid.UUID,
) error {
	if err := u.repository.DeleteById(ctx, id, deletedBy); err != nil {
		return err
	}

	return u.cache.InvalidateAll(ctx)
}
