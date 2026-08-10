package applicationinfraredcodergeneration

import (
	"context"
	"encoding/json"
	"testing"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

func TestWriteRetryCasesParsesStructuredResponse(t *testing.T) {
	powerId := uuid.New()
	states := []domainmodels.InfraredState{{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}
	coder := domainmodels.InfraredStateCoder{SummaryReadme: "summary", DetailReadme: "detail"}
	failedState := map[uuid.UUID]string{powerId: "OFF"}

	responseBody, _ := json.Marshal([]map[string]interface{}{
		{"description": "Re-record POWER OFF.", "states": map[string]string{"POWER": "OFF"}},
	})
	client := &fakeLlmClient{responseText: string(responseBody)}

	plans, err := WriteRetryCases(context.Background(), client, "Polytron", "PAC-09HDN", coder, failedState, states)
	if err != nil {
		t.Fatalf("WriteRetryCases() error = %v, want nil", err)
	}
	if len(plans) != 1 {
		t.Fatalf("WriteRetryCases() returned %d plans, want 1", len(plans))
	}
	if plans[0].States[powerId] != "OFF" {
		t.Fatalf("WriteRetryCases() plan[0].States[POWER] = %q, want OFF", plans[0].States[powerId])
	}
}

func TestWriteRetryCasesPropagatesLlmError(t *testing.T) {
	client := &fakeLlmClient{err: context.DeadlineExceeded}
	_, err := WriteRetryCases(context.Background(), client, "Polytron", "PAC-09HDN", domainmodels.InfraredStateCoder{}, nil, nil)
	if err == nil {
		t.Fatal("WriteRetryCases() error = nil, want propagated error")
	}
}
