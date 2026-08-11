package domainusecasesadmin

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type LlmConfigManagement interface {
	Get(ctx context.Context) (*domainmodels.LlmConfig, error)
	Update(ctx context.Context, request UpdateLlmConfigRequest) error
	TestConnection(ctx context.Context) (domainmodels.LlmClientStatus, error)
	TestConnectionWithConfig(ctx context.Context, request TestLlmConnectionWithConfigRequest) (domainmodels.LlmClientStatus, error)
}

type UpdateLlmConfigRequest struct {
	Provider  domainmodels.LlmProvider
	Model     string
	ApiKey    *string
	BaseURL   *string
	UpdatedBy *uuid.UUID
}

type TestLlmConnectionWithConfigRequest struct {
	Provider domainmodels.LlmProvider
	Model    string
	ApiKey   *string // nil/omitted = reuse the currently stored key
	BaseURL  *string
}
