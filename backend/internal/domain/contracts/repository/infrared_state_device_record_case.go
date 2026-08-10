package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredStateDeviceRecordCase interface {
	CreateWithStates(ctx context.Context, sessionId uuid.UUID, step int32, description string, states []domainmodels.InfraredStateDeviceRecordState, createdBy *uuid.UUID) (caseId uuid.UUID, err error)
	ListBySessionId(ctx context.Context, sessionId uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordCase, error)
	GetById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredStateDeviceRecordCase, error)
	UpdateStatusById(ctx context.Context, id uuid.UUID, status domainmodels.InfraredRecordCaseStatus, updatedBy *uuid.UUID) error
	DeleteById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
	ListStatesByCaseId(ctx context.Context, caseId uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordState, error)
	DeleteStateById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
	CreateRaw(ctx context.Context, caseId uuid.UUID, rawData []byte, createdBy *uuid.UUID) (rawId uuid.UUID, err error)
	ListRawByCaseId(ctx context.Context, caseId uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordRaw, error)
	UpdateRawStatusById(ctx context.Context, id uuid.UUID, status domainmodels.InfraredRecordRawStatus, discardedReason *string, updatedBy *uuid.UUID) error
	DeleteRawById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
	CountAcceptedRawByCaseId(ctx context.Context, caseId uuid.UUID) (int, error)
	GetRawById(ctx context.Context, rawId uuid.UUID) (*domainmodels.InfraredStateDeviceRecordRaw, error)
}
