package applicationnodeota

import (
	"context"

	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsnode "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/node"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
	"github.com/google/uuid"
)

type usecase struct {
	node      domainusecasesrepocache.Node
	firmware  domainusecasesrepocache.Firmware
	publisher domaincontractsnode.Publish
	logger    domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	node domainusecasesrepocache.Node,
	firmware domainusecasesrepocache.Firmware,
	publisher domaincontractsnode.Publish,
	logger domaincontractslogger.Leveled,
) domainusecasesnode.Ota {
	return &usecase{
		node:      node,
		firmware:  firmware,
		publisher: publisher,
		logger:    logger,
	}
}

func (u *usecase) DispatchByNodeId(
	ctx context.Context,
	request domainusecasesnode.DispatchOtaByNodeIdRequest,
) error {
	const tag = "node/ota/DispatchByNodeId"

	node, err := u.node.ReadById(ctx, request.NodeId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node", domainmodels.LoggerMeta{
			"err":      err,
			"node_id":  request.NodeId,
			"actor_id": request.ActorId,
		})
		return err
	}

	return u.dispatch(ctx, tag, *node, request.FirmwareName, request.FirmwareUrl, request.ActorId)
}

func (u *usecase) DispatchByNodeDeviceId(
	ctx context.Context,
	request domainusecasesnode.DispatchOtaByNodeDeviceIdRequest,
) error {
	const tag = "node/ota/DispatchByNodeDeviceId"

	node, err := u.node.ReadByDeviceId(ctx, request.NodeDeviceId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node", domainmodels.LoggerMeta{
			"err":       err,
			"device_id": request.NodeDeviceId,
			"actor_id":  request.ActorId,
		})
		return err
	}

	return u.dispatch(ctx, tag, *node, request.FirmwareName, request.FirmwareUrl, request.ActorId)
}

func (u *usecase) dispatch(
	ctx context.Context,
	tag string,
	node domainmodels.Node,
	firmwareName string,
	firmwareUrl string,
	actorId *uuid.UUID,
) error {
	firmware, err := u.firmware.ReadByName(ctx, firmwareName)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read firmware", domainmodels.LoggerMeta{
			"err":           err,
			"node_id":       node.Id,
			"firmware_name": firmwareName,
			"actor_id":      actorId,
		})
		return err
	}

	if err := u.publisher.Ota(ctx, node.DeviceId, firmwareUrl, firmware.Size, firmware.Checksum); err != nil {
		u.logger.Error(ctx, tag, "failed to publish ota", domainmodels.LoggerMeta{
			"err":           err,
			"node_id":       node.Id,
			"firmware_name": firmwareName,
			"actor_id":      actorId,
		})
		return err
	}

	return nil
}
