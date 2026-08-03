package presentationmqtthandler

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	presentationmqttdto "github.com/MateWorkspace/mate-things/backend/internal/presentation/mqtt/dto"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func (h *Handler) Telemetry(ctx context.Context, msg mqtt.Message, deviceId string) {
	const tag = path + "/Telemetry"

	request, err := presentationmqttdto.DecodeTelemetry(deviceId, msg.Payload())
	if err != nil {
		h.Logger.Warn(ctx, tag, "invalid telemetry message", domainmodels.LoggerMeta{
			"err":       err,
			"topic":     msg.Topic(),
			"device_id": deviceId,
		})
		return
	}

	if err := h.MessagingCallback.Telemetry(ctx, request); err != nil {
		h.Logger.Error(ctx, tag, "failed to handle telemetry message", domainmodels.LoggerMeta{
			"err":       err,
			"topic":     msg.Topic(),
			"device_id": deviceId,
		})
		return
	}
}
