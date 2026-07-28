package domainmodels

import (
	"encoding/json"
	"time"
)

type TelemetryRecord struct {
	Id                   int64           `db:"id" json:"id"`
	NodeDeviceId         string          `db:"node_device_id" json:"node_device_id"`
	MetricName           string          `db:"metric_name" json:"metric_name"`
	PayloadSchemaName    string          `db:"payload_schema_name" json:"payload_schema_name"`
	PayloadSchemaVersion int32           `db:"payload_schema_version" json:"payload_schema_version"`
	Payload              json.RawMessage `db:"payload" json:"payload"`
	RecordedAt           time.Time       `db:"recorded_at" json:"recorded_at"`
	CreatedAt            time.Time       `db:"created_at" json:"created_at"`
}
