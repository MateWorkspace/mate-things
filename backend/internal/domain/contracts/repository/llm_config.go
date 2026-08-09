package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type LlmConfig interface {
	// Get returns (nil, nil) when no configuration has been set yet — this
	// is a normal, expected state, not an error.
	Get(ctx context.Context) (*domainmodels.LlmConfig, error)
	Upsert(ctx context.Context, provider domainmodels.LlmProvider, model string, apiKeyEncrypted []byte, baseURL *string, updatedBy *uuid.UUID) (id uuid.UUID, err error)
}
