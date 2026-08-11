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
		existing, err := u.repository.Get(ctx)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to read existing llm config", domainmodels.LoggerMeta{"err": err})
			return err
		}
		if existing == nil {
			return domainmodels.NewError("api_key is required", domainmodels.ErrTypeValidation, nil)
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
