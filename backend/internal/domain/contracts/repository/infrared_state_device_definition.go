package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredStateDeviceDefinition interface {
	// CreateMany inserts one row per definition. All definitions must share
	// the given infraredDeviceId; this is validated by the usecase, not here.
	CreateMany(ctx context.Context, definitions []domainmodels.InfraredStateDeviceDefinition) error
	ListByDeviceId(ctx context.Context, infraredDeviceId uuid.UUID) ([]domainmodels.InfraredStateDeviceDefinition, error)
}
