package applicationtelemetrybroadcast

import (
	"context"
	"net/http"

	domaincontractsbroadcaster "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/broadcaster"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasestelemetry "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/telemetry"
	"github.com/google/uuid"
)

type usecase struct {
	broadcaster     domaincontractsbroadcaster.Telemetry
	telemetryRecord domaincontractsrepository.TelemetryRecord
	logger          domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	broadcaster domaincontractsbroadcaster.Telemetry,
	telemetryRecord domaincontractsrepository.TelemetryRecord,
	logger domaincontractslogger.Leveled,
) domainusecasestelemetry.Broadcast {
	return &usecase{
		broadcaster:     broadcaster,
		telemetryRecord: telemetryRecord,
		logger:          logger,
	}
}

func (u *usecase) Register(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	userId uuid.UUID,
	nodeDeviceId *string,
	metricName *string,
) error {
	const tag = "telemetry/broadcast/Register"

	initial, err := u.telemetryRecord.ReadLatest(ctx, nodeDeviceId, metricName)
	if err != nil {
		u.logger.Warn(ctx, tag, "failed to read latest telemetry record for new session", domainmodels.LoggerMeta{
			"err":            err,
			"user_id":        userId,
			"node_device_id": nodeDeviceId,
			"metric_name":    metricName,
		})
		initial = nil
	}

	if err := u.broadcaster.Register(ctx, w, r, userId, nodeDeviceId, metricName, initial); err != nil {
		u.logger.Warn(ctx, tag, "failed to register broadcast session", domainmodels.LoggerMeta{
			"err":            err,
			"user_id":        userId,
			"node_device_id": nodeDeviceId,
			"metric_name":    metricName,
		})
		return err
	}

	return nil
}

func (u *usecase) SessionList(ctx context.Context) ([]domainmodels.BroadcastSessionTelemetry, error) {
	const tag = "telemetry/broadcast/SessionList"

	sessions, err := u.broadcaster.SessionList(ctx)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list broadcast sessions", domainmodels.LoggerMeta{
			"err": err,
		})
		return nil, err
	}

	return sessions, nil
}
