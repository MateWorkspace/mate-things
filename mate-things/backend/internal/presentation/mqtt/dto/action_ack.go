package presentationmqttdto

import (
	"encoding/json"
	"strings"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	domainusecasesnode "github.com/ABA-Developer/nusapala-things/backend/internal/domain/usecases/node"
	"github.com/google/uuid"
)

type ActionAck struct {
	ExecutionId string `json:"execution_id"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

func DecodeActionAck(deviceId string, payload []byte) (domainusecasesnode.NodeActionAckMessageRequest, error) {
	var dto ActionAck
	if err := json.Unmarshal(payload, &dto); err != nil {
		return domainusecasesnode.NodeActionAckMessageRequest{}, domainmodels.NewError("invalid action ack payload", domainmodels.ErrTypeValidation, err)
	}

	executionId, err := uuid.Parse(strings.TrimSpace(dto.ExecutionId))
	if err != nil {
		return domainusecasesnode.NodeActionAckMessageRequest{}, domainmodels.NewError("invalid action ack execution id", domainmodels.ErrTypeValidation, err)
	}

	status := domainmodels.ActionStatus(strings.ToUpper(strings.TrimSpace(dto.Status)))
	switch status {
	case domainmodels.ActionStatusSuccess, domainmodels.ActionStatusFailed:
	default:
		return domainusecasesnode.NodeActionAckMessageRequest{}, domainmodels.NewError("invalid action ack status", domainmodels.ErrTypeValidation, nil)
	}

	message := strings.TrimSpace(dto.Message)
	var messagePtr *string
	if message != "" {
		messagePtr = &message
	}

	return domainusecasesnode.NodeActionAckMessageRequest{
		DeviceId:      deviceId,
		ExecutionId:   executionId,
		ActionStatus:  status,
		ActionMessage: messagePtr,
	}, nil
}
