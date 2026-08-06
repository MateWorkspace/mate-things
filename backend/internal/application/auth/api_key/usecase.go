package applicationauthapikey

import (
	"context"
	"time"

	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesauth "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/auth"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
)

type usecase struct {
	apiKey    domainusecasesrepocache.ApiKey
	user      domainusecasesrepocache.User
	role      domainusecasesrepocache.Role
	generator domaincontractsutility.ApiKey
	logger    domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	apiKey domainusecasesrepocache.ApiKey,
	user domainusecasesrepocache.User,
	role domainusecasesrepocache.Role,
	generator domaincontractsutility.ApiKey,
	logger domaincontractslogger.Leveled,
) domainusecasesauth.ApiKey {
	return &usecase{
		apiKey:    apiKey,
		user:      user,
		role:      role,
		generator: generator,
		logger:    logger,
	}
}

func (u *usecase) Authenticate(ctx context.Context, rawKey string) (*domainmodels.TokenClaimsAccess, error) {
	const tag = "auth/api_key/Authenticate"

	hash := u.generator.Hash(rawKey)

	key, err := u.apiKey.ReadByKeyHash(ctx, hash)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read api key", domainmodels.LoggerMeta{"err": err})
		return nil, err
	}
	if key == nil || key.RevokedAt != nil || (key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now())) {
		return nil, domainmodels.NewError("invalid api key", domainmodels.ErrTypeUnauthorized, nil)
	}

	user, err := u.user.ReadById(ctx, key.UserId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read api key user", domainmodels.LoggerMeta{
			"err":     err,
			"user_id": key.UserId,
		})
		return nil, err
	}

	role, err := u.role.ReadById(ctx, user.RoleId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read api key user role", domainmodels.LoggerMeta{
			"err":     err,
			"user_id": user.Id,
			"role_id": user.RoleId,
		})
		return nil, err
	}

	permissions, err := u.user.ReadPermissions(ctx, user.Id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read api key user permissions", domainmodels.LoggerMeta{
			"err":     err,
			"user_id": user.Id,
		})
		return nil, err
	}

	permissionNames := make([]string, len(permissions))
	for i, permission := range permissions {
		permissionNames[i] = permission.Name
	}

	return &domainmodels.TokenClaimsAccess{
		UserId:      user.Id,
		Name:        user.Name,
		Username:    user.Username,
		Role:        role.Name,
		Permissions: permissionNames,
	}, nil
}
