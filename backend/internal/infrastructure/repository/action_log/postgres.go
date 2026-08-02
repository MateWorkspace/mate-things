package infrastructurerepositoryactionlog

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructurerepositoryshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/shared"
	"github.com/MateWorkspace/mate-things/backend/pkg/pgxdt"
	"github.com/google/uuid"
)

type postgresImpl struct {
	infrastructurerepositoryshared.BasePostgres
}

func NewPostgresImpl(
	dt pgxdt.Pgxdt,
	sqrQuestion *squirrel.StatementBuilderType,
	sqrDollar *squirrel.StatementBuilderType,
) domaincontractsrepository.ActionLog {
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
	executionId uuid.UUID,
	actionId uuid.UUID,
	nodeId *uuid.UUID,
	actionStatus domainmodels.ActionStatus,
	actionMessage *string,
	payload json.RawMessage,
	executedAt time.Time,
) (id int64, err error) {
	query, args, err := p.queryCreate(executionId, actionId, nodeId, string(actionStatus), actionMessage, payload, executedAt)
	if err != nil {
		return 0, infrastructurerepositoryshared.QueryBuildError("failed to build create action log query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return 0, infrastructurerepositoryshared.MapPgxError("failed to create action log", err)
	}

	return id, nil
}

func (p *postgresImpl) UpdateStatusByExecutionId(
	ctx context.Context,
	executionId uuid.UUID,
	actionStatus domainmodels.ActionStatus,
	actionMessage *string,
) (err error) {
	query, args, err := p.queryUpdateStatusByExecutionId(executionId, string(actionStatus), actionMessage)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build update action log status query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to update action log status", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("action log not found", nil)
	}

	return nil
}

func (p *postgresImpl) ReadByFilter(
	ctx context.Context,
	executedAtStart *time.Time,
	executedAtEnd *time.Time,
	actionId *uuid.UUID,
	nodeId *uuid.UUID,
	executionId *uuid.UUID,
	actionStatus *domainmodels.ActionStatus,
	page int,
	limit int,
) (actionLogs []domainmodels.ActionLogListItem, total int, err error) {
	totalQuery, totalArgs, query, queryArgs, err := p.queryReadByFilter(executedAtStart, executedAtEnd, actionId, nodeId, executionId, actionStatus, page, limit)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.QueryBuildError("failed to build read action logs query", err)
	}

	if err := p.Dt.QueryRow(ctx, totalQuery, totalArgs...).Scan(&total); err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to count action logs", err)
	}
	if total == 0 {
		return []domainmodels.ActionLogListItem{}, 0, nil
	}

	rows, err := p.Dt.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to read action logs", err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanPgxActionLogListItems(rows)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to scan action logs", err)
	}

	return items, total, nil
}

func (p *postgresImpl) DeleteByFilter(
	ctx context.Context,
	executedAtStart *time.Time,
	executedAtEnd *time.Time,
	actionId *uuid.UUID,
	nodeId *uuid.UUID,
	executionId *uuid.UUID,
	actionStatus *domainmodels.ActionStatus,
) (total int, err error) {
	query, args, err := p.queryDeleteByFilter(executedAtStart, executedAtEnd, actionId, nodeId, executionId, actionStatus)
	if err != nil {
		return 0, infrastructurerepositoryshared.QueryBuildError("failed to build delete action logs query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return 0, infrastructurerepositoryshared.MapPgxError("failed to delete action logs", err)
	}

	return int(commandTag.RowsAffected()), nil
}
