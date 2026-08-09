package presentationmqtthandler

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	presentationmqttdto "github.com/MateWorkspace/mate-things/backend/internal/presentation/mqtt/dto"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func (h *Handler) IrCapture(ctx context.Context, msg mqtt.Message, deviceId string) {
	const tag = path + "/IrCapture"

	request, err := presentationmqttdto.DecodeIrCapture(deviceId, msg.Payload())
	if err != nil {
		h.Logger.Warn(ctx, tag, "invalid ir capture payload", domainmodels.LoggerMeta{"err": err, "device_id": deviceId})
		return
	}

	if err := h.InfraredRecordSession.CaptureIrRaw(ctx, request); err != nil {
		h.Logger.Error(ctx, tag, "failed to handle ir capture", domainmodels.LoggerMeta{"err": err, "device_id": deviceId})
		return
	}
}
