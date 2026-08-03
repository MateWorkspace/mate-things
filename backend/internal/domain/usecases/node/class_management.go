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
	AssignAction(ctx context.Context, request AssignNodeClassActionRequest) (uuid.UUID, error)
	RevokeAction(ctx context.Context, request RevokeNodeClassActionRequest) error
	ReadActions(ctx context.Context, request ReadNodeClassActionsRequest) ([]domainmodels.Action, error)
	ReadNodeClassActionById(ctx context.Context, request ReadNodeClassActionByIdRequest) (NodeClassActionResult, error)
	ReadNodeClassActionByNodeClassIdAndActionId(ctx context.Context, request ReadNodeClassActionByNodeClassIdAndActionIdRequest) (NodeClassActionResult, error)
	ReadNodeClassActionsByPagination(ctx context.Context, request ReadNodeClassActionsByPaginationRequest) ([]domainmodels.NodeClassAction, []domainmodels.NodeClass, []domainmodels.Action, int, error)
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

type AssignNodeClassActionRequest struct {
	NodeClassId uuid.UUID
	ActionId    uuid.UUID
	CreatedBy   *uuid.UUID
}

type RevokeNodeClassActionRequest struct {
	NodeClassId uuid.UUID
	ActionId    uuid.UUID
}

type ReadNodeClassActionsRequest struct {
	NodeClassId uuid.UUID
}

type ReadNodeClassActionByIdRequest struct {
	Id uuid.UUID
}

type ReadNodeClassActionByNodeClassIdAndActionIdRequest struct {
	NodeClassId uuid.UUID
	ActionId    uuid.UUID
}

type ReadNodeClassActionsByPaginationRequest struct {
	Page        int
	Limit       int
	NodeClassId *uuid.UUID
	ActionId    *uuid.UUID
}

type NodeClassActionResult struct {
	NodeClassAction domainmodels.NodeClassAction
	NodeClass       domainmodels.NodeClass
	Action          domainmodels.Action
}
