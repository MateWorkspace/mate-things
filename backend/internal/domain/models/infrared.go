package domainmodels

import (
	"time"

	"github.com/google/uuid"
)

type InfraredDeviceType struct {
	Id        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID
	DeletedBy *uuid.UUID
}

type InfraredDevice struct {
	Id                   uuid.UUID
	InfraredDeviceTypeId uuid.UUID
	Brand                string
	Model                string
	CreatedAt            time.Time
	UpdatedAt            *time.Time
	DeletedAt            *time.Time
	CreatedBy            *uuid.UUID
	UpdatedBy            *uuid.UUID
	DeletedBy            *uuid.UUID
}

const (
	InfraredRecordingStateDraft               = "DRAFT"
	InfraredRecordingStateCasesGenerating     = "CASES_GENERATING"
	InfraredRecordingStateRecording           = "RECORDING"
	InfraredRecordingStateAnalyzing           = "ANALYZING"
	InfraredRecordingStateFunctionGenerating  = "FUNCTION_GENERATING"
	InfraredRecordingStateTestCasesGenerating = "TEST_CASES_GENERATING"
	InfraredRecordingStateTesting             = "TESTING"
	InfraredRecordingStateCompleted           = "COMPLETED"
	InfraredRecordingStateFailed              = "FAILED"
)

var validInfraredRecordingStates = map[string]struct{}{
	InfraredRecordingStateDraft:               {},
	InfraredRecordingStateCasesGenerating:     {},
	InfraredRecordingStateRecording:           {},
	InfraredRecordingStateAnalyzing:           {},
	InfraredRecordingStateFunctionGenerating:  {},
	InfraredRecordingStateTestCasesGenerating: {},
	InfraredRecordingStateTesting:             {},
	InfraredRecordingStateCompleted:           {},
	InfraredRecordingStateFailed:              {},
}

func IsValidInfraredRecordingState(state string) bool {
	_, ok := validInfraredRecordingStates[state]
	return ok
}

type InfraredRecordSession struct {
	Id                          uuid.UUID
	NodeId                      uuid.UUID
	InfraredDeviceId            uuid.UUID
	RecordingState              string
	CurrentRecordCaseId         *uuid.UUID
	IsCompleted                 bool
	ChecksumClarificationUsedAt *time.Time
	CreatedAt                   time.Time
	DeletedAt                   *time.Time
	CreatedBy                   *uuid.UUID
	DeletedBy                   *uuid.UUID
}

type InfraredRecordSessionListItem struct {
	Id                   uuid.UUID
	RecordingState       string
	IsCompleted          bool
	InfraredDeviceId     uuid.UUID
	Brand                string
	Model                string
	InfraredDeviceTypeId uuid.UUID
	DeviceTypeName       string
	CreatedAt            time.Time
}

type InfraredStateType string

const (
	InfraredStateTypeRange InfraredStateType = "RANGE"
	InfraredStateTypeEnum  InfraredStateType = "ENUM"
)

type InfraredState struct {
	Id                   uuid.UUID
	InfraredDeviceTypeId uuid.UUID
	Name                 string
	Type                 InfraredStateType
	CreatedAt            time.Time
	UpdatedAt            *time.Time
	DeletedAt            *time.Time
	CreatedBy            *uuid.UUID
	UpdatedBy            *uuid.UUID
	DeletedBy            *uuid.UUID
}

type InfraredStateDeviceDefinition struct {
	Id               uuid.UUID
	InfraredDeviceId uuid.UUID
	InfraredStateId  uuid.UUID
	Options          []string
	Minimum          *float64
	Maximum          *float64
	Step             *float64
	CreatedAt        time.Time
	UpdatedAt        *time.Time
	DeletedAt        *time.Time
	CreatedBy        *uuid.UUID
	UpdatedBy        *uuid.UUID
	DeletedBy        *uuid.UUID
}

type InfraredRecordCaseStatus string

const (
	InfraredRecordCaseStatusPending  InfraredRecordCaseStatus = "PENDING"
	InfraredRecordCaseStatusActive   InfraredRecordCaseStatus = "ACTIVE"
	InfraredRecordCaseStatusAccepted InfraredRecordCaseStatus = "ACCEPTED"
)

type InfraredStateDeviceRecordCase struct {
	Id                      uuid.UUID
	InfraredRecordSessionId uuid.UUID
	Step                    int32
	Description             string
	Status                  InfraredRecordCaseStatus
	CreatedAt               time.Time
	DeletedAt               *time.Time
	DeletedBy               *uuid.UUID
}

type InfraredStateDeviceRecordState struct {
	Id                              uuid.UUID
	InfraredStateDeviceRecordCaseId uuid.UUID
	InfraredStateId                 uuid.UUID
	StateValue                      string
	CreatedAt                       time.Time
	DeletedAt                       *time.Time
	DeletedBy                       *uuid.UUID
}

type InfraredRecordRawStatus string

const (
	InfraredRecordRawStatusCaptured  InfraredRecordRawStatus = "CAPTURED"
	InfraredRecordRawStatusAccepted  InfraredRecordRawStatus = "ACCEPTED"
	InfraredRecordRawStatusDiscarded InfraredRecordRawStatus = "DISCARDED"
)

type InfraredStateDeviceRecordRaw struct {
	Id                              uuid.UUID
	InfraredStateDeviceRecordCaseId uuid.UUID
	RawData                         []byte
	Status                          InfraredRecordRawStatus
	DiscardedReason                 *string
	CreatedAt                       time.Time
	DeletedAt                       *time.Time
	DeletedBy                       *uuid.UUID
}

type InfraredRecordSessionEvent struct {
	SessionId           uuid.UUID
	RecordingState      string
	CurrentRecordCaseId *uuid.UUID
}

type InfraredStateCoderStatus string

const (
	InfraredStateCoderStatusUnverified InfraredStateCoderStatus = "UNVERIFIED"
	InfraredStateCoderStatusActive     InfraredStateCoderStatus = "ACTIVE"
	InfraredStateCoderStatusSuperseded InfraredStateCoderStatus = "SUPERSEDED"
)

type InfraredStateCoder struct {
	Id                      uuid.UUID
	InfraredDeviceId        uuid.UUID
	InfraredRecordSessionId uuid.UUID
	EncoderSource           string
	DecoderSource           string
	SummaryReadme           string
	DetailReadme            string
	Status                  InfraredStateCoderStatus
	CreatedAt               time.Time
	DeletedAt               *time.Time
	DeletedBy               *uuid.UUID
}

type InfraredTestCaseStatus string

const (
	InfraredTestCaseStatusPending InfraredTestCaseStatus = "PENDING"
	InfraredTestCaseStatusPassed  InfraredTestCaseStatus = "PASSED"
	InfraredTestCaseStatusFailed  InfraredTestCaseStatus = "FAILED"
)

type InfraredTestCase struct {
	Id                   uuid.UUID
	InfraredStateCoderId uuid.UUID
	Step                 int32
	Description          string
	Status               InfraredTestCaseStatus
	CreatedAt            time.Time
	DeletedAt            *time.Time
	DeletedBy            *uuid.UUID
}

type InfraredTestCaseState struct {
	Id                 uuid.UUID
	InfraredTestCaseId uuid.UUID
	InfraredStateId    uuid.UUID
	StateValue         string
	CreatedAt          time.Time
	DeletedAt          *time.Time
	DeletedBy          *uuid.UUID
}
