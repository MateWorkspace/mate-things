package presentationmqtthandler

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	presentationmqttdto "github.com/MateWorkspace/mate-things/backend/internal/presentation/mqtt/dto"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func (h *Handler) ActionAck(ctx context.Context, msg mqtt.Message, deviceId string) {
	const tag = path + "/ActionAck"

	request, err := presentationmqttdto.DecodeActionAck(deviceId, msg.Payload())
	if err != nil {
		h.Logger.Warn(ctx, tag, "invalid action ack message", domainmodels.LoggerMeta{
			"err":       err,
			"topic":     msg.Topic(),
			"device_id": deviceId,
		})
		return
	}

	if err := h.MessagingCallback.ActionAck(ctx, request); err != nil {
		h.Logger.Error(ctx, tag, "failed to handle action ack message", domainmodels.LoggerMeta{
			"err":          err,
			"topic":        msg.Topic(),
			"device_id":    deviceId,
			"execution_id": request.ExecutionId,
		})
		return
	}
}
