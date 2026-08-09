package applicationshared

import (
	"errors"
	"testing"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

func TestRequiredInfraredRecordingStateRejectsUnknownValue(t *testing.T) {
	_, err := RequiredInfraredRecordingState("SLEEPING", "recording_state")
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("RequiredInfraredRecordingState() error = %v, want validation error", err)
	}
}

func TestRequiredInfraredRecordingStateAcceptsKnownValues(t *testing.T) {
	values := []string{
		domainmodels.InfraredRecordingStateDraft,
		domainmodels.InfraredRecordingStateCasesGenerating,
		domainmodels.InfraredRecordingStateRecording,
		domainmodels.InfraredRecordingStateAnalyzing,
		domainmodels.InfraredRecordingStateFunctionGenerating,
		domainmodels.InfraredRecordingStateTestCasesGenerating,
		domainmodels.InfraredRecordingStateTesting,
		domainmodels.InfraredRecordingStateCompleted,
		domainmodels.InfraredRecordingStateFailed,
	}
	for _, value := range values {
		if _, err := RequiredInfraredRecordingState(value, "recording_state"); err != nil {
			t.Errorf("RequiredInfraredRecordingState(%q) error = %v, want nil", value, err)
		}
	}
}

func TestRequiredInfraredBrandRejectsEmpty(t *testing.T) {
	_, err := RequiredInfraredBrand("", "brand")
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("RequiredInfraredBrand(\"\") error = %v, want validation error", err)
	}
}

func TestRequiredInfraredModelRejectsEmpty(t *testing.T) {
	_, err := RequiredInfraredModel("", "model")
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("RequiredInfraredModel(\"\") error = %v, want validation error", err)
	}
}
