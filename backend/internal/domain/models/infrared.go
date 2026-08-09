package domainmodels

import (
	"time"

	"github.com/google/uuid"
)

type InfraredDeviceType struct {
	Id   uuid.UUID
	Name string
}

type InfraredDevice struct {
	Id                   uuid.UUID
	InfraredDeviceTypeId uuid.UUID
	Brand                string
	Model                string
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

type InfraredRecordSession struct {
	Id                  uuid.UUID
	NodeId              uuid.UUID
	InfraredDeviceId    uuid.UUID
	RecordingState      string
	CurrentRecordCaseId *uuid.UUID
	IsCompleted         bool
	CreatedAt           time.Time
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
}

type InfraredStateDeviceDefinition struct {
	Id               uuid.UUID
	InfraredDeviceId uuid.UUID
	InfraredStateId  uuid.UUID
	Options          []string
	Minimum          *float64
	Maximum          *float64
	Step             *float64
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
}

type InfraredStateDeviceRecordState struct {
	Id                              uuid.UUID
	InfraredStateDeviceRecordCaseId uuid.UUID
	InfraredStateId                 uuid.UUID
	StateValue                      string
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
}
