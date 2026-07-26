package domaincontractscache

import (
	"context"

	"github.com/google/uuid"
)

type RolePermission interface {
	GetById(ctx context.Context, id uuid.UUID) (item *RolePermissionItem, hit bool, err error)
	SetById(ctx context.Context, id uuid.UUID, item *RolePermissionItem) error
	DeleteById(ctx context.Context, id uuid.UUID) error
	GetByRoleIdAndPermissionId(ctx context.Context, roleId uuid.UUID, permissionId uuid.UUID) (item *RolePermissionItem, hit bool, err error)
	SetByRoleIdAndPermissionId(ctx context.Context, roleId uuid.UUID, permissionId uuid.UUID, item *RolePermissionItem) error
	DeleteByRoleIdAndPermissionId(ctx context.Context, roleId uuid.UUID, permissionId uuid.UUID) error
	GetPagination(ctx context.Context, page int, limit int, roleId *uuid.UUID, permissionId *uuid.UUID) (pagination Pagination[RolePermissionItem], hit bool, err error)
	SetPagination(ctx context.Context, page int, limit int, roleId *uuid.UUID, permissionId *uuid.UUID, pagination Pagination[RolePermissionItem]) error
	InvalidatePagination(ctx context.Context) error
	InvalidateByRoleId(ctx context.Context, roleId uuid.UUID) error
	InvalidateByPermissionId(ctx context.Context, permissionId uuid.UUID) error
	InvalidateAll(ctx context.Context) error
}
