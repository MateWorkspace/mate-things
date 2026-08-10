package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredTestCase interface {
	CreateWithStates(ctx context.Context, coderId uuid.UUID, step int32, description string, states []domainmodels.InfraredTestCaseState, createdBy *uuid.UUID) (id uuid.UUID, err error)
	ListByCoderId(ctx context.Context, coderId uuid.UUID) ([]domainmodels.InfraredTestCase, error)
	GetById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredTestCase, error)
	UpdateStatusById(ctx context.Context, id uuid.UUID, status domainmodels.InfraredTestCaseStatus, updatedBy *uuid.UUID) error
	DeleteById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
	ListStatesByTestCaseId(ctx context.Context, testCaseId uuid.UUID) ([]domainmodels.InfraredTestCaseState, error)
	DeleteStateById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
}
