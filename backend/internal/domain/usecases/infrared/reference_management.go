package domainusecasesinfrared

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

// ReferenceManagement covers the 4 infrared tables that describe reference
// data (device types, devices, states, and their per-device definitions)
// rather than an in-flight recording session — RecordSessionManagement owns
// the session lifecycle and creates InfraredDevice/InfraredStateDeviceDefinition
// rows as a side effect of Start, but device types and states themselves are
// managed independently of any one session.
type ReferenceManagement interface {
	CreateDeviceType(ctx context.Context, name string, createdBy *uuid.UUID) (uuid.UUID, error)
	ListDeviceTypes(ctx context.Context) ([]domainmodels.InfraredDeviceType, error)
	DeleteDeviceTypeById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
	CreateState(ctx context.Context, deviceTypeId uuid.UUID, name string, stateType domainmodels.InfraredStateType, createdBy *uuid.UUID) (uuid.UUID, error)
	ListStatesByDeviceTypeId(ctx context.Context, deviceTypeId uuid.UUID) ([]domainmodels.InfraredState, error)
	DeleteStateById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
	DeleteDeviceById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
	GetDeviceById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredDevice, error)
	ListDefinitionsByDeviceId(ctx context.Context, deviceId uuid.UUID) ([]domainmodels.InfraredStateDeviceDefinition, error)
	DeleteDefinitionById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
}
