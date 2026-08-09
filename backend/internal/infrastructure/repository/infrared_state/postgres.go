package infrastructurerepositoryinfraredstate

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
) domaincontractsrepository.InfraredState {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func (p *postgresImpl) ListByDeviceTypeId(ctx context.Context, infraredDeviceTypeId uuid.UUID) ([]domainmodels.InfraredState, error) {
	query, args, err := p.queryListByDeviceTypeId(infraredDeviceTypeId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build infrared_state list query", err)
	}

	rows, err := p.Dt.Query(ctx, query, args...)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to list infrared_state", err)
	}
	defer rows.Close()

	var result []domainmodels.InfraredState
	for rows.Next() {
		var item domainmodels.InfraredState
		var stateType string
		if err := rows.Scan(&item.Id, &item.InfraredDeviceTypeId, &item.Name, &stateType); err != nil {
			return nil, infrastructurerepositoryshared.MapPgxError("failed to scan infrared_state", err)
		}
		item.Type = domainmodels.InfraredStateType(stateType)
		result = append(result, item)
	}
	return result, nil
}

func (p *postgresImpl) ReadByDeviceTypeIdAndName(ctx context.Context, infraredDeviceTypeId uuid.UUID, name string) (*domainmodels.InfraredState, error) {
	query, args, err := p.queryReadByDeviceTypeIdAndName(infraredDeviceTypeId, name)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build infrared_state read query", err)
	}

	var item domainmodels.InfraredState
	var stateType string
	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&item.Id, &item.InfraredDeviceTypeId, &item.Name, &stateType); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("infrared_state not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read infrared_state", err)
	}
	item.Type = domainmodels.InfraredStateType(stateType)
	return &item, nil
}

func (p *postgresImpl) Create(ctx context.Context, infraredDeviceTypeId uuid.UUID, name string, stateType domainmodels.InfraredStateType) (id uuid.UUID, err error) {
	query, args, err := p.queryCreate(infraredDeviceTypeId, name, stateType)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create infrared_state query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create infrared_state", err)
	}
	return id, nil
}
