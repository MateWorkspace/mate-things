package domainmodels

import (
	"time"

	"github.com/google/uuid"
)

type BroadcastSessionTelemetry struct {
	Id           uuid.UUID `json:"id"`
	UserId       uuid.UUID `json:"user_id"`
	NodeDeviceId *string   `json:"node_device_id"`
	MetricName   *string   `json:"metric_name"`
	RemoteAddr   string    `json:"remote_addr"`
	ConnectedAt  time.Time `json:"connected_at"`
}
