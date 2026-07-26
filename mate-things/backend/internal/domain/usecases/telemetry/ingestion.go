package domainusecasestelemetry

import (
	"context"
	"encoding/json"
	"time"
)

type Ingestion interface {
	Record(ctx context.Context, request RecordTelemetryRequest) (int64, error)
}

type RecordTelemetryRequest struct {
	NodeDeviceId         string
	MetricName           string
	PayloadSchemaName    string
	PayloadSchemaVersion int32
	Payload              json.RawMessage
	RecordedAt           time.Time
}
