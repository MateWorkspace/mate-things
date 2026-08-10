package infrastructurerepositoryinfrareddevice

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
) domaincontractsrepository.InfraredDevice {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func (p *postgresImpl) Create(ctx context.Context, infraredDeviceTypeId uuid.UUID, brand string, model string, createdBy *uuid.UUID) (id uuid.UUID, err error) {
	query, args, err := p.queryCreate(infraredDeviceTypeId, brand, model, createdBy)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create infrared_device query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create infrared_device", err)
	}
	return id, nil
}

func (p *postgresImpl) GetById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredDevice, error) {
	query, args, err := p.queryGetById(id)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build infrared_device get query", err)
	}

	var item domainmodels.InfraredDevice
	if err := p.Dt.QueryRow(ctx, query, args...).Scan(
		&item.Id, &item.InfraredDeviceTypeId, &item.Brand, &item.Model,
		&item.CreatedAt, &item.UpdatedAt, &item.DeletedAt,
		&item.CreatedBy, &item.UpdatedBy, &item.DeletedBy,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("infrared_device not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read infrared_device", err)
	}
	return &item, nil
}

func (p *postgresImpl) DeleteById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	query, args, err := p.queryDeleteById(id, deletedBy)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build delete infrared_device query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to delete infrared_device", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("infrared_device not found", nil)
	}
	return nil
}
