package applicationnodeclassmanagement

import (
	"context"

	applicationshared "github.com/MateWorkspace/mate-things/backend/internal/application/shared"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
	"github.com/google/uuid"
)

type usecase struct {
	nodeClass       domainusecasesrepocache.NodeClass
	nodeClassAction domainusecasesrepocache.NodeClassAction
	logger          domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	nodeClass domainusecasesrepocache.NodeClass,
	nodeClassAction domainusecasesrepocache.NodeClassAction,
	logger domaincontractslogger.Leveled,
) domainusecasesnode.ClassManagement {
	return &usecase{
		nodeClass:       nodeClass,
		nodeClassAction: nodeClassAction,
		logger:          logger,
	}
}

func (u *usecase) Create(ctx context.Context, request domainusecasesnode.CreateNodeClassRequest) (uuid.UUID, error) {
	const tag = "node/class_management/Create"

	name, err := applicationshared.RequiredNodeClassName(request.Name, "name")
	if err != nil {
		return uuid.Nil, err
	}

	id, err := u.nodeClass.Create(ctx, name, request.Description, request.CreatedBy)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to create node class", domainmodels.LoggerMeta{
			"err":        err,
			"created_by": request.CreatedBy,
		})
		return uuid.Nil, err
	}

	return id, nil
}

func (u *usecase) ReadById(
	ctx context.Context,
	request domainusecasesnode.ReadNodeClassByIdRequest,
) (*domainmodels.NodeClass, error) {
	const tag = "node/class_management/ReadById"

	nodeClass, err := u.nodeClass.ReadById(ctx, request.Id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node class", domainmodels.LoggerMeta{
			"err": err,
			"id":  request.Id,
		})
		return nil, err
	}

	return nodeClass, nil
}

func (u *usecase) ReadByName(
	ctx context.Context,
	request domainusecasesnode.ReadNodeClassByNameRequest,
) (*domainmodels.NodeClass, error) {
	const tag = "node/class_management/ReadByName"

	nodeClass, err := u.nodeClass.ReadByName(ctx, request.Name)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node class", domainmodels.LoggerMeta{
			"err":  err,
			"name": request.Name,
		})
		return nil, err
	}

	return nodeClass, nil
}

func (u *usecase) ReadByPagination(
	ctx context.Context,
	request domainusecasesnode.ReadNodeClassesByPaginationRequest,
) ([]domainmodels.NodeClass, int, error) {
	const tag = "node/class_management/ReadByPagination"

	nodeClasses, total, err := u.nodeClass.ReadByPagination(ctx, request.Page, request.Limit, request.Search)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node classes", domainmodels.LoggerMeta{
			"err":   err,
			"page":  request.Page,
			"limit": request.Limit,
		})
		return nil, 0, err
	}

	return nodeClasses, total, nil
}

func (u *usecase) UpdateById(ctx context.Context, request domainusecasesnode.UpdateNodeClassRequest) error {
	const tag = "node/class_management/UpdateById"

	name, err := applicationshared.OptionalNodeClassName(request.Name, "name")
	if err != nil {
		return err
	}

	if err := u.nodeClass.UpdateById(ctx, request.Id, name, request.Description, nil, request.UpdatedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to update node class", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	return nil
}

func (u *usecase) DeleteById(ctx context.Context, request domainusecasesnode.DeleteNodeClassRequest) error {
	const tag = "node/class_management/DeleteById"

	if err := u.nodeClass.DeleteById(ctx, request.Id, request.DeletedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to delete node class", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"deleted_by": request.DeletedBy,
		})
		return err
	}

	return nil
}

func (u *usecase) AssignAction(
	ctx context.Context,
	request domainusecasesnode.AssignNodeClassActionRequest,
) (uuid.UUID, error) {
	const tag = "node/class_management/AssignAction"

	id, err := u.nodeClassAction.Create(ctx, request.NodeClassId, request.ActionId, request.CreatedBy)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to assign node class action", domainmodels.LoggerMeta{
			"err":           err,
			"node_class_id": request.NodeClassId,
			"action_id":     request.ActionId,
			"created_by":    request.CreatedBy,
		})
		return uuid.Nil, err
	}

	return id, nil
}

