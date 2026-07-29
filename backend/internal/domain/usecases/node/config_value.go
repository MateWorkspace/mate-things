package domainusecasesnode

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type ConfigValue interface {
	ReadByNodeId(ctx context.Context, request ReadConfigValuesByNodeIdRequest) ([]domainmodels.NodeConfigValue, error)
	SetByNodeId(ctx context.Context, request SetConfigValueRequest) error
}

type ReadConfigValuesByNodeIdRequest struct {
	NodeId uuid.UUID
}

type SetConfigValueRequest struct {
	NodeId  uuid.UUID
	Key     string
	Value   string
	ActorId *uuid.UUID
}
