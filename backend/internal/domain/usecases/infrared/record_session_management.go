package domainusecasesinfrared

import (
	"context"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type RecordSessionManagement interface {
	Start(ctx context.Context, request StartRecordSessionRequest) (sessionId uuid.UUID, err error)
	GetById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredRecordSession, error)
	ListByFilter(ctx context.Context, request ListRecordSessionsRequest) (items []domainmodels.InfraredRecordSessionListItem, total int, err error)
	ListCases(ctx context.Context, sessionId uuid.UUID) ([]CaseWithStatesAndRaw, error)
	AcceptRaw(ctx context.Context, rawId uuid.UUID) error
	DiscardRaw(ctx context.Context, rawId uuid.UUID, reason string) error
	RetryCase(ctx context.Context, caseId uuid.UUID) error
	SetCurrentCase(ctx context.Context, sessionId uuid.UUID, caseId uuid.UUID) error
	CaptureIrRaw(ctx context.Context, request CaptureIrRawRequest) error
	GetCoderBySessionId(ctx context.Context, sessionId uuid.UUID) (*domainmodels.InfraredStateCoder, error)
	ListTestCases(ctx context.Context, sessionId uuid.UUID) ([]TestCaseWithStates, error)
	TransmitTestCase(ctx context.Context, testCaseId uuid.UUID) error
	RecordTestCaseResult(ctx context.Context, testCaseId uuid.UUID, passed bool) error
	DeleteRecordSessionById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
	DeleteRecordCaseById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
	DeleteRecordStateById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
	DeleteRecordRawById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
	DeleteStateCoderById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
	DeleteTestCaseById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
	DeleteTestCaseStateById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
}

type ListRecordSessionsRequest struct {
	RecordingState       *string
	InfraredDeviceTypeId *uuid.UUID
	CreatedAtStart       *time.Time
	CreatedAtEnd         *time.Time
	Page                 int
	Limit                int
}

type StartRecordSessionRequest struct {
	NodeId               uuid.UUID
	InfraredDeviceTypeId uuid.UUID
	Brand                string
	Model                string
	Definitions          []StartRecordSessionDefinition
	CreatedBy            *uuid.UUID
}

type StartRecordSessionDefinition struct {
	InfraredStateId uuid.UUID
	Options         []string
	Minimum         *float64
	Maximum         *float64
	Step            *float64
}

type CaseWithStatesAndRaw struct {
	Case   domainmodels.InfraredStateDeviceRecordCase
	States []domainmodels.InfraredStateDeviceRecordState
	Raw    []domainmodels.InfraredStateDeviceRecordRaw
}

type TestCaseWithStates struct {
	TestCase domainmodels.InfraredTestCase
	States   []domainmodels.InfraredTestCaseState
}

type CaptureIrRawRequest struct {
	NodeDeviceId string
	RawData      []int32
}
