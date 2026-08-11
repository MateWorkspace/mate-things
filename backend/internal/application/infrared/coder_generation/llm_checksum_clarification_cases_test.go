package applicationinfraredcodergeneration

import (
	"context"
	"testing"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

func TestWriteChecksumClarificationCasesParsesResponse(t *testing.T) {
	powerId := uuid.New()
	states := []domainmodels.InfraredState{{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}
	coder := Coder{SummaryReadme: "summary", DetailReadme: "detail"}

	responseBody := `{"entries": [{"description": "repeat POWER=ON twice", "states": [{"name": "POWER", "value": "ON"}]}]}`
	client := &fakeLlmClient{responseText: responseBody}

	plans, err := WriteChecksumClarificationCases(context.Background(), client, "Polytron", "PAC-09HDN", coder, states)
	if err != nil {
		t.Fatalf("WriteChecksumClarificationCases() error = %v, want nil", err)
	}
	if len(plans) != 1 {
		t.Fatalf("len(plans) = %d, want 1", len(plans))
	}
	if plans[0].States[powerId] != "ON" {
		t.Fatalf("plans[0].States[powerId] = %q, want \"ON\"", plans[0].States[powerId])
	}
	if client.lastRequest.ResponseSchema == nil {
		t.Fatal("GenerateText() request had no ResponseSchema")
	}
}

func TestWriteChecksumClarificationCasesPropagatesLlmError(t *testing.T) {
	client := &fakeLlmClient{err: errPlaceholder}
	_, err := WriteChecksumClarificationCases(context.Background(), client, "Polytron", "PAC-09HDN", Coder{}, nil)
	if err == nil {
		t.Fatal("WriteChecksumClarificationCases() error = nil, want propagated error")
	}
}

func TestWriteChecksumClarificationCasesRejectsUnknownState(t *testing.T) {
	states := []domainmodels.InfraredState{{Id: uuid.New(), Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}
	client := &fakeLlmClient{responseText: `{"entries": [{"description": "x", "states": [{"name": "NOT_A_REAL_STATE", "value": "ON"}]}]}`}

	_, err := WriteChecksumClarificationCases(context.Background(), client, "Polytron", "PAC-09HDN", Coder{}, states)
	if err == nil {
		t.Fatal("WriteChecksumClarificationCases() error = nil, want error for unknown state name")
	}
}
