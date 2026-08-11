package applicationinfraredcodergeneration

import (
	"context"
	"encoding/json"
	"fmt"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

// WriteChecksumClarificationCases asks the LLM to propose additional
// recording scenarios likely to help pin down a checksum algorithm it
// couldn't fully infer — used when Validate() reports a ChecksumOnlyGap:
// every owned bit is correct, but the checksum bits aren't, which is a
// data problem (too few examples), not something a repair prompt can fix.
// Runs before any coder is persisted, so it takes the in-memory Coder
// (this attempt's own unpersisted output), not a DB-backed
// domainmodels.InfraredStateCoder. Reuses testCaseResponseSchema's shape
// — identical output contract to WriteTestCases/WriteRetryCases.
func WriteChecksumClarificationCases(
	ctx context.Context,
	client domaincontractsllm.Client,
	deviceBrand string,
	deviceModel string,
	coder Coder,
	states []domainmodels.InfraredState,
) ([]RetryCasePlan, error) {
	stateIdByName := make(map[string]uuid.UUID, len(states))
	for _, state := range states {
		stateIdByName[state.Name] = state.Id
	}

	prompt := fmt.Sprintf(
		"Device: %s %s\n\nProtocol summary: %s\n\nProtocol detail: %s\n\nYour encoder correctly reproduces every bit whose meaning is already known, but could not be verified to compute the checksum bits correctly from the examples recorded so far — there isn't enough data yet to confirm the checksum algorithm. Propose new targeted recording scenarios (full target state plus a short description) likely to reveal the checksum pattern — for example, scenarios that change exactly ONE state from the baseline to a value not recorded yet, or repeat an existing case to confirm determinism. Every scenario must differ from the baseline in exactly one state — never vary multiple states together. Respond as JSON matching the given schema.",
		deviceBrand, deviceModel, coder.SummaryReadme, coder.DetailReadme,
	)

	result, err := client.GenerateText(ctx, domaincontractsllm.GenerateTextRequest{
		System:          "You are an expert at diagnosing infrared remote control encoder failures and proposing corrective recording scenarios.",
		Prompt:          prompt,
		MaxOutputTokens: 4096,
		ResponseSchema:  []byte(testCaseResponseSchema),
	})
	if err != nil {
		return nil, domainmodels.NewError("failed to generate checksum clarification cases", domainmodels.ErrTypeFailure, err)
	}

	var wrapper testCasePlanListResponse
	if err := json.Unmarshal([]byte(result.Text), &wrapper); err != nil {
		return nil, domainmodels.NewError("llm returned malformed checksum clarification response", domainmodels.ErrTypeFailure, err)
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
