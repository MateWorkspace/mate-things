package domaincontractsbroadcaster

import (
	"context"
	"net/http"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type Telemetry interface {
	Register(
		ctx context.Context,
		w http.ResponseWriter,
		r *http.Request,
		userId uuid.UUID,
		nodeDeviceId *string,
		metricName *string,
	) (err error)

	Send(
		ctx context.Context,
		record domainmodels.TelemetryRecord,
	) (err error)

	SessionList(
		ctx context.Context,
	) (sessions []domainmodels.BroadcastSessionTelemetry, err error)
}
