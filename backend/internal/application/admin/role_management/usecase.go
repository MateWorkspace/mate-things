package applicationadminrolemanagement

import (
	"context"

	applicationshared "github.com/MateWorkspace/mate-things/backend/internal/application/shared"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesadmin "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/admin"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
	"github.com/google/uuid"
)

type usecase struct {
	role           domainusecasesrepocache.Role
	rolePermission domainusecasesrepocache.RolePermission
	logger         domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	role domainusecasesrepocache.Role,
	rolePermission domainusecasesrepocache.RolePermission,
	logger domaincontractslogger.Leveled,
) domainusecasesadmin.RoleManagement {
	return &usecase{
		role:           role,
		rolePermission: rolePermission,
		logger:         logger,
	}
}

func (u *usecase) Create(ctx context.Context, request domainusecasesadmin.CreateRoleRequest) (uuid.UUID, error) {
	const tag = "admin/role_management/Create"

	name, err := applicationshared.RequiredRoleName(request.Name, "name")
	if err != nil {
		return uuid.Nil, err
	}

	id, err := u.role.Create(ctx, name, request.Description, nil, request.CreatedBy)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to create role", domainmodels.LoggerMeta{
			"err":        err,
			"created_by": request.CreatedBy,
		})
		return uuid.Nil, err
	}

	return id, nil
}

func (u *usecase) ReadById(
	ctx context.Context,
	request domainusecasesadmin.ReadRoleByIdRequest,
) (*domainmodels.Role, error) {
	const tag = "admin/role_management/ReadById"

	role, err := u.role.ReadById(ctx, request.Id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read role", domainmodels.LoggerMeta{
			"err": err,
			"id":  request.Id,
		})
		return nil, err
	}

	return role, nil
}

func (u *usecase) ReadByName(
	ctx context.Context,
	request domainusecasesadmin.ReadRoleByNameRequest,
) (*domainmodels.Role, error) {
	const tag = "admin/role_management/ReadByName"

	role, err := u.role.ReadByName(ctx, request.Name)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read role", domainmodels.LoggerMeta{
			"err":  err,
			"name": request.Name,
		})
		return nil, err
	}

	return role, nil
}

func (u *usecase) ReadDefault(
	ctx context.Context,
	_ domainusecasesadmin.ReadDefaultRoleRequest,
) (*domainmodels.Role, error) {
	const tag = "admin/role_management/ReadDefault"

	role, err := u.role.ReadDefault(ctx)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read default role", domainmodels.LoggerMeta{
			"err": err,
		})
		return nil, err
	}

	return role, nil
}

func (u *usecase) ReadPermissions(
	ctx context.Context,
	request domainusecasesadmin.ReadRolePermissionsRequest,
) ([]domainmodels.Permission, error) {
	const tag = "admin/role_management/ReadPermissions"

	permissions, err := u.role.ReadPermissions(ctx, request.RoleId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read role permissions", domainmodels.LoggerMeta{
			"err":     err,
			"role_id": request.RoleId,
		})
		return nil, err
	}

	return permissions, nil
}

func (u *usecase) ReadByPagination(
	ctx context.Context,
	request domainusecasesadmin.ReadRolesByPaginationRequest,
) ([]domainmodels.Role, int, error) {
	const tag = "admin/role_management/ReadByPagination"

	roles, total, err := u.role.ReadByPagination(ctx, request.Page, request.Limit, request.Search)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read roles", domainmodels.LoggerMeta{
			"err":   err,
			"page":  request.Page,
			"limit": request.Limit,
		})
		return nil, 0, err
	}

	return roles, total, nil
}

