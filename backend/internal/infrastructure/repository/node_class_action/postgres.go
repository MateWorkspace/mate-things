package infrastructurerepositorynodeclassaction

import (
	"context"
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
) domaincontractsrepository.NodeClassAction {
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
	actionId uuid.UUID,
	createdBy *uuid.UUID,
) (id uuid.UUID, err error) {
	query, args, err := p.queryCreate(nodeClassId, actionId, createdBy)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create node class action query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError(
			"failed to create node class action", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "node_class_id_action_id", Type: domainmodels.ErrTypeNodeClassActionExists},
		)
	}

	return id, nil
}

func (p *postgresImpl) ReadById(
	ctx context.Context,
	id uuid.UUID,
) (nodeClassAction *domainmodels.NodeClassAction, nodeClass *domainmodels.NodeClass, action *domainmodels.Action, err error) {
	query, args, err := p.queryReadById(id)
	if err != nil {
		return nil, nil, nil, infrastructurerepositoryshared.QueryBuildError("failed to build read node class action query", err)
	}

	nca, nc, a, err := scanPgxNodeClassActionJoined(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, nil, infrastructurerepositoryshared.NotFound("node class action not found", err)
		}
		return nil, nil, nil, infrastructurerepositoryshared.MapPgxError("failed to read node class action", err)
	}

	return &nca, &nc, &a, nil
}

func (p *postgresImpl) ReadByNodeClassIdAndActionId(
	ctx context.Context,
	nodeClassId uuid.UUID,
	actionId uuid.UUID,
) (nodeClassAction *domainmodels.NodeClassAction, nodeClass *domainmodels.NodeClass, action *domainmodels.Action, err error) {
	query, args, err := p.queryReadByNodeClassIdAndActionId(nodeClassId, actionId)
	if err != nil {
		return nil, nil, nil, infrastructurerepositoryshared.QueryBuildError("failed to build read node class action query", err)
	}

	nca, nc, a, err := scanPgxNodeClassActionJoined(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, nil, infrastructurerepositoryshared.NotFound("node class action not found", err)
		}
		return nil, nil, nil, infrastructurerepositoryshared.MapPgxError("failed to read node class action", err)
	}

	return &nca, &nc, &a, nil
}

func (p *postgresImpl) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	nodeClassId *uuid.UUID,
	actionId *uuid.UUID,
) (nodeClassActions []domainmodels.NodeClassAction, nodeClasses []domainmodels.NodeClass, actions []domainmodels.Action, total int, err error) {
	totalQuery, totalArgs, query, queryArgs, err := p.queryReadByPagination(page, limit, nodeClassId, actionId)
	if err != nil {
		return nil, nil, nil, 0, infrastructurerepositoryshared.QueryBuildError("failed to build read node class actions query", err)
	}

	if err := p.Dt.QueryRow(ctx, totalQuery, totalArgs...).Scan(&total); err != nil {
		return nil, nil, nil, 0, infrastructurerepositoryshared.MapPgxError("failed to count node class actions", err)
	}
	if total == 0 {
		return []domainmodels.NodeClassAction{}, []domainmodels.NodeClass{}, []domainmodels.Action{}, 0, nil
	}

	rows, err := p.Dt.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, nil, nil, 0, infrastructurerepositoryshared.MapPgxError("failed to read node class actions", err)
	}
	defer rows.Close()

	nodeClassActions = make([]domainmodels.NodeClassAction, 0, total)
	nodeClasses = make([]domainmodels.NodeClass, 0, total)
	actions = make([]domainmodels.Action, 0, total)
	for rows.Next() {
		nca, nc, a, err := scanPgxNodeClassActionJoined(rows)
		if err != nil {
			return nil, nil, nil, 0, infrastructurerepositoryshared.MapPgxError("failed to scan node class action", err)
		}

		nodeClassActions = append(nodeClassActions, nca)
		nodeClasses = append(nodeClasses, nc)
		actions = append(actions, a)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, 0, infrastructurerepositoryshared.MapPgxError("failed to iterate node class actions", err)
	}

	return nodeClassActions, nodeClasses, actions, total, nil
}

func (p *postgresImpl) DeleteById(ctx context.Context, id uuid.UUID) (err error) {
	query, args, err := p.queryDeleteById(id)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build delete node class action query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to delete node class action", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("node class action not found", nil)
	}

	return nil
}

func (p *postgresImpl) DeleteByNodeClassIdAndActionId(
	ctx context.Context,
	nodeClassId *uuid.UUID,
	actionId *uuid.UUID,
) (err error) {
	query, args, err := p.queryDeleteByNodeClassIdAndActionId(nodeClassId, actionId)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build delete node class action query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to delete node class action", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("node class action not found", nil)
	}

	return nil
}

func scanPgxNodeClassActionJoined(row pgx.Row) (
	nodeClassAction domainmodels.NodeClassAction,
	nodeClass domainmodels.NodeClass,
	action domainmodels.Action,
	err error,
) {
	err = row.Scan(
		&nodeClassAction.Id,
		&nodeClassAction.NodeClassId,
		&nodeClassAction.ActionId,
		&nodeClassAction.CreatedAt,
		&nodeClassAction.CreatedBy,
		&nodeClass.Id,
		&nodeClass.Name,
		&nodeClass.Description,
		&nodeClass.Preferences,
		&nodeClass.CreatedAt,
		&nodeClass.UpdatedAt,
		&nodeClass.DeletedAt,
		&nodeClass.CreatedBy,
		&nodeClass.UpdatedBy,
		&nodeClass.DeletedBy,
		&action.Id,
		&action.Name,
		&action.Description,
		&action.PayloadSchemaName,
		&action.PayloadSchemaVersion,
		&action.Preferences,
		&action.CreatedAt,
		&action.UpdatedAt,
		&action.DeletedAt,
		&action.CreatedBy,
		&action.UpdatedBy,
		&action.DeletedBy,
	)
	return
}
