package presentationmqttdto

import (
	"encoding/json"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	domainusecasesnode "github.com/ABA-Developer/nusapala-things/backend/internal/domain/usecases/node"
)

type Status struct {
	IsConnected *bool `json:"is_connected"`
}

func DecodeStatus(deviceId string, payload []byte) (domainusecasesnode.NodeStatusMessageRequest, error) {
	var dto Status
	if err := json.Unmarshal(payload, &dto); err != nil {
		return domainusecasesnode.NodeStatusMessageRequest{}, domainmodels.NewError("invalid status payload", domainmodels.ErrTypeValidation, err)
	}
	if dto.IsConnected == nil {
		return domainusecasesnode.NodeStatusMessageRequest{}, domainmodels.NewError("invalid status payload", domainmodels.ErrTypeValidation, nil)
	}

	return domainusecasesnode.NodeStatusMessageRequest{
		DeviceId:    deviceId,
		IsConnected: *dto.IsConnected,
	}, nil
}
