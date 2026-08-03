package applicationrepocachenodeclass

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
	repository domaincontractsrepository.NodeClass
	cache      domaincontractscache.NodeClass
}

func NewRepoCacheImpl(
	repository domaincontractsrepository.NodeClass,
	cache domaincontractscache.NodeClass,
) domainusecasesrepocache.NodeClass {
	return &usecase{
		repository: repository,
		cache:      cache,
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

	if err := u.cache.InvalidateAll(ctx); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (u *usecase) ReadById(ctx context.Context, id uuid.UUID) (*domainmodels.NodeClass, error) {
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

func (u *usecase) ReadByName(ctx context.Context, name string) (*domainmodels.NodeClass, error) {
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

func (u *usecase) ReadActions(ctx context.Context, nodeClassId uuid.UUID) ([]domainmodels.Action, error) {
	if items, hit, err := u.cache.GetActions(ctx, nodeClassId); err != nil {
		return nil, err
	} else if hit {
		return items, nil
	}

	items, err := u.repository.ReadActions(ctx, nodeClassId)
	if err != nil {
		return nil, err
	}

	if err := u.cache.SetActions(ctx, nodeClassId, items); err != nil {
		return nil, err
	}

	return items, nil
}

func (u *usecase) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	search *string,
) ([]domainmodels.NodeClass, int, error) {
	if pagination, hit, err := u.cache.GetPagination(ctx, page, limit, search); err != nil {
		return nil, 0, err
	} else if hit {
		return pagination.Items, pagination.Total, nil
	}

	items, total, err := u.repository.ReadByPagination(ctx, page, limit, search)
	if err != nil {
		return nil, 0, err
	}

	if err := u.cache.SetPagination(ctx, page, limit, search, domaincontractscache.Pagination[domainmodels.NodeClass]{
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
