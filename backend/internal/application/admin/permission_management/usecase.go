package applicationadminpermissionmanagement

import (
	"context"

	applicationshared "github.com/MateWorkspace/mate-things/backend/internal/application/shared"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesadmin "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/admin"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
	"github.com/google/uuid"
)

type usecase struct {
	permission domainusecasesrepocache.Permission
	logger     domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	permission domainusecasesrepocache.Permission,
	logger domaincontractslogger.Leveled,
) domainusecasesadmin.PermissionManagement {
	return &usecase{
		permission: permission,
		logger:     logger,
	}
}

func (u *usecase) Create(ctx context.Context, request domainusecasesadmin.CreatePermissionRequest) (uuid.UUID, error) {
	const tag = "admin/permission_management/Create"

	name, err := applicationshared.RequiredPermissionName(request.Name, "name")
	if err != nil {
		return uuid.Nil, err
	}

	id, err := u.permission.Create(ctx, name, request.Description, request.CreatedBy)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to create permission", domainmodels.LoggerMeta{
			"err":        err,
			"created_by": request.CreatedBy,
		})
		return uuid.Nil, err
	}

	return id, nil
}

func (u *usecase) ReadById(
	ctx context.Context,
	request domainusecasesadmin.ReadPermissionByIdRequest,
) (*domainmodels.Permission, error) {
	const tag = "admin/permission_management/ReadById"

	permission, err := u.permission.ReadById(ctx, request.Id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read permission", domainmodels.LoggerMeta{
			"err": err,
			"id":  request.Id,
		})
		return nil, err
	}

	return permission, nil
}

func (u *usecase) ReadByName(
	ctx context.Context,
	request domainusecasesadmin.ReadPermissionByNameRequest,
) (*domainmodels.Permission, error) {
	const tag = "admin/permission_management/ReadByName"

	permission, err := u.permission.ReadByName(ctx, request.Name)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read permission", domainmodels.LoggerMeta{
			"err":  err,
			"name": request.Name,
		})
		return nil, err
	}

	return permission, nil
}

func (u *usecase) ReadByPagination(
	ctx context.Context,
	request domainusecasesadmin.ReadPermissionsByPaginationRequest,
) ([]domainmodels.Permission, int, error) {
	const tag = "admin/permission_management/ReadByPagination"

	permissions, total, err := u.permission.ReadByPagination(ctx, request.Page, request.Limit, request.Search)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read permissions", domainmodels.LoggerMeta{
			"err":   err,
			"page":  request.Page,
			"limit": request.Limit,
		})
		return nil, 0, err
	}

	return permissions, total, nil
}

func (u *usecase) UpdateById(ctx context.Context, request domainusecasesadmin.UpdatePermissionRequest) error {
	const tag = "admin/permission_management/UpdateById"

	name, err := applicationshared.OptionalPermissionName(request.Name, "name")
	if err != nil {
		return err
	}

	if err := u.permission.UpdateById(
		ctx,
		request.Id,
		name,
		request.Description,
		nil,
		request.UpdatedBy,
	); err != nil {
		u.logger.Error(ctx, tag, "failed to update permission", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	return nil
}

func (u *usecase) DeleteById(ctx context.Context, request domainusecasesadmin.DeletePermissionRequest) error {
	const tag = "admin/permission_management/DeleteById"

	if err := u.permission.DeleteById(ctx, request.Id, request.DeletedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to delete permission", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"deleted_by": request.DeletedBy,
		})
		return err
	}

	return nil
}
