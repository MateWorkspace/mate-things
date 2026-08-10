package domainusecasesnode

import (
	"context"
	"encoding/json"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type MessagingCallback interface {
	Register(ctx context.Context, request RegisterNodeMessageRequest) error
	Status(ctx context.Context, request NodeStatusMessageRequest) error
	ActionAck(ctx context.Context, request NodeActionAckMessageRequest) error
	Log(ctx context.Context, request NodeLogMessageRequest) error
	Telemetry(ctx context.Context, request NodeTelemetryMessageRequest) error
	Resubscribe(ctx context.Context) error
}

type RegisterNodeMessageRequest struct {
	DeviceId      string
	DeviceInfo    string
	NodeClassName string
	FirmwareName  string
	Config        map[string]string
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

type NodeTelemetryMessageRequest struct {
	DeviceId             string
	MetricName           string
	PayloadSchemaName    string
	PayloadSchemaVersion int32
	Payload              json.RawMessage
	RecordedAt           time.Time
}
