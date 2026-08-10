package presentationmqtthandler

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	presentationmqttdto "github.com/MateWorkspace/mate-things/backend/internal/presentation/mqtt/dto"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func (h *Handler) IrTransmitAck(ctx context.Context, msg mqtt.Message, deviceId string) {
	const tag = path + "/IrTransmitAck"

	event, err := presentationmqttdto.DecodeIrTransmitAck(deviceId, msg.Payload())
	if err != nil {
		h.Logger.Warn(ctx, tag, "invalid ir transmit ack payload", domainmodels.LoggerMeta{"err": err, "device_id": deviceId})
		return
	}

	if event.Status != "SUCCESS" {
		h.Logger.Warn(ctx, tag, "node reported ir transmit failure", domainmodels.LoggerMeta{
			"device_id": deviceId, "execution_id": event.ExecutionId, "message": event.Message,
		})
		return
	}
	h.Logger.Debug(ctx, tag, "ir transmit acknowledged", domainmodels.LoggerMeta{"device_id": deviceId, "execution_id": event.ExecutionId})
}
