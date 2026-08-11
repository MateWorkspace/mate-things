package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredStateDeviceDefinition interface {
	CreateMany(ctx context.Context, definitions []domainmodels.InfraredStateDeviceDefinition, createdBy *uuid.UUID) error
	ReadListByDeviceId(ctx context.Context, infraredDeviceId uuid.UUID) ([]domainmodels.InfraredStateDeviceDefinition, error)
	DeleteById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
}
