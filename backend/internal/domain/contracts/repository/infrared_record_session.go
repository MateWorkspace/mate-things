package domaincontractsrepository

import (
	"context"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredRecordSession interface {
	Create(ctx context.Context, nodeId uuid.UUID, infraredDeviceId uuid.UUID, createdBy *uuid.UUID) (id uuid.UUID, err error)
	ReadById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredRecordSession, error)
	ReadActiveByNodeId(ctx context.Context, nodeId uuid.UUID) (*domainmodels.InfraredRecordSession, error)
	// ReadByFilter lists sessions joined against their device and device
	// type for display purposes (brand/model/device-type name), filtered
	// and paginated — the list-page counterpart to ReadById's single-item
	// lookup.
	ReadByFilter(
		ctx context.Context,
		recordingState *string,
		infraredDeviceTypeId *uuid.UUID,
		createdAtStart *time.Time,
		createdAtEnd *time.Time,
		page int,
		limit int,
	) (items []domainmodels.InfraredRecordSessionListItem, total int, err error)
	UpdateRecordingStateById(ctx context.Context, id uuid.UUID, recordingState string, isCompleted bool) error
	UpdateCurrentRecordCaseIdById(ctx context.Context, id uuid.UUID, currentRecordCaseId *uuid.UUID) error
	MarkChecksumClarificationUsedById(ctx context.Context, id uuid.UUID) error
	DeleteById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
}
