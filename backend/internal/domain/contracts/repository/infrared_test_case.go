package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredTestCase interface {
	CreateWithStates(ctx context.Context, coderId uuid.UUID, step int32, description string, states []domainmodels.InfraredTestCaseState) (id uuid.UUID, err error)
	ReadListByCoderId(ctx context.Context, coderId uuid.UUID) ([]domainmodels.InfraredTestCase, error)
	ReadById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredTestCase, error)
	UpdateStatusById(ctx context.Context, id uuid.UUID, status domainmodels.InfraredTestCaseStatus) error
	DeleteById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
	ReadListStatesByTestCaseId(ctx context.Context, testCaseId uuid.UUID) ([]domainmodels.InfraredTestCaseState, error)
	DeleteStateById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
}
