package domaincontractsrepository

import (
	"context"
	"encoding/json"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type ActionLog interface {
	Create(
		ctx context.Context,
		executionId uuid.UUID,
		actionId uuid.UUID,
		nodeId *uuid.UUID,
		actionStatus domainmodels.ActionStatus,
		actionMessage *string,
		payload json.RawMessage,
		executedAt time.Time,
	) (id int64, err error)

	UpdateStatusByExecutionId(
		ctx context.Context,
		executionId uuid.UUID,
		actionStatus domainmodels.ActionStatus,
		actionMessage *string,
	) (err error)

	ReadByFilter(
		ctx context.Context,
		executedAtStart *time.Time,
		executedAtEnd *time.Time,
		actionId *uuid.UUID,
		nodeId *uuid.UUID,
	) (actionLogs []domainmodels.ActionLog, total int, err error)

	DeleteByFilter(
		ctx context.Context,
		executedAtStart *time.Time,
		executedAtEnd *time.Time,
		actionId *uuid.UUID,
		nodeId *uuid.UUID,
	) (total int, err error)
}
