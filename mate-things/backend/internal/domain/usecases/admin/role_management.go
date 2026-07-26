package domainusecasesadmin

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type RoleManagement interface {
	Create(ctx context.Context, request CreateRoleRequest) (uuid.UUID, error)
	ReadById(ctx context.Context, request ReadRoleByIdRequest) (*domainmodels.Role, error)
	ReadByName(ctx context.Context, request ReadRoleByNameRequest) (*domainmodels.Role, error)
	ReadDefault(ctx context.Context, request ReadDefaultRoleRequest) (*domainmodels.Role, error)
	ReadPermissions(ctx context.Context, request ReadRolePermissionsRequest) ([]domainmodels.Permission, error)
	ReadByPagination(ctx context.Context, request ReadRolesByPaginationRequest) ([]domainmodels.Role, int, error)
	UpdateById(ctx context.Context, request UpdateRoleRequest) error
	SetDefaultRole(ctx context.Context, request SetDefaultRoleRequest) error
	DeleteById(ctx context.Context, request DeleteRoleRequest) error
	AssignPermission(ctx context.Context, request AssignRolePermissionRequest) (uuid.UUID, error)
	RevokePermission(ctx context.Context, request RevokeRolePermissionRequest) error
	ReadRolePermissionById(ctx context.Context, request ReadRolePermissionByIdRequest) (RolePermissionResult, error)
	ReadRolePermissionByRoleIdAndPermissionId(ctx context.Context, request ReadRolePermissionByRoleIdAndPermissionIdRequest) (RolePermissionResult, error)
	ReadRolePermissionsByPagination(ctx context.Context, request ReadRolePermissionsByPaginationRequest) ([]domainmodels.RolePermission, []domainmodels.Role, []domainmodels.Permission, int, error)
}

type CreateRoleRequest struct {
	Name        string
	Description *string
	CreatedBy   *uuid.UUID
}

type ReadRoleByIdRequest struct {
	Id uuid.UUID
}

type ReadRoleByNameRequest struct {
	Name string
}

type ReadDefaultRoleRequest struct{}

type ReadRolePermissionsRequest struct {
	RoleId uuid.UUID
}

type ReadRolesByPaginationRequest struct {
	Page   int
	Limit  int
	Search *string
}

type UpdateRoleRequest struct {
	Id          uuid.UUID
	Name        *string
	Description *string
	UpdatedBy   *uuid.UUID
}

type SetDefaultRoleRequest struct {
	Id        uuid.UUID
	UpdatedBy *uuid.UUID
}

type DeleteRoleRequest struct {
	Id        uuid.UUID
	DeletedBy *uuid.UUID
}

type AssignRolePermissionRequest struct {
	RoleId       uuid.UUID
	PermissionId uuid.UUID
	CreatedBy    *uuid.UUID
}

type RevokeRolePermissionRequest struct {
	RoleId       uuid.UUID
	PermissionId uuid.UUID
}

type ReadRolePermissionByIdRequest struct {
	Id uuid.UUID
}

type ReadRolePermissionByRoleIdAndPermissionIdRequest struct {
	RoleId       uuid.UUID
	PermissionId uuid.UUID
}

type ReadRolePermissionsByPaginationRequest struct {
	Page         int
	Limit        int
	RoleId       *uuid.UUID
	PermissionId *uuid.UUID
}

type RolePermissionResult struct {
	RolePermission domainmodels.RolePermission
	Role           domainmodels.Role
	Permission     domainmodels.Permission
}
