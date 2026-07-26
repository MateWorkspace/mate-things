package applicationtelemetryingestion

import (
	"context"

	domaincontractslogger "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/logger"
	domaincontractsrepository "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/repository"
	domaincontractsutility "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	domainusecasesrepocache "github.com/ABA-Developer/nusapala-things/backend/internal/domain/usecases/repocache"
	domainusecasestelemetry "github.com/ABA-Developer/nusapala-things/backend/internal/domain/usecases/telemetry"
)

type usecase struct {
	telemetryRecord domaincontractsrepository.TelemetryRecord
	payloadSchema   domainusecasesrepocache.PayloadSchema
	validator       domaincontractsutility.PayloadSchemaValidator
	logger          domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	telemetryRecord domaincontractsrepository.TelemetryRecord,
	payloadSchema domainusecasesrepocache.PayloadSchema,
	validator domaincontractsutility.PayloadSchemaValidator,
	logger domaincontractslogger.Leveled,
) domainusecasestelemetry.Ingestion {
	return &usecase{
		telemetryRecord: telemetryRecord,
		payloadSchema:   payloadSchema,
		validator:       validator,
		logger:          logger,
	}
}

func (u *usecase) Record(ctx context.Context, request domainusecasestelemetry.RecordTelemetryRequest) (int64, error) {
	const tag = "telemetry/ingestion/Record"

	payloadSchema, err := u.payloadSchema.ReadByNameAndVersion(
		ctx,
		request.PayloadSchemaName,
		request.PayloadSchemaVersion,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read payload schema", domainmodels.LoggerMeta{
			"err":                    err,
			"node_device_id":         request.NodeDeviceId,
			"metric_name":            request.MetricName,
			"payload_schema_name":    request.PayloadSchemaName,
			"payload_schema_version": request.PayloadSchemaVersion,
		})
		return 0, err
	}

	if err := u.validator.Validate(ctx, payloadSchema.Definition, request.Payload); err != nil {
		u.logger.Error(ctx, tag, "failed to validate telemetry payload", domainmodels.LoggerMeta{
			"err":                    err,
			"node_device_id":         request.NodeDeviceId,
			"metric_name":            request.MetricName,
			"payload_schema_name":    request.PayloadSchemaName,
			"payload_schema_version": request.PayloadSchemaVersion,
		})
		return 0, err
	}

	id, err := u.telemetryRecord.Create(
		ctx,
		request.NodeDeviceId,
		request.MetricName,
		request.PayloadSchemaName,
		request.PayloadSchemaVersion,
		request.Payload,
		request.RecordedAt,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to create telemetry record", domainmodels.LoggerMeta{
			"err":                    err,
			"node_device_id":         request.NodeDeviceId,
			"metric_name":            request.MetricName,
			"payload_schema_name":    request.PayloadSchemaName,
			"payload_schema_version": request.PayloadSchemaVersion,
			"recorded_at":            request.RecordedAt,
		})
		return 0, err
	}

	return id, nil
}
