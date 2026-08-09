package domainmodels

import (
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

type InfraredRecordSession struct {
	Id               uuid.UUID
	NodeId           uuid.UUID
	InfraredDeviceId uuid.UUID
	RecordingState   string
	IsCompleted      bool
}

// __________[ STATEFUL PART (STATE) ]__________

type InfraredState struct {
	Id                   uuid.UUID
	InfraredDeviceTypeId uuid.UUID
	Name                 string
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

type InfraredStateDeviceRecordCase struct {
	Id               uuid.UUID
	InfraredDeviceId uuid.UUID
	Step             int32
	Description      string
}

type InfraredStateDeviceRecordState struct {
	Id                              uuid.UUID
	InfraredStateDeviceRecordCaseId uuid.UUID
	InfraredStateId                 uuid.UUID
	StateValue                      string
}

type InfraredStateDeviceRecordRaw struct {
	Id                              uuid.UUID
	InfraredStateDeviceRecordCaseId uuid.UUID
	RawData                         []byte
}
