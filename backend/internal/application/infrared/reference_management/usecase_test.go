package applicationinfraredreferencemanagement

import (
	"context"
	"errors"
	"testing"

	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type fakeDeviceTypeRepository struct {
	domaincontractsrepository.InfraredDeviceType
	createdName string
	createdBy   *uuid.UUID
	createErr   error
	listResult  []domainmodels.InfraredDeviceType
	deletedId   uuid.UUID
	deletedBy   *uuid.UUID
	deleteErr   error
}

func (f *fakeDeviceTypeRepository) Create(_ context.Context, name string, createdBy *uuid.UUID) (uuid.UUID, error) {
	f.createdName, f.createdBy = name, createdBy
	if f.createErr != nil {
		return uuid.Nil, f.createErr
	}
	return uuid.New(), nil
}
func (f *fakeDeviceTypeRepository) List(_ context.Context) ([]domainmodels.InfraredDeviceType, error) {
	return f.listResult, nil
}
func (f *fakeDeviceTypeRepository) DeleteById(_ context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	f.deletedId, f.deletedBy = id, deletedBy
	return f.deleteErr
}

type fakeDeviceRepository struct {
	domaincontractsrepository.InfraredDevice
	deletedId uuid.UUID
	deletedBy *uuid.UUID
	deleteErr error
}

func (f *fakeDeviceRepository) DeleteById(_ context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	f.deletedId, f.deletedBy = id, deletedBy
	return f.deleteErr
}

type fakeStateRepository struct {
	domaincontractsrepository.InfraredState
	createdDeviceTypeId uuid.UUID
	createdName         string
	createdType         domainmodels.InfraredStateType
	createdBy           *uuid.UUID
	listResult          []domainmodels.InfraredState
	deletedId           uuid.UUID
	deletedBy           *uuid.UUID
	deleteErr           error
}

func (f *fakeStateRepository) Create(_ context.Context, deviceTypeId uuid.UUID, name string, stateType domainmodels.InfraredStateType, createdBy *uuid.UUID) (uuid.UUID, error) {
	f.createdDeviceTypeId, f.createdName, f.createdType, f.createdBy = deviceTypeId, name, stateType, createdBy
	return uuid.New(), nil
}
func (f *fakeStateRepository) ListByDeviceTypeId(_ context.Context, _ uuid.UUID) ([]domainmodels.InfraredState, error) {
	return f.listResult, nil
}
func (f *fakeStateRepository) DeleteById(_ context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	f.deletedId, f.deletedBy = id, deletedBy
	return f.deleteErr
}

type fakeDefinitionRepository struct {
	domaincontractsrepository.InfraredStateDeviceDefinition
	listResult []domainmodels.InfraredStateDeviceDefinition
	deletedId  uuid.UUID
	deletedBy  *uuid.UUID
	deleteErr  error
}

func (f *fakeDefinitionRepository) ListByDeviceId(_ context.Context, _ uuid.UUID) ([]domainmodels.InfraredStateDeviceDefinition, error) {
	return f.listResult, nil
}
func (f *fakeDefinitionRepository) DeleteById(_ context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	f.deletedId, f.deletedBy = id, deletedBy
	return f.deleteErr
}

func TestCreateDeviceTypeForwardsToRepository(t *testing.T) {
	deviceTypeRepo := &fakeDeviceTypeRepository{}
	createdBy := uuid.New()
	usecase := NewUsecaseImpl(deviceTypeRepo, &fakeDeviceRepository{}, &fakeStateRepository{}, &fakeDefinitionRepository{})

	if _, err := usecase.CreateDeviceType(context.Background(), "Air Conditioner", &createdBy); err != nil {
		t.Fatalf("CreateDeviceType() error = %v, want nil", err)
	}
	if deviceTypeRepo.createdName != "Air Conditioner" || deviceTypeRepo.createdBy != &createdBy {
		t.Fatalf("Create() called with (%q, %v), want (%q, %v)", deviceTypeRepo.createdName, deviceTypeRepo.createdBy, "Air Conditioner", &createdBy)
	}
}

func TestListDeviceTypesForwardsToRepository(t *testing.T) {
	want := []domainmodels.InfraredDeviceType{{Id: uuid.New(), Name: "Air Conditioner"}}
	deviceTypeRepo := &fakeDeviceTypeRepository{listResult: want}
	usecase := NewUsecaseImpl(deviceTypeRepo, &fakeDeviceRepository{}, &fakeStateRepository{}, &fakeDefinitionRepository{})

	got, err := usecase.ListDeviceTypes(context.Background())
	if err != nil {
		t.Fatalf("ListDeviceTypes() error = %v, want nil", err)
	}
	if len(got) != 1 || got[0].Id != want[0].Id {
		t.Fatalf("ListDeviceTypes() = %v, want %v", got, want)
	}
}

func TestDeleteDeviceTypeByIdForwardsToRepository(t *testing.T) {
	id, deletedBy := uuid.New(), uuid.New()
	deviceTypeRepo := &fakeDeviceTypeRepository{}
	usecase := NewUsecaseImpl(deviceTypeRepo, &fakeDeviceRepository{}, &fakeStateRepository{}, &fakeDefinitionRepository{})

	if err := usecase.DeleteDeviceTypeById(context.Background(), id, &deletedBy); err != nil {
		t.Fatalf("DeleteDeviceTypeById() error = %v, want nil", err)
	}
	if deviceTypeRepo.deletedId != id {
		t.Fatalf("DeleteById() called with id %v, want %v", deviceTypeRepo.deletedId, id)
	}

	deviceTypeRepo.deleteErr = errors.New("not found")
	if err := usecase.DeleteDeviceTypeById(context.Background(), id, &deletedBy); !errors.Is(err, deviceTypeRepo.deleteErr) {
		t.Fatalf("DeleteDeviceTypeById() error = %v, want the repository's error propagated", err)
	}
}

func TestCreateStateForwardsToRepository(t *testing.T) {
	deviceTypeId, createdBy := uuid.New(), uuid.New()
	stateRepo := &fakeStateRepository{}
	usecase := NewUsecaseImpl(&fakeDeviceTypeRepository{}, &fakeDeviceRepository{}, stateRepo, &fakeDefinitionRepository{})

	if _, err := usecase.CreateState(context.Background(), deviceTypeId, "POWER", domainmodels.InfraredStateTypeEnum, &createdBy); err != nil {
		t.Fatalf("CreateState() error = %v, want nil", err)
	}
	if stateRepo.createdDeviceTypeId != deviceTypeId || stateRepo.createdName != "POWER" || stateRepo.createdType != domainmodels.InfraredStateTypeEnum {
		t.Fatalf("Create() called with (%v, %q, %q), want (%v, %q, %q)", stateRepo.createdDeviceTypeId, stateRepo.createdName, stateRepo.createdType, deviceTypeId, "POWER", domainmodels.InfraredStateTypeEnum)
	}
}

func TestListStatesByDeviceTypeIdForwardsToRepository(t *testing.T) {
	want := []domainmodels.InfraredState{{Id: uuid.New(), Name: "POWER"}}
	stateRepo := &fakeStateRepository{listResult: want}
	usecase := NewUsecaseImpl(&fakeDeviceTypeRepository{}, &fakeDeviceRepository{}, stateRepo, &fakeDefinitionRepository{})

	got, err := usecase.ListStatesByDeviceTypeId(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("ListStatesByDeviceTypeId() error = %v, want nil", err)
	}
	if len(got) != 1 || got[0].Id != want[0].Id {
		t.Fatalf("ListStatesByDeviceTypeId() = %v, want %v", got, want)
	}
}

func TestDeleteStateByIdForwardsToRepository(t *testing.T) {
	id, deletedBy := uuid.New(), uuid.New()
	stateRepo := &fakeStateRepository{}
	usecase := NewUsecaseImpl(&fakeDeviceTypeRepository{}, &fakeDeviceRepository{}, stateRepo, &fakeDefinitionRepository{})

	if err := usecase.DeleteStateById(context.Background(), id, &deletedBy); err != nil {
		t.Fatalf("DeleteStateById() error = %v, want nil", err)
	}
	if stateRepo.deletedId != id {
		t.Fatalf("DeleteById() called with id %v, want %v", stateRepo.deletedId, id)
	}
}

func TestDeleteDeviceByIdForwardsToRepository(t *testing.T) {
	id, deletedBy := uuid.New(), uuid.New()
	deviceRepo := &fakeDeviceRepository{}
	usecase := NewUsecaseImpl(&fakeDeviceTypeRepository{}, deviceRepo, &fakeStateRepository{}, &fakeDefinitionRepository{})

	if err := usecase.DeleteDeviceById(context.Background(), id, &deletedBy); err != nil {
		t.Fatalf("DeleteDeviceById() error = %v, want nil", err)
	}
	if deviceRepo.deletedId != id {
		t.Fatalf("DeleteById() called with id %v, want %v", deviceRepo.deletedId, id)
	}
}

func TestListDefinitionsByDeviceIdForwardsToRepository(t *testing.T) {
	want := []domainmodels.InfraredStateDeviceDefinition{{Id: uuid.New()}}
	definitionRepo := &fakeDefinitionRepository{listResult: want}
	usecase := NewUsecaseImpl(&fakeDeviceTypeRepository{}, &fakeDeviceRepository{}, &fakeStateRepository{}, definitionRepo)

	got, err := usecase.ListDefinitionsByDeviceId(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("ListDefinitionsByDeviceId() error = %v, want nil", err)
	}
	if len(got) != 1 || got[0].Id != want[0].Id {
		t.Fatalf("ListDefinitionsByDeviceId() = %v, want %v", got, want)
	}
}

func TestDeleteDefinitionByIdForwardsToRepository(t *testing.T) {
	id, deletedBy := uuid.New(), uuid.New()
	definitionRepo := &fakeDefinitionRepository{}
	usecase := NewUsecaseImpl(&fakeDeviceTypeRepository{}, &fakeDeviceRepository{}, &fakeStateRepository{}, definitionRepo)

	if err := usecase.DeleteDefinitionById(context.Background(), id, &deletedBy); err != nil {
		t.Fatalf("DeleteDefinitionById() error = %v, want nil", err)
	}
	if definitionRepo.deletedId != id {
		t.Fatalf("DeleteById() called with id %v, want %v", definitionRepo.deletedId, id)
	}
}
