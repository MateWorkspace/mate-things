package applicationrepocachefirmware

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
	repository domaincontractsrepository.Firmware
	cache      domaincontractscache.Firmware
}

func NewRepoCacheImpl(
	repository domaincontractsrepository.Firmware,
	cache domaincontractscache.Firmware,
) domainusecasesrepocache.Firmware {
	return &usecase{
		repository: repository,
		cache:      cache,
	}
}

func (u *usecase) Create(
	ctx context.Context,
	nodeClassId uuid.UUID,
	name string,
	size int32,
	checksum string,
	binaryPath string,
	createdBy *uuid.UUID,
) (uuid.UUID, error) {
	id, err := u.repository.Create(ctx, nodeClassId, name, size, checksum, binaryPath, createdBy)
	if err != nil {
		return uuid.Nil, err
	}

	if err := u.cache.InvalidateAll(ctx); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (u *usecase) ReadById(ctx context.Context, id uuid.UUID) (*domainmodels.Firmware, error) {
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

func (u *usecase) ReadByName(ctx context.Context, name string) (*domainmodels.Firmware, error) {
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

func (u *usecase) ReadByNodeClassIdAndPagination(
	ctx context.Context,
	nodeClassId uuid.UUID,
	page int,
	limit int,
	search *string,
) ([]domainmodels.Firmware, int, error) {
	if pagination, hit, err := u.cache.GetPagination(ctx, page, limit, search, &nodeClassId); err != nil {
		return nil, 0, err
	} else if hit {
		return pagination.Items, pagination.Total, nil
	}

	items, total, err := u.repository.ReadByNodeClassIdAndPagination(ctx, nodeClassId, page, limit, search)
	if err != nil {
		return nil, 0, err
	}

	if err := u.cache.SetPagination(ctx, page, limit, search, &nodeClassId, domaincontractscache.Pagination[domainmodels.Firmware]{
		Items: items,
		Total: total,
	}); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (u *usecase) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	search *string,
	nodeClassId *uuid.UUID,
) ([]domainmodels.Firmware, int, error) {
	if pagination, hit, err := u.cache.GetPagination(ctx, page, limit, search, nodeClassId); err != nil {
		return nil, 0, err
	} else if hit {
		return pagination.Items, pagination.Total, nil
	}

	items, total, err := u.repository.ReadByPagination(ctx, page, limit, search, nodeClassId)
	if err != nil {
		return nil, 0, err
	}

	if err := u.cache.SetPagination(ctx, page, limit, search, nodeClassId, domaincontractscache.Pagination[domainmodels.Firmware]{
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
	nodeClassId *uuid.UUID,
	name *string,
	size *int32,
	checksum *string,
	binaryPath *string,
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) error {
	if err := u.repository.UpdateById(ctx, id, nodeClassId, name, size, checksum, binaryPath, preferences, updatedBy); err != nil {
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
