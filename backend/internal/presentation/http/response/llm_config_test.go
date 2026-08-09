package presentationhttpresponse

import (
	"encoding/json"
	"strings"
	"testing"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

func TestLlmConfigNilProducesUnsetResponse(t *testing.T) {
	got := LlmConfig(nil)

	want := LlmConfigResponse{ApiKeySet: false}
	if got != want {
		t.Fatalf("LlmConfig(nil) = %+v, want %+v", got, want)
	}
}

func TestLlmConfigJSONNeverContainsApiKeyMaterialOrField(t *testing.T) {
	model := &domainmodels.LlmConfig{
		Provider:        domainmodels.LlmProviderClaude,
		Model:           "claude-opus-5",
		ApiKeyEncrypted: []byte("sk-ant-super-secret-value"),
	}

	body, err := json.Marshal(LlmConfig(model))
	if err != nil {
		t.Fatalf("json.Marshal() error = %v, want nil", err)
	}

	bodyLower := strings.ToLower(string(body))
	if strings.Contains(string(body), "sk-ant-super-secret-value") {
		t.Fatalf("LlmConfig JSON output contains the plaintext API key: %s", body)
	}
	// Checks for the literal JSON key "api_key" (quoted), not the bare
	// substring — "api_key_set" is an intentional, non-secret field.
	if strings.Contains(bodyLower, `"api_key"`) {
		t.Fatalf("LlmConfig JSON output contains an api_key field: %s", body)
	}
}
