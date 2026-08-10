package applicationinfraredcodergeneration

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

func TestWriteTestCasesParsesStructuredResponse(t *testing.T) {
	powerId := uuid.New()
	states := []domainmodels.InfraredState{{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}
	definitions := []domainmodels.InfraredStateDeviceDefinition{{InfraredStateId: powerId, Options: []string{"ON", "OFF"}}}
	coder := domainmodels.InfraredStateCoder{SummaryReadme: "summary", DetailReadme: "detail"}

	responseBody, _ := json.Marshal([]map[string]interface{}{
		{"description": "Turn the unit on.", "states": map[string]string{"POWER": "ON"}},
		{"description": "Turn the unit off.", "states": map[string]string{"POWER": "OFF"}},
	})
	client := &fakeLlmClient{responseText: string(responseBody)}

	plans, err := WriteTestCases(context.Background(), client, "Polytron", "PAC-09HDN", coder, states, definitions)
	if err != nil {
		t.Fatalf("WriteTestCases() error = %v, want nil", err)
	}
	if len(plans) != 2 {
		t.Fatalf("WriteTestCases() returned %d plans, want 2", len(plans))
	}
	if plans[0].Description == "" {
		t.Fatal("WriteTestCases() plan has empty description")
	}
	if plans[0].States[powerId] != "ON" {
		t.Fatalf("WriteTestCases() plan[0].States[POWER] = %q, want ON", plans[0].States[powerId])
	}

	if client.lastRequest.ResponseSchema == nil {
		t.Fatal("GenerateText() request had no ResponseSchema — structured output was not requested")
	}
}

func TestWriteTestCasesPropagatesLlmError(t *testing.T) {
	client := &fakeLlmClient{err: context.DeadlineExceeded}
	_, err := WriteTestCases(context.Background(), client, "Polytron", "PAC-09HDN", domainmodels.InfraredStateCoder{}, nil, nil)
	if err == nil {
		t.Fatal("WriteTestCases() error = nil, want propagated error")
	}
}

func TestWriteTestCasesErrorsOnUnknownStateName(t *testing.T) {
	states := []domainmodels.InfraredState{{Id: uuid.New(), Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}
	responseBody, _ := json.Marshal([]map[string]interface{}{
		{"description": "bad", "states": map[string]string{"NONEXISTENT": "X"}},
	})
	client := &fakeLlmClient{responseText: string(responseBody)}

	_, err := WriteTestCases(context.Background(), client, "Polytron", "PAC-09HDN", domainmodels.InfraredStateCoder{}, states, nil)
	if err == nil {
		t.Fatal("WriteTestCases() error = nil, want an error for a state name the LLM invented")
	}
}

func TestWriteTestCasesPromptEnumeratesStateNamesAndValues(t *testing.T) {
	powerId := uuid.New()
	states := []domainmodels.InfraredState{{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}
	definitions := []domainmodels.InfraredStateDeviceDefinition{{InfraredStateId: powerId, Options: []string{"ON", "OFF"}}}
	coder := domainmodels.InfraredStateCoder{SummaryReadme: "summary", DetailReadme: "detail"}

	responseBody, _ := json.Marshal([]map[string]interface{}{
		{"description": "Turn the unit on.", "states": map[string]string{"POWER": "ON"}},
	})
	client := &fakeLlmClient{responseText: string(responseBody)}

	if _, err := WriteTestCases(context.Background(), client, "Polytron", "PAC-09HDN", coder, states, definitions); err != nil {
		t.Fatalf("WriteTestCases() error = %v, want nil", err)
	}

	if !strings.Contains(client.lastRequest.Prompt, "POWER") {
		t.Fatalf("prompt = %q, want it to mention state name POWER", client.lastRequest.Prompt)
	}
	if !strings.Contains(client.lastRequest.Prompt, "ON") || !strings.Contains(client.lastRequest.Prompt, "OFF") {
		t.Fatalf("prompt = %q, want it to mention legal values ON/OFF", client.lastRequest.Prompt)
	}
}
