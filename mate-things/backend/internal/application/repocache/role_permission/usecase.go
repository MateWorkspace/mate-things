package applicationrepocacherolepermission

import (
	"context"

	domaincontractscache "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/cache"
	domaincontractsrepository "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	domainusecasesrepocache "github.com/ABA-Developer/nusapala-things/backend/internal/domain/usecases/repocache"
	"github.com/google/uuid"
)

type usecase struct {
	repository domaincontractsrepository.RolePermission
	cache      domaincontractscache.RolePermission
	roleCache  domaincontractscache.Role
	userCache  domaincontractscache.User
}

func NewRepoCacheImpl(
	repository domaincontractsrepository.RolePermission,
	cache domaincontractscache.RolePermission,
	roleCache domaincontractscache.Role,
	userCache domaincontractscache.User,
) domainusecasesrepocache.RolePermission {
	return &usecase{
		repository: repository,
		cache:      cache,
		roleCache:  roleCache,
		userCache:  userCache,
	}
}

func (u *usecase) Create(
	ctx context.Context,
	roleId uuid.UUID,
	permissionId uuid.UUID,
	createdBy *uuid.UUID,
) (uuid.UUID, error) {
	id, err := u.repository.Create(ctx, roleId, permissionId, createdBy)
	if err != nil {
		return uuid.Nil, err
	}

	if err := u.invalidateRelations(ctx); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (u *usecase) ReadById(
	ctx context.Context,
	id uuid.UUID,
) (*domainmodels.RolePermission, *domainmodels.Role, *domainmodels.Permission, error) {
	if item, hit, err := u.cache.GetById(ctx, id); err != nil {
		return nil, nil, nil, err
	} else if hit {
		return &item.RolePermission, &item.Role, &item.Permission, nil
	}

	rolePermission, role, permission, err := u.repository.ReadById(ctx, id)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := u.cache.SetById(ctx, id, &domaincontractscache.RolePermissionItem{
		RolePermission: *rolePermission,
		Role:           *role,
		Permission:     *permission,
	}); err != nil {
		return nil, nil, nil, err
	}

	return rolePermission, role, permission, nil
}

func (u *usecase) ReadByRoleIdAndPermissionId(
	ctx context.Context,
	roleId uuid.UUID,
	permissionId uuid.UUID,
) (*domainmodels.RolePermission, *domainmodels.Role, *domainmodels.Permission, error) {
	if item, hit, err := u.cache.GetByRoleIdAndPermissionId(ctx, roleId, permissionId); err != nil {
		return nil, nil, nil, err
	} else if hit {
		return &item.RolePermission, &item.Role, &item.Permission, nil
	}

	rolePermission, role, permission, err := u.repository.ReadByRoleIdAndPermissionId(ctx, roleId, permissionId)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := u.cache.SetByRoleIdAndPermissionId(ctx, roleId, permissionId, &domaincontractscache.RolePermissionItem{
		RolePermission: *rolePermission,
		Role:           *role,
		Permission:     *permission,
	}); err != nil {
		return nil, nil, nil, err
	}

	return rolePermission, role, permission, nil
}

func (u *usecase) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	roleId *uuid.UUID,
	permissionId *uuid.UUID,
) ([]domainmodels.RolePermission, []domainmodels.Role, []domainmodels.Permission, int, error) {
	if pagination, hit, err := u.cache.GetPagination(ctx, page, limit, roleId, permissionId); err != nil {
		return nil, nil, nil, 0, err
	} else if hit {
		rolePermissions, roles, permissions := splitRolePermissionItems(pagination.Items)
		return rolePermissions, roles, permissions, pagination.Total, nil
	}

	rolePermissions, roles, permissions, total, err := u.repository.ReadByPagination(ctx, page, limit, roleId, permissionId)
	if err != nil {
		return nil, nil, nil, 0, err
	}

	if err := u.cache.SetPagination(ctx, page, limit, roleId, permissionId, domaincontractscache.Pagination[domaincontractscache.RolePermissionItem]{
		Items: joinRolePermissionItems(rolePermissions, roles, permissions),
		Total: total,
	}); err != nil {
		return nil, nil, nil, 0, err
	}

	return rolePermissions, roles, permissions, total, nil
}

func (u *usecase) DeleteById(ctx context.Context, id uuid.UUID) error {
	if err := u.repository.DeleteById(ctx, id); err != nil {
		return err
	}

	return u.invalidateRelations(ctx)
}

func (u *usecase) DeleteByRoleIdAndPermissionId(
	ctx context.Context,
	roleId *uuid.UUID,
	permissionId *uuid.UUID,
) error {
	if err := u.repository.DeleteByRoleIdAndPermissionId(ctx, roleId, permissionId); err != nil {
		return err
	}

	return u.invalidateRelations(ctx)
}

func (u *usecase) invalidateRelations(ctx context.Context) error {
	if err := u.cache.InvalidateAll(ctx); err != nil {
		return err
	}
	if err := u.roleCache.InvalidatePermissions(ctx); err != nil {
		return err
	}
	if err := u.userCache.InvalidatePermissions(ctx); err != nil {
		return err
	}
	return nil
}

func joinRolePermissionItems(
	rolePermissions []domainmodels.RolePermission,
	roles []domainmodels.Role,
	permissions []domainmodels.Permission,
) []domaincontractscache.RolePermissionItem {
	items := make([]domaincontractscache.RolePermissionItem, len(rolePermissions))
	for i := range rolePermissions {
		items[i].RolePermission = rolePermissions[i]
		if i < len(roles) {
			items[i].Role = roles[i]
		}
		if i < len(permissions) {
			items[i].Permission = permissions[i]
		}
	}
	return items
}

func splitRolePermissionItems(
	items []domaincontractscache.RolePermissionItem,
) ([]domainmodels.RolePermission, []domainmodels.Role, []domainmodels.Permission) {
	rolePermissions := make([]domainmodels.RolePermission, len(items))
	roles := make([]domainmodels.Role, len(items))
	permissions := make([]domainmodels.Permission, len(items))

	for i, item := range items {
		rolePermissions[i] = item.RolePermission
		roles[i] = item.Role
		permissions[i] = item.Permission
	}

	return rolePermissions, roles, permissions
}
