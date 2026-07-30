package domainusecasesnode

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type DeviceManagement interface {
	ReadById(ctx context.Context, request ReadNodeByIdRequest) (*domainmodels.Node, error)
	ReadByDeviceId(ctx context.Context, request ReadNodeByDeviceIdRequest) (*domainmodels.Node, error)
	ReadByPagination(ctx context.Context, request ReadNodesByPaginationRequest) ([]domainmodels.Node, int, error)
	UpdateById(ctx context.Context, request UpdateNodeRequest) error
	AssignFirmware(ctx context.Context, request AssignNodeFirmwareRequest) error
	DeleteById(ctx context.Context, request DeleteNodeRequest) error
}

type ReadNodeByIdRequest struct {
	Id uuid.UUID
}

type ReadNodeByDeviceIdRequest struct {
	DeviceId string
}

type ReadNodesByPaginationRequest struct {
	Page        int
	Limit       int
	Search      *string
	NodeClassId *uuid.UUID
	FirmwareId  *uuid.UUID
}

type UpdateNodeRequest struct {
	Id          uuid.UUID
	NodeClassId *uuid.UUID
	Name        *string
	FirmwareId  *uuid.UUID
	Description *string
	UpdatedBy   *uuid.UUID
}

type AssignNodeFirmwareRequest struct {
	Id         uuid.UUID
	FirmwareId uuid.UUID
	UpdatedBy  *uuid.UUID
}

type DeleteNodeRequest struct {
	Id        uuid.UUID
	DeletedBy *uuid.UUID
}
