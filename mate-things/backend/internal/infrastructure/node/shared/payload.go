package infrastructurenodeshared

import (
	"encoding/json"

	"github.com/google/uuid"
)

type OtaPayload struct {
	FirmwareUrl      string `json:"firmware_url"`
	FirmwareSize     int32  `json:"firmware_size"`
	FirmwareChecksum string `json:"firmware_checksum"`
}

type ActionPayload struct {
	ExecutionId uuid.UUID       `json:"execution_id"`
	Action      string          `json:"action"`
	Payload     json.RawMessage `json:"payload"`
}
