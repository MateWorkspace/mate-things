package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredState interface {
	ListByDeviceTypeId(ctx context.Context, infraredDeviceTypeId uuid.UUID) ([]domainmodels.InfraredState, error)
	ReadByDeviceTypeIdAndName(ctx context.Context, infraredDeviceTypeId uuid.UUID, name string) (*domainmodels.InfraredState, error)
	Create(ctx context.Context, infraredDeviceTypeId uuid.UUID, name string, stateType domainmodels.InfraredStateType) (id uuid.UUID, err error)
}
