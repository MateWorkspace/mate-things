package applicationinfraredcasegeneration

import (
	"context"
	"testing"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type fakeLlmClient struct {
	responseText string
	err          error
	lastRequest  domaincontractsllm.GenerateTextRequest
}

func (f *fakeLlmClient) GenerateText(_ context.Context, req domaincontractsllm.GenerateTextRequest) (domaincontractsllm.GenerateTextResult, error) {
	f.lastRequest = req
	return domaincontractsllm.GenerateTextResult{Text: f.responseText}, f.err
}

func TestWriteScriptAppliesDescriptionsAndReorderingFromLlmResponse(t *testing.T) {
	powerStateId := uuid.New()
	states := []domainmodels.InfraredState{{Id: powerStateId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}
	cases := []GeneratedCase{
		{Step: 1, States: map[uuid.UUID]string{powerStateId: "ON"}},
		{Step: 2, States: map[uuid.UUID]string{powerStateId: "OFF"}},
	}

	client := &fakeLlmClient{responseText: `[
		{"case_index": 0, "description": "Set the remote to POWER ON.", "order": 2},
		{"case_index": 1, "description": "Set the remote to POWER OFF.", "order": 1}
	]`}

	result, err := WriteScript(context.Background(), client, "Polytron", "PAC-09HDN", states, cases)
	if err != nil {
		t.Fatalf("WriteScript() error = %v, want nil", err)
	}
	if len(result) != 2 {
		t.Fatalf("WriteScript() returned %d cases, want 2", len(result))
	}

	// case_index 1 (order 1) comes first now, at Step 1; case_index 0 (order 2) is Step 2.
	if result[0].Description != "Set the remote to POWER OFF." || result[0].Step != 1 {
		t.Errorf("result[0] = %+v, want reordered OFF case at step 1", result[0])
	}
	if result[1].Description != "Set the remote to POWER ON." || result[1].Step != 2 {
		t.Errorf("result[1] = %+v, want reordered ON case at step 2", result[1])
	}

	if client.lastRequest.ResponseSchema == nil {
		t.Fatalf("GenerateText() request had no ResponseSchema — structured output was not requested")
	}
	if client.lastRequest.MaxOutputTokens <= 0 {
		t.Fatalf("GenerateText() request had MaxOutputTokens = %d, want > 0", client.lastRequest.MaxOutputTokens)
	}
}

func TestWriteScriptPropagatesLlmError(t *testing.T) {
	client := &fakeLlmClient{err: context.DeadlineExceeded}
	_, err := WriteScript(context.Background(), client, "Polytron", "PAC-09HDN", nil, []GeneratedCase{{Step: 1, States: map[uuid.UUID]string{}}})
	if err == nil {
		t.Fatalf("WriteScript() error = nil, want propagated error")
	}
}
