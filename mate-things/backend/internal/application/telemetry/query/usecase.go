package applicationtelemetryquery

import (
	"context"

	domaincontractslogger "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/logger"
	domaincontractsrepository "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	domainusecasestelemetry "github.com/ABA-Developer/nusapala-things/backend/internal/domain/usecases/telemetry"
)

type usecase struct {
	telemetryRecord domaincontractsrepository.TelemetryRecord
	logger          domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	telemetryRecord domaincontractsrepository.TelemetryRecord,
	logger domaincontractslogger.Leveled,
) domainusecasestelemetry.Query {
	return &usecase{
		telemetryRecord: telemetryRecord,
		logger:          logger,
	}
}

func (u *usecase) ReadByFilter(
	ctx context.Context,
	request domainusecasestelemetry.ReadTelemetryByFilterRequest,
) ([]domainmodels.TelemetryRecord, int, error) {
	const tag = "telemetry/query/ReadByFilter"

	telemetryRecords, total, err := u.telemetryRecord.ReadByFilter(
		ctx,
		request.RecordedAtStart,
		request.RecordedAtEnd,
		request.NodeDeviceId,
		request.MetricName,
		request.PayloadSchemaName,
		request.PayloadSchemaVersion,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read telemetry records", domainmodels.LoggerMeta{
			"err":                    err,
			"node_device_id":         request.NodeDeviceId,
			"metric_name":            request.MetricName,
			"payload_schema_name":    request.PayloadSchemaName,
			"payload_schema_version": request.PayloadSchemaVersion,
		})
		return nil, 0, err
	}

	return telemetryRecords, total, nil
}

func (u *usecase) DeleteByFilter(
	ctx context.Context,
	request domainusecasestelemetry.DeleteTelemetryByFilterRequest,
) (int, error) {
	const tag = "telemetry/query/DeleteByFilter"

	total, err := u.telemetryRecord.DeleteByFilter(
		ctx,
		request.RecordedAtStart,
		request.RecordedAtEnd,
		request.NodeDeviceId,
		request.MetricName,
		request.PayloadSchemaName,
		request.PayloadSchemaVersion,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to delete telemetry records", domainmodels.LoggerMeta{
			"err":                    err,
			"node_device_id":         request.NodeDeviceId,
			"metric_name":            request.MetricName,
			"payload_schema_name":    request.PayloadSchemaName,
			"payload_schema_version": request.PayloadSchemaVersion,
		})
		return 0, err
	}

	return total, nil
}
