package applicationrepocachenode

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
	repository domaincontractsrepository.Node
	cache      domaincontractscache.Node
}

func NewRepoCacheImpl(
	repository domaincontractsrepository.Node,
	cache domaincontractscache.Node,
) domainusecasesrepocache.Node {
	return &usecase{
		repository: repository,
		cache:      cache,
	}
}

func (u *usecase) Create(
	ctx context.Context,
	nodeClassId uuid.UUID,
	deviceId string,
	deviceInfo string,
	name string,
	firmwareId uuid.UUID,
	description *string,
	isConnected bool,
	createdBy *uuid.UUID,
) (uuid.UUID, error) {
	id, err := u.repository.Create(ctx, nodeClassId, deviceId, deviceInfo, name, firmwareId, description, isConnected, createdBy)
	if err != nil {
		return uuid.Nil, err
	}

	if err := u.cache.InvalidateAll(ctx); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (u *usecase) UpsertRegistration(
	ctx context.Context,
	deviceId string,
	deviceInfo string,
	firmwareName string,
) (*domainmodels.Node, bool, error) {
	item, created, err := u.repository.UpsertRegistration(ctx, deviceId, deviceInfo, firmwareName)
	if err != nil {
		return nil, false, err
	}

	if err := u.cache.InvalidateAll(ctx); err != nil {
		return nil, false, err
	}

	return item, created, nil
}

func (u *usecase) ReadById(ctx context.Context, id uuid.UUID) (*domainmodels.Node, error) {
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

func (u *usecase) ReadByDeviceId(ctx context.Context, deviceId string) (*domainmodels.Node, error) {
	if item, hit, err := u.cache.GetByDeviceId(ctx, deviceId); err != nil {
		return nil, err
	} else if hit {
		return item, nil
	}

	item, err := u.repository.ReadByDeviceId(ctx, deviceId)
	if err != nil {
		return nil, err
	}

	if err := u.cache.SetByDeviceId(ctx, deviceId, item); err != nil {
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
	firmwareId *uuid.UUID,
) ([]domainmodels.Node, int, error) {
	if pagination, hit, err := u.cache.GetPagination(ctx, page, limit, search, nodeClassId, firmwareId); err != nil {
		return nil, 0, err
	} else if hit {
		return pagination.Items, pagination.Total, nil
	}

	items, total, err := u.repository.ReadByPagination(ctx, page, limit, search, nodeClassId, firmwareId)
	if err != nil {
		return nil, 0, err
	}

	if err := u.cache.SetPagination(ctx, page, limit, search, nodeClassId, firmwareId, domaincontractscache.Pagination[domainmodels.Node]{
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
	deviceId *string,
	deviceInfo *string,
	name *string,
	firmwareId *uuid.UUID,
	description *string,
	isConnected *bool,
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) error {
	if err := u.repository.UpdateById(ctx, id, nodeClassId, deviceId, deviceInfo, name, firmwareId, description, isConnected, preferences, updatedBy); err != nil {
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
