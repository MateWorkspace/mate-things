package domaincontractsllm

import (
	"context"
	"encoding/json"
)

// Client is a provider-agnostic interface for making a single LLM generation
// call. Implementations live in internal/infrastructure/llm/<provider>/ and
// are constructed per-call by the ClientFactory (internal/infrastructure/llm),
// never held as a process-lifetime singleton.
type Client interface {
	// GenerateText makes one generation call. When req.ResponseSchema is nil,
	// the result is free-form text (e.g. generated code). When it is set, the
	// result's Text field is a JSON string conforming to that schema.
	GenerateText(ctx context.Context, req GenerateTextRequest) (GenerateTextResult, error)
}

type GenerateTextRequest struct {
	System          string
	Prompt          string
	MaxOutputTokens int32
	ResponseSchema  json.RawMessage
}

type GenerateTextResult struct {
	Text  string
	Usage Usage
}

type Usage struct {
	InputTokens  int32
	OutputTokens int32
}
