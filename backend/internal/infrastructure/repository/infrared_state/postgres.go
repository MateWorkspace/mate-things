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

func scanInfraredState(row pgx.Row, item *domainmodels.InfraredState) error {
	var stateType string
	if err := row.Scan(
		&item.Id, &item.InfraredDeviceTypeId, &item.Name, &stateType,
		&item.CreatedAt, &item.UpdatedAt, &item.DeletedAt,
		&item.CreatedBy, &item.UpdatedBy, &item.DeletedBy,
	); err != nil {
		return err
	}
	item.Type = domainmodels.InfraredStateType(stateType)
	return nil
}

func (p *postgresImpl) ReadListByDeviceTypeId(ctx context.Context, infraredDeviceTypeId uuid.UUID) ([]domainmodels.InfraredState, error) {
	query, args, err := p.queryReadListByDeviceTypeId(infraredDeviceTypeId)
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
		if err := scanInfraredState(rows, &item); err != nil {
			return nil, infrastructurerepositoryshared.MapPgxError("failed to scan infrared_state", err)
		}
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
	if err := scanInfraredState(p.Dt.QueryRow(ctx, query, args...), &item); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("infrared_state not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read infrared_state", err)
	}
	return &item, nil
}

func (p *postgresImpl) Create(ctx context.Context, infraredDeviceTypeId uuid.UUID, name string, stateType domainmodels.InfraredStateType, createdBy *uuid.UUID) (id uuid.UUID, err error) {
	query, args, err := p.queryCreate(infraredDeviceTypeId, name, stateType, createdBy)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create infrared_state query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create infrared_state", err)
	}
	return id, nil
}

func (p *postgresImpl) DeleteById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	query, args, err := p.queryDeleteById(id, deletedBy)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build delete infrared_state query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to delete infrared_state", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("infrared_state not found", nil)
	}
	return nil
}
