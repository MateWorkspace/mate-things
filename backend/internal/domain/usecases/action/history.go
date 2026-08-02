package domainusecasesaction

import (
	"context"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type History interface {
	ReadByFilter(ctx context.Context, request ReadActionLogsByFilterRequest) ([]domainmodels.ActionLogListItem, int, error)
	DeleteByFilter(ctx context.Context, request DeleteActionLogsByFilterRequest) (int, error)
}

type ReadActionLogsByFilterRequest struct {
	ExecutedAtStart *time.Time
	ExecutedAtEnd   *time.Time
	ActionId        *uuid.UUID
	NodeId          *uuid.UUID
	ExecutionId     *uuid.UUID
	ActionStatus    *domainmodels.ActionStatus
	Page            int
	Limit           int
}

type DeleteActionLogsByFilterRequest struct {
	ExecutedAtStart *time.Time
	ExecutedAtEnd   *time.Time
	ActionId        *uuid.UUID
	NodeId          *uuid.UUID
	ExecutionId     *uuid.UUID
	ActionStatus    *domainmodels.ActionStatus
}
