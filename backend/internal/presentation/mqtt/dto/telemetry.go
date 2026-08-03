package presentationmqttdto

import (
	"encoding/json"
	"strings"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
)

type Telemetry struct {
	MetricName           string          `json:"metric_name"`
	PayloadSchemaName    string          `json:"payload_schema_name"`
	PayloadSchemaVersion int32           `json:"payload_schema_version"`
	Payload              json.RawMessage `json:"payload"`
	RecordedAt           string          `json:"recorded_at"`
}

func DecodeTelemetry(deviceId string, payload []byte) (domainusecasesnode.NodeTelemetryMessageRequest, error) {
	var dto Telemetry
	if err := json.Unmarshal(payload, &dto); err != nil {
		return domainusecasesnode.NodeTelemetryMessageRequest{}, domainmodels.NewError("invalid telemetry payload", domainmodels.ErrTypeValidation, err)
	}

	metricName := strings.TrimSpace(dto.MetricName)
	schemaName := strings.TrimSpace(dto.PayloadSchemaName)
	if metricName == "" || schemaName == "" || dto.PayloadSchemaVersion <= 0 || len(dto.Payload) == 0 {
		return domainusecasesnode.NodeTelemetryMessageRequest{}, domainmodels.NewError("invalid telemetry message", domainmodels.ErrTypeValidation, nil)
	}

	// recorded_at is optional - devices without synced clocks simply omit
	// it and the backend timestamps at ingestion instead.
	recordedAt := time.Now().UTC()
	if trimmed := strings.TrimSpace(dto.RecordedAt); trimmed != "" {
		if parsed, err := time.Parse(time.RFC3339, trimmed); err == nil {
			recordedAt = parsed
		}
	}

	return domainusecasesnode.NodeTelemetryMessageRequest{
		DeviceId:             deviceId,
		MetricName:           metricName,
		PayloadSchemaName:    schemaName,
		PayloadSchemaVersion: dto.PayloadSchemaVersion,
		Payload:              dto.Payload,
		RecordedAt:           recordedAt,
	}, nil
}