func (u *usecase) RevokeAction(
	ctx context.Context,
	request domainusecasesnode.RevokeNodeClassActionRequest,
) error {
	const tag = "node/class_management/RevokeAction"

	if err := u.nodeClassAction.DeleteByNodeClassIdAndActionId(ctx, &request.NodeClassId, &request.ActionId); err != nil {
		u.logger.Error(ctx, tag, "failed to revoke node class action", domainmodels.LoggerMeta{
			"err":           err,
			"node_class_id": request.NodeClassId,
			"action_id":     request.ActionId,
		})
		return err
	}

	return nil
}

func (u *usecase) ReadActions(
	ctx context.Context,
	request domainusecasesnode.ReadNodeClassActionsRequest,
) ([]domainmodels.Action, error) {
	const tag = "node/class_management/ReadActions"

	actions, err := u.nodeClass.ReadActions(ctx, request.NodeClassId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node class actions", domainmodels.LoggerMeta{
			"err":           err,
			"node_class_id": request.NodeClassId,
		})
		return nil, err
	}

	return actions, nil
}

func (u *usecase) ReadNodeClassActionById(
	ctx context.Context,
	request domainusecasesnode.ReadNodeClassActionByIdRequest,
) (domainusecasesnode.NodeClassActionResult, error) {
	const tag = "node/class_management/ReadNodeClassActionById"

	nodeClassAction, nodeClass, action, err := u.nodeClassAction.ReadById(ctx, request.Id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node class action", domainmodels.LoggerMeta{
			"err": err,
			"id":  request.Id,
		})
		return domainusecasesnode.NodeClassActionResult{}, err
	}

	return domainusecasesnode.NodeClassActionResult{
		NodeClassAction: *nodeClassAction,
		NodeClass:       *nodeClass,
		Action:          *action,
	}, nil
}

func (u *usecase) ReadNodeClassActionByNodeClassIdAndActionId(
	ctx context.Context,
	request domainusecasesnode.ReadNodeClassActionByNodeClassIdAndActionIdRequest,
) (domainusecasesnode.NodeClassActionResult, error) {
	const tag = "node/class_management/ReadNodeClassActionByNodeClassIdAndActionId"

	nodeClassAction, nodeClass, action, err := u.nodeClassAction.ReadByNodeClassIdAndActionId(
		ctx,
		request.NodeClassId,
		request.ActionId,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node class action", domainmodels.LoggerMeta{
			"err":           err,
			"node_class_id": request.NodeClassId,
			"action_id":     request.ActionId,
		})
		return domainusecasesnode.NodeClassActionResult{}, err
	}

	return domainusecasesnode.NodeClassActionResult{
		NodeClassAction: *nodeClassAction,
		NodeClass:       *nodeClass,
		Action:          *action,
	}, nil
}

func (u *usecase) ReadNodeClassActionsByPagination(
	ctx context.Context,
	request domainusecasesnode.ReadNodeClassActionsByPaginationRequest,
) ([]domainmodels.NodeClassAction, []domainmodels.NodeClass, []domainmodels.Action, int, error) {
	const tag = "node/class_management/ReadNodeClassActionsByPagination"

	nodeClassActions, nodeClasses, actions, total, err := u.nodeClassAction.ReadByPagination(
		ctx,
		request.Page,
		request.Limit,
		request.NodeClassId,
		request.ActionId,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node class actions", domainmodels.LoggerMeta{
			"err":           err,
			"page":          request.Page,
			"limit":         request.Limit,
			"node_class_id": request.NodeClassId,
			"action_id":     request.ActionId,
		})
		return nil, nil, nil, 0, err
	}

	return nodeClassActions, nodeClasses, actions, total, nil
}
