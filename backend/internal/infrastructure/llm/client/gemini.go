package infrastructurellmclient

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/genai"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
)

type geminiClient struct {
	apiKey  string
	baseURL *string
	model   string
}

func NewGeminiClient(apiKey string, baseURL *string, model string) domaincontractsllm.Client {
	return &geminiClient{apiKey: apiKey, baseURL: baseURL, model: model}
}

func (c *geminiClient) GenerateText(ctx context.Context, req domaincontractsllm.GenerateTextRequest) (domaincontractsllm.GenerateTextResult, error) {
	if req.MaxOutputTokens <= 0 {
		return domaincontractsllm.GenerateTextResult{}, fmt.Errorf("MaxOutputTokens must be greater than zero, got %d", req.MaxOutputTokens)
	}

	clientConfig := &genai.ClientConfig{APIKey: c.apiKey}
	if c.baseURL != nil {
		clientConfig.HTTPOptions = genai.HTTPOptions{BaseURL: *c.baseURL}
	}
	sdk, err := genai.NewClient(ctx, clientConfig)
	if err != nil {
		return domaincontractsllm.GenerateTextResult{}, fmt.Errorf("gemini client init failed: %w", err)
	}

	config := &genai.GenerateContentConfig{MaxOutputTokens: req.MaxOutputTokens}
	if req.System != "" {
		config.SystemInstruction = genai.NewContentFromText(req.System, genai.RoleUser)
	}
	if len(req.ResponseSchema) > 0 {
		var schema genai.Schema
		if err := json.Unmarshal(req.ResponseSchema, &schema); err != nil {
			return domaincontractsllm.GenerateTextResult{}, fmt.Errorf("invalid response schema: %w", err)
		}
		config.ResponseMIMEType = "application/json"
		config.ResponseSchema = &schema
	}

	response, err := sdk.Models.GenerateContent(
		ctx,
		c.model,
		[]*genai.Content{genai.NewContentFromText(req.Prompt, genai.RoleUser)},
		config,
	)
	if err != nil {
		return domaincontractsllm.GenerateTextResult{}, fmt.Errorf("gemini generation failed: %w", err)
	}

	result := domaincontractsllm.GenerateTextResult{Text: response.Text()}
	if response.UsageMetadata != nil {
		result.Usage = domaincontractsllm.Usage{
			InputTokens:  response.UsageMetadata.PromptTokenCount,
			OutputTokens: response.UsageMetadata.CandidatesTokenCount,
		}
	}
	return result, nil
}
