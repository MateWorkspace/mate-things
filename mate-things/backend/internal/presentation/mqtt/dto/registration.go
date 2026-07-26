package presentationmqttdto

import (
	"encoding/json"
	"strings"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	domainusecasesnode "github.com/ABA-Developer/nusapala-things/backend/internal/domain/usecases/node"
)

type Registration struct {
	DeviceId     string `json:"device_id"`
	DeviceInfo   string `json:"device_info"`
	FirmwareName string `json:"firmware_name"`
}

func DecodeRegistration(payload []byte) (domainusecasesnode.RegisterNodeMessageRequest, error) {
	var dto Registration
	if err := json.Unmarshal(payload, &dto); err != nil {
		return domainusecasesnode.RegisterNodeMessageRequest{}, domainmodels.NewError("invalid registration payload", domainmodels.ErrTypeValidation, err)
	}

	dto.DeviceId = strings.TrimSpace(dto.DeviceId)
	dto.DeviceInfo = strings.TrimSpace(dto.DeviceInfo)
	dto.FirmwareName = strings.TrimSpace(dto.FirmwareName)
	if dto.DeviceId == "" || dto.DeviceInfo == "" || dto.FirmwareName == "" {
		return domainusecasesnode.RegisterNodeMessageRequest{}, domainmodels.NewError("invalid registration payload", domainmodels.ErrTypeValidation, nil)
	}

	return domainusecasesnode.RegisterNodeMessageRequest{
		DeviceId:     dto.DeviceId,
		DeviceInfo:   dto.DeviceInfo,
		FirmwareName: dto.FirmwareName,
	}, nil
}
