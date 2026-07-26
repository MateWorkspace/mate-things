package domainusecasesadmin

import (
	"context"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type UserManagement interface {
	Create(ctx context.Context, request CreateUserRequest) (uuid.UUID, error)
	ReadById(ctx context.Context, request ReadUserByIdRequest) (*domainmodels.User, error)
	ReadByUsername(ctx context.Context, request ReadUserByUsernameRequest) (*domainmodels.User, error)
	ReadPermissions(ctx context.Context, request ReadUserPermissionsRequest) ([]domainmodels.Permission, error)
	ReadByPagination(ctx context.Context, request ReadUsersByPaginationRequest) ([]domainmodels.User, int, error)
	UpdateById(ctx context.Context, request UpdateUserRequest) error
	ResetPassword(ctx context.Context, request ResetUserPasswordRequest) error
	DeleteById(ctx context.Context, request DeleteUserRequest) error
}

type CreateUserRequest struct {
	RoleId    uuid.UUID
	Name      string
	Bio       *string
	Username  string
	Password  string
	CreatedBy *uuid.UUID
}

type ReadUserByIdRequest struct {
	Id uuid.UUID
}

type ReadUserByUsernameRequest struct {
	Username string
}

type ReadUserPermissionsRequest struct {
	UserId uuid.UUID
}

type ReadUsersByPaginationRequest struct {
	Page   int
	Limit  int
	Search *string
	RoleId *uuid.UUID
}

type UpdateUserRequest struct {
	Id        uuid.UUID
	RoleId    *uuid.UUID
	Name      *string
	Bio       *string
	Username  *string
	UpdatedBy *uuid.UUID
}

type ResetUserPasswordRequest struct {
	Id        uuid.UUID
	Password  string
	UpdatedBy *uuid.UUID
}

type DeleteUserRequest struct {
	Id        uuid.UUID
	DeletedBy *uuid.UUID
}