func (u *usecase) UpdateById(ctx context.Context, request domainusecasesadmin.UpdateRoleRequest) error {
	const tag = "admin/role_management/UpdateById"

	name, err := applicationshared.OptionalRoleName(request.Name, "name")
	if err != nil {
		return err
	}

	if err := u.role.UpdateById(
		ctx,
		request.Id,
		name,
		request.Description,
		nil,
		nil,
		request.UpdatedBy,
	); err != nil {
		u.logger.Error(ctx, tag, "failed to update role", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	return nil
}

func (u *usecase) SetDefaultRole(ctx context.Context, request domainusecasesadmin.SetDefaultRoleRequest) error {
	const tag = "admin/role_management/SetDefaultRole"

	isDefault := true
	if err := u.role.UpdateById(ctx, request.Id, nil, nil, &isDefault, nil, request.UpdatedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to update default role", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	return nil
}

func (u *usecase) DeleteById(ctx context.Context, request domainusecasesadmin.DeleteRoleRequest) error {
	const tag = "admin/role_management/DeleteById"

	if err := u.role.DeleteById(ctx, request.Id, request.DeletedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to delete role", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"deleted_by": request.DeletedBy,
		})
		return err
	}

	return nil
}

func (u *usecase) AssignPermission(
	ctx context.Context,
	request domainusecasesadmin.AssignRolePermissionRequest,
) (uuid.UUID, error) {
	const tag = "admin/role_management/AssignPermission"

	id, err := u.rolePermission.Create(ctx, request.RoleId, request.PermissionId, request.CreatedBy)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to assign role permission", domainmodels.LoggerMeta{
			"err":           err,
			"role_id":       request.RoleId,
			"permission_id": request.PermissionId,
			"created_by":    request.CreatedBy,
		})
		return uuid.Nil, err
	}

	return id, nil
}

func (u *usecase) RevokePermission(
	ctx context.Context,
	request domainusecasesadmin.RevokeRolePermissionRequest,
) error {
	const tag = "admin/role_management/RevokePermission"

	if err := u.rolePermission.DeleteByRoleIdAndPermissionId(ctx, &request.RoleId, &request.PermissionId); err != nil {
		u.logger.Error(ctx, tag, "failed to revoke role permission", domainmodels.LoggerMeta{
			"err":           err,
			"role_id":       request.RoleId,
			"permission_id": request.PermissionId,
		})
		return err
	}

	return nil
}

func (u *usecase) ReadRolePermissionById(
	ctx context.Context,
	request domainusecasesadmin.ReadRolePermissionByIdRequest,
) (domainusecasesadmin.RolePermissionResult, error) {
	const tag = "admin/role_management/ReadRolePermissionById"

	rolePermission, role, permission, err := u.rolePermission.ReadById(ctx, request.Id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read role permission", domainmodels.LoggerMeta{
			"err": err,
			"id":  request.Id,
		})
		return domainusecasesadmin.RolePermissionResult{}, err
	}

	return domainusecasesadmin.RolePermissionResult{
		RolePermission: *rolePermission,
		Role:           *role,
		Permission:     *permission,
	}, nil
}

func (u *usecase) ReadRolePermissionByRoleIdAndPermissionId(
	ctx context.Context,
	request domainusecasesadmin.ReadRolePermissionByRoleIdAndPermissionIdRequest,
) (domainusecasesadmin.RolePermissionResult, error) {
	const tag = "admin/role_management/ReadRolePermissionByRoleIdAndPermissionId"

	rolePermission, role, permission, err := u.rolePermission.ReadByRoleIdAndPermissionId(
		ctx,
		request.RoleId,
		request.PermissionId,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read role permission", domainmodels.LoggerMeta{
			"err":           err,
			"role_id":       request.RoleId,
			"permission_id": request.PermissionId,
		})
		return domainusecasesadmin.RolePermissionResult{}, err
	}

	return domainusecasesadmin.RolePermissionResult{
		RolePermission: *rolePermission,
		Role:           *role,
		Permission:     *permission,
	}, nil
}

func (u *usecase) ReadRolePermissionsByPagination(
	ctx context.Context,
	request domainusecasesadmin.ReadRolePermissionsByPaginationRequest,
) ([]domainmodels.RolePermission, []domainmodels.Role, []domainmodels.Permission, int, error) {
	const tag = "admin/role_management/ReadRolePermissionsByPagination"

	rolePermissions, roles, permissions, total, err := u.rolePermission.ReadByPagination(
		ctx,
		request.Page,
		request.Limit,
		request.RoleId,
		request.PermissionId,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read role permissions", domainmodels.LoggerMeta{
			"err":           err,
			"page":          request.Page,
			"limit":         request.Limit,
			"role_id":       request.RoleId,
			"permission_id": request.PermissionId,
		})
		return nil, nil, nil, 0, err
	}

	return rolePermissions, roles, permissions, total, nil
}
