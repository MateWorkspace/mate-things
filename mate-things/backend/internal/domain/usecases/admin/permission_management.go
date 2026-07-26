package domainusecasesadmin

import (
	"context"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type PermissionManagement interface {
	Create(ctx context.Context, request CreatePermissionRequest) (uuid.UUID, error)
	ReadById(ctx context.Context, request ReadPermissionByIdRequest) (*domainmodels.Permission, error)
	ReadByName(ctx context.Context, request ReadPermissionByNameRequest) (*domainmodels.Permission, error)
	ReadByPagination(ctx context.Context, request ReadPermissionsByPaginationRequest) ([]domainmodels.Permission, int, error)
	UpdateById(ctx context.Context, request UpdatePermissionRequest) error
	DeleteById(ctx context.Context, request DeletePermissionRequest) error
}

type CreatePermissionRequest struct {
	Name        string
	Description *string
	CreatedBy   *uuid.UUID
}

type ReadPermissionByIdRequest struct {
	Id uuid.UUID
}

type ReadPermissionByNameRequest struct {
	Name string
}

type ReadPermissionsByPaginationRequest struct {
	Page   int
	Limit  int
	Search *string
}

type UpdatePermissionRequest struct {
	Id          uuid.UUID
	Name        *string
	Description *string
	UpdatedBy   *uuid.UUID
}

type DeletePermissionRequest struct {
	Id        uuid.UUID
	DeletedBy *uuid.UUID
}
