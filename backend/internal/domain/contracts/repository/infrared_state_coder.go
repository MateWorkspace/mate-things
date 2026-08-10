package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredStateCoder interface {
	Create(ctx context.Context, coder domainmodels.InfraredStateCoder) (id uuid.UUID, err error)
	GetBySessionId(ctx context.Context, sessionId uuid.UUID) (*domainmodels.InfraredStateCoder, error)
	GetById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredStateCoder, error)
	Activate(ctx context.Context, coderId uuid.UUID, deviceId uuid.UUID) error
}
