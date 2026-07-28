package presentationhttpresponse

import (
	"encoding/json"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type FirmwareResponse struct {
	Id          string          `json:"id"`
	NodeClassId string          `json:"node_class_id"`
	Name        string          `json:"name"`
	Size        int32           `json:"size"`
	Checksum    string          `json:"checksum"`
	BinaryPath  string          `json:"binary_path"`
	Preferences json.RawMessage `json:"preferences"`
	AuditResponse
}

type FirmwareBinaryStatResponse struct {
	BinaryPath string `json:"binary_path"`
	Size       int32  `json:"size"`
	Checksum   string `json:"checksum"`
}

type FirmwareCreateResponse struct {
	Id         string `json:"id"`
	Size       int32  `json:"size"`
	Checksum   string `json:"checksum"`
	BinaryPath string `json:"binary_path"`
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
