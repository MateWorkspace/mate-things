package presentationhttpresponse

import (
	"encoding/json"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type TelemetryRecordResponse struct {
	Id                   int64           `json:"id"`
	NodeDeviceId         string          `json:"node_device_id"`
	MetricName           string          `json:"metric_name"`
	PayloadSchemaName    string          `json:"payload_schema_name"`
	PayloadSchemaVersion int32           `json:"payload_schema_version"`
	Payload              json.RawMessage `json:"payload"`
	RecordedAt           time.Time       `json:"recorded_at"`
	CreatedAt            time.Time       `json:"created_at"`
}

func TelemetryRecord(telemetryRecord domainmodels.TelemetryRecord) TelemetryRecordResponse {
	return TelemetryRecordResponse{
		Id:                   telemetryRecord.Id,
		NodeDeviceId:         telemetryRecord.NodeDeviceId,
		MetricName:           telemetryRecord.MetricName,
		PayloadSchemaName:    telemetryRecord.PayloadSchemaName,
		PayloadSchemaVersion: telemetryRecord.PayloadSchemaVersion,
		Payload:              NormalizeJSON(telemetryRecord.Payload),
		RecordedAt:           telemetryRecord.RecordedAt,
		CreatedAt:            telemetryRecord.CreatedAt,
	}
}

func TelemetryRecords(telemetryRecords []domainmodels.TelemetryRecord) []TelemetryRecordResponse {
	result := make([]TelemetryRecordResponse, 0, len(telemetryRecords))
	for _, telemetryRecord := range telemetryRecords {
		result = append(result, TelemetryRecord(telemetryRecord))
	}
	return result
}
