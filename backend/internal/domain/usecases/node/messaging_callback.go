package domainusecasesnode

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type MessagingCallback interface {
	Register(ctx context.Context, request RegisterNodeMessageRequest) error
	Status(ctx context.Context, request NodeStatusMessageRequest) error
	ActionAck(ctx context.Context, request NodeActionAckMessageRequest) error
	Log(ctx context.Context, request NodeLogMessageRequest) error
	Resubscribe(ctx context.Context) error
}

type RegisterNodeMessageRequest struct {
	DeviceId     string
	DeviceInfo   string
	FirmwareName string
	Config       map[string]string
}

type NodeStatusMessageRequest struct {
	DeviceId    string
	IsConnected bool
}

type NodeActionAckMessageRequest struct {
	DeviceId      string
	ExecutionId   uuid.UUID
	ActionStatus  domainmodels.ActionStatus
	ActionMessage *string
}

type NodeLogMessageRequest struct {
	DeviceId string
	Payload  []byte
}
