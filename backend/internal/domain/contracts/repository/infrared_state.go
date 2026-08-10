package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredState interface {
	ReadListByDeviceTypeId(ctx context.Context, infraredDeviceTypeId uuid.UUID) ([]domainmodels.InfraredState, error)
	ReadByDeviceTypeIdAndName(ctx context.Context, infraredDeviceTypeId uuid.UUID, name string) (*domainmodels.InfraredState, error)
	Create(ctx context.Context, infraredDeviceTypeId uuid.UUID, name string, stateType domainmodels.InfraredStateType, createdBy *uuid.UUID) (id uuid.UUID, err error)
	DeleteById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
}
