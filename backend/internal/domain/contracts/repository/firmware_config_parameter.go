package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type FirmwareConfigParameter interface {
	// ReplaceForFirmwareId soft-deletes every existing row for firmwareId
	// not present in params, then upserts each entry in params
	// (ON CONFLICT (firmware_id, key) DO UPDATE), all in one transaction.
	ReplaceForFirmwareId(
		ctx context.Context,
		firmwareId uuid.UUID,
		params []FirmwareConfigParameterInput,
		actorId *uuid.UUID,
	) (err error)

	ReadByFirmwareId(
		ctx context.Context,
		firmwareId uuid.UUID,
	) (params []domainmodels.FirmwareConfigParameter, err error)
}

type FirmwareConfigParameterInput struct {
	Key       string
	ValueType string
}
