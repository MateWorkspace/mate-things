package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type LlmConfig interface {
	Get(ctx context.Context) (*domainmodels.LlmConfig, error)
	Upsert(ctx context.Context, provider domainmodels.LlmProvider, model string, apiKeyEncrypted []byte, baseURL *string, updatedBy *uuid.UUID) (id uuid.UUID, err error)
}
