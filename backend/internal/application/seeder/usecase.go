package applicationseeder

import (
	"context"
	"errors"

	seederdata "github.com/MateWorkspace/mate-things/backend/database/seeder"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesseeder "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/seeder"
	"github.com/google/uuid"
)

type usecase struct {
	permission         domaincontractsrepository.Permission
	role               domaincontractsrepository.Role
	rolePermission     domaincontractsrepository.RolePermission
	nodeClass          domaincontractsrepository.NodeClass
	nodeClassAction    domaincontractsrepository.NodeClassAction
	payloadSchema      domaincontractsrepository.PayloadSchema
	action             domaincontractsrepository.Action
	user               domaincontractsrepository.User
	infraredDeviceType domaincontractsrepository.InfraredDeviceType
	infraredState      domaincontractsrepository.InfraredState
	password           domaincontractsutility.Password
	logger             domaincontractslogger.Leveled
	data               seederdata.Data
}

func NewUsecaseImpl(
	permission domaincontractsrepository.Permission,
	role domaincontractsrepository.Role,
	rolePermission domaincontractsrepository.RolePermission,
	nodeClass domaincontractsrepository.NodeClass,
	nodeClassAction domaincontractsrepository.NodeClassAction,
	payloadSchema domaincontractsrepository.PayloadSchema,
	action domaincontractsrepository.Action,
	user domaincontractsrepository.User,
	infraredDeviceType domaincontractsrepository.InfraredDeviceType,
	infraredState domaincontractsrepository.InfraredState,
	password domaincontractsutility.Password,
	logger domaincontractslogger.Leveled,
	data seederdata.Data,
) domainusecasesseeder.Seeder {
	return &usecase{
		permission:         permission,
		role:               role,
		rolePermission:     rolePermission,
		nodeClass:          nodeClass,
		nodeClassAction:    nodeClassAction,
		payloadSchema:      payloadSchema,
		action:             action,
		user:               user,
		infraredDeviceType: infraredDeviceType,
		infraredState:      infraredState,
		password:           password,
		logger:             logger,
		data:               data,
	}
}

func (u *usecase) Run(ctx context.Context) error {
	permissionIds, err := u.seedPermissions(ctx)
	if err != nil {
		return err
	}

	roleIds, err := u.seedRoles(ctx)
	if err != nil {
		return err
	}

	if err := u.seedRolePermissions(ctx, roleIds, permissionIds); err != nil {
		return err
	}

	nodeClassIds, err := u.seedNodeClasses(ctx)
	if err != nil {
		return err
	}

	if err := u.seedPayloadSchemas(ctx); err != nil {
		return err
	}

	actionIds, err := u.seedActions(ctx, nodeClassIds)
	if err != nil {
		return err
	}

	if err := u.seedNodeClassActions(ctx, nodeClassIds, actionIds); err != nil {
		return err
	}

	if err := u.seedUsers(ctx, roleIds); err != nil {
		return err
	}

	deviceTypeIds, err := u.seedInfraredDeviceTypes(ctx)
	if err != nil {
		return err
	}

	if err := u.seedInfraredStates(ctx, deviceTypeIds); err != nil {
		return err
	}

	return nil
}

func (u *usecase) seedPermissions(ctx context.Context) (map[string]uuid.UUID, error) {
	const tag = "seeder/seedPermissions"

	ids := make(map[string]uuid.UUID, len(u.data.Permissions))
	for _, permission := range u.data.Permissions {
		existing, err := u.permission.ReadByName(ctx, permission.Name)
		if err == nil {
			u.logger.Debug(ctx, tag, "permission already exists, skipping", domainmodels.LoggerMeta{"name": permission.Name})
			ids[permission.Name] = existing.Id
			continue
		}
		if !errors.Is(err, domainmodels.ErrTypeNotFound) {
			u.logger.Error(ctx, tag, "failed to read permission", domainmodels.LoggerMeta{
				"err":  err,
				"name": permission.Name,
			})
			return nil, err
		}

		description := permission.Description
		id, err := u.permission.Create(ctx, permission.Name, &description, nil)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to create permission", domainmodels.LoggerMeta{
				"err":  err,
				"name": permission.Name,
			})
			return nil, err
		}

		u.logger.Info(ctx, tag, "permission created", domainmodels.LoggerMeta{"name": permission.Name})
		ids[permission.Name] = id
	}

	return ids, nil
}

