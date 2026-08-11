package applicationadminllmconfigmanagement

import (
	"context"

	applicationshared "github.com/MateWorkspace/mate-things/backend/internal/application/shared"
	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesadmin "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/admin"
)

type usecase struct {
	repository    domaincontractsrepository.LlmConfig
	encryptor     domaincontractsutility.Encryptor
	clientFactory domaincontractsllm.ClientFactory
	logger        domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	repository domaincontractsrepository.LlmConfig,
	encryptor domaincontractsutility.Encryptor,
	clientFactory domaincontractsllm.ClientFactory,
	logger domaincontractslogger.Leveled,
) domainusecasesadmin.LlmConfigManagement {
	return &usecase{repository: repository, encryptor: encryptor, clientFactory: clientFactory, logger: logger}
}

func (u *usecase) Get(ctx context.Context) (*domainmodels.LlmConfig, error) {
	const tag = "admin/llm_config_management/Get"

	config, err := u.repository.Get(ctx)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read llm config", domainmodels.LoggerMeta{"err": err})
		return nil, err
	}
	return config, nil
}

func (u *usecase) Update(ctx context.Context, request domainusecasesadmin.UpdateLlmConfigRequest) error {
	const tag = "admin/llm_config_management/Update"

	provider, err := applicationshared.RequiredLlmProvider(request.Provider, "provider")
	if err != nil {
		return err
	}
	model, err := applicationshared.RequiredLlmModel(request.Model, "model")
	if err != nil {
		return err
	}
	baseURL, err := applicationshared.OptionalLlmBaseURL(request.BaseURL, "base_url")
	if err != nil {
		return err
	}

	var encryptedApiKey []byte
	if apiKey := applicationshared.OptionalLlmApiKey(request.ApiKey); apiKey != nil {
		encryptedApiKey, err = u.encryptor.Encrypt(*apiKey)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to encrypt api key", domainmodels.LoggerMeta{"err": err})
			return err
		}
	} else {
		existing, err := u.requireExistingConfig(ctx, tag)
		if err != nil {
			return err
		}
		encryptedApiKey = existing.ApiKeyEncrypted
	}

	if _, err := u.repository.Upsert(ctx, provider, model, encryptedApiKey, baseURL, request.UpdatedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to upsert llm config", domainmodels.LoggerMeta{"err": err})
		return err
	}
	return nil
}

func (u *usecase) TestConnection(ctx context.Context) (domainmodels.LlmClientStatus, error) {
	const tag = "admin/llm_config_management/TestConnection"

	client, err := u.clientFactory.Current(ctx)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to resolve llm client", domainmodels.LoggerMeta{"err": err})
		return "", err
	}

	return u.ping(ctx, tag, client)
}

func (u *usecase) TestConnectionWithConfig(ctx context.Context, request domainusecasesadmin.TestLlmConnectionWithConfigRequest) (domainmodels.LlmClientStatus, error) {
	const tag = "admin/llm_config_management/TestConnectionWithConfig"

	provider, err := applicationshared.RequiredLlmProvider(request.Provider, "provider")
	if err != nil {
		return "", err
	}
	model, err := applicationshared.RequiredLlmModel(request.Model, "model")
	if err != nil {
		return "", err
	}
	baseURL, err := applicationshared.OptionalLlmBaseURL(request.BaseURL, "base_url")
	if err != nil {
		return "", err
	}

	var apiKey string
	if provided := applicationshared.OptionalLlmApiKey(request.ApiKey); provided != nil {
		apiKey = *provided
	} else {
		existing, err := u.requireExistingConfig(ctx, tag)
		if err != nil {
			return "", err
		}
		apiKey, err = u.encryptor.Decrypt(existing.ApiKeyEncrypted)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to decrypt existing api key", domainmodels.LoggerMeta{"err": err})
			return "", err
		}
	}

	client, err := u.clientFactory.FromCredentials(provider, apiKey, baseURL, model)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to resolve llm client", domainmodels.LoggerMeta{"err": err})
		return "", err
	}

	return u.ping(ctx, tag, client)
}

// requireExistingConfig reads the currently stored config, used whenever a
// caller omits the API key and means to fall back to what's already saved.
func (u *usecase) requireExistingConfig(ctx context.Context, tag string) (*domainmodels.LlmConfig, error) {
	existing, err := u.repository.Get(ctx)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read existing llm config", domainmodels.LoggerMeta{"err": err})
		return nil, err
	}
	if existing == nil {
		return nil, domainmodels.NewError("api_key is required", domainmodels.ErrTypeValidation, nil)
	}
	return existing, nil
}

func (u *usecase) ping(ctx context.Context, tag string, client domaincontractsllm.Client) (domainmodels.LlmClientStatus, error) {
	if _, err := client.GenerateText(ctx, domaincontractsllm.GenerateTextRequest{
		System:          "You are a connectivity test.",
		Prompt:          "Reply with the single word OK.",
		MaxOutputTokens: 16,
	}); err != nil {
		u.logger.Error(ctx, tag, "llm connectivity test failed", domainmodels.LoggerMeta{"err": err})
		return "", err
	}

	return domainmodels.LlmClientStatusConnected, nil
}
