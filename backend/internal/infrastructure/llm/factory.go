package infrastructurellm

import (
	"context"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructurellmclaude "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/llm/claude"
	infrastructurellmopenai "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/llm/openai"
)

type newClientFunc func(apiKey string, baseURL *string, model string) domaincontractsllm.Client

// ClientFactory resolves the current LlmConfig and builds the matching
// adapter fresh on every call — it deliberately holds no client instance
// between calls, so a config change takes effect on the very next one.
type ClientFactory struct {
	repository domaincontractsrepository.LlmConfig
	encryptor  domaincontractsutility.Encryptor

	newClaudeClient newClientFunc
	newOpenAIClient newClientFunc
}

func NewClientFactory(repository domaincontractsrepository.LlmConfig, encryptor domaincontractsutility.Encryptor) *ClientFactory {
	return &ClientFactory{
		repository:      repository,
		encryptor:       encryptor,
		newClaudeClient: infrastructurellmclaude.NewClient,
		newOpenAIClient: infrastructurellmopenai.NewClient,
	}
}

func (f *ClientFactory) Current(ctx context.Context) (domaincontractsllm.Client, error) {
	config, err := f.repository.Get(ctx)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, domainmodels.ErrTypeLlmConfigNotConfigured
	}

	apiKey, err := f.encryptor.Decrypt(config.ApiKeyEncrypted)
	if err != nil {
		return nil, err
	}

	switch config.Provider {
	case domainmodels.LlmProviderClaude:
		return f.newClaudeClient(apiKey, config.BaseURL, config.Model), nil
	case domainmodels.LlmProviderOpenAI:
		return f.newOpenAIClient(apiKey, config.BaseURL, config.Model), nil
	default:
		return nil, domainmodels.NewError("llm_config has an unrecognized provider", domainmodels.ErrTypeFailure, nil)
	}
}
