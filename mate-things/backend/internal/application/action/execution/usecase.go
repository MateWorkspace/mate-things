package applicationactionexecution

import (
	"context"
	"errors"

	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsnode "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/node"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesaction "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/action"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
	"github.com/google/uuid"
)

type usecase struct {
	action        domainusecasesrepocache.Action
	node          domainusecasesrepocache.Node
	payloadSchema domainusecasesrepocache.PayloadSchema
	actionLog     domaincontractsrepository.ActionLog
	publisher     domaincontractsnode.Publish
	validator     domaincontractsutility.PayloadSchemaValidator
	logger        domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	action domainusecasesrepocache.Action,
	node domainusecasesrepocache.Node,
	payloadSchema domainusecasesrepocache.PayloadSchema,
	actionLog domaincontractsrepository.ActionLog,
	publisher domaincontractsnode.Publish,
	validator domaincontractsutility.PayloadSchemaValidator,
	logger domaincontractslogger.Leveled,
) domainusecasesaction.Execution {
	return &usecase{
		action:        action,
		node:          node,
		payloadSchema: payloadSchema,
		actionLog:     actionLog,
		publisher:     publisher,
		validator:     validator,
		logger:        logger,
	}
}

func (u *usecase) Dispatch(
	ctx context.Context,
	request domainusecasesaction.DispatchActionRequest,
) (*domainmodels.ActionLog, error) {
	const tag = "action/execution/Dispatch"

	action, err := u.action.ReadById(ctx, request.ActionId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read action", domainmodels.LoggerMeta{
			"err":       err,
			"action_id": request.ActionId,
			"actor_id":  request.ActorId,
		})
		return nil, err
	}

	executionId := uuid.New()
	node, err := u.node.ReadById(ctx, request.NodeId)
	if err != nil {
		if errors.Is(err, domainmodels.ErrTypeNotFound) {
			return u.createActionLog(
				ctx,
				tag,
				executionId,
				request,
				nil,
				domainmodels.ActionStatusUnexecuted,
				new("node not found"),
			)
		}
		u.logger.Error(ctx, tag, "failed to read node", domainmodels.LoggerMeta{
			"err":      err,
			"node_id":  request.NodeId,
			"actor_id": request.ActorId,
		})
		return nil, err
	}
	if !node.IsConnected {
		return u.createActionLog(
			ctx,
			tag,
			executionId,
			request,
			&node.Id,
			domainmodels.ActionStatusUnexecuted,
			new("node is not connected"),
		)
	}
	if node.NodeClassId != action.NodeClassId {
		return u.createActionLog(
			ctx,
			tag,
			executionId,
			request,
			&node.Id,
			domainmodels.ActionStatusUnexecuted,
			new("node class does not match action's node class"),
		)
	}

	payloadSchema, err := u.payloadSchema.ReadByNameAndVersion(
		ctx,
		action.PayloadSchemaName,
		action.PayloadSchemaVersion,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read payload schema", domainmodels.LoggerMeta{
			"err":                    err,
			"action_id":              request.ActionId,
			"payload_schema_name":    action.PayloadSchemaName,
			"payload_schema_version": action.PayloadSchemaVersion,
			"actor_id":               request.ActorId,
		})
		return nil, err
	}

	if err := u.validator.Validate(ctx, payloadSchema.Definition, request.Payload); err != nil {
		u.logger.Error(ctx, tag, "failed to validate action payload", domainmodels.LoggerMeta{
			"err":                    err,
			"action_id":              request.ActionId,
			"node_id":                request.NodeId,
			"payload_schema_name":    action.PayloadSchemaName,
			"payload_schema_version": action.PayloadSchemaVersion,
			"actor_id":               request.ActorId,
		})
		return nil, err
	}

	actionLog, err := u.createActionLog(
		ctx,
		tag,
		executionId,
		request,
		&node.Id,
		domainmodels.ActionStatusUnresponded,
		nil,
	)
	if err != nil {
		return nil, err
	}

	if err := u.publisher.Action(ctx, node.DeviceId, executionId, action.Name, request.Payload); err != nil {
		u.logger.Error(ctx, tag, "failed to publish action", domainmodels.LoggerMeta{
			"err":          err,
			"execution_id": executionId,
			"action_id":    request.ActionId,
			"node_id":      request.NodeId,
			"actor_id":     request.ActorId,
		})
		message := err.Error()
		if updateErr := u.actionLog.UpdateStatusByExecutionId(ctx, executionId, domainmodels.ActionStatusUnexecuted, &message); updateErr != nil {
			u.logger.Error(ctx, tag, "failed to update action log status", domainmodels.LoggerMeta{
				"err":          updateErr,
				"execution_id": executionId,
				"action_id":    request.ActionId,
				"node_id":      request.NodeId,
				"actor_id":     request.ActorId,
			})
			return nil, updateErr
		}
		actionLog.ActionStatus = domainmodels.ActionStatusUnexecuted
		actionLog.ActionMessage = &message
	}

	return actionLog, nil
}

func (u *usecase) createActionLog(
	ctx context.Context,
	tag string,
	executionId uuid.UUID,
	request domainusecasesaction.DispatchActionRequest,
	nodeId *uuid.UUID,
	actionStatus domainmodels.ActionStatus,
	actionMessage *string,
) (*domainmodels.ActionLog, error) {
	id, err := u.actionLog.Create(
		ctx,
		executionId,
		request.ActionId,
		nodeId,
		actionStatus,
		actionMessage,
		request.Payload,
		request.ExecutedAt,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to create action log", domainmodels.LoggerMeta{
			"err":           err,
			"execution_id":  executionId,
			"action_id":     request.ActionId,
			"node_id":       nodeId,
			"action_status": actionStatus,
			"executed_at":   request.ExecutedAt,
			"actor_id":      request.ActorId,
		})
		return nil, err
	}

	return &domainmodels.ActionLog{
		Id:            id,
		ExecutionId:   executionId,
		ActionId:      request.ActionId,
		NodeId:        nodeId,
		ActionStatus:  actionStatus,
		ActionMessage: actionMessage,
		Payload:       request.Payload,
		ExecutedAt:    request.ExecutedAt,
	}, nil
}
