package presentationhttpresponse

import (
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type BroadcastSessionTelemetryResponse struct {
	Id           uuid.UUID `json:"id"`
	UserId       uuid.UUID `json:"user_id"`
	NodeDeviceId *string   `json:"node_device_id" example:"ESP32-BARISTA-07"`
	MetricName   *string   `json:"metric_name" example:"water_temperature_celsius"`
	RemoteAddr   string    `json:"remote_addr" example:"172.18.0.1:53214"`
	ConnectedAt  time.Time `json:"connected_at" example:"2026-07-28T08:14:58Z"`
}

func BroadcastSessionTelemetry(session domainmodels.BroadcastSessionTelemetry) BroadcastSessionTelemetryResponse {
	return BroadcastSessionTelemetryResponse{
		Id:           session.Id,
		UserId:       session.UserId,
		NodeDeviceId: session.NodeDeviceId,
		MetricName:   session.MetricName,
		RemoteAddr:   session.RemoteAddr,
		ConnectedAt:  session.ConnectedAt,
	}
}

func BroadcastSessionTelemetries(sessions []domainmodels.BroadcastSessionTelemetry) []BroadcastSessionTelemetryResponse {
	result := make([]BroadcastSessionTelemetryResponse, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, BroadcastSessionTelemetry(session))
	}
	return result
}
