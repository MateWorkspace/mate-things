package presentationmqtthandler

import (
	"context"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	presentationmqttdto "github.com/ABA-Developer/nusapala-things/backend/internal/presentation/mqtt/dto"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func (h *Handler) Status(ctx context.Context, msg mqtt.Message, deviceId string) {
	const tag = path + "/Status"

	request, err := presentationmqttdto.DecodeStatus(deviceId, msg.Payload())
	if err != nil {
		h.Logger.Warn(ctx, tag, "invalid status message", domainmodels.LoggerMeta{
			"err":       err,
			"topic":     msg.Topic(),
			"device_id": deviceId,
		})
		return
	}

	if err := h.MessagingCallback.Status(ctx, request); err != nil {
		h.Logger.Error(ctx, tag, "failed to handle status message", domainmodels.LoggerMeta{
			"err":       err,
			"topic":     msg.Topic(),
			"device_id": deviceId,
		})
		return
	}
}
