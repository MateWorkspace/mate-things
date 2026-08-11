package applicationinfraredcodergeneration

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	applicationinfraredanalysis "github.com/MateWorkspace/mate-things/backend/internal/application/infrared/analysis"
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

func TestWriteCoderParsesStructuredResponse(t *testing.T) {
	powerId := uuid.New()
	states := []domainmodels.InfraredState{{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}
	definitions := []domainmodels.InfraredStateDeviceDefinition{{InfraredStateId: powerId, Options: []string{"ON", "OFF"}}}
	payload := applicationinfraredanalysis.AnalysisPayload{
		FrameBitLength: 4,
		BaselineBits:   []int{0, 0, 0, 0},
		States:         []applicationinfraredanalysis.StateAttribution{{StateId: powerId, BitOffsets: []int{0}, ValueBits: map[string][]int{"ON": {0}, "OFF": {1}}}},
	}

	responseBody, _ := json.Marshal(map[string]string{
		"encoder_source": "function encode(state) { return [9000, 4500]; }",
		"decoder_source": "function decode(raw) { return {}; }",
		"summary_readme": "Polytron PAC-09HDN infrared protocol.",
		"detail_readme":  "POWER is encoded at bit offset 0.",
	})
	client := &fakeLlmClient{responseText: string(responseBody)}

	baselineValues := map[string]string{"POWER": "OFF"}
	coder, err := WriteCoder(context.Background(), client, "Polytron", "PAC-09HDN", states, definitions, payload, baselineValues)
	if err != nil {
		t.Fatalf("WriteCoder() error = %v, want nil", err)
	}
	if coder.EncoderSource == "" || coder.DecoderSource == "" || coder.SummaryReadme == "" || coder.DetailReadme == "" {
		t.Fatalf("WriteCoder() = %+v, want all four fields populated", coder)
	}

	if client.lastRequest.ResponseSchema == nil {
		t.Fatalf("GenerateText() request had no ResponseSchema — structured output was not requested")
	}
	if client.lastRequest.MaxOutputTokens <= 0 {
		t.Fatalf("GenerateText() request had MaxOutputTokens = %d, want > 0", client.lastRequest.MaxOutputTokens)
	}

	if !strings.Contains(client.lastRequest.Prompt, "POWER=OFF") {
		t.Fatalf("prompt = %q, want it to state the baseline value POWER=OFF explicitly", client.lastRequest.Prompt)
	}
}

func TestWriteCoderPropagatesLlmError(t *testing.T) {
	client := &fakeLlmClient{err: context.DeadlineExceeded}
	_, err := WriteCoder(context.Background(), client, "Polytron", "PAC-09HDN", nil, nil, applicationinfraredanalysis.AnalysisPayload{}, nil)
	if err == nil {
		t.Fatal("WriteCoder() error = nil, want propagated error")
	}
}

func TestWriteCoderErrorsOnMalformedResponse(t *testing.T) {
	client := &fakeLlmClient{responseText: "not json"}
	_, err := WriteCoder(context.Background(), client, "Polytron", "PAC-09HDN", nil, nil, applicationinfraredanalysis.AnalysisPayload{}, nil)
	if err == nil {
		t.Fatal("WriteCoder() error = nil, want a malformed-response error")
	}
}
