package presentationhttpresponse

import (
	"encoding/json"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type FirmwareResponse struct {
	Id          string          `json:"id" example:"9c4e2b7a-1f3d-4a6c-8b5e-2d7f9a1c3e08"`
	NodeClassId string          `json:"node_class_id" example:"3f1c9a2e-6d4b-4e7a-8c2f-1a9b3d5e7f01"`
	Name        string          `json:"name" example:"espresso-fw"`
	Size        int32           `json:"size" example:"482112"`
	Checksum    string          `json:"checksum" example:"3b1e7c9a2f5d8b4e6c1a9f3d7e2b5c8a1d4e6f9b2c5a8d1e4f7b0c3a6d9e2f5"`
	BinaryPath  string          `json:"binary_path" example:"firmwares/espresso-fw/v3/firmware.bin"`
	Preferences json.RawMessage `json:"preferences" swaggertype:"object"`
	AuditResponse
}

type FirmwareBinaryStatResponse struct {
	BinaryPath string `json:"binary_path" example:"firmwares/espresso-fw/v3/firmware.bin"`
	Size       int32  `json:"size" example:"482112"`
	Checksum   string `json:"checksum" example:"3b1e7c9a2f5d8b4e6c1a9f3d7e2b5c8a1d4e6f9b2c5a8d1e4f7b0c3a6d9e2f5"`
}

type FirmwareCreateResponse struct {
	Id         string `json:"id" example:"9c4e2b7a-1f3d-4a6c-8b5e-2d7f9a1c3e08"`
	Size       int32  `json:"size" example:"482112"`
	Checksum   string `json:"checksum" example:"3b1e7c9a2f5d8b4e6c1a9f3d7e2b5c8a1d4e6f9b2c5a8d1e4f7b0c3a6d9e2f5"`
	BinaryPath string `json:"binary_path" example:"firmwares/espresso-fw/v3/firmware.bin"`
}

func Firmware(firmware domainmodels.Firmware) FirmwareResponse {
	return FirmwareResponse{
		Id:          UUIDString(firmware.Id),
		NodeClassId: UUIDString(firmware.NodeClassId),
		Name:        firmware.Name,
		Size:        firmware.Size,
		Checksum:    firmware.Checksum,
		BinaryPath:  firmware.BinaryPath,
		Preferences: NormalizeJSON(firmware.Preferences),
		AuditResponse: Audit(
			firmware.CreatedAt,
			firmware.UpdatedAt,
			firmware.DeletedAt,
			firmware.CreatedBy,
			firmware.UpdatedBy,
			firmware.DeletedBy,
		),
	}
}

func Firmwares(firmwares []domainmodels.Firmware) []FirmwareResponse {
	result := make([]FirmwareResponse, 0, len(firmwares))
	for _, firmware := range firmwares {
		result = append(result, Firmware(firmware))
	}
	return result
}
