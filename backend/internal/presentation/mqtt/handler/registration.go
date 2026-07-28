package presentationmqtthandler

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	presentationmqttdto "github.com/MateWorkspace/mate-things/backend/internal/presentation/mqtt/dto"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func (h *Handler) Registration(ctx context.Context, msg mqtt.Message) {
	const tag = path + "/Registration"

	request, err := presentationmqttdto.DecodeRegistration(msg.Payload())
	if err != nil {
		h.Logger.Warn(ctx, tag, "invalid registration message", domainmodels.LoggerMeta{
			"err":   err,
			"topic": msg.Topic(),
		})
		return
	}

	if err := h.MessagingCallback.Register(ctx, request); err != nil {
		h.Logger.Error(ctx, tag, "failed to handle registration message", domainmodels.LoggerMeta{
			"err":       err,
			"topic":     msg.Topic(),
			"device_id": request.DeviceId,
		})
		return
	}
}
