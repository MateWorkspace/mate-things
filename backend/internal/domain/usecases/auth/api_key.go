package domainusecasesauth

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type ApiKey interface {
	Authenticate(ctx context.Context, rawKey string) (*domainmodels.TokenClaimsAccess, error)
}
