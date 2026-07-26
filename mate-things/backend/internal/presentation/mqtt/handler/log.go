package presentationmqtthandler

import (
	"context"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	domainusecasesnode "github.com/ABA-Developer/nusapala-things/backend/internal/domain/usecases/node"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func (h *Handler) Log(ctx context.Context, msg mqtt.Message, deviceId string) {
	const tag = path + "/Log"

	if err := h.MessagingCallback.Log(ctx, domainusecasesnode.NodeLogMessageRequest{
		DeviceId: deviceId,
		Payload:  msg.Payload(),
	}); err != nil {
		h.Logger.Error(ctx, tag, "failed to handle log message", domainmodels.LoggerMeta{
			"err":       err,
			"topic":     msg.Topic(),
			"device_id": deviceId,
		})
		return
	}
}
