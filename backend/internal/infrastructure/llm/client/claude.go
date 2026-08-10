package infrastructurellmclient

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
)

type claudeClient struct {
	sdk   anthropic.Client
	model string
}

func NewClaudeClient(apiKey string, baseURL *string, model string) domaincontractsllm.Client {
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != nil {
		opts = append(opts, option.WithBaseURL(*baseURL))
	}
	return &claudeClient{
		sdk:   anthropic.NewClient(opts...),
		model: model,
	}
}

func (c *claudeClient) GenerateText(ctx context.Context, req domaincontractsllm.GenerateTextRequest) (domaincontractsllm.GenerateTextResult, error) {
	if req.MaxOutputTokens <= 0 {
		return domaincontractsllm.GenerateTextResult{}, fmt.Errorf("MaxOutputTokens must be greater than zero, got %d", req.MaxOutputTokens)
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: int64(req.MaxOutputTokens),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(req.Prompt)),
		},
	}
	if req.System != "" {
		params.System = []anthropic.TextBlockParam{{Text: req.System}}
	}
	if len(req.ResponseSchema) > 0 {
		var schema map[string]any
		if err := json.Unmarshal(req.ResponseSchema, &schema); err != nil {
			return domaincontractsllm.GenerateTextResult{}, fmt.Errorf("invalid response schema: %w", err)
		}
		params.OutputConfig = anthropic.OutputConfigParam{
			Format: anthropic.JSONOutputFormatParam{
				Schema: schema,
			},
		}
	}

	message, err := c.sdk.Messages.New(ctx, params)
	if err != nil {
		return domaincontractsllm.GenerateTextResult{}, fmt.Errorf("claude generation failed: %w", err)
	}

	var text string
	for _, block := range message.Content {
		if textBlock, ok := block.AsAny().(anthropic.TextBlock); ok {
			text = textBlock.Text
			break
		}
	}

	return domaincontractsllm.GenerateTextResult{
		Text: text,
		Usage: domaincontractsllm.Usage{
			InputTokens:  int32(message.Usage.InputTokens),
			OutputTokens: int32(message.Usage.OutputTokens),
		},
	}, nil
}
