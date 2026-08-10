package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredRecordSession interface {
	Create(ctx context.Context, nodeId uuid.UUID, infraredDeviceId uuid.UUID, createdBy *uuid.UUID) (id uuid.UUID, err error)
	ReadById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredRecordSession, error)
	// ReadActiveByNodeId returns the session currently RECORDING against the
	// given node, or (nil, nil) if none is active — there is deliberately
	// no error for the common "nothing recording right now" case.
	ReadActiveByNodeId(ctx context.Context, nodeId uuid.UUID) (*domainmodels.InfraredRecordSession, error)
	UpdateRecordingStateById(ctx context.Context, id uuid.UUID, recordingState string, isCompleted bool) error
	UpdateCurrentRecordCaseIdById(ctx context.Context, id uuid.UUID, currentRecordCaseId *uuid.UUID) error
	DeleteById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
}
