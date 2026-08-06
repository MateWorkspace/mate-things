package applicationadminapikeymanagement

import (
	"context"

	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesadmin "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/admin"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
)

type usecase struct {
	apiKey    domainusecasesrepocache.ApiKey
	generator domaincontractsutility.ApiKey
	logger    domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	apiKey domainusecasesrepocache.ApiKey,
	generator domaincontractsutility.ApiKey,
	logger domaincontractslogger.Leveled,
) domainusecasesadmin.ApiKeyManagement {
	return &usecase{
		apiKey:    apiKey,
		generator: generator,
		logger:    logger,
	}
}

func (u *usecase) Create(ctx context.Context, request domainusecasesadmin.CreateApiKeyRequest) (string, error) {
	const tag = "admin/api_key_management/Create"

	raw, hash, lastFour, err := u.generator.Generate()
	if err != nil {
		u.logger.Error(ctx, tag, "failed to generate api key", domainmodels.LoggerMeta{
			"err":     err,
			"user_id": request.UserId,
		})
		return "", err
	}

	if _, err := u.apiKey.Create(ctx, request.UserId, hash, lastFour, request.ExpiresAt, request.CreatedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to create api key", domainmodels.LoggerMeta{
			"err":     err,
			"user_id": request.UserId,
		})
		return "", err
	}

	return raw, nil
}

func (u *usecase) Regenerate(ctx context.Context, request domainusecasesadmin.RegenerateApiKeyRequest) (string, error) {
	const tag = "admin/api_key_management/Regenerate"

	raw, hash, lastFour, err := u.generator.Generate()
	if err != nil {
		u.logger.Error(ctx, tag, "failed to generate api key", domainmodels.LoggerMeta{
			"err": err,
			"id":  request.Id,
		})
		return "", err
	}

	if err := u.apiKey.Regenerate(ctx, request.Id, hash, lastFour, request.ExpiresAt, request.UpdatedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to regenerate api key", domainmodels.LoggerMeta{
			"err": err,
			"id":  request.Id,
		})
		return "", err
	}

	return raw, nil
}

func (u *usecase) Revoke(ctx context.Context, request domainusecasesadmin.RevokeApiKeyRequest) error {
	const tag = "admin/api_key_management/Revoke"

	if err := u.apiKey.Revoke(ctx, request.Id, request.UpdatedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to revoke api key", domainmodels.LoggerMeta{
			"err": err,
			"id":  request.Id,
		})
		return err
	}

	return nil
}

func (u *usecase) ReadByPagination(
	ctx context.Context,
	request domainusecasesadmin.ReadApiKeysByPaginationRequest,
) ([]domainmodels.ApiKeyWithUser, int, error) {
	const tag = "admin/api_key_management/ReadByPagination"

	apiKeys, total, err := u.apiKey.ReadByPagination(ctx, request.Page, request.Limit, request.Search, request.Status)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read api keys", domainmodels.LoggerMeta{
			"err":    err,
			"page":   request.Page,
			"limit":  request.Limit,
			"status": request.Status,
		})
		return nil, 0, err
	}

	return apiKeys, total, nil
}

func (u *usecase) DeleteById(ctx context.Context, request domainusecasesadmin.DeleteApiKeyRequest) error {
	const tag = "admin/api_key_management/DeleteById"

	if err := u.apiKey.DeleteById(ctx, request.Id); err != nil {
		u.logger.Error(ctx, tag, "failed to delete api key", domainmodels.LoggerMeta{
			"err": err,
			"id":  request.Id,
		})
		return err
	}

	return nil
}
