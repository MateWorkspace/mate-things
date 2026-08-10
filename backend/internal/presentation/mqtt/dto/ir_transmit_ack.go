package presentationmqttdto

import (
	"encoding/json"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type IrTransmitAck struct {
	ExecutionId uuid.UUID `json:"execution_id"`
	Status      string    `json:"status"`
	Message     string    `json:"message,omitempty"`
}

type IrTransmitAckEvent struct {
	DeviceId    string
	ExecutionId uuid.UUID
	Status      string
	Message     string
}

func DecodeIrTransmitAck(deviceId string, payload []byte) (IrTransmitAckEvent, error) {
	var dto IrTransmitAck
	if err := json.Unmarshal(payload, &dto); err != nil {
		return IrTransmitAckEvent{}, domainmodels.NewError("invalid ir transmit ack payload", domainmodels.ErrTypeValidation, err)
	}
	return IrTransmitAckEvent{
		DeviceId:    deviceId,
		ExecutionId: dto.ExecutionId,
		Status:      dto.Status,
		Message:     dto.Message,
	}, nil
}
