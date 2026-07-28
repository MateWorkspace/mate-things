package applicationrepocachepayloadschema

import (
	"context"
	"encoding/json"
	"time"

	domaincontractscache "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/cache"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
	"github.com/google/uuid"
)

type usecase struct {
	repository domaincontractsrepository.PayloadSchema
	cache      domaincontractscache.PayloadSchema
}

func NewRepoCacheImpl(
	repository domaincontractsrepository.PayloadSchema,
	cache domaincontractscache.PayloadSchema,
) domainusecasesrepocache.PayloadSchema {
	return &usecase{
		repository: repository,
		cache:      cache,
	}
}

func (u *usecase) Create(
	ctx context.Context,
	name string,
	version int32,
	definition json.RawMessage,
	validFrom *time.Time,
	validTo *time.Time,
	createdBy *uuid.UUID,
) (uuid.UUID, error) {
	id, err := u.repository.Create(ctx, name, version, definition, validFrom, validTo, createdBy)
	if err != nil {
		return uuid.Nil, err
	}

	if err := u.cache.InvalidateAll(ctx); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (u *usecase) ReadById(ctx context.Context, id uuid.UUID) (*domainmodels.PayloadSchema, error) {
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

func (u *usecase) ReadByNameAndVersion(
	ctx context.Context,
	name string,
	version int32,
) (*domainmodels.PayloadSchema, error) {
	if item, hit, err := u.cache.GetByNameAndVersion(ctx, name, version); err != nil {
		return nil, err
	} else if hit {
		return item, nil
	}

	item, err := u.repository.ReadByNameAndVersion(ctx, name, version)
	if err != nil {
		return nil, err
	}

	if err := u.cache.SetByNameAndVersion(ctx, name, version, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (u *usecase) ReadLatestByName(ctx context.Context, name string) (*domainmodels.PayloadSchema, error) {
	if item, hit, err := u.cache.GetLatestByName(ctx, name); err != nil {
		return nil, err
	} else if hit {
		return item, nil
	}

	item, err := u.repository.ReadLatestByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if err := u.cache.SetLatestByName(ctx, name, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (u *usecase) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	search *string,
	validAt *time.Time,
) ([]domainmodels.PayloadSchema, int, error) {
	if pagination, hit, err := u.cache.GetPagination(ctx, page, limit, search, validAt); err != nil {
		return nil, 0, err
	} else if hit {
		return pagination.Items, pagination.Total, nil
	}

	items, total, err := u.repository.ReadByPagination(ctx, page, limit, search, validAt)
	if err != nil {
		return nil, 0, err
	}

	if err := u.cache.SetPagination(ctx, page, limit, search, validAt, domaincontractscache.Pagination[domainmodels.PayloadSchema]{
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
	version *int32,
	definition *json.RawMessage,
	validFrom *time.Time,
	validTo *time.Time,
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) error {
	if err := u.repository.UpdateById(ctx, id, name, version, definition, validFrom, validTo, preferences, updatedBy); err != nil {
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
