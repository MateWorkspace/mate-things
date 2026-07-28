package applicationpreferencesupdate

import (
	"context"

	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasespreferences "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/preferences"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
)

type usecase struct {
	action        domainusecasesrepocache.Action
	firmware      domainusecasesrepocache.Firmware
	node          domainusecasesrepocache.Node
	nodeClass     domainusecasesrepocache.NodeClass
	payloadSchema domainusecasesrepocache.PayloadSchema
	permission    domainusecasesrepocache.Permission
	role          domainusecasesrepocache.Role
	user          domainusecasesrepocache.User
	logger        domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	action domainusecasesrepocache.Action,
	firmware domainusecasesrepocache.Firmware,
	node domainusecasesrepocache.Node,
	nodeClass domainusecasesrepocache.NodeClass,
	payloadSchema domainusecasesrepocache.PayloadSchema,
	permission domainusecasesrepocache.Permission,
	role domainusecasesrepocache.Role,
	user domainusecasesrepocache.User,
	logger domaincontractslogger.Leveled,
) domainusecasespreferences.Update {
	return &usecase{
		action:        action,
		firmware:      firmware,
		node:          node,
		nodeClass:     nodeClass,
		payloadSchema: payloadSchema,
		permission:    permission,
		role:          role,
		user:          user,
		logger:        logger,
	}
}

func (u *usecase) Action(
	ctx context.Context,
	request domainusecasespreferences.UpdateActionPreferencesRequest,
) error {
	const tag = "preferences/update/Action"

	if err := u.action.UpdateById(ctx, request.Id, nil, nil, nil, nil, nil, &request.Preferences, request.UpdatedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to update action preferences", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	return nil
}

func (u *usecase) Firmware(
	ctx context.Context,
	request domainusecasespreferences.UpdateFirmwarePreferencesRequest,
) error {
	const tag = "preferences/update/Firmware"

	if err := u.firmware.UpdateById(ctx, request.Id, nil, nil, nil, nil, nil, &request.Preferences, request.UpdatedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to update firmware preferences", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	return nil
}

func (u *usecase) Node(
	ctx context.Context,
	request domainusecasespreferences.UpdateNodePreferencesRequest,
) error {
	const tag = "preferences/update/Node"

	if err := u.node.UpdateById(ctx, request.Id, nil, nil, nil, nil, nil, nil, nil, &request.Preferences, request.UpdatedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to update node preferences", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	return nil
}

func (u *usecase) NodeClass(
	ctx context.Context,
	request domainusecasespreferences.UpdateNodeClassPreferencesRequest,
) error {
	const tag = "preferences/update/NodeClass"

	if err := u.nodeClass.UpdateById(ctx, request.Id, nil, nil, &request.Preferences, request.UpdatedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to update node class preferences", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	return nil
}

func (u *usecase) PayloadSchema(
	ctx context.Context,
	request domainusecasespreferences.UpdatePayloadSchemaPreferencesRequest,
) error {
	const tag = "preferences/update/PayloadSchema"

	if err := u.payloadSchema.UpdateById(ctx, request.Id, nil, nil, nil, nil, nil, &request.Preferences, request.UpdatedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to update payload schema preferences", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	return nil
}

func (u *usecase) Permission(
	ctx context.Context,
	request domainusecasespreferences.UpdatePermissionPreferencesRequest,
) error {
	const tag = "preferences/update/Permission"

	if err := u.permission.UpdateById(ctx, request.Id, nil, nil, &request.Preferences, request.UpdatedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to update permission preferences", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	return nil
}

func (u *usecase) Role(
	ctx context.Context,
	request domainusecasespreferences.UpdateRolePreferencesRequest,
) error {
	const tag = "preferences/update/Role"

	if err := u.role.UpdateById(ctx, request.Id, nil, nil, nil, &request.Preferences, request.UpdatedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to update role preferences", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	return nil
}

func (u *usecase) User(
	ctx context.Context,
	request domainusecasespreferences.UpdateUserPreferencesRequest,
) error {
	const tag = "preferences/update/User"

	if err := u.user.UpdateById(ctx, request.Id, nil, nil, nil, nil, nil, &request.Preferences, request.UpdatedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to update user preferences", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	return nil
}
