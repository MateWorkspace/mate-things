package presentationmqttdto

import (
	"encoding/json"
	"strings"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
)

type Status struct {
	Status *domainmodels.NodeStatus `json:"status"`
}

func DecodeStatus(deviceId string, payload []byte) (domainusecasesnode.NodeStatusMessageRequest, error) {
	var dto Status
	if err := json.Unmarshal(payload, &dto); err != nil {
		return domainusecasesnode.NodeStatusMessageRequest{}, domainmodels.NewError("invalid status payload", domainmodels.ErrTypeValidation, err)
	}
	if dto.Status == nil {
		return domainusecasesnode.NodeStatusMessageRequest{}, domainmodels.NewError("invalid status payload", domainmodels.ErrTypeValidation, nil)
	}

	status := domainmodels.NodeStatus(strings.ToUpper(strings.TrimSpace(string(*dto.Status))))
	if status != domainmodels.NodeStatusOnline && status != domainmodels.NodeStatusOffline {
		return domainusecasesnode.NodeStatusMessageRequest{}, domainmodels.NewError("invalid status payload", domainmodels.ErrTypeValidation, nil)
	}

	return domainusecasesnode.NodeStatusMessageRequest{
		DeviceId:    deviceId,
		IsConnected: status == domainmodels.NodeStatusOnline,
	}, nil
}
