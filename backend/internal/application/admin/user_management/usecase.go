package applicationadminusermanagement

import (
	"context"

	applicationshared "github.com/MateWorkspace/mate-things/backend/internal/application/shared"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesadmin "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/admin"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
	"github.com/google/uuid"
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
) domainusecasesadmin.UserManagement {
	return &usecase{
		user:     user,
		password: password,
		logger:   logger,
	}
}

func (u *usecase) Create(ctx context.Context, request domainusecasesadmin.CreateUserRequest) (uuid.UUID, error) {
	const tag = "admin/user_management/Create"

	name, err := applicationshared.RequiredPersonName(request.Name, "name")
	if err != nil {
		return uuid.Nil, err
	}
	username, err := applicationshared.RequiredUsername(request.Username, "username")
	if err != nil {
		return uuid.Nil, err
	}
	password, err := applicationshared.RequiredPassword(request.Password, "password")
	if err != nil {
		return uuid.Nil, err
	}

	passwordHash, err := u.password.Hash(password)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to hash user password", domainmodels.LoggerMeta{
			"err":        err,
			"created_by": request.CreatedBy,
		})
		return uuid.Nil, err
	}

	id, err := u.user.Create(
		ctx,
		request.RoleId,
		name,
		request.Bio,
		username,
		passwordHash,
		request.CreatedBy,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to create user", domainmodels.LoggerMeta{
			"err":        err,
			"role_id":    request.RoleId,
			"created_by": request.CreatedBy,
		})
		return uuid.Nil, err
	}

	return id, nil
}

func (u *usecase) ReadById(
	ctx context.Context,
	request domainusecasesadmin.ReadUserByIdRequest,
) (*domainmodels.User, error) {
	const tag = "admin/user_management/ReadById"

	user, err := u.user.ReadById(ctx, request.Id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read user", domainmodels.LoggerMeta{
			"err": err,
			"id":  request.Id,
		})
		return nil, err
	}

	return user, nil
}

func (u *usecase) ReadByUsername(
	ctx context.Context,
	request domainusecasesadmin.ReadUserByUsernameRequest,
) (*domainmodels.User, error) {
	const tag = "admin/user_management/ReadByUsername"

	user, err := u.user.ReadByUsername(ctx, request.Username)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read user", domainmodels.LoggerMeta{
			"err": err,
		})
		return nil, err
	}

	return user, nil
}

func (u *usecase) ReadPermissions(
	ctx context.Context,
	request domainusecasesadmin.ReadUserPermissionsRequest,
) ([]domainmodels.Permission, error) {
	const tag = "admin/user_management/ReadPermissions"

	permissions, err := u.user.ReadPermissions(ctx, request.UserId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read user permissions", domainmodels.LoggerMeta{
			"err":     err,
			"user_id": request.UserId,
		})
		return nil, err
	}

	return permissions, nil
}

func (u *usecase) ReadByPagination(
	ctx context.Context,
	request domainusecasesadmin.ReadUsersByPaginationRequest,
) ([]domainmodels.User, int, error) {
	const tag = "admin/user_management/ReadByPagination"

	users, total, err := u.user.ReadByPagination(ctx, request.Page, request.Limit, request.Search, request.RoleId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read users", domainmodels.LoggerMeta{
			"err":     err,
			"page":    request.Page,
			"limit":   request.Limit,
			"role_id": request.RoleId,
		})
		return nil, 0, err
	}

	return users, total, nil
}

func (u *usecase) UpdateById(ctx context.Context, request domainusecasesadmin.UpdateUserRequest) error {
	const tag = "admin/user_management/UpdateById"

	name, err := applicationshared.OptionalPersonName(request.Name, "name")
	if err != nil {
		return err
	}
	username, err := applicationshared.OptionalUsername(request.Username, "username")
	if err != nil {
		return err
	}

	if err := u.user.UpdateById(
		ctx,
		request.Id,
		request.RoleId,
		name,
		request.Bio,
		username,
		nil,
		nil,
		request.UpdatedBy,
	); err != nil {
		u.logger.Error(ctx, tag, "failed to update user", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	return nil
}

func (u *usecase) ResetPassword(ctx context.Context, request domainusecasesadmin.ResetUserPasswordRequest) error {
	const tag = "admin/user_management/ResetPassword"

	password, err := applicationshared.RequiredPassword(request.Password, "password")
	if err != nil {
		return err
	}

	passwordHash, err := u.password.Hash(password)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to hash user password", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	if err := u.user.UpdateById(ctx, request.Id, nil, nil, nil, nil, &passwordHash, nil, request.UpdatedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to reset user password", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	return nil
}

func (u *usecase) DeleteById(ctx context.Context, request domainusecasesadmin.DeleteUserRequest) error {
	const tag = "admin/user_management/DeleteById"

	if err := u.user.DeleteById(ctx, request.Id, request.DeletedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to delete user", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"deleted_by": request.DeletedBy,
		})
		return err
	}

	return nil
}
