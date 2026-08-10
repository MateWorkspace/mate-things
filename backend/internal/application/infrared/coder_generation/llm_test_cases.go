package applicationinfraredcodergeneration

import (
	"context"
	"encoding/json"
	"fmt"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

const testCaseResponseSchema = `{
	"type": "array",
	"items": {
		"type": "object",
		"properties": {
			"description": {"type": "string"},
			"states": {"type": "object", "additionalProperties": {"type": "string"}}
		},
		"required": ["description", "states"],
		"additionalProperties": false
	}
}`

type TestCasePlan struct {
	Description string
	States      map[uuid.UUID]string
}

type testCasePlanResponse struct {
	Description string            `json:"description"`
	States      map[string]string `json:"states"`
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

	prompt := fmt.Sprintf(
		"Device: %s %s\n\nProtocol summary: %s\n\nProtocol detail: %s\n\nPropose a minimal but sufficient set of test scenarios to verify this encoder works correctly on real hardware. Each scenario names a full target state (every field) and a short description of what a technician should observe. Respond as a JSON array matching the given schema.",
		deviceBrand, deviceModel, coder.SummaryReadme, coder.DetailReadme,
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

	var responses []testCasePlanResponse
	if err := json.Unmarshal([]byte(result.Text), &responses); err != nil {
		return nil, domainmodels.NewError("llm returned malformed test case response", domainmodels.ErrTypeFailure, err)
	}

	plans := make([]TestCasePlan, 0, len(responses))
	for _, r := range responses {
		states := make(map[uuid.UUID]string, len(r.States))
		for name, value := range r.States {
			stateId, ok := stateIdByName[name]
			if !ok {
				return nil, domainmodels.NewError(fmt.Sprintf("llm referenced unknown state %q", name), domainmodels.ErrTypeFailure, nil)
			}
			states[stateId] = value
		}
		plans = append(plans, TestCasePlan{Description: r.Description, States: states})
	}
	return plans, nil
}