func (u *usecase) seedRoles(ctx context.Context) (map[string]uuid.UUID, error) {
	const tag = "seeder/seedRoles"

	ids := make(map[string]uuid.UUID, len(u.data.Roles))
	for _, role := range u.data.Roles {
		existing, err := u.role.ReadByName(ctx, role.Name)
		if err == nil {
			u.logger.Debug(ctx, tag, "role already exists, skipping", domainmodels.LoggerMeta{"name": role.Name})
			ids[role.Name] = existing.Id
			continue
		}
		if !errors.Is(err, domainmodels.ErrTypeNotFound) {
			u.logger.Error(ctx, tag, "failed to read role", domainmodels.LoggerMeta{
				"err":  err,
				"name": role.Name,
			})
			return nil, err
		}

		description := role.Description
		isDefault := role.IsDefault
		id, err := u.role.Create(ctx, role.Name, &description, &isDefault, nil)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to create role", domainmodels.LoggerMeta{
				"err":  err,
				"name": role.Name,
			})
			return nil, err
		}

		u.logger.Info(ctx, tag, "role created", domainmodels.LoggerMeta{"name": role.Name})
		ids[role.Name] = id
	}

	return ids, nil
}

func (u *usecase) seedRolePermissions(
	ctx context.Context,
	roleIds map[string]uuid.UUID,
	permissionIds map[string]uuid.UUID,
) error {
	const tag = "seeder/seedRolePermissions"

	for _, role := range u.data.Roles {
		roleId, ok := roleIds[role.Name]
		if !ok {
			err := domainmodels.NewError("role not found for role_permission seeding", domainmodels.ErrTypeNotFound, nil)
			u.logger.Error(ctx, tag, "unknown role", domainmodels.LoggerMeta{"err": err, "role": role.Name})
			return err
		}

		for _, permissionName := range role.Permissions {
			permissionId, ok := permissionIds[permissionName]
			if !ok {
				err := domainmodels.NewError("permission not found for role_permission seeding", domainmodels.ErrTypeNotFound, nil)
				u.logger.Error(ctx, tag, "unknown permission", domainmodels.LoggerMeta{
					"err":        err,
					"role":       role.Name,
					"permission": permissionName,
				})
				return err
			}

			_, _, _, err := u.rolePermission.ReadByRoleIdAndPermissionId(ctx, roleId, permissionId)
			if err == nil {
				u.logger.Debug(ctx, tag, "role permission already assigned, skipping", domainmodels.LoggerMeta{
					"role":       role.Name,
					"permission": permissionName,
				})
				continue
			}
			if !errors.Is(err, domainmodels.ErrTypeNotFound) {
				u.logger.Error(ctx, tag, "failed to read role permission", domainmodels.LoggerMeta{
					"err":        err,
					"role":       role.Name,
					"permission": permissionName,
				})
				return err
			}

			if _, err := u.rolePermission.Create(ctx, roleId, permissionId, nil); err != nil {
				u.logger.Error(ctx, tag, "failed to create role permission", domainmodels.LoggerMeta{
					"err":        err,
					"role":       role.Name,
					"permission": permissionName,
				})
				return err
			}

			u.logger.Info(ctx, tag, "role permission assigned", domainmodels.LoggerMeta{
				"role":       role.Name,
				"permission": permissionName,
			})
		}
	}

	return nil
}

func (u *usecase) seedNodeClasses(ctx context.Context) (map[string]uuid.UUID, error) {
	const tag = "seeder/seedNodeClasses"

	ids := make(map[string]uuid.UUID, len(u.data.NodeClasses))
	for _, nodeClass := range u.data.NodeClasses {
		existing, err := u.nodeClass.ReadByName(ctx, nodeClass.Name)
		if err == nil {
			u.logger.Debug(ctx, tag, "node class already exists, skipping", domainmodels.LoggerMeta{"name": nodeClass.Name})
			ids[nodeClass.Name] = existing.Id
			continue
		}
		if !errors.Is(err, domainmodels.ErrTypeNotFound) {
			u.logger.Error(ctx, tag, "failed to read node class", domainmodels.LoggerMeta{
				"err":  err,
				"name": nodeClass.Name,
			})
			return nil, err
		}

		description := nodeClass.Description
		id, err := u.nodeClass.Create(ctx, nodeClass.Name, &description, nil)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to create node class", domainmodels.LoggerMeta{
				"err":  err,
				"name": nodeClass.Name,
			})
			return nil, err
		}

		u.logger.Info(ctx, tag, "node class created", domainmodels.LoggerMeta{"name": nodeClass.Name})
		ids[nodeClass.Name] = id
	}

	return ids, nil
}

