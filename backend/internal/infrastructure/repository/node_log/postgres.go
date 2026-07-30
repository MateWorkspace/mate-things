package infrastructurerepositorynodelog

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructurerepositoryshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/shared"
	"github.com/MateWorkspace/mate-things/backend/pkg/pgxdt"
)

type postgresImpl struct {
	infrastructurerepositoryshared.BasePostgres
}

func NewPostgresImpl(
	dt pgxdt.Pgxdt,
	sqrQuestion *squirrel.StatementBuilderType,
	sqrDollar *squirrel.StatementBuilderType,
) domaincontractsrepository.NodeLog {
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
	nodeDeviceId string,
	level domainmodels.NodeLogLevel,
	tag string,
	message string,
	loggedAt time.Time,
) (id int64, err error) {
	query, args, err := p.queryCreate(nodeDeviceId, level, tag, message, loggedAt)
	if err != nil {
		return 0, infrastructurerepositoryshared.QueryBuildError("failed to build create node log query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return 0, infrastructurerepositoryshared.MapPgxError("failed to create node log", err)
	}

	return id, nil
}

func (p *postgresImpl) ReadByFilter(
	ctx context.Context,
	loggedAtStart *time.Time,
	loggedAtEnd *time.Time,
	nodeDeviceId *string,
	level *domainmodels.NodeLogLevel,
) (nodeLogs []domainmodels.NodeLog, total int, err error) {
	totalQuery, totalArgs, query, queryArgs, err := p.queryReadByFilter(loggedAtStart, loggedAtEnd, nodeDeviceId, level)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.QueryBuildError("failed to build read node logs query", err)
	}

	if err := p.Dt.QueryRow(ctx, totalQuery, totalArgs...).Scan(&total); err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to count node logs", err)
	}
	if total == 0 {
		return []domainmodels.NodeLog{}, 0, nil
	}

	rows, err := p.Dt.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to read node logs", err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanPgxNodeLogs(rows)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to scan node logs", err)
	}

	return items, total, nil
}

func (p *postgresImpl) DeleteByFilter(
	ctx context.Context,
	loggedAtStart *time.Time,
	loggedAtEnd *time.Time,
	nodeDeviceId *string,
	level *domainmodels.NodeLogLevel,
) (total int, err error) {
	query, args, err := p.queryDeleteByFilter(loggedAtStart, loggedAtEnd, nodeDeviceId, level)
	if err != nil {
		return 0, infrastructurerepositoryshared.QueryBuildError("failed to build delete node logs query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return 0, infrastructurerepositoryshared.MapPgxError("failed to delete node logs", err)
	}

	return int(commandTag.RowsAffected()), nil
}
