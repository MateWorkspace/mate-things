package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredDevice interface {
	Create(ctx context.Context, infraredDeviceTypeId uuid.UUID, brand string, model string, createdBy *uuid.UUID) (id uuid.UUID, err error)
	GetById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredDevice, error)
	DeleteById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
}
