package applicationinfraredcodergeneration

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

// The schema wraps the array under an "entries" object property, and
// represents each case's states as an array of {name, value} pairs rather
// than a free-form object map: OpenAI's Structured Outputs mode rejects
// both an array-rooted schema and any "additionalProperties" holding a
// schema (open-ended dictionaries aren't representable in strict mode -
// every property has to be named in "required"), while Claude and Gemini
// accept both shapes fine. One schema, working for every provider, beats
// branching per client.
const testCaseResponseSchema = `{
	"type": "object",
	"properties": {
		"entries": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
					"description": {"type": "string"},
					"states": {
						"type": "array",
						"items": {
							"type": "object",
							"properties": {
								"name": {"type": "string"},
								"value": {"type": "string"}
							},
							"required": ["name", "value"],
							"additionalProperties": false
						}
					}
				},
				"required": ["description", "states"],
				"additionalProperties": false
			}
		}
	},
	"required": ["entries"],
	"additionalProperties": false
}`

type TestCasePlan struct {
	Description string
	States      map[uuid.UUID]string
}

type testCaseStateResponse struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type testCasePlanResponse struct {
	Description string                  `json:"description"`
	States      []testCaseStateResponse `json:"states"`
}

type testCasePlanListResponse struct {
	Entries []testCasePlanResponse `json:"entries"`
}

// WriteTestCases asks the LLM to propose a minimal-but-sufficient set of
// test scenarios for the finished encoder, given the coder's own
// self-description (its README fields) and the device's state/definition
// list. This is judgment about WHICH combinations to exercise, not a
// re-derivation of bit layout — the encoder itself is treated as a black
// box here.
func WriteTestCases(
	ctx context.Context,
	client domaincontractsllm.Client,
	deviceBrand string,
	deviceModel string,
	coder domainmodels.InfraredStateCoder,
	states []domainmodels.InfraredState,
	definitions []domainmodels.InfraredStateDeviceDefinition,
) ([]TestCasePlan, error) {
	stateIdByName := make(map[string]uuid.UUID, len(states))
	for _, state := range states {
		stateIdByName[state.Name] = state.Id
	}

	definitionByStateId := make(map[uuid.UUID]domainmodels.InfraredStateDeviceDefinition, len(definitions))
	for _, definition := range definitions {
		definitionByStateId[definition.InfraredStateId] = definition
	}

	var stateCatalog strings.Builder
	for _, state := range states {
		fmt.Fprintf(&stateCatalog, "- %s: %s\n", state.Name, describeLegalValues(definitionByStateId[state.Id]))
	}

	prompt := fmt.Sprintf(
		"Device: %s %s\n\nProtocol summary: %s\n\nProtocol detail: %s\n\nDevice states and their legal values:\n%s\nPropose a minimal but sufficient set of test scenarios to verify this encoder works correctly on real hardware. Each scenario names a full target state (every field, using exactly the state names given above) and a short description of what a technician should observe. Respond as JSON matching the given schema.",
		deviceBrand, deviceModel, coder.SummaryReadme, coder.DetailReadme, stateCatalog.String(),
	)

	result, err := client.GenerateText(ctx, domaincontractsllm.GenerateTextRequest{
		System:          "You are an expert at designing minimal, high-coverage test plans for infrared remote control encoders.",
		Prompt:          prompt,
		MaxOutputTokens: 4096,
		ResponseSchema:  []byte(testCaseResponseSchema),
	})
	if err != nil {
		return nil, domainmodels.NewError("failed to generate test cases", domainmodels.ErrTypeFailure, err)
	}

	var wrapper testCasePlanListResponse
	if err := json.Unmarshal([]byte(result.Text), &wrapper); err != nil {
		return nil, domainmodels.NewError("llm returned malformed test case response", domainmodels.ErrTypeFailure, err)
	}

	plans := make([]TestCasePlan, 0, len(wrapper.Entries))
	for _, r := range wrapper.Entries {
		states := make(map[uuid.UUID]string, len(r.States))
		for _, s := range r.States {
			stateId, ok := stateIdByName[s.Name]
			if !ok {
				return nil, domainmodels.NewError(fmt.Sprintf("llm referenced unknown state %q", s.Name), domainmodels.ErrTypeFailure, nil)
			}
			states[stateId] = s.Value
		}
		plans = append(plans, TestCasePlan{Description: r.Description, States: states})
	}
	return plans, nil
}

func describeLegalValues(definition domainmodels.InfraredStateDeviceDefinition) string {
	if len(definition.Options) > 0 {
		return fmt.Sprintf("one of %v", definition.Options)
	}
	if definition.Minimum != nil && definition.Maximum != nil {
		return fmt.Sprintf("a number between %g and %g", *definition.Minimum, *definition.Maximum)
	}
	return "any string value"
}
