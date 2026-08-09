package infrastructurellmopenai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
)

type openAIClient struct {
	sdk   openai.Client
	model string
}

// NewClient constructs an OpenAI-backed llm.Client. Called per-call by the
// ClientFactory (internal/infrastructure/llm) — never held as a singleton.
func NewClient(apiKey string, baseURL *string, model string) domaincontractsllm.Client {
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != nil {
		opts = append(opts, option.WithBaseURL(*baseURL))
	}
	return &openAIClient{
		sdk:   openai.NewClient(opts...),
		model: model,
	}
}

func (c *openAIClient) GenerateText(ctx context.Context, req domaincontractsllm.GenerateTextRequest) (domaincontractsllm.GenerateTextResult, error) {
	if req.MaxOutputTokens <= 0 {
		return domaincontractsllm.GenerateTextResult{}, fmt.Errorf("MaxOutputTokens must be greater than zero, got %d", req.MaxOutputTokens)
	}

	params := openai.ChatCompletionNewParams{
		Model: c.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(req.Prompt),
		},
		MaxTokens: openai.Int(int64(req.MaxOutputTokens)),
	}
	if req.System != "" {
		params.Messages = append([]openai.ChatCompletionMessageParamUnion{openai.SystemMessage(req.System)}, params.Messages...)
	}
	if len(req.ResponseSchema) > 0 {
		var schema map[string]any
		if err := json.Unmarshal(req.ResponseSchema, &schema); err != nil {
			return domaincontractsllm.GenerateTextResult{}, fmt.Errorf("invalid response schema: %w", err)
		}
		params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:   "response",
					Schema: schema,
					Strict: openai.Bool(true),
				},
			},
		}
	}

	completion, err := c.sdk.Chat.Completions.New(ctx, params)
	if err != nil {
		return domaincontractsllm.GenerateTextResult{}, fmt.Errorf("openai generation failed: %w", err)
	}

	var text string
	if len(completion.Choices) > 0 {
		text = completion.Choices[0].Message.Content
	}

	return domaincontractsllm.GenerateTextResult{
		Text: text,
		Usage: domaincontractsllm.Usage{
			InputTokens:  int32(completion.Usage.PromptTokens),
			OutputTokens: int32(completion.Usage.CompletionTokens),
		},
	}, nil
}
