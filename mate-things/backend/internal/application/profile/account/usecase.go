package applicationprofileaccount

import (
	"context"

	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesprofile "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/profile"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
)

type usecase struct {
	user   domainusecasesrepocache.User
	logger domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	user domainusecasesrepocache.User,
	logger domaincontractslogger.Leveled,
) domainusecasesprofile.Account {
	return &usecase{
		user:   user,
		logger: logger,
	}
}

func (u *usecase) UpdateProfile(ctx context.Context, request domainusecasesprofile.UpdateProfileRequest) error {
	const tag = "profile/account/UpdateProfile"

	if err := u.user.UpdateById(
		ctx,
		request.UserId,
		nil,
		request.Name,
		request.Bio,
		request.Username,
		nil,
		nil,
		request.UpdatedBy,
	); err != nil {
		u.logger.Error(ctx, tag, "failed to update profile", domainmodels.LoggerMeta{
			"err":        err,
			"user_id":    request.UserId,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	return nil
}
