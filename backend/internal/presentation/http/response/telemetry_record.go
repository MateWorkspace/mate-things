package presentationhttpresponse

import (
	"encoding/json"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type TelemetryRecordResponse struct {
	Id                   int64           `json:"id" example:"918273"`
	NodeDeviceId         string          `json:"node_device_id" example:"ESP32-BARISTA-07"`
	MetricName           string          `json:"metric_name" example:"water_temperature_celsius"`
	PayloadSchemaName    string          `json:"payload_schema_name" example:"brew_command"`
	PayloadSchemaVersion int32           `json:"payload_schema_version" example:"1"`
	Payload              json.RawMessage `json:"payload" swaggertype:"object"`
	RecordedAt           time.Time       `json:"recorded_at" example:"2026-07-28T08:14:58Z"`
	CreatedAt            time.Time       `json:"created_at" example:"2026-07-28T08:15:00Z"`
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
