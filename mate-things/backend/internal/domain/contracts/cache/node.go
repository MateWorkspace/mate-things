package domaincontractscache

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type Node interface {
	GetById(ctx context.Context, id uuid.UUID) (node *domainmodels.Node, hit bool, err error)
	SetById(ctx context.Context, id uuid.UUID, node *domainmodels.Node) error
	DeleteById(ctx context.Context, id uuid.UUID) error
	GetByDeviceId(ctx context.Context, deviceId string) (node *domainmodels.Node, hit bool, err error)
	SetByDeviceId(ctx context.Context, deviceId string, node *domainmodels.Node) error
	DeleteByDeviceId(ctx context.Context, deviceId string) error
	GetPagination(ctx context.Context, page int, limit int, search *string, nodeClassId *uuid.UUID, firmwareName *string) (pagination Pagination[domainmodels.Node], hit bool, err error)
	SetPagination(ctx context.Context, page int, limit int, search *string, nodeClassId *uuid.UUID, firmwareName *string, pagination Pagination[domainmodels.Node]) error
	InvalidatePagination(ctx context.Context) error
	InvalidateAll(ctx context.Context) error
}
