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

func scanInfraredDeviceType(row pgx.Row, item *domainmodels.InfraredDeviceType) error {
	return row.Scan(
		&item.Id, &item.Name,
		&item.CreatedAt, &item.UpdatedAt, &item.DeletedAt,
		&item.CreatedBy, &item.UpdatedBy, &item.DeletedBy,
	)
}

func (p *postgresImpl) ReadList(ctx context.Context) ([]domainmodels.InfraredDeviceType, error) {
	query, args, err := p.queryReadList()
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
		if err := scanInfraredDeviceType(rows, &item); err != nil {
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
	if err := scanInfraredDeviceType(p.Dt.QueryRow(ctx, query, args...), &item); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("infrared_device_type not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read infrared_device_type", err)
	}
	return &item, nil
}

func (p *postgresImpl) Create(ctx context.Context, name string, createdBy *uuid.UUID) (id uuid.UUID, err error) {
	query, args, err := p.queryCreate(name, createdBy)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create infrared_device_type query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create infrared_device_type", err)
	}
	return id, nil
}

func (p *postgresImpl) DeleteById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	query, args, err := p.queryDeleteById(id, deletedBy)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build delete infrared_device_type query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to delete infrared_device_type", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("infrared_device_type not found", nil)
	}
	return nil
}
