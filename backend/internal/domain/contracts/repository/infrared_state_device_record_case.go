package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredStateDeviceRecordCase interface {
	CreateWithStates(ctx context.Context, sessionId uuid.UUID, step int32, description string, states []domainmodels.InfraredStateDeviceRecordState) (caseId uuid.UUID, err error)
	ListBySessionId(ctx context.Context, sessionId uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordCase, error)
	GetById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredStateDeviceRecordCase, error)
	UpdateStatusById(ctx context.Context, id uuid.UUID, status domainmodels.InfraredRecordCaseStatus) error
	ListStatesByCaseId(ctx context.Context, caseId uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordState, error)
	CreateRaw(ctx context.Context, caseId uuid.UUID, rawData []byte) (rawId uuid.UUID, err error)
	ListRawByCaseId(ctx context.Context, caseId uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordRaw, error)
	UpdateRawStatusById(ctx context.Context, id uuid.UUID, status domainmodels.InfraredRecordRawStatus, discardedReason *string) error
	CountAcceptedRawByCaseId(ctx context.Context, caseId uuid.UUID) (int, error)
}
