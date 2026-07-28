package domaincontractsrepository

import (
	"context"
	"encoding/json"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type TelemetryRecord interface {
	Create(
		ctx context.Context,
		nodeDeviceId string,
		metricName string,
		payloadSchemaName string,
		payloadSchemaVersion int32,
		payload json.RawMessage,
		recordedAt time.Time,
	) (id int64, err error)

	ReadByFilter(
		ctx context.Context,
		recordedAtStart *time.Time,
		recordedAtEnd *time.Time,
		nodeDeviceId *string,
		metricName *string,
		payloadSchemaName *string,
		payloadSchemaVersion *int32,
	) (telemetryRecords []domainmodels.TelemetryRecord, total int, err error)

	DeleteByFilter(
		ctx context.Context,
		recordedAtStart *time.Time,
		recordedAtEnd *time.Time,
		nodeDeviceId *string,
		metricName *string,
		payloadSchemaName *string,
		payloadSchemaVersion *int32,
	) (total int, err error)
}
