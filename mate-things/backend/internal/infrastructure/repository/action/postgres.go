package infrastructurerepositoryaction

import (
	"context"
	"encoding/json"
	"errors"

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
) domaincontractsrepository.Action {
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
	nodeClassId uuid.UUID,
	name string,
	description *string,
	payloadSchemaName string,
	payloadSchemaVersion int32,
	createdBy *uuid.UUID,
) (id uuid.UUID, err error) {
	query, args, err := p.queryCreate(nodeClassId, name, description, payloadSchemaName, payloadSchemaVersion, createdBy)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create action query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create action", err)
	}

	return id, nil
}

func (p *postgresImpl) ReadById(ctx context.Context, id uuid.UUID) (action *domainmodels.Action, err error) {
	query, args, err := p.queryReadById(id)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read action query", err)
	}

	item, err := infrastructurerepositoryshared.ScanPgxAction(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("action not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read action", err)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByName(ctx context.Context, name string) (action *domainmodels.Action, err error) {
	query, args, err := p.queryReadByName(name)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read action query", err)
	}

	item, err := infrastructurerepositoryshared.ScanPgxAction(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("action not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read action", err)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	search *string,
	nodeClassId *uuid.UUID,
	payloadSchemaName *string,
	payloadSchemaVersion *int32,
) (actions []domainmodels.Action, total int, err error) {
	totalQuery, totalArgs, query, queryArgs, err := p.queryReadByPagination(page, limit, search, nodeClassId, payloadSchemaName, payloadSchemaVersion)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.QueryBuildError("failed to build read actions query", err)
	}

	if err := p.Dt.QueryRow(ctx, totalQuery, totalArgs...).Scan(&total); err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to count actions", err)
	}
	if total == 0 {
		return []domainmodels.Action{}, 0, nil
	}

	rows, err := p.Dt.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to read actions", err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanPgxActions(rows)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to scan actions", err)
	}

	return items, total, nil
}

func (p *postgresImpl) UpdateById(
	ctx context.Context,
	id uuid.UUID,
	nodeClassId *uuid.UUID,
	name *string,
	description *string,
	payloadSchemaName *string,
	payloadSchemaVersion *int32,
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) (err error) {
	query, args, err := p.queryUpdateById(id, nodeClassId, name, description, payloadSchemaName, payloadSchemaVersion, preferences, updatedBy)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build update action query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to update action", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("action not found", nil)
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
		return infrastructurerepositoryshared.QueryBuildError("failed to build delete action query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to delete action", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("action not found", nil)
	}

	return nil
}
