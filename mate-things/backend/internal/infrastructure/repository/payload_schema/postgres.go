package infrastructurerepositorypayloadschema

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	domaincontractsrepository "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	infrastructurerepositoryshared "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/repository/shared"
	"github.com/ABA-Developer/nusapala-things/backend/pkg/pgxdt"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type postgresImpl struct {
	infrastructurerepositoryshared.BasePostgres
}

func NewPostgresImpl(
	dt pgxdt.Pgxdt,
	sqrQuestion *squirrel.StatementBuilderType,
	sqrDollar *squirrel.StatementBuilderType,
) domaincontractsrepository.PayloadSchema {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func (p *postgresImpl) Create(
	ctx context.Context,
	name string,
	version int32,
	definition json.RawMessage,
	validFrom *time.Time,
	validTo *time.Time,
	createdBy *uuid.UUID,
) (id uuid.UUID, err error) {
	query, args, err := p.queryCreate(name, version, definition, validFrom, validTo, createdBy)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create payload schema query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create payload schema", err)
	}

	return id, nil
}

func (p *postgresImpl) ReadById(ctx context.Context, id uuid.UUID) (payloadSchema *domainmodels.PayloadSchema, err error) {
	query, args, err := p.queryReadById(id)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read payload schema query", err)
	}

	item, err := infrastructurerepositoryshared.ScanPgxPayloadSchema(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("payload schema not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read payload schema", err)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByNameAndVersion(
	ctx context.Context,
	name string,
	version int32,
) (payloadSchema *domainmodels.PayloadSchema, err error) {
	query, args, err := p.queryReadByNameAndVersion(name, version)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read payload schema query", err)
	}

	item, err := infrastructurerepositoryshared.ScanPgxPayloadSchema(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("payload schema not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read payload schema", err)
	}

	return &item, nil
}

func (p *postgresImpl) ReadLatestByName(ctx context.Context, name string) (payloadSchema *domainmodels.PayloadSchema, err error) {
	query, args, err := p.queryReadLatestByName(name)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read payload schema query", err)
	}

	item, err := infrastructurerepositoryshared.ScanPgxPayloadSchema(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("payload schema not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read payload schema", err)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	search *string,
	validAt *time.Time,
) (payloadSchemas []domainmodels.PayloadSchema, total int, err error) {
	totalQuery, totalArgs, query, queryArgs, err := p.queryReadByPagination(page, limit, search, validAt)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.QueryBuildError("failed to build read payload schemas query", err)
	}

	if err := p.Dt.QueryRow(ctx, totalQuery, totalArgs...).Scan(&total); err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to count payload schemas", err)
	}
	if total == 0 {
		return []domainmodels.PayloadSchema{}, 0, nil
	}

	rows, err := p.Dt.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to read payload schemas", err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanPgxPayloadSchemas(rows)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to scan payload schemas", err)
	}

	return items, total, nil
}

func (p *postgresImpl) UpdateById(
	ctx context.Context,
	id uuid.UUID,
	name *string,
	version *int32,
	definition *json.RawMessage,
	validFrom *time.Time,
	validTo *time.Time,
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) (err error) {
	query, args, err := p.queryUpdateById(id, name, version, definition, validFrom, validTo, preferences, updatedBy)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build update payload schema query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to update payload schema", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("payload schema not found", nil)
	}

	return nil
}

func (p *postgresImpl) DeleteById(
	ctx context.Context,
	id uuid.UUID,
	deletedBy *uuid.UUID,
) (err error) {
	query, args, err := p.queryDeleteById(id, deletedBy)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build delete payload schema query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to delete payload schema", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("payload schema not found", nil)
	}

	return nil
}
