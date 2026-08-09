package infrastructurellmopenai

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
			"id": "chatcmpl_test",
			"object": "chat.completion",
			"model": "gpt-test",
			"choices": [{"index": 0, "finish_reason": "stop", "message": {"role": "assistant", "content": "function encode(state) { return []; }"}}],
			"usage": {"prompt_tokens": 100, "completion_tokens": 40, "total_tokens": 140}
		}`))
	}))
	defer server.Close()

	baseURL := server.URL
	client := NewClient("test-api-key", &baseURL, "gpt-test")

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
	if result.Usage.InputTokens != 100 || result.Usage.OutputTokens != 40 {
		t.Fatalf("GenerateText() Usage = %+v, want {100 40}", result.Usage)
	}
}

func TestGenerateTextStructuredSendsResponseFormat(t *testing.T) {
	var capturedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": "chatcmpl_test",
			"object": "chat.completion",
			"model": "gpt-test",
			"choices": [{"index": 0, "finish_reason": "stop", "message": {"role": "assistant", "content": "{\"description\":\"set to cool\"}"}}],
			"usage": {"prompt_tokens": 60, "completion_tokens": 15, "total_tokens": 75}
		}`))
	}))
	defer server.Close()

	baseURL := server.URL
	client := NewClient("test-api-key", &baseURL, "gpt-test")

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

	responseFormat, ok := capturedBody["response_format"].(map[string]any)
	if !ok || responseFormat["type"] != "json_schema" {
		t.Fatalf("response_format = %+v, want {type: json_schema, ...}", responseFormat)
	}
}
