package infrastructurerepositorynodeconfigvalue

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
) domaincontractsrepository.NodeConfigValue {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func (p *postgresImpl) ReadByNodeId(ctx context.Context, nodeId uuid.UUID) ([]domainmodels.NodeConfigValue, error) {
	query, args, err := p.queryReadByNodeId(nodeId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read node config values query", err)
	}

	rows, err := p.Dt.Query(ctx, query, args...)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read node config values", err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanPgxNodeConfigValues(rows)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to scan node config values", err)
	}

	return items, nil
}

func (p *postgresImpl) ReadByNodeIdAndKey(ctx context.Context, nodeId uuid.UUID, key string) (*domainmodels.NodeConfigValue, error) {
	query, args, err := p.queryReadByNodeIdAndKey(nodeId, key)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read node config value query", err)
	}

	item, err := infrastructurerepositoryshared.ScanPgxNodeConfigValue(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("node config value not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read node config value", err)
	}

	return &item, nil
}

func (p *postgresImpl) Upsert(
	ctx context.Context,
	nodeId uuid.UUID,
	firmwareId uuid.UUID,
	key string,
	value string,
	actorId *uuid.UUID,
) error {
	query, args, err := p.queryUpsert(nodeId, firmwareId, key, value, actorId)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build upsert node config value query", err)
	}

	if _, err := p.Dt.Exec(ctx, query, args...); err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to upsert node config value", err)
	}

	return nil
}
