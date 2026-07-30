package applicationnodedevicemanagement

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
	"github.com/google/uuid"
)

func TestUpdateByIdAllowsSafePersistedNodeNames(t *testing.T) {
	for _, name := range []string{"node_AC276E5E030C", "node-AC276E5E030C"} {
		t.Run(name, func(t *testing.T) {
			repository := &recordingNodeRepository{}
			usecase := NewUsecaseImpl(repository, &deviceLogger{})

			err := usecase.UpdateById(context.Background(), domainusecasesnode.UpdateNodeRequest{
				Id:   uuid.New(),
				Name: &name,
			})

			if err != nil {
				t.Fatalf("UpdateById() error = %v, want nil", err)
			}
			if repository.updateCalls != 1 {
				t.Fatalf("repository UpdateById() calls = %d, want 1", repository.updateCalls)
			}
			if repository.name == nil || *repository.name != name {
				t.Fatalf("repository name = %v, want %q", repository.name, name)
			}
			if repository.deviceID != nil {
				t.Fatalf("repository device ID = %v, want immutable nil update", repository.deviceID)
			}
		})
	}
}

func TestUpdateByIdRejectsPathLikeNodeName(t *testing.T) {
	name := "node_../../secrets"
	repository := &recordingNodeRepository{}
	usecase := NewUsecaseImpl(repository, &deviceLogger{})

	err := usecase.UpdateById(context.Background(), domainusecasesnode.UpdateNodeRequest{
		Id:   uuid.New(),
		Name: &name,
	})

	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("UpdateById() error = %v, want validation error", err)
	}
	if repository.updateCalls != 0 {
		t.Fatalf("repository UpdateById() calls = %d, want 0", repository.updateCalls)
	}
}

type recordingNodeRepository struct {
	domainusecasesrepocache.Node
	updateCalls int
	name        *string
	deviceID    *string
}

func (r *recordingNodeRepository) UpdateById(
	_ context.Context,
	_ uuid.UUID,
	_ *uuid.UUID,
	deviceID *string,
	_ *string,
	name *string,
	_ *uuid.UUID,
	_ *string,
	_ *bool,
	_ *json.RawMessage,
	_ *uuid.UUID,
) error {
	r.updateCalls++
	r.name = name
	r.deviceID = deviceID
	return nil
}

type deviceLogger struct {
	domaincontractslogger.Leveled
}
