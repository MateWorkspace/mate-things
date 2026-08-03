package domaincontractscache

import (
	"context"

	"github.com/google/uuid"
)

type NodeClassAction interface {
	GetById(ctx context.Context, id uuid.UUID) (item *NodeClassActionItem, hit bool, err error)
	SetById(ctx context.Context, id uuid.UUID, item *NodeClassActionItem) error
	DeleteById(ctx context.Context, id uuid.UUID) error
	GetByNodeClassIdAndActionId(ctx context.Context, nodeClassId uuid.UUID, actionId uuid.UUID) (item *NodeClassActionItem, hit bool, err error)
	SetByNodeClassIdAndActionId(ctx context.Context, nodeClassId uuid.UUID, actionId uuid.UUID, item *NodeClassActionItem) error
	DeleteByNodeClassIdAndActionId(ctx context.Context, nodeClassId uuid.UUID, actionId uuid.UUID) error
	GetPagination(ctx context.Context, page int, limit int, nodeClassId *uuid.UUID, actionId *uuid.UUID) (pagination Pagination[NodeClassActionItem], hit bool, err error)
	SetPagination(ctx context.Context, page int, limit int, nodeClassId *uuid.UUID, actionId *uuid.UUID, pagination Pagination[NodeClassActionItem]) error
	InvalidatePagination(ctx context.Context) error
	InvalidateByNodeClassId(ctx context.Context, nodeClassId uuid.UUID) error
	InvalidateByActionId(ctx context.Context, actionId uuid.UUID) error
	InvalidateAll(ctx context.Context) error
}
