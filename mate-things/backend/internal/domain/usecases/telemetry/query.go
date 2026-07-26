package domainusecasestelemetry

import (
	"context"
	"time"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
)

type Query interface {
	ReadByFilter(ctx context.Context, request ReadTelemetryByFilterRequest) ([]domainmodels.TelemetryRecord, int, error)
	DeleteByFilter(ctx context.Context, request DeleteTelemetryByFilterRequest) (int, error)
}

type ReadTelemetryByFilterRequest struct {
	RecordedAtStart      *time.Time
	RecordedAtEnd        *time.Time
	NodeDeviceId         *string
	MetricName           *string
	PayloadSchemaName    *string
	PayloadSchemaVersion *int32
}

type DeleteTelemetryByFilterRequest struct {
	RecordedAtStart      *time.Time
	RecordedAtEnd        *time.Time
	NodeDeviceId         *string
	MetricName           *string
	PayloadSchemaName    *string
	PayloadSchemaVersion *int32
}
