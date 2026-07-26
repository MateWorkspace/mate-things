package domainusecasesaction

import (
	"context"
	"encoding/json"
	"time"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type Execution interface {
	Dispatch(ctx context.Context, request DispatchActionRequest) (*domainmodels.ActionLog, error)
}

type DispatchActionRequest struct {
	ActionId   uuid.UUID
	NodeId     uuid.UUID
	Payload    json.RawMessage
	ExecutedAt time.Time
	ActorId    *uuid.UUID
}
