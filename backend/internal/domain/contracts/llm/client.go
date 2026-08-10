package domaincontractsllm

import (
	"context"
	"encoding/json"
)

type Client interface {
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
