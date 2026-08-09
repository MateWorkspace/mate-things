package applicationinfraredcasegeneration

import (
	"testing"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

func TestGenerateProducesBaselinePlusOneCasePerStateValue(t *testing.T) {
	powerStateId := uuid.New()
	modeStateId := uuid.New()

	states := []domainmodels.InfraredState{
		{Id: powerStateId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum},
		{Id: modeStateId, Name: "MODE", Type: domainmodels.InfraredStateTypeEnum},
	}
	definitions := []domainmodels.InfraredStateDeviceDefinition{
		{InfraredStateId: powerStateId, Options: []string{"ON", "OFF"}},
		{InfraredStateId: modeStateId, Options: []string{"COOL", "HEAT", "FAN"}},
	}

	cases := Generate(states, definitions)

	// Baseline (first value of every state) + one case per NON-baseline value:
	// POWER has 1 non-baseline value (OFF), MODE has 2 (HEAT, FAN) => 1 + 1 + 2 = 4.
	if len(cases) != 4 {
		t.Fatalf("Generate() returned %d cases, want 4", len(cases))
	}

	baseline := cases[0]
	if baseline.States[powerStateId] != "ON" || baseline.States[modeStateId] != "COOL" {
		t.Fatalf("baseline states = %v, want POWER=ON MODE=COOL", baseline.States)
	}
	if baseline.Step != 1 {
		t.Fatalf("baseline step = %d, want 1", baseline.Step)
	}

	for i, c := range cases {
		if c.Step != int32(i+1) {
			t.Fatalf("case[%d].Step = %d, want %d (steps must be contiguous starting at 1)", i, c.Step, i+1)
		}
	}
}

func TestGenerateEachNonBaselineCaseDiffersFromBaselineInExactlyOneState(t *testing.T) {
	powerStateId := uuid.New()
	modeStateId := uuid.New()

	states := []domainmodels.InfraredState{
		{Id: powerStateId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum},
		{Id: modeStateId, Name: "MODE", Type: domainmodels.InfraredStateTypeEnum},
	}
	definitions := []domainmodels.InfraredStateDeviceDefinition{
		{InfraredStateId: powerStateId, Options: []string{"ON", "OFF"}},
		{InfraredStateId: modeStateId, Options: []string{"COOL", "HEAT"}},
	}

	cases := Generate(states, definitions)
	baseline := cases[0]

	for _, c := range cases[1:] {
		differences := 0
		for stateId, value := range c.States {
			if baseline.States[stateId] != value {
				differences++
			}
		}
		if differences != 1 {
			t.Errorf("case at step %d differs from baseline in %d states, want exactly 1 (states = %v)", c.Step, differences, c.States)
		}
	}
}

func TestGenerateRangeStateSweepsFromMinimumToMaximumByStep(t *testing.T) {
	temperatureStateId := uuid.New()

	states := []domainmodels.InfraredState{
		{Id: temperatureStateId, Name: "TEMPERATURE", Type: domainmodels.InfraredStateTypeRange},
	}
	minimum, maximum, step := 16.0, 18.0, 1.0
	definitions := []domainmodels.InfraredStateDeviceDefinition{
		{InfraredStateId: temperatureStateId, Minimum: &minimum, Maximum: &maximum, Step: &step},
	}

	cases := Generate(states, definitions)

	// Baseline = minimum (16); non-baseline sweep = 17, 18 => 3 cases total.
	if len(cases) != 3 {
		t.Fatalf("Generate() returned %d cases, want 3", len(cases))
	}
	want := []string{"16", "17", "18"}
	for i, c := range cases {
		if c.States[temperatureStateId] != want[i] {
			t.Errorf("case[%d] TEMPERATURE = %q, want %q", i, c.States[temperatureStateId], want[i])
		}
	}
}

func TestGenerateReturnsOnlyBaselineWhenNoStatesGiven(t *testing.T) {
	cases := Generate(nil, nil)
	if len(cases) != 1 {
		t.Fatalf("Generate() returned %d cases, want 1 (baseline only)", len(cases))
	}
	if len(cases[0].States) != 0 {
		t.Fatalf("baseline States = %v, want empty", cases[0].States)
	}
}
