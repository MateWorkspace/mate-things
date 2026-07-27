package applicationactiondefinition

import (
	"context"

	applicationshared "github.com/MateWorkspace/mate-things/backend/internal/application/shared"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesaction "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/action"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
	"github.com/google/uuid"
)

type usecase struct {
	action domainusecasesrepocache.Action
	logger domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	action domainusecasesrepocache.Action,
	logger domaincontractslogger.Leveled,
) domainusecasesaction.Definition {
	return &usecase{
		action: action,
		logger: logger,
	}
}

func (u *usecase) Create(ctx context.Context, request domainusecasesaction.CreateActionRequest) (uuid.UUID, error) {
	const tag = "action/definition/Create"

	name, err := applicationshared.RequiredActionName(request.Name, "name")
	if err != nil {
		return uuid.Nil, err
	}
	payloadSchemaName, err := applicationshared.RequiredSnakeCaseName(request.PayloadSchemaName, "payload_schema_name")
	if err != nil {
		return uuid.Nil, err
	}
	payloadSchemaVersion, err := applicationshared.RequiredPositiveVersion(request.PayloadSchemaVersion, "payload_schema_version")
	if err != nil {
		return uuid.Nil, err
	}

	id, err := u.action.Create(
		ctx,
		request.NodeClassId,
		name,
		request.Description,
		payloadSchemaName,
		payloadSchemaVersion,
		request.CreatedBy,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to create action", domainmodels.LoggerMeta{
			"err":        err,
			"created_by": request.CreatedBy,
		})
		return uuid.Nil, err
	}

	return id, nil
}

func (u *usecase) ReadById(
	ctx context.Context,
	request domainusecasesaction.ReadActionByIdRequest,
) (*domainmodels.Action, error) {
	const tag = "action/definition/ReadById"

	action, err := u.action.ReadById(ctx, request.Id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read action", domainmodels.LoggerMeta{
			"err": err,
			"id":  request.Id,
		})
		return nil, err
	}

	return action, nil
}

func (u *usecase) ReadByName(
	ctx context.Context,
	request domainusecasesaction.ReadActionByNameRequest,
) (*domainmodels.Action, error) {
	const tag = "action/definition/ReadByName"

	action, err := u.action.ReadByName(ctx, request.Name)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read action", domainmodels.LoggerMeta{
			"err":  err,
			"name": request.Name,
		})
		return nil, err
	}

	return action, nil
}

func (u *usecase) ReadByPagination(
	ctx context.Context,
	request domainusecasesaction.ReadActionsByPaginationRequest,
) ([]domainmodels.Action, int, error) {
	const tag = "action/definition/ReadByPagination"

	actions, total, err := u.action.ReadByPagination(
		ctx,
		request.Page,
		request.Limit,
		request.Search,
		request.NodeClassId,
		request.PayloadSchemaName,
		request.PayloadSchemaVersion,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read actions", domainmodels.LoggerMeta{
			"err":   err,
			"page":  request.Page,
			"limit": request.Limit,
		})
		return nil, 0, err
	}

	return actions, total, nil
}

func (u *usecase) UpdateById(ctx context.Context, request domainusecasesaction.UpdateActionRequest) error {
	const tag = "action/definition/UpdateById"

	name, err := applicationshared.OptionalActionName(request.Name, "name")
	if err != nil {
		return err
	}
	payloadSchemaName, err := applicationshared.OptionalSnakeCaseName(request.PayloadSchemaName, "payload_schema_name")
	if err != nil {
		return err
	}
	payloadSchemaVersion, err := applicationshared.OptionalPositiveVersion(request.PayloadSchemaVersion, "payload_schema_version")
	if err != nil {
		return err
	}

	if err := u.action.UpdateById(
		ctx,
		request.Id,
		request.NodeClassId,
		name,
		request.Description,
		payloadSchemaName,
		payloadSchemaVersion,
		nil,
		request.UpdatedBy,
	); err != nil {
		u.logger.Error(ctx, tag, "failed to update action", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	return nil
}

func (u *usecase) DeleteById(ctx context.Context, request domainusecasesaction.DeleteActionRequest) error {
	const tag = "action/definition/DeleteById"

	if err := u.action.DeleteById(ctx, request.Id, request.DeletedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to delete action", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"deleted_by": request.DeletedBy,
		})
		return err
	}

	return nil
}
