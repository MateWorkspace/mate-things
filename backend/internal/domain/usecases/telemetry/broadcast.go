package domainusecasestelemetry

import (
	"context"
	"net/http"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type Broadcast interface {
	Register(
		ctx context.Context,
		w http.ResponseWriter,
		r *http.Request,
		userId uuid.UUID,
		nodeDeviceId *string,
		metricName *string,
	) (err error)

	SessionList(ctx context.Context) (sessions []domainmodels.BroadcastSessionTelemetry, err error)
}
