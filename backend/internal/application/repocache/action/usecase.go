package applicationrepocacheaction

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
	repository           domaincontractsrepository.Action
	cache                domaincontractscache.Action
	nodeClassActionCache domaincontractscache.NodeClassAction
}

func NewRepoCacheImpl(
	repository domaincontractsrepository.Action,
	cache domaincontractscache.Action,
	nodeClassActionCache domaincontractscache.NodeClassAction,
) domainusecasesrepocache.Action {
	return &usecase{
		repository:           repository,
		cache:                cache,
		nodeClassActionCache: nodeClassActionCache,
	}
}

func (u *usecase) Create(
	ctx context.Context,
	name string,
	description *string,
	payloadSchemaName string,
	payloadSchemaVersion int32,
	createdBy *uuid.UUID,
) (uuid.UUID, error) {
	id, err := u.repository.Create(ctx, name, description, payloadSchemaName, payloadSchemaVersion, createdBy)
	if err != nil {
		return uuid.Nil, err
	}

	if err := u.invalidateAll(ctx); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (u *usecase) ReadById(ctx context.Context, id uuid.UUID) (*domainmodels.Action, error) {
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

func (u *usecase) ReadByName(ctx context.Context, name string) (*domainmodels.Action, error) {
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
	nodeClassId *uuid.UUID,
	payloadSchemaName *string,
	payloadSchemaVersion *int32,
) ([]domainmodels.ActionListItem, int, error) {
	if pagination, hit, err := u.cache.GetPagination(ctx, page, limit, search, nodeClassId, payloadSchemaName, payloadSchemaVersion); err != nil {
		return nil, 0, err
	} else if hit {
		return pagination.Items, pagination.Total, nil
	}

	items, total, err := u.repository.ReadByPagination(ctx, page, limit, search, nodeClassId, payloadSchemaName, payloadSchemaVersion)
	if err != nil {
		return nil, 0, err
	}

	if err := u.cache.SetPagination(ctx, page, limit, search, nodeClassId, payloadSchemaName, payloadSchemaVersion, domaincontractscache.Pagination[domainmodels.ActionListItem]{
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
	payloadSchemaName *string,
	payloadSchemaVersion *int32,
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) error {
	if err := u.repository.UpdateById(ctx, id, name, description, payloadSchemaName, payloadSchemaVersion, preferences, updatedBy); err != nil {
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
	return u.nodeClassActionCache.InvalidateAll(ctx)
}
