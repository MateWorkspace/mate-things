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

type RetryCasePlan struct {
	Description string
	States      map[uuid.UUID]string
}

// WriteRetryCases asks the LLM to propose new targeted recording scenarios
// after a test case failed hardware verification, given the failing
// scenario's target state and the current coder's own self-description.
// Reuses testCaseResponseSchema's shape (description + states object) —
// the LLM's output contract is identical to WriteTestCases', even though
// the two calls serve different points in the session lifecycle.
func WriteRetryCases(
	ctx context.Context,
	client domaincontractsllm.Client,
	deviceBrand string,
	deviceModel string,
	coder domainmodels.InfraredStateCoder,
	failedTestCaseStates []map[uuid.UUID]string,
	states []domainmodels.InfraredState,
) ([]RetryCasePlan, error) {
	stateNameById := make(map[string]string, len(states))
	stateIdByName := make(map[string]uuid.UUID, len(states))
	for _, state := range states {
		stateNameById[state.Id.String()] = state.Name
		stateIdByName[state.Name] = state.Id
	}

	var failedDescription strings.Builder
	for i, failedStates := range failedTestCaseStates {
		fmt.Fprintf(&failedDescription, "Failure %d: ", i+1)
		for stateId, value := range failedStates {
			fmt.Fprintf(&failedDescription, "%s=%s ", stateNameById[stateId.String()], value)
		}
		failedDescription.WriteString("\n")
	}

	prompt := fmt.Sprintf(
		"Device: %s %s\n\nProtocol summary: %s\n\nProtocol detail: %s\n\nA hardware test transmitting this target state failed: %s\n\nPropose new targeted recording scenarios (full target state plus a short description) likely to explain the discrepancy and improve the encoder. Respond as JSON matching the given schema.",
		deviceBrand, deviceModel, coder.SummaryReadme, coder.DetailReadme, failedDescription.String(),
	)

	result, err := client.GenerateText(ctx, domaincontractsllm.GenerateTextRequest{
		System:          "You are an expert at diagnosing infrared remote control encoder failures and proposing corrective recording scenarios.",
		Prompt:          prompt,
		MaxOutputTokens: 4096,
		ResponseSchema:  []byte(testCaseResponseSchema),
	})
	if err != nil {
		return nil, domainmodels.NewError("failed to generate retry cases", domainmodels.ErrTypeFailure, err)
	}

	var wrapper testCasePlanListResponse
	if err := json.Unmarshal([]byte(result.Text), &wrapper); err != nil {
		return nil, domainmodels.NewError("llm returned malformed retry case response", domainmodels.ErrTypeFailure, err)
	}

	plans := make([]RetryCasePlan, 0, len(wrapper.Entries))
	for _, r := range wrapper.Entries {
		planStates := make(map[uuid.UUID]string, len(r.States))
		for _, s := range r.States {
			stateId, ok := stateIdByName[s.Name]
			if !ok {
				return nil, domainmodels.NewError(fmt.Sprintf("llm referenced unknown state %q", s.Name), domainmodels.ErrTypeFailure, nil)
			}
			planStates[stateId] = s.Value
		}
		plans = append(plans, RetryCasePlan{Description: r.Description, States: planStates})
	}
	return plans, nil
}
