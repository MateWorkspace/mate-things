package presentationmqttdto

import (
	"encoding/json"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesinfrared "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/infrared"
)

type IrCapture struct {
	RawData []int32 `json:"raw_data"`
}

func DecodeIrCapture(deviceId string, payload []byte) (domainusecasesinfrared.CaptureIrRawRequest, error) {
	var dto IrCapture
	if err := json.Unmarshal(payload, &dto); err != nil {
		return domainusecasesinfrared.CaptureIrRawRequest{}, domainmodels.NewError("invalid ir capture payload", domainmodels.ErrTypeValidation, err)
	}
	if len(dto.RawData) == 0 {
		return domainusecasesinfrared.CaptureIrRawRequest{}, domainmodels.NewError("ir capture payload has no raw_data", domainmodels.ErrTypeValidation, nil)
	}
	return domainusecasesinfrared.CaptureIrRawRequest{
		NodeDeviceId: deviceId,
		RawData:      dto.RawData,
	}, nil
}
