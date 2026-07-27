package applicationprofilesecurity

import (
	"context"

	applicationshared "github.com/MateWorkspace/mate-things/backend/internal/application/shared"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesprofile "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/profile"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
)

type usecase struct {
	user     domainusecasesrepocache.User
	password domaincontractsutility.Password
	logger   domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	user domainusecasesrepocache.User,
	password domaincontractsutility.Password,
	logger domaincontractslogger.Leveled,
) domainusecasesprofile.Security {
	return &usecase{
		user:     user,
		password: password,
		logger:   logger,
	}
}

func (u *usecase) ChangePassword(ctx context.Context, request domainusecasesprofile.ChangePasswordRequest) error {
	const tag = "profile/security/ChangePassword"

	user, err := u.user.ReadById(ctx, request.UserId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read user", domainmodels.LoggerMeta{
			"err":     err,
			"user_id": request.UserId,
		})
		return err
	}
	if err := u.password.Compare(user.PasswordHash, request.CurrentPassword); err != nil {
		u.logger.Error(ctx, tag, "failed to compare current password", domainmodels.LoggerMeta{
			"err":     err,
			"user_id": request.UserId,
		})
		return err
	}

	newPassword, err := applicationshared.RequiredPassword(request.NewPassword, "new_password")
	if err != nil {
		return err
	}

	passwordHash, err := u.password.Hash(newPassword)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to hash new password", domainmodels.LoggerMeta{
			"err":     err,
			"user_id": request.UserId,
		})
		return err
	}

	if err := u.user.UpdateById(ctx, request.UserId, nil, nil, nil, nil, &passwordHash, nil, request.UpdatedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to update user password", domainmodels.LoggerMeta{
			"err":        err,
			"user_id":    request.UserId,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	return nil
}
