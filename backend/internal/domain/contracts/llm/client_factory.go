package domaincontractsllm

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type ClientFactory interface {
	Current(ctx context.Context) (Client, error)
	FromCredentials(provider domainmodels.LlmProvider, apiKey string, baseURL *string, model string) (Client, error)
}
