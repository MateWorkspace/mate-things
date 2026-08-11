package infrastructureloggerleveled

import (
	"errors"
	"testing"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

func TestNormalizeMetaConvertsErrorsToTheirMessage(t *testing.T) {
	source := errors.New("openai generation failed: model not found")
	wrapped := domainmodels.NewError("failed to generate case script", domainmodels.ErrTypeFailure, source)

	got := normalizeMeta(domainmodels.LoggerMeta{
		"err":        wrapped,
		"session_id": "abc30223-1008-4f28-a28f-7403b461c112",
	})

	wantErr := wrapped.Error()
	if got["err"] != wantErr {
		t.Fatalf("normalizeMeta()[\"err\"] = %v, want %q", got["err"], wantErr)
	}
	if got["session_id"] != "abc30223-1008-4f28-a28f-7403b461c112" {
		t.Fatalf("normalizeMeta()[\"session_id\"] = %v, want unchanged", got["session_id"])
	}
}

func TestNormalizeMetaLeavesNonErrorValuesUnchanged(t *testing.T) {
	got := normalizeMeta(domainmodels.LoggerMeta{
		"count": 3,
		"name":  "case",
	})

	if got["count"] != 3 || got["name"] != "case" {
		t.Fatalf("normalizeMeta() mutated non-error values: %+v", got)
	}
}

func TestNormalizeMetaHandlesNilError(t *testing.T) {
	var nilErr error
	got := normalizeMeta(domainmodels.LoggerMeta{"err": nilErr})

	if got["err"] != nilErr {
		t.Fatalf("normalizeMeta()[\"err\"] = %v, want nil error left as-is", got["err"])
	}
}
