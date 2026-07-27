package applicationnodedevicemanagement

import (
	"context"

	applicationshared "github.com/MateWorkspace/mate-things/backend/internal/application/shared"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
)

type usecase struct {
	node   domainusecasesrepocache.Node
	logger domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	node domainusecasesrepocache.Node,
	logger domaincontractslogger.Leveled,
) domainusecasesnode.DeviceManagement {
	return &usecase{
		node:   node,
		logger: logger,
	}
}

func (u *usecase) ReadById(
	ctx context.Context,
	request domainusecasesnode.ReadNodeByIdRequest,
) (*domainmodels.Node, error) {
	const tag = "node/device_management/ReadById"

	node, err := u.node.ReadById(ctx, request.Id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node", domainmodels.LoggerMeta{
			"err": err,
			"id":  request.Id,
		})
		return nil, err
	}

	return node, nil
}

func (u *usecase) ReadByDeviceId(
	ctx context.Context,
	request domainusecasesnode.ReadNodeByDeviceIdRequest,
) (*domainmodels.Node, error) {
	const tag = "node/device_management/ReadByDeviceId"

	node, err := u.node.ReadByDeviceId(ctx, request.DeviceId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node", domainmodels.LoggerMeta{
			"err":       err,
			"device_id": request.DeviceId,
		})
		return nil, err
	}

	return node, nil
}

func (u *usecase) ReadByPagination(
	ctx context.Context,
	request domainusecasesnode.ReadNodesByPaginationRequest,
) ([]domainmodels.Node, int, error) {
	const tag = "node/device_management/ReadByPagination"

	nodes, total, err := u.node.ReadByPagination(
		ctx,
		request.Page,
		request.Limit,
		request.Search,
		request.NodeClassId,
		request.FirmwareId,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read nodes", domainmodels.LoggerMeta{
			"err":           err,
			"page":          request.Page,
			"limit":         request.Limit,
			"node_class_id": request.NodeClassId,
			"firmware_id":   request.FirmwareId,
		})
		return nil, 0, err
	}

	return nodes, total, nil
}

func (u *usecase) UpdateById(ctx context.Context, request domainusecasesnode.UpdateNodeRequest) error {
	const tag = "node/device_management/UpdateById"

	name, err := applicationshared.OptionalNodeName(request.Name, "name")
	if err != nil {
		return err
	}

	if err := u.node.UpdateById(
		ctx,
		request.Id,
		request.NodeClassId,
		request.DeviceId,
		nil,
		name,
		request.FirmwareId,
		request.Description,
		nil,
		nil,
		request.UpdatedBy,
	); err != nil {
		u.logger.Error(ctx, tag, "failed to update node", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	return nil
}

func (u *usecase) AssignFirmware(ctx context.Context, request domainusecasesnode.AssignNodeFirmwareRequest) error {
	const tag = "node/device_management/AssignFirmware"

	if err := u.node.UpdateById(
		ctx,
		request.Id,
		nil,
		nil,
		nil,
		nil,
		&request.FirmwareId,
		nil,
		nil,
		nil,
		request.UpdatedBy,
	); err != nil {
		u.logger.Error(ctx, tag, "failed to assign node firmware", domainmodels.LoggerMeta{
			"err":         err,
			"id":          request.Id,
			"firmware_id": request.FirmwareId,
			"updated_by":  request.UpdatedBy,
		})
		return err
	}

	return nil
}

func (u *usecase) DeleteById(ctx context.Context, request domainusecasesnode.DeleteNodeRequest) error {
	const tag = "node/device_management/DeleteById"

	if err := u.node.DeleteById(ctx, request.Id, request.DeletedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to delete node", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"deleted_by": request.DeletedBy,
		})
		return err
	}

	return nil
}
