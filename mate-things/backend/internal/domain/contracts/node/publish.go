package domaincontractsnode

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

type Publish interface {
	RegistrationAck(
		ctx context.Context,
		nodeDeviceId string,
	) (err error)

	Ota(
		ctx context.Context,
		nodeDeviceId string,
		firmwareUrl string,
		firmwareSize int32,
		firmwareChecksum string,
	) (err error)

	Action(
		ctx context.Context,
		nodeDeviceId string,
		executionId uuid.UUID,
		payload json.RawMessage,
	) (err error)
}
