package applicationnodeclassmanagement

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

func TestUpdateByIdAllowsSafePersistedNodeClassNames(t *testing.T) {
	for _, name := range []string{"base_node", "base-node"} {
		t.Run(name, func(t *testing.T) {
			repository := &recordingNodeClassRepository{}
			usecase := NewUsecaseImpl(repository, &stubNodeClassAction{}, &classLogger{})

			err := usecase.UpdateById(context.Background(), domainusecasesnode.UpdateNodeClassRequest{
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
		})
	}
}

func TestUpdateByIdRejectsPathLikeNodeClassName(t *testing.T) {
	name := "base_node/../../secrets"
	repository := &recordingNodeClassRepository{}
	usecase := NewUsecaseImpl(repository, &stubNodeClassAction{}, &classLogger{})

	err := usecase.UpdateById(context.Background(), domainusecasesnode.UpdateNodeClassRequest{
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

type recordingNodeClassRepository struct {
	domainusecasesrepocache.NodeClass
	updateCalls int
	name        *string
}

func (r *recordingNodeClassRepository) UpdateById(
	_ context.Context,
	_ uuid.UUID,
	name *string,
	_ *string,
	_ *json.RawMessage,
	_ *uuid.UUID,
) error {
	r.updateCalls++
	r.name = name
	return nil
}

type classLogger struct {
	domaincontractslogger.Leveled
}

type stubNodeClassAction struct {
	domainusecasesrepocache.NodeClassAction
}
