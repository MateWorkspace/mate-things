package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredDeviceType interface {
	List(ctx context.Context) ([]domainmodels.InfraredDeviceType, error)
	ReadByName(ctx context.Context, name string) (*domainmodels.InfraredDeviceType, error)
	Create(ctx context.Context, name string, createdBy *uuid.UUID) (id uuid.UUID, err error)
	DeleteById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
}
