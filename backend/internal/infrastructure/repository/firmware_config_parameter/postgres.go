package infrastructurerepositoryfirmwareconfigparameter

import (
	"context"

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
) domaincontractsrepository.FirmwareConfigParameter {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func (p *postgresImpl) ReplaceForFirmwareId(
	ctx context.Context,
	firmwareId uuid.UUID,
	params []domaincontractsrepository.FirmwareConfigParameterInput,
	actorId *uuid.UUID,
) error {
	keys := make([]string, 0, len(params))
	for _, param := range params {
		keys = append(keys, param.Key)
	}

	return p.Dt.WithTx(ctx, func(ctx context.Context) error {
		deleteQuery, deleteArgs, err := p.querySoftDeleteMissing(firmwareId, keys, actorId)
		if err != nil {
			return infrastructurerepositoryshared.QueryBuildError("failed to build delete firmware config parameters query", err)
		}
		if _, err := p.Dt.Exec(ctx, deleteQuery, deleteArgs...); err != nil {
			return infrastructurerepositoryshared.MapPgxError("failed to delete stale firmware config parameters", err)
		}

		for _, param := range params {
			upsertQuery, upsertArgs, err := p.queryUpsert(firmwareId, param.Key, param.ValueType, actorId)
			if err != nil {
				return infrastructurerepositoryshared.QueryBuildError("failed to build upsert firmware config parameter query", err)
			}
			if _, err := p.Dt.Exec(ctx, upsertQuery, upsertArgs...); err != nil {
				return infrastructurerepositoryshared.MapPgxError(
					"failed to upsert firmware config parameter", err,
					infrastructurerepositoryshared.ConflictMatch{Contains: "firmware_id_key", Type: domainmodels.ErrTypeFirmwareConfigKeyExists},
				)
			}
		}

		return nil
	})
}

func (p *postgresImpl) ReadByFirmwareId(
	ctx context.Context,
	firmwareId uuid.UUID,
) ([]domainmodels.FirmwareConfigParameter, error) {
	query, args, err := p.queryReadByFirmwareId(firmwareId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read firmware config parameters query", err)
	}

	rows, err := p.Dt.Query(ctx, query, args...)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read firmware config parameters", err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanPgxFirmwareConfigParameters(rows)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to scan firmware config parameters", err)
	}

	return items, nil
}
