package applicationadminllmconfigmanagement

import (
	"context"
	"encoding/json"

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

// pingResponseSchema requests the same class of response real generation
// calls always use (structured JSON output via an object-rooted schema) -
// a plain-text ping can succeed against a provider/model/base_url
// combination that then fails every real generation call, because
// structured output support is a narrower capability than plain chat.
// Confirmed live against OpenRouter: a working chat completion still
// rejected a schema-constrained request with "schema must be type object".
const pingResponseSchema = `{
	"type": "object",
	"properties": {
		"status": {"type": "string"}
	},
	"required": ["status"],
	"additionalProperties": false
}`

type pingResponse struct {
	Status string `json:"status"`
}

// pingMaxOutputTokens has to clear a reasoning model's hidden "thinking"
// budget, not just the visible answer: confirmed live against OpenRouter's
// gpt-5-mini, a 32-token cap left finish_reason="length" with empty
// content (spent entirely on invisible reasoning tokens before any visible
// output), while 512 succeeded with room to spare (147 completion tokens,
// 128 of them reasoning). This mirrors the real generation calls
// elsewhere in this codebase, which already use several thousand for the
// same reason.
const pingMaxOutputTokens = 512

func (u *usecase) ping(ctx context.Context, tag string, client domaincontractsllm.Client) (domainmodels.LlmClientStatus, error) {
	result, err := client.GenerateText(ctx, domaincontractsllm.GenerateTextRequest{
		System:          "You are a connectivity test.",
		Prompt:          `Respond with JSON matching the given schema, setting "status" to the single word OK.`,
		MaxOutputTokens: pingMaxOutputTokens,
		ResponseSchema:  []byte(pingResponseSchema),
	})
	if err != nil {
		u.logger.Error(ctx, tag, "llm connectivity test failed", domainmodels.LoggerMeta{"err": err})
		return "", err
	}

	var parsed pingResponse
	if err := json.Unmarshal([]byte(result.Text), &parsed); err != nil {
		u.logger.Error(ctx, tag, "llm connectivity test returned malformed structured output", domainmodels.LoggerMeta{"err": err})
		return "", domainmodels.NewError("llm connectivity test returned malformed structured output", domainmodels.ErrTypeFailure, err)
	}

	return domainmodels.LlmClientStatusConnected, nil
}
