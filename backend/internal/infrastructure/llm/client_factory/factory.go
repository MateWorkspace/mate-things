package infrastructurellmclientfactory

import (
	"context"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructurellmclient "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/llm/client"
)

type newClientFunc func(apiKey string, baseURL *string, model string) domaincontractsllm.Client

type ClientFactory struct {
	repository domaincontractsrepository.LlmConfig
	encryptor  domaincontractsutility.Encryptor

	newClaudeClient newClientFunc
	newOpenAIClient newClientFunc
	newGeminiClient newClientFunc
}

func NewClientFactory(repository domaincontractsrepository.LlmConfig, encryptor domaincontractsutility.Encryptor) domaincontractsllm.ClientFactory {
	return &ClientFactory{
		repository:      repository,
		encryptor:       encryptor,
		newClaudeClient: infrastructurellmclient.NewClaudeClient,
		newOpenAIClient: infrastructurellmclient.NewOpenAIClient,
		newGeminiClient: infrastructurellmclient.NewGeminiClient,
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

	return f.FromCredentials(config.Provider, apiKey, config.BaseURL, config.Model)
}

func (f *ClientFactory) FromCredentials(provider domainmodels.LlmProvider, apiKey string, baseURL *string, model string) (domaincontractsllm.Client, error) {
	switch provider {
	case domainmodels.LlmProviderClaude:
		return f.newClaudeClient(apiKey, baseURL, model), nil
	case domainmodels.LlmProviderOpenAI:
		return f.newOpenAIClient(apiKey, baseURL, model), nil
	case domainmodels.LlmProviderGemini:
		return f.newGeminiClient(apiKey, baseURL, model), nil
	default:
		return nil, domainmodels.NewError("llm_config has an unrecognized provider", domainmodels.ErrTypeFailure, nil)
	}
}
