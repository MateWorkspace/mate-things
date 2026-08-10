package applicationinfraredcasegeneration

import (
	"fmt"
	"sort"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type GeneratedCase struct {
	Step        int32
	States      map[uuid.UUID]string
	Description string
}

// Generate enumerates one baseline case (the first value of every state,
// enum options in declared order / range at its minimum) plus one case per
// non-baseline value of every state, changing exactly one state at a time
// from the baseline. This one-factor-at-a-time sweep is what makes every
// later bit-attribution step (Plan B2) sound: any bit that changes between
// a case and the baseline is attributable to that case's single changed
// state.
func Generate(states []domainmodels.InfraredState, definitions []domainmodels.InfraredStateDeviceDefinition) []GeneratedCase {
	definitionByStateId := make(map[uuid.UUID]domainmodels.InfraredStateDeviceDefinition, len(definitions))
	for _, definition := range definitions {
		definitionByStateId[definition.InfraredStateId] = definition
	}

	// Stable, deterministic ordering: sort states by name so re-running
	// Generate on the same input always produces the same case sequence.
	orderedStates := make([]domainmodels.InfraredState, len(states))
	copy(orderedStates, states)
	sort.Slice(orderedStates, func(i, j int) bool { return orderedStates[i].Name < orderedStates[j].Name })

	baseline := Baseline(states, definitions)
	valuesByStateId := make(map[uuid.UUID][]string, len(orderedStates))
	for _, state := range orderedStates {
		values := allValues(definitionByStateId[state.Id])
		if len(values) == 0 {
			continue
		}
		valuesByStateId[state.Id] = values
	}

	cases := []GeneratedCase{{Step: 1, States: cloneMap(baseline)}}
	step := int32(2)
	for _, state := range orderedStates {
		values := valuesByStateId[state.Id]
		for _, value := range values[1:] {
			caseStates := cloneMap(baseline)
			caseStates[state.Id] = value
			cases = append(cases, GeneratedCase{Step: step, States: caseStates})
			step++
		}
	}
	return cases
}

// Baseline computes the same first-value-of-every-state map Generate uses
// for its own baseline case (cases[0]). It's exported so callers that only
// have the persisted case rows (not the original GeneratedCase slice) can
// identify which recorded case is the true baseline by matching its stored
// state values against this map — the case's Step is not reliable for this,
// since a later re-ordering step (e.g. an LLM-chosen press order) can
// reassign Step away from 1.
func Baseline(states []domainmodels.InfraredState, definitions []domainmodels.InfraredStateDeviceDefinition) map[uuid.UUID]string {
	definitionByStateId := make(map[uuid.UUID]domainmodels.InfraredStateDeviceDefinition, len(definitions))
	for _, definition := range definitions {
		definitionByStateId[definition.InfraredStateId] = definition
	}

	baseline := make(map[uuid.UUID]string, len(states))
	for _, state := range states {
		values := allValues(definitionByStateId[state.Id])
		if len(values) == 0 {
			continue
		}
		baseline[state.Id] = values[0]
	}
	return baseline
}

func allValues(definition domainmodels.InfraredStateDeviceDefinition) []string {
	if len(definition.Options) > 0 {
		return definition.Options
	}
	if definition.Minimum == nil || definition.Maximum == nil || definition.Step == nil || *definition.Step <= 0 {
		return nil
	}
	var values []string
	for v := *definition.Minimum; v <= *definition.Maximum; v += *definition.Step {
		values = append(values, formatRangeValue(v))
	}
	return values
}

func formatRangeValue(v float64) string {
	if v == float64(int64(v)) {
		return fmt.Sprintf("%d", int64(v))
	}
	return fmt.Sprintf("%g", v)
}

func cloneMap(m map[uuid.UUID]string) map[uuid.UUID]string {
	clone := make(map[uuid.UUID]string, len(m))
	for k, v := range m {
		clone[k] = v
	}
	return clone
}
