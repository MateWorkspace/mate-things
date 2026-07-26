package domainusecasesnode

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type ClassManagement interface {
	Create(ctx context.Context, request CreateNodeClassRequest) (uuid.UUID, error)
	ReadById(ctx context.Context, request ReadNodeClassByIdRequest) (*domainmodels.NodeClass, error)
	ReadByName(ctx context.Context, request ReadNodeClassByNameRequest) (*domainmodels.NodeClass, error)
	ReadByPagination(ctx context.Context, request ReadNodeClassesByPaginationRequest) ([]domainmodels.NodeClass, int, error)
	UpdateById(ctx context.Context, request UpdateNodeClassRequest) error
	DeleteById(ctx context.Context, request DeleteNodeClassRequest) error
}

type CreateNodeClassRequest struct {
	Name        string
	Description *string
	CreatedBy   *uuid.UUID
}

type ReadNodeClassByIdRequest struct {
	Id uuid.UUID
}

type ReadNodeClassByNameRequest struct {
	Name string
}

type ReadNodeClassesByPaginationRequest struct {
	Page   int
	Limit  int
	Search *string
}

type UpdateNodeClassRequest struct {
	Id          uuid.UUID
	Name        *string
	Description *string
	UpdatedBy   *uuid.UUID
}

type DeleteNodeClassRequest struct {
	Id        uuid.UUID
	DeletedBy *uuid.UUID
}
