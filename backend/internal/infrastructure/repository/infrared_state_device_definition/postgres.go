package infrastructurerepositoryinfraredstatedevicedefinition

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
) domaincontractsrepository.InfraredStateDeviceDefinition {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func (p *postgresImpl) CreateMany(ctx context.Context, definitions []domainmodels.InfraredStateDeviceDefinition) error {
	for _, definition := range definitions {
		query, args, err := p.queryCreate(definition)
		if err != nil {
			return infrastructurerepositoryshared.QueryBuildError("failed to build create infrared_state_device_definition query", err)
		}

		var id uuid.UUID
		if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
			return infrastructurerepositoryshared.MapPgxError("failed to create infrared_state_device_definition", err)
		}
	}
	return nil
}

func (p *postgresImpl) ListByDeviceId(ctx context.Context, infraredDeviceId uuid.UUID) ([]domainmodels.InfraredStateDeviceDefinition, error) {
	query, args, err := p.queryListByDeviceId(infraredDeviceId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build infrared_state_device_definition list query", err)
	}

	rows, err := p.Dt.Query(ctx, query, args...)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to list infrared_state_device_definition", err)
	}
	defer rows.Close()

	var result []domainmodels.InfraredStateDeviceDefinition
	for rows.Next() {
		var item domainmodels.InfraredStateDeviceDefinition
		if err := rows.Scan(&item.Id, &item.InfraredDeviceId, &item.InfraredStateId, &item.Options, &item.Minimum, &item.Maximum, &item.Step); err != nil {
			return nil, infrastructurerepositoryshared.MapPgxError("failed to scan infrared_state_device_definition", err)
		}
		result = append(result, item)
	}
	return result, nil
}
