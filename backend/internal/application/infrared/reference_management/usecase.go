package applicationinfraredreferencemanagement

import (
	"context"

	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesinfrared "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/infrared"
	"github.com/google/uuid"
)

type usecase struct {
	deviceType domaincontractsrepository.InfraredDeviceType
	device     domaincontractsrepository.InfraredDevice
	state      domaincontractsrepository.InfraredState
	definition domaincontractsrepository.InfraredStateDeviceDefinition
}

func NewUsecaseImpl(
	deviceType domaincontractsrepository.InfraredDeviceType,
	device domaincontractsrepository.InfraredDevice,
	state domaincontractsrepository.InfraredState,
	definition domaincontractsrepository.InfraredStateDeviceDefinition,
) domainusecasesinfrared.ReferenceManagement {
	return &usecase{deviceType: deviceType, device: device, state: state, definition: definition}
}

func (u *usecase) CreateDeviceType(ctx context.Context, name string, createdBy *uuid.UUID) (uuid.UUID, error) {
	return u.deviceType.Create(ctx, name, createdBy)
}

func (u *usecase) ListDeviceTypes(ctx context.Context) ([]domainmodels.InfraredDeviceType, error) {
	return u.deviceType.List(ctx)
}

func (u *usecase) DeleteDeviceTypeById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	return u.deviceType.DeleteById(ctx, id, deletedBy)
}

func (u *usecase) CreateState(ctx context.Context, deviceTypeId uuid.UUID, name string, stateType domainmodels.InfraredStateType, createdBy *uuid.UUID) (uuid.UUID, error) {
	return u.state.Create(ctx, deviceTypeId, name, stateType, createdBy)
}

func (u *usecase) ListStatesByDeviceTypeId(ctx context.Context, deviceTypeId uuid.UUID) ([]domainmodels.InfraredState, error) {
	return u.state.ListByDeviceTypeId(ctx, deviceTypeId)
}

func (u *usecase) DeleteStateById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	return u.state.DeleteById(ctx, id, deletedBy)
}

func (u *usecase) DeleteDeviceById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	return u.device.DeleteById(ctx, id, deletedBy)
}

func (u *usecase) ListDefinitionsByDeviceId(ctx context.Context, deviceId uuid.UUID) ([]domainmodels.InfraredStateDeviceDefinition, error) {
	return u.definition.ListByDeviceId(ctx, deviceId)
}

func (u *usecase) DeleteDefinitionById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	return u.definition.DeleteById(ctx, id, deletedBy)
}
