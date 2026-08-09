package infrastructurellmclaude

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
)

func TestGenerateTextFreeForm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": "msg_test",
			"type": "message",
			"role": "assistant",
			"model": "claude-opus-5",
			"content": [{"type": "text", "text": "function encode(state) { return []; }"}],
			"stop_reason": "end_turn",
			"usage": {"input_tokens": 120, "output_tokens": 45}
		}`))
	}))
	defer server.Close()

	baseURL := server.URL
	client := NewClient("test-api-key", &baseURL, "claude-opus-5")

	result, err := client.GenerateText(context.Background(), domaincontractsllm.GenerateTextRequest{
		System:          "You write JavaScript IR encoders.",
		Prompt:          "Write an encoder for POWER only.",
		MaxOutputTokens: 4096,
	})
	if err != nil {
		t.Fatalf("GenerateText() error = %v, want nil", err)
	}
	if result.Text != "function encode(state) { return []; }" {
		t.Fatalf("GenerateText() Text = %q, want the generated function body", result.Text)
	}
	if result.Usage.InputTokens != 120 || result.Usage.OutputTokens != 45 {
		t.Fatalf("GenerateText() Usage = %+v, want {120 45}", result.Usage)
	}
}

func TestGenerateTextRejectsZeroMaxOutputTokens(t *testing.T) {
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
	}))
	defer server.Close()

	baseURL := server.URL
	client := NewClient("test-api-key", &baseURL, "claude-opus-5")

	_, err := client.GenerateText(context.Background(), domaincontractsllm.GenerateTextRequest{
		Prompt:          "Describe the case.",
		MaxOutputTokens: 0,
	})
	if err == nil {
		t.Fatalf("GenerateText() error = nil, want an error for MaxOutputTokens = 0")
	}
	if requestCount != 0 {
		t.Fatalf("GenerateText() hit the server %d times, want 0", requestCount)
	}
}

func TestGenerateTextStructuredSendsOutputConfigFormat(t *testing.T) {
	var capturedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": "msg_test",
			"type": "message",
			"role": "assistant",
			"model": "claude-opus-5",
			"content": [{"type": "text", "text": "{\"description\":\"set to cool\"}"}],
			"stop_reason": "end_turn",
			"usage": {"input_tokens": 80, "output_tokens": 20}
		}`))
	}))
	defer server.Close()

	baseURL := server.URL
	client := NewClient("test-api-key", &baseURL, "claude-opus-5")

	schema := json.RawMessage(`{"type":"object","properties":{"description":{"type":"string"}},"required":["description"],"additionalProperties":false}`)
	result, err := client.GenerateText(context.Background(), domaincontractsllm.GenerateTextRequest{
		Prompt:          "Describe the case.",
		MaxOutputTokens: 256,
		ResponseSchema:  schema,
	})
	if err != nil {
		t.Fatalf("GenerateText() error = %v, want nil", err)
	}
	if result.Text != `{"description":"set to cool"}` {
		t.Fatalf("GenerateText() Text = %q, want the raw JSON string", result.Text)
	}

	outputConfig, ok := capturedBody["output_config"].(map[string]any)
	if !ok {
		t.Fatalf("request body has no output_config field; body = %+v", capturedBody)
	}
	format, ok := outputConfig["format"].(map[string]any)
	if !ok || format["type"] != "json_schema" {
		t.Fatalf("output_config.format = %+v, want {type: json_schema, ...}", format)
	}
}
