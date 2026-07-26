package domainusecasesnode

import (
	"context"
	"io"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type FirmwareManagement interface {
	Create(ctx context.Context, request CreateFirmwareRequest) (CreateFirmwareResult, error)
	ReadById(ctx context.Context, request ReadFirmwareByIdRequest) (*domainmodels.Firmware, error)
	ReadByName(ctx context.Context, request ReadFirmwareByNameRequest) (*domainmodels.Firmware, error)
	ReadAvailableByNodeId(ctx context.Context, request ReadAvailableFirmwaresByNodeIdRequest) ([]domainmodels.Firmware, int, error)
	ReadByNodeClassIdAndPagination(ctx context.Context, request ReadFirmwaresByNodeClassIdAndPaginationRequest) ([]domainmodels.Firmware, int, error)
	ReadByPagination(ctx context.Context, request ReadFirmwaresByPaginationRequest) ([]domainmodels.Firmware, int, error)
	UpdateById(ctx context.Context, request UpdateFirmwareRequest) error
	ReplaceBinaryById(ctx context.Context, request ReplaceFirmwareBinaryByIdRequest) (FirmwareBinaryStatResult, error)
	OpenBinaryById(ctx context.Context, request OpenFirmwareBinaryByIdRequest) (OpenFirmwareBinaryResult, error)
	OpenBinaryByName(ctx context.Context, request OpenFirmwareBinaryByNameRequest) (OpenFirmwareBinaryResult, error)
	StatBinaryByName(ctx context.Context, request StatFirmwareBinaryByNameRequest) (FirmwareBinaryStatResult, error)
	DeleteById(ctx context.Context, request DeleteFirmwareRequest) error
}

type CreateFirmwareRequest struct {
	NodeClassId uuid.UUID
	Name        string
	Content     io.Reader
	CreatedBy   *uuid.UUID
}

type CreateFirmwareResult struct {
	Id         uuid.UUID
	Size       int32
	Checksum   string
	BinaryPath string
}

type ReadFirmwareByIdRequest struct {
	Id uuid.UUID
}

type ReadFirmwareByNameRequest struct {
	Name string
}

type ReadAvailableFirmwaresByNodeIdRequest struct {
	NodeId uuid.UUID
	Page   int
	Limit  int
	Search *string
}

type ReadFirmwaresByNodeClassIdAndPaginationRequest struct {
	NodeClassId uuid.UUID
	Page        int
	Limit       int
	Search      *string
}

type ReadFirmwaresByPaginationRequest struct {
	Page        int
	Limit       int
	Search      *string
	NodeClassId *uuid.UUID
}

type UpdateFirmwareRequest struct {
	Id          uuid.UUID
	NodeClassId *uuid.UUID
	Name        *string
	UpdatedBy   *uuid.UUID
}

type ReplaceFirmwareBinaryByIdRequest struct {
	Id        uuid.UUID
	Content   io.Reader
	UpdatedBy *uuid.UUID
}

type OpenFirmwareBinaryByIdRequest struct {
	Id uuid.UUID
}

type OpenFirmwareBinaryByNameRequest struct {
	Name string
}

type OpenFirmwareBinaryResult struct {
	Firmware domainmodels.Firmware
	Content  io.ReadCloser
}

type StatFirmwareBinaryByNameRequest struct {
	Name string
}

type FirmwareBinaryStatResult struct {
	BinaryPath string
	Size       int32
	Checksum   string
}

type DeleteFirmwareRequest struct {
	Id        uuid.UUID
	DeletedBy *uuid.UUID
}