func (u *usecase) seedPayloadSchemas(ctx context.Context) error {
	const tag = "seeder/seedPayloadSchemas"

	for _, payloadSchema := range u.data.PayloadSchemas {
		_, err := u.payloadSchema.ReadByNameAndVersion(ctx, payloadSchema.Name, payloadSchema.Version)
		if err == nil {
			u.logger.Debug(ctx, tag, "payload schema already exists, skipping", domainmodels.LoggerMeta{
				"name":    payloadSchema.Name,
				"version": payloadSchema.Version,
			})
			continue
		}
		if !errors.Is(err, domainmodels.ErrTypeNotFound) {
			u.logger.Error(ctx, tag, "failed to read payload schema", domainmodels.LoggerMeta{
				"err":     err,
				"name":    payloadSchema.Name,
				"version": payloadSchema.Version,
			})
			return err
		}

		if _, err := u.payloadSchema.Create(ctx, payloadSchema.Name, payloadSchema.Version, payloadSchema.Definition, nil, nil, nil); err != nil {
			u.logger.Error(ctx, tag, "failed to create payload schema", domainmodels.LoggerMeta{
				"err":     err,
				"name":    payloadSchema.Name,
				"version": payloadSchema.Version,
			})
			return err
		}

		u.logger.Info(ctx, tag, "payload schema created", domainmodels.LoggerMeta{
			"name":    payloadSchema.Name,
			"version": payloadSchema.Version,
		})
	}

	return nil
}

func (u *usecase) seedActions(ctx context.Context, nodeClassIds map[string]uuid.UUID) (map[string]uuid.UUID, error) {
	const tag = "seeder/seedActions"

	ids := make(map[string]uuid.UUID, len(u.data.Actions))
	for _, action := range u.data.Actions {
		for _, nodeClassName := range action.NodeClassNames {
			if _, ok := nodeClassIds[nodeClassName]; !ok {
				err := domainmodels.NewError("node class not found for action seeding", domainmodels.ErrTypeNotFound, nil)
				u.logger.Error(ctx, tag, "unknown node class", domainmodels.LoggerMeta{
					"err":        err,
					"action":     action.Name,
					"node_class": nodeClassName,
				})
				return nil, err
			}
		}

		existing, err := u.action.ReadByName(ctx, action.Name)
		if err == nil {
			u.logger.Debug(ctx, tag, "action already exists, skipping", domainmodels.LoggerMeta{"name": action.Name})
			ids[action.Name] = existing.Id
			continue
		}
		if !errors.Is(err, domainmodels.ErrTypeNotFound) {
			u.logger.Error(ctx, tag, "failed to read action", domainmodels.LoggerMeta{
				"err":  err,
				"name": action.Name,
			})
			return nil, err
		}

		description := action.Description
		id, err := u.action.Create(
			ctx,
			action.Name,
			&description,
			action.PayloadSchemaName,
			action.PayloadSchemaVersion,
			nil,
		)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to create action", domainmodels.LoggerMeta{
				"err":  err,
				"name": action.Name,
			})
			return nil, err
		}

		u.logger.Info(ctx, tag, "action created", domainmodels.LoggerMeta{"name": action.Name})
		ids[action.Name] = id
	}

	return ids, nil
}

func (u *usecase) seedNodeClassActions(
	ctx context.Context,
	nodeClassIds map[string]uuid.UUID,
	actionIds map[string]uuid.UUID,
) error {
	const tag = "seeder/seedNodeClassActions"

	for _, action := range u.data.Actions {
		actionId, ok := actionIds[action.Name]
		if !ok {
			err := domainmodels.NewError("action not found for node_class_action seeding", domainmodels.ErrTypeNotFound, nil)
			u.logger.Error(ctx, tag, "unknown action", domainmodels.LoggerMeta{"err": err, "action": action.Name})
			return err
		}

		for _, nodeClassName := range action.NodeClassNames {
			nodeClassId, ok := nodeClassIds[nodeClassName]
			if !ok {
				err := domainmodels.NewError("node class not found for node_class_action seeding", domainmodels.ErrTypeNotFound, nil)
				u.logger.Error(ctx, tag, "unknown node class", domainmodels.LoggerMeta{
					"err":        err,
					"action":     action.Name,
					"node_class": nodeClassName,
				})
				return err
			}

			_, _, _, err := u.nodeClassAction.ReadByNodeClassIdAndActionId(ctx, nodeClassId, actionId)
			if err == nil {
				u.logger.Debug(ctx, tag, "node class action already assigned, skipping", domainmodels.LoggerMeta{
					"action":     action.Name,
					"node_class": nodeClassName,
				})
				continue
			}
			if !errors.Is(err, domainmodels.ErrTypeNotFound) {
				u.logger.Error(ctx, tag, "failed to read node class action", domainmodels.LoggerMeta{
					"err":        err,
					"action":     action.Name,
					"node_class": nodeClassName,
				})
				return err
			}

			if _, err := u.nodeClassAction.Create(ctx, nodeClassId, actionId, nil); err != nil {
				u.logger.Error(ctx, tag, "failed to create node class action", domainmodels.LoggerMeta{
					"err":        err,
					"action":     action.Name,
					"node_class": nodeClassName,
				})
				return err
			}

			u.logger.Info(ctx, tag, "node class action assigned", domainmodels.LoggerMeta{
				"action":     action.Name,
				"node_class": nodeClassName,
			})
		}
	}

	return nil
}

