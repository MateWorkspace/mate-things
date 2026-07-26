package domainusecasesaction

import (
	"context"
	"time"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type History interface {
	ReadByFilter(ctx context.Context, request ReadActionLogsByFilterRequest) ([]domainmodels.ActionLog, int, error)
	DeleteByFilter(ctx context.Context, request DeleteActionLogsByFilterRequest) (int, error)
}

type ReadActionLogsByFilterRequest struct {
	ExecutedAtStart *time.Time
	ExecutedAtEnd   *time.Time
	ActionId        *uuid.UUID
	NodeId          *uuid.UUID
}

type DeleteActionLogsByFilterRequest struct {
	ExecutedAtStart *time.Time
	ExecutedAtEnd   *time.Time
	ActionId        *uuid.UUID
	NodeId          *uuid.UUID
}
