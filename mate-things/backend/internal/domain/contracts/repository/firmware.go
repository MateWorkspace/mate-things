package domaincontractsrepository

import (
	"context"
	"encoding/json"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type Firmware interface {
	Create(
		ctx context.Context,
		nodeClassId uuid.UUID,
		name string,
		size int32,
		checksum string,
		binaryPath string,
		createdBy *uuid.UUID,
	) (id uuid.UUID, err error)

	ReadById(
		ctx context.Context,
		id uuid.UUID,
	) (firmware *domainmodels.Firmware, err error)

	ReadByName(
		ctx context.Context,
		name string,
	) (firmware *domainmodels.Firmware, err error)

	ReadByNodeClassIdAndPagination(
		ctx context.Context,
		nodeClassId uuid.UUID,
		page int,
		limit int,
		search *string,
	) (firmwares []domainmodels.Firmware, total int, err error)

	ReadByPagination(
		ctx context.Context,
		page int,
		limit int,
		search *string,
		nodeClassId *uuid.UUID,
	) (firmwares []domainmodels.Firmware, total int, err error)

	UpdateById(
		ctx context.Context,
		id uuid.UUID,
		nodeClassId *uuid.UUID,
		name *string,
		size *int32,
		checksum *string,
		binaryPath *string,
		preferences *json.RawMessage,
		updatedBy *uuid.UUID,
	) (err error)

	DeleteById(
		ctx context.Context,
		id uuid.UUID,
		deletedBy *uuid.UUID,
	) (err error)
}