func (u *usecase) seedUsers(ctx context.Context, roleIds map[string]uuid.UUID) error {
	const tag = "seeder/seedUsers"

	for _, user := range u.data.Users {
		roleId, ok := roleIds[user.RoleName]
		if !ok {
			err := domainmodels.NewError("role not found for user seeding", domainmodels.ErrTypeNotFound, nil)
			u.logger.Error(ctx, tag, "unknown role", domainmodels.LoggerMeta{
				"err":      err,
				"username": user.Username,
				"role":     user.RoleName,
			})
			return err
		}

		_, err := u.user.ReadByUsername(ctx, user.Username)
		if err == nil {
			u.logger.Debug(ctx, tag, "user already exists, skipping", domainmodels.LoggerMeta{"username": user.Username})
			continue
		}
		if !errors.Is(err, domainmodels.ErrTypeNotFound) {
			u.logger.Error(ctx, tag, "failed to read user", domainmodels.LoggerMeta{
				"err":      err,
				"username": user.Username,
			})
			return err
		}

		passwordHash, err := u.password.Hash(user.Password)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to hash user password", domainmodels.LoggerMeta{
				"err":      err,
				"username": user.Username,
			})
			return err
		}

		if _, err := u.user.Create(ctx, roleId, user.Name, user.Bio, user.Username, passwordHash, nil); err != nil {
			u.logger.Error(ctx, tag, "failed to create user", domainmodels.LoggerMeta{
				"err":      err,
				"username": user.Username,
			})
			return err
		}

		u.logger.Info(ctx, tag, "user created", domainmodels.LoggerMeta{"username": user.Username})
	}

	return nil
}

func (u *usecase) seedInfraredDeviceTypes(ctx context.Context) (map[string]uuid.UUID, error) {
	const tag = "seeder/seedInfraredDeviceTypes"

	ids := make(map[string]uuid.UUID, len(u.data.InfraredDeviceTypes))
	for _, deviceType := range u.data.InfraredDeviceTypes {
		existing, err := u.infraredDeviceType.ReadByName(ctx, deviceType.Name)
		if err == nil {
			u.logger.Debug(ctx, tag, "infrared device type already exists, skipping", domainmodels.LoggerMeta{"name": deviceType.Name})
			ids[deviceType.Name] = existing.Id
			continue
		}
		if !errors.Is(err, domainmodels.ErrTypeNotFound) {
			u.logger.Error(ctx, tag, "failed to read infrared device type", domainmodels.LoggerMeta{
				"err":  err,
				"name": deviceType.Name,
			})
			return nil, err
		}

		id, err := u.infraredDeviceType.Create(ctx, deviceType.Name)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to create infrared device type", domainmodels.LoggerMeta{
				"err":  err,
				"name": deviceType.Name,
			})
			return nil, err
		}

		u.logger.Info(ctx, tag, "infrared device type created", domainmodels.LoggerMeta{"name": deviceType.Name})
		ids[deviceType.Name] = id
	}

	return ids, nil
}

func (u *usecase) seedInfraredStates(ctx context.Context, deviceTypeIds map[string]uuid.UUID) error {
	const tag = "seeder/seedInfraredStates"

	for _, state := range u.data.InfraredStates {
		deviceTypeId, ok := deviceTypeIds[state.DeviceTypeName]
		if !ok {
			err := domainmodels.NewError("infrared device type not found for infrared_state seeding", domainmodels.ErrTypeNotFound, nil)
			u.logger.Error(ctx, tag, "unknown infrared device type", domainmodels.LoggerMeta{
				"err":         err,
				"device_type": state.DeviceTypeName,
			})
			return err
		}

		existing, err := u.infraredState.ReadByDeviceTypeIdAndName(ctx, deviceTypeId, state.Name)
		if err == nil {
			u.logger.Debug(ctx, tag, "infrared state already exists, skipping", domainmodels.LoggerMeta{"name": state.Name})
			_ = existing
			continue
		}
		if !errors.Is(err, domainmodels.ErrTypeNotFound) {
			u.logger.Error(ctx, tag, "failed to read infrared state", domainmodels.LoggerMeta{
				"err":  err,
				"name": state.Name,
			})
			return err
		}

		if _, err := u.infraredState.Create(ctx, deviceTypeId, state.Name, domainmodels.InfraredStateType(state.Type)); err != nil {
			u.logger.Error(ctx, tag, "failed to create infrared state", domainmodels.LoggerMeta{
				"err":  err,
				"name": state.Name,
			})
			return err
		}

		u.logger.Info(ctx, tag, "infrared state created", domainmodels.LoggerMeta{"name": state.Name})
	}

	return nil
}
