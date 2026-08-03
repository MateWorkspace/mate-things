package domainusecasesaction

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type Definition interface {
	Create(ctx context.Context, request CreateActionRequest) (uuid.UUID, error)
	ReadById(ctx context.Context, request ReadActionByIdRequest) (*domainmodels.Action, error)
	ReadByName(ctx context.Context, request ReadActionByNameRequest) (*domainmodels.Action, error)
	ReadByPagination(ctx context.Context, request ReadActionsByPaginationRequest) ([]domainmodels.ActionListItem, int, error)
	UpdateById(ctx context.Context, request UpdateActionRequest) error
	DeleteById(ctx context.Context, request DeleteActionRequest) error
}

type CreateActionRequest struct {
	Name                 string
	Description          *string
	PayloadSchemaName    string
	PayloadSchemaVersion int32
	CreatedBy            *uuid.UUID
}

type ReadActionByIdRequest struct {
	Id uuid.UUID
}

type ReadActionByNameRequest struct {
	Name string
}

type ReadActionsByPaginationRequest struct {
	Page                 int
	Limit                int
	Search               *string
	NodeClassId          *uuid.UUID
	PayloadSchemaName    *string
	PayloadSchemaVersion *int32
}

type UpdateActionRequest struct {
	Id                   uuid.UUID
	Name                 *string
	Description          *string
	PayloadSchemaName    *string
	PayloadSchemaVersion *int32
	UpdatedBy            *uuid.UUID
}

type DeleteActionRequest struct {
	Id        uuid.UUID
	DeletedBy *uuid.UUID
}
