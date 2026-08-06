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
		success bool,
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
		actionName string,
		payload json.RawMessage,
	) (err error)

	Config(
		ctx context.Context,
		nodeDeviceId string,
		key string,
		value string,
	) (err error)
}
