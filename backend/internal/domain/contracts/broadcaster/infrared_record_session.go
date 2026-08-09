package domaincontractsbroadcaster

import (
	"context"
	"net/http"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredRecordSession interface {
	Register(ctx context.Context, w http.ResponseWriter, r *http.Request, sessionId uuid.UUID) (err error)
	Send(ctx context.Context, event domainmodels.InfraredRecordSessionEvent) (err error)
}
