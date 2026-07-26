package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type RolePermission interface {
	Create(
		ctx context.Context,
		roleId uuid.UUID,
		permissionId uuid.UUID,
		createdBy *uuid.UUID,
	) (id uuid.UUID, err error)

	ReadById(
		ctx context.Context,
		id uuid.UUID,
	) (rolePermission *domainmodels.RolePermission, role *domainmodels.Role, permission *domainmodels.Permission, err error)

	ReadByRoleIdAndPermissionId(
		ctx context.Context,
		roleId uuid.UUID,
		permissionId uuid.UUID,
	) (rolePermission *domainmodels.RolePermission, role *domainmodels.Role, permission *domainmodels.Permission, err error)

	ReadByPagination(
		ctx context.Context,
		page int,
		limit int,
		roleId *uuid.UUID,
		permissionId *uuid.UUID,
	) (rolePermissions []domainmodels.RolePermission, roles []domainmodels.Role, permissions []domainmodels.Permission, total int, err error)

	DeleteById(
		ctx context.Context,
		id uuid.UUID,
	) (err error)

	DeleteByRoleIdAndPermissionId(
		ctx context.Context,
		roleId *uuid.UUID,
		permissionId *uuid.UUID,
	) (err error)
}
