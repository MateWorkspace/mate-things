package applicationprofileme

import (
	"context"

	domaincontractslogger "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/logger"
	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	domainusecasesprofile "github.com/ABA-Developer/nusapala-things/backend/internal/domain/usecases/profile"
	domainusecasesrepocache "github.com/ABA-Developer/nusapala-things/backend/internal/domain/usecases/repocache"
)

type usecase struct {
	user   domainusecasesrepocache.User
	logger domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	user domainusecasesrepocache.User,
	logger domaincontractslogger.Leveled,
) domainusecasesprofile.Me {
	return &usecase{
		user:   user,
		logger: logger,
	}
}

func (u *usecase) GetProfile(
	ctx context.Context,
	request domainusecasesprofile.GetProfileRequest,
) (*domainmodels.User, error) {
	const tag = "profile/me/GetProfile"

	user, err := u.user.ReadById(ctx, request.UserId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read profile", domainmodels.LoggerMeta{
			"err":     err,
			"user_id": request.UserId,
		})
		return nil, err
	}

	return user, nil
}

func (u *usecase) GetPermissions(
	ctx context.Context,
	request domainusecasesprofile.GetProfilePermissionsRequest,
) ([]domainmodels.Permission, error) {
	const tag = "profile/me/GetPermissions"

	permissions, err := u.user.ReadPermissions(ctx, request.UserId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read profile permissions", domainmodels.LoggerMeta{
			"err":     err,
			"user_id": request.UserId,
		})
		return nil, err
	}

	return permissions, nil
}
