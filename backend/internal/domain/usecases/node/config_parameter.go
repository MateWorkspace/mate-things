package domainusecasesnode

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type ConfigParameter interface {
	// ValidateSchema checks a config schema without persisting it, so callers
	// can reject a bad schema before any irreversible side effect.
	ValidateSchema(parameters []ConfigParameterInput) error
	ReplaceForFirmware(ctx context.Context, request ReplaceConfigParametersRequest) error
	ReadByFirmwareId(ctx context.Context, request ReadConfigParametersByFirmwareIdRequest) ([]domainmodels.FirmwareConfigParameter, error)
}

type ConfigParameterInput struct {
	Key       string
	ValueType string
}

type ReplaceConfigParametersRequest struct {
	FirmwareId uuid.UUID
	Parameters []ConfigParameterInput
	ActorId    *uuid.UUID
}

type ReadConfigParametersByFirmwareIdRequest struct {
	FirmwareId uuid.UUID
}
