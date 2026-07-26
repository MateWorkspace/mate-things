package applicationauthsession

import (
	"context"

	domaincontractslogger "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/logger"
	domaincontractsutility "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	domainusecasesauth "github.com/ABA-Developer/nusapala-things/backend/internal/domain/usecases/auth"
	domainusecasesrepocache "github.com/ABA-Developer/nusapala-things/backend/internal/domain/usecases/repocache"
)

type usecase struct {
	user     domainusecasesrepocache.User
	role     domainusecasesrepocache.Role
	password domaincontractsutility.Password
	token    domaincontractsutility.Token
	logger   domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	user domainusecasesrepocache.User,
	role domainusecasesrepocache.Role,
	password domaincontractsutility.Password,
	token domaincontractsutility.Token,
	logger domaincontractslogger.Leveled,
) domainusecasesauth.Session {
	return &usecase{
		user:     user,
		role:     role,
		password: password,
		token:    token,
		logger:   logger,
	}
}

func (u *usecase) Login(ctx context.Context, request domainusecasesauth.LoginRequest) (domainusecasesauth.LoginResult, error) {
	const tag = "auth/session/Login"

	user, err := u.user.ReadByUsername(ctx, request.Username)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read user", domainmodels.LoggerMeta{
			"err": err,
		})
		return domainusecasesauth.LoginResult{}, err
	}
	if err := u.password.Compare(user.PasswordHash, request.Password); err != nil {
		u.logger.Error(ctx, tag, "failed to compare user password", domainmodels.LoggerMeta{
			"err":     err,
			"user_id": user.Id,
		})
		return domainusecasesauth.LoginResult{}, err
	}

	return u.buildLoginResult(ctx, *user, tag)
}

func (u *usecase) Refresh(
	ctx context.Context,
	request domainusecasesauth.RefreshRequest,
) (domainusecasesauth.LoginResult, error) {
	const tag = "auth/session/Refresh"

	claims, err := u.token.ValidateRefresh(request.RefreshToken)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to validate refresh token", domainmodels.LoggerMeta{
			"err": err,
		})
		return domainusecasesauth.LoginResult{}, err
	}

	user, err := u.user.ReadById(ctx, claims.UserId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read user", domainmodels.LoggerMeta{
			"err":     err,
			"user_id": claims.UserId,
		})
		return domainusecasesauth.LoginResult{}, err
	}

	return u.buildLoginResult(ctx, *user, tag)
}

func (u *usecase) buildLoginResult(ctx context.Context, user domainmodels.User, tag string) (domainusecasesauth.LoginResult, error) {
	role, err := u.role.ReadById(ctx, user.RoleId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read user role", domainmodels.LoggerMeta{
			"err":     err,
			"user_id": user.Id,
			"role_id": user.RoleId,
		})
		return domainusecasesauth.LoginResult{}, err
	}

	permissions, err := u.user.ReadPermissions(ctx, user.Id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read user permissions", domainmodels.LoggerMeta{
			"err":     err,
			"user_id": user.Id,
		})
		return domainusecasesauth.LoginResult{}, err
	}

	permissionNames := make([]string, len(permissions))
	for i, permission := range permissions {
		permissionNames[i] = permission.Name
	}

	accessToken, err := u.token.GenerateAccess(&domainmodels.TokenClaimsAccess{
		UserId:      user.Id,
		Name:        user.Name,
		Username:    user.Username,
		Role:        role.Name,
		Permissions: permissionNames,
	})
	if err != nil {
		u.logger.Error(ctx, tag, "failed to generate access token", domainmodels.LoggerMeta{
			"err":     err,
			"user_id": user.Id,
			"role_id": user.RoleId,
		})
		return domainusecasesauth.LoginResult{}, err
	}

	refreshToken, err := u.token.GenerateRefresh(&domainmodels.TokenClaimsRefresh{
		UserId: user.Id,
	})
	if err != nil {
		u.logger.Error(ctx, tag, "failed to generate refresh token", domainmodels.LoggerMeta{
			"err":     err,
			"user_id": user.Id,
		})
		return domainusecasesauth.LoginResult{}, err
	}

	return domainusecasesauth.LoginResult{
		User:         user,
		Role:         *role,
		Permissions:  permissions,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
