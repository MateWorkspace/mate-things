package applicationrepocacheapikey

import (
	"context"
	"time"

	domaincontractscache "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/cache"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
	"github.com/google/uuid"
)

type usecase struct {
	repository domaincontractsrepository.ApiKey
	cache      domaincontractscache.ApiKey
}

func NewRepoCacheImpl(
	repository domaincontractsrepository.ApiKey,
	cache domaincontractscache.ApiKey,
) domainusecasesrepocache.ApiKey {
	return &usecase{
		repository: repository,
		cache:      cache,
	}
}

func (u *usecase) Create(
	ctx context.Context,
	userId uuid.UUID,
	keyHash string,
	keyLastFour string,
	expiresAt *time.Time,
	createdBy *uuid.UUID,
) (uuid.UUID, error) {
	id, err := u.repository.Create(ctx, userId, keyHash, keyLastFour, expiresAt, createdBy)
	if err != nil {
		return uuid.Nil, err
	}

	if err := u.cache.InvalidateAll(ctx); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (u *usecase) ReadByKeyHash(ctx context.Context, keyHash string) (*domainmodels.ApiKey, error) {
	if item, hit, err := u.cache.GetByKeyHash(ctx, keyHash); err != nil {
		return nil, err
	} else if hit {
		return item, nil
	}

	item, err := u.repository.ReadByKeyHash(ctx, keyHash)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}

	if err := u.cache.SetByKeyHash(ctx, keyHash, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (u *usecase) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	search *string,
	status *string,
) ([]domainmodels.ApiKeyWithUser, int, error) {
	if pagination, hit, err := u.cache.GetPagination(ctx, page, limit, search, status); err != nil {
		return nil, 0, err
	} else if hit {
		return pagination.Items, pagination.Total, nil
	}

	items, total, err := u.repository.ReadByPagination(ctx, page, limit, search, status)
	if err != nil {
		return nil, 0, err
	}

	if err := u.cache.SetPagination(ctx, page, limit, search, status, domaincontractscache.Pagination[domainmodels.ApiKeyWithUser]{
		Items: items,
		Total: total,
	}); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (u *usecase) Regenerate(
	ctx context.Context,
	id uuid.UUID,
	keyHash string,
	keyLastFour string,
	expiresAt *time.Time,
	updatedBy *uuid.UUID,
) error {
	if err := u.repository.Regenerate(ctx, id, keyHash, keyLastFour, expiresAt, updatedBy); err != nil {
		return err
	}

	return u.cache.InvalidateAll(ctx)
}

func (u *usecase) Revoke(ctx context.Context, id uuid.UUID, updatedBy *uuid.UUID) error {
	if err := u.repository.Revoke(ctx, id, updatedBy); err != nil {
		return err
	}

	return u.cache.InvalidateAll(ctx)
}

func (u *usecase) DeleteById(ctx context.Context, id uuid.UUID) error {
	if err := u.repository.DeleteById(ctx, id); err != nil {
		return err
	}

	return u.cache.InvalidateAll(ctx)
}
