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
}

func NewClientFactory(repository domaincontractsrepository.LlmConfig, encryptor domaincontractsutility.Encryptor) domaincontractsllm.ClientFactory {
	return &ClientFactory{
		repository:      repository,
		encryptor:       encryptor,
		newClaudeClient: infrastructurellmclient.NewClaudeClient,
		newOpenAIClient: infrastructurellmclient.NewOpenAIClient,
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
