package applicationinfraredcasegeneration

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

const responseSchema = `{
	"type": "array",
	"items": {
		"type": "object",
		"properties": {
			"case_index": {"type": "integer"},
			"description": {"type": "string"},
			"order": {"type": "integer"}
		},
		"required": ["case_index", "description", "order"],
		"additionalProperties": false
	}
}`

type scriptEntry struct {
	CaseIndex   int    `json:"case_index"`
	Description string `json:"description"`
	Order       int    `json:"order"`
}

// WriteScript asks the LLM to write a human-facing instruction for every
// case and choose a press-order minimizing physical remote interaction. It
// never changes which state values a case targets — only Description and
// Step, per this plan's contract that case existence is deterministic and
// the LLM's role is language and ordering only.
func WriteScript(
	ctx context.Context,
	client domaincontractsllm.Client,
	deviceBrand string,
	deviceModel string,
	states []domainmodels.InfraredState,
	cases []GeneratedCase,
) ([]GeneratedCase, error) {
	stateNameById := make(map[string]string, len(states))
	for _, state := range states {
		stateNameById[state.Id.String()] = state.Name
	}

	var promptBuilder strings.Builder
	fmt.Fprintf(&promptBuilder, "Device: %s %s\n\n", deviceBrand, deviceModel)
	promptBuilder.WriteString("Below is a numbered list of remote-control states to physically set up, one per case. ")
	promptBuilder.WriteString("For each case, write a short, clear instruction telling a technician the FULL state to set the remote to " +
		"(every field, not just what changed from the previous case), then choose a press-order (\"order\", 1-based) that groups similar " +
		"button sequences together to minimize physical button presses across the whole list. Respond as a JSON array matching the given schema, " +
		"one entry per case_index below.\n\n")
	for i, c := range cases {
		fmt.Fprintf(&promptBuilder, "case_index %d: ", i)
		fields := make([]string, 0, len(c.States))
		for stateId, value := range c.States {
			fields = append(fields, fmt.Sprintf("%s=%s", stateNameById[stateId.String()], value))
		}
		sort.Strings(fields)
		promptBuilder.WriteString(strings.Join(fields, ", "))
		promptBuilder.WriteString("\n")
	}

	result, err := client.GenerateText(ctx, domaincontractsllm.GenerateTextRequest{
		System:          "You write concise, technician-facing instructions for recording infrared remote-control signals.",
		Prompt:          promptBuilder.String(),
		MaxOutputTokens: 4096,
		ResponseSchema:  []byte(responseSchema),
	})
	if err != nil {
		return nil, domainmodels.NewError("failed to generate case script", domainmodels.ErrTypeFailure, err)
	}

	var entries []scriptEntry
	if err := json.Unmarshal([]byte(result.Text), &entries); err != nil {
		return nil, domainmodels.NewError("llm returned malformed case script", domainmodels.ErrTypeFailure, err)
	}
	if len(entries) != len(cases) {
		return nil, domainmodels.NewError(
			fmt.Sprintf("llm returned %d case script entries, want %d", len(entries), len(cases)),
			domainmodels.ErrTypeFailure, nil,
		)
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Order < entries[j].Order })

	reordered := make([]GeneratedCase, len(cases))
	for newStep, entry := range entries {
		if entry.CaseIndex < 0 || entry.CaseIndex >= len(cases) {
			return nil, domainmodels.NewError(fmt.Sprintf("llm referenced out-of-range case_index %d", entry.CaseIndex), domainmodels.ErrTypeFailure, nil)
		}
		original := cases[entry.CaseIndex]
		reordered[newStep] = GeneratedCase{
			Step:        int32(newStep + 1),
			States:      original.States,
			Description: entry.Description,
		}
	}
	return reordered, nil
}
