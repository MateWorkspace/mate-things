package applicationadminschemaregistry

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
	payloadSchema domainusecasesrepocache.PayloadSchema
	logger        domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	payloadSchema domainusecasesrepocache.PayloadSchema,
	logger domaincontractslogger.Leveled,
) domainusecasesadmin.SchemaRegistry {
	return &usecase{
		payloadSchema: payloadSchema,
		logger:        logger,
	}
}

func (u *usecase) Create(ctx context.Context, request domainusecasesadmin.CreatePayloadSchemaRequest) (uuid.UUID, error) {
	const tag = "admin/schema_registry/Create"

	name, err := applicationshared.RequiredSnakeCaseName(request.Name, "name")
	if err != nil {
		return uuid.Nil, err
	}
	version, err := applicationshared.RequiredPositiveVersion(request.Version, "version")
	if err != nil {
		return uuid.Nil, err
	}
	definition, err := applicationshared.RequiredPayloadSchemaDefinition(request.Definition, "definition")
	if err != nil {
		return uuid.Nil, err
	}
	if err := applicationshared.ValidateTimeWindow(request.ValidFrom, request.ValidTo); err != nil {
		return uuid.Nil, err
	}

	id, err := u.payloadSchema.Create(
		ctx,
		name,
		version,
		definition,
		request.ValidFrom,
		request.ValidTo,
		request.CreatedBy,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to create payload schema", domainmodels.LoggerMeta{
			"err":        err,
			"name":       request.Name,
			"version":    request.Version,
			"created_by": request.CreatedBy,
		})
		return uuid.Nil, err
	}

	return id, nil
}

func (u *usecase) ReadById(
	ctx context.Context,
	request domainusecasesadmin.ReadPayloadSchemaByIdRequest,
) (*domainmodels.PayloadSchema, error) {
	const tag = "admin/schema_registry/ReadById"

	payloadSchema, err := u.payloadSchema.ReadById(ctx, request.Id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read payload schema", domainmodels.LoggerMeta{
			"err": err,
			"id":  request.Id,
		})
		return nil, err
	}

	return payloadSchema, nil
}

func (u *usecase) ReadByNameAndVersion(
	ctx context.Context,
	request domainusecasesadmin.ReadPayloadSchemaByNameAndVersionRequest,
) (*domainmodels.PayloadSchema, error) {
	const tag = "admin/schema_registry/ReadByNameAndVersion"

	payloadSchema, err := u.payloadSchema.ReadByNameAndVersion(ctx, request.Name, request.Version)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read payload schema", domainmodels.LoggerMeta{
			"err":     err,
			"name":    request.Name,
			"version": request.Version,
		})
		return nil, err
	}

	return payloadSchema, nil
}

func (u *usecase) ReadLatestByName(
	ctx context.Context,
	request domainusecasesadmin.ReadLatestPayloadSchemaByNameRequest,
) (*domainmodels.PayloadSchema, error) {
	const tag = "admin/schema_registry/ReadLatestByName"

	payloadSchema, err := u.payloadSchema.ReadLatestByName(ctx, request.Name)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read latest payload schema", domainmodels.LoggerMeta{
			"err":  err,
			"name": request.Name,
		})
		return nil, err
	}

	return payloadSchema, nil
}

func (u *usecase) ReadByPagination(
	ctx context.Context,
	request domainusecasesadmin.ReadPayloadSchemasByPaginationRequest,
) ([]domainmodels.PayloadSchema, int, error) {
	const tag = "admin/schema_registry/ReadByPagination"

	payloadSchemas, total, err := u.payloadSchema.ReadByPagination(ctx, request.Page, request.Limit, request.Search, request.ValidAt)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read payload schemas", domainmodels.LoggerMeta{
			"err":      err,
			"page":     request.Page,
			"limit":    request.Limit,
			"valid_at": request.ValidAt,
		})
		return nil, 0, err
	}

	return payloadSchemas, total, nil
}

func (u *usecase) UpdateById(ctx context.Context, request domainusecasesadmin.UpdatePayloadSchemaRequest) error {
	const tag = "admin/schema_registry/UpdateById"

	name, err := applicationshared.OptionalSnakeCaseName(request.Name, "name")
	if err != nil {
		return err
	}
	version, err := applicationshared.OptionalPositiveVersion(request.Version, "version")
	if err != nil {
		return err
	}
	definition, err := applicationshared.OptionalPayloadSchemaDefinition(request.Definition, "definition")
	if err != nil {
		return err
	}
	if err := applicationshared.ValidateTimeWindow(request.ValidFrom, request.ValidTo); err != nil {
		return err
	}

	if err := u.payloadSchema.UpdateById(
		ctx,
		request.Id,
		name,
		version,
		definition,
		request.ValidFrom,
		request.ValidTo,
		nil,
		request.UpdatedBy,
	); err != nil {
		u.logger.Error(ctx, tag, "failed to update payload schema", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	return nil
}

func (u *usecase) DeleteById(ctx context.Context, request domainusecasesadmin.DeletePayloadSchemaRequest) error {
	const tag = "admin/schema_registry/DeleteById"

	if err := u.payloadSchema.DeleteById(ctx, request.Id, request.DeletedBy); err != nil {
		u.logger.Error(ctx, tag, "failed to delete payload schema", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"deleted_by": request.DeletedBy,
		})
		return err
	}

	return nil
}
