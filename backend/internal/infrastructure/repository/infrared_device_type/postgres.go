package infrastructurerepositoryinfrareddevicetype

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
) domaincontractsrepository.InfraredDeviceType {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func (p *postgresImpl) List(ctx context.Context) ([]domainmodels.InfraredDeviceType, error) {
	query, args, err := p.queryList()
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build infrared_device_type list query", err)
	}

	rows, err := p.Dt.Query(ctx, query, args...)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to list infrared_device_type", err)
	}
	defer rows.Close()

	var result []domainmodels.InfraredDeviceType
	for rows.Next() {
		var item domainmodels.InfraredDeviceType
		if err := rows.Scan(&item.Id, &item.Name); err != nil {
			return nil, infrastructurerepositoryshared.MapPgxError("failed to scan infrared_device_type", err)
		}
		result = append(result, item)
	}
	return result, nil
}

func (p *postgresImpl) ReadByName(ctx context.Context, name string) (*domainmodels.InfraredDeviceType, error) {
	query, args, err := p.queryReadByName(name)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build infrared_device_type read query", err)
	}

	var item domainmodels.InfraredDeviceType
	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&item.Id, &item.Name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("infrared_device_type not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read infrared_device_type", err)
	}
	return &item, nil
}

func (p *postgresImpl) Create(ctx context.Context, name string) (id uuid.UUID, err error) {
	query, args, err := p.queryCreate(name)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create infrared_device_type query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create infrared_device_type", err)
	}
	return id, nil
}
