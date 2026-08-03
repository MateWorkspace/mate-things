package infrastructurerepositorynodeclass

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructurerepositoryshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/shared"
	"github.com/MateWorkspace/mate-things/backend/pkg/pgxdt"
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
) domaincontractsrepository.NodeClass {
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
	description *string,
	createdBy *uuid.UUID,
) (id uuid.UUID, err error) {
	query, args, err := p.queryCreate(name, description, createdBy)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create node class query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError(
			"failed to create node class", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "name", Type: domainmodels.ErrTypeNodeClassNameExists},
		)
	}

	return id, nil
}

func (p *postgresImpl) ReadById(ctx context.Context, id uuid.UUID) (nodeClass *domainmodels.NodeClass, err error) {
	query, args, err := p.queryReadById(id)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read node class query", err)
	}

	item, err := infrastructurerepositoryshared.ScanPgxNodeClass(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("node class not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read node class", err)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByName(ctx context.Context, name string) (nodeClass *domainmodels.NodeClass, err error) {
	query, args, err := p.queryReadByName(name)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read node class query", err)
	}

	item, err := infrastructurerepositoryshared.ScanPgxNodeClass(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("node class not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read node class", err)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	search *string,
) (nodeClasses []domainmodels.NodeClass, total int, err error) {
	totalQuery, totalArgs, query, queryArgs, err := p.queryReadByPagination(page, limit, search)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.QueryBuildError("failed to build read node classes query", err)
	}

	if err := p.Dt.QueryRow(ctx, totalQuery, totalArgs...).Scan(&total); err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to count node classes", err)
	}
	if total == 0 {
		return []domainmodels.NodeClass{}, 0, nil
	}

	rows, err := p.Dt.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to read node classes", err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanPgxNodeClasses(rows)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to scan node classes", err)
	}

	return items, total, nil
}

func (p *postgresImpl) ReadActions(ctx context.Context, nodeClassId uuid.UUID) (actions []domainmodels.Action, err error) {
	query, args, err := p.queryReadActions(nodeClassId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read node class actions query", err)
	}

	rows, err := p.Dt.Query(ctx, query, args...)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read node class actions", err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanPgxActions(rows)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to scan node class actions", err)
	}

	return items, nil
}

func (p *postgresImpl) UpdateById(
	ctx context.Context,
	id uuid.UUID,
	name *string,
	description *string,
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) (err error) {
	query, args, err := p.queryUpdateById(id, name, description, preferences, updatedBy)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build update node class query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError(
			"failed to update node class", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "name", Type: domainmodels.ErrTypeNodeClassNameExists},
		)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("node class not found", nil)
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
		return infrastructurerepositoryshared.QueryBuildError("failed to build delete node class query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to delete node class", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("node class not found", nil)
	}

	return nil
}
