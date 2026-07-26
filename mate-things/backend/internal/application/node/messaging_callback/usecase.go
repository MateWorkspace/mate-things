package applicationnodemessagingcallback

import (
	"context"

	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsnode "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/node"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
)

const resubscribePageLimit = 100

type usecase struct {
	node          domainusecasesrepocache.Node
	actionLog     domaincontractsrepository.ActionLog
	publisher     domaincontractsnode.Publish
	subscriptions domaincontractsnode.Subscriptions
	logger        domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	node domainusecasesrepocache.Node,
	actionLog domaincontractsrepository.ActionLog,
	publisher domaincontractsnode.Publish,
	subscriptions domaincontractsnode.Subscriptions,
	logger domaincontractslogger.Leveled,
) domainusecasesnode.MessagingCallback {
	return &usecase{
		node:          node,
		actionLog:     actionLog,
		publisher:     publisher,
		subscriptions: subscriptions,
		logger:        logger,
	}
}

func (u *usecase) Register(ctx context.Context, request domainusecasesnode.RegisterNodeMessageRequest) error {
	const tag = "node/messaging_callback/Register"

	node, created, err := u.node.UpsertRegistration(ctx, request.DeviceId, request.DeviceInfo, request.FirmwareName)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to upsert node registration", domainmodels.LoggerMeta{
			"err":           err,
			"device_id":     request.DeviceId,
			"firmware_name": request.FirmwareName,
		})
		return err
	}

	if created {
		if err := u.subscribeNode(ctx, node.DeviceId); err != nil {
			u.logger.Error(ctx, tag, "failed to subscribe node topics", domainmodels.LoggerMeta{
				"err":       err,
				"device_id": node.DeviceId,
				"node_id":   node.Id,
			})
			return err
		}
	}

	if err := u.publisher.RegistrationAck(ctx, node.DeviceId); err != nil {
		u.logger.Error(ctx, tag, "failed to publish registration ack", domainmodels.LoggerMeta{
			"err":       err,
			"device_id": node.DeviceId,
			"node_id":   node.Id,
		})
		return err
	}

	return nil
}

func (u *usecase) Status(ctx context.Context, request domainusecasesnode.NodeStatusMessageRequest) error {
	const tag = "node/messaging_callback/Status"

	node, err := u.node.ReadByDeviceId(ctx, request.DeviceId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node", domainmodels.LoggerMeta{
			"err":       err,
			"device_id": request.DeviceId,
		})
		return err
	}

	if err := u.node.UpdateById(ctx, node.Id, nil, nil, nil, nil, nil, nil, &request.IsConnected, nil, nil); err != nil {
		u.logger.Error(ctx, tag, "failed to update node status", domainmodels.LoggerMeta{
			"err":          err,
			"device_id":    request.DeviceId,
			"node_id":      node.Id,
			"is_connected": request.IsConnected,
		})
		return err
	}

	return nil
}

func (u *usecase) ActionAck(ctx context.Context, request domainusecasesnode.NodeActionAckMessageRequest) error {
	const tag = "node/messaging_callback/ActionAck"

	switch request.ActionStatus {
	case domainmodels.ActionStatusSuccess, domainmodels.ActionStatusFailed:
	default:
		err := domainmodels.NewError("invalid action ack status", domainmodels.ErrTypeValidation, nil)
		u.logger.Error(ctx, tag, "invalid action ack status", domainmodels.LoggerMeta{
			"err":           err,
			"device_id":     request.DeviceId,
			"execution_id":  request.ExecutionId,
			"action_status": request.ActionStatus,
		})
		return err
	}

	if _, err := u.node.ReadByDeviceId(ctx, request.DeviceId); err != nil {
		u.logger.Error(ctx, tag, "failed to read node", domainmodels.LoggerMeta{
			"err":          err,
			"device_id":    request.DeviceId,
			"execution_id": request.ExecutionId,
		})
		return err
	}

	if err := u.actionLog.UpdateStatusByExecutionId(ctx, request.ExecutionId, request.ActionStatus, request.ActionMessage); err != nil {
		u.logger.Error(ctx, tag, "failed to update action log status", domainmodels.LoggerMeta{
			"err":           err,
			"device_id":     request.DeviceId,
			"execution_id":  request.ExecutionId,
			"action_status": request.ActionStatus,
		})
		return err
	}

	return nil
}

func (u *usecase) Log(_ context.Context, _ domainusecasesnode.NodeLogMessageRequest) error {
	return nil
}

func (u *usecase) Resubscribe(ctx context.Context) error {
	const tag = "node/messaging_callback/Resubscribe"

	if err := u.subscriptions.Registration(ctx); err != nil {
		u.logger.Error(ctx, tag, "failed to subscribe registration topic", domainmodels.LoggerMeta{
			"err": err,
		})
		return err
	}

	page := 1
	for {
		nodes, total, err := u.node.ReadByPagination(ctx, page, resubscribePageLimit, nil, nil, nil)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to read nodes", domainmodels.LoggerMeta{
				"err":   err,
				"page":  page,
				"limit": resubscribePageLimit,
			})
			return err
		}
		for _, node := range nodes {
			if err := u.subscribeNode(ctx, node.DeviceId); err != nil {
				u.logger.Error(ctx, tag, "failed to subscribe node topics", domainmodels.LoggerMeta{
					"err":       err,
					"device_id": node.DeviceId,
					"node_id":   node.Id,
				})
				return err
			}
		}

		if page*resubscribePageLimit >= total {
			break
		}
		page++
	}

	return nil
}

func (u *usecase) subscribeNode(ctx context.Context, deviceId string) error {
	if err := u.subscriptions.ActionAck(ctx, deviceId); err != nil {
		return err
	}
	if err := u.subscriptions.Status(ctx, deviceId); err != nil {
		return err
	}
	if err := u.subscriptions.Log(ctx, deviceId); err != nil {
		return err
	}
	return nil
}
