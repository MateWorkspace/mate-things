package infrastructurerepositoryinfraredstatecoder

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
) domaincontractsrepository.InfraredStateCoder {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func (p *postgresImpl) Create(ctx context.Context, coder domainmodels.InfraredStateCoder) (id uuid.UUID, err error) {
	query, args, err := p.queryCreate(coder)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create infrared_state_coder query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create infrared_state_coder", err)
	}
	return id, nil
}

func scanInfraredStateCoder(row pgx.Row, item *domainmodels.InfraredStateCoder) error {
	var status string
	if err := row.Scan(
		&item.Id, &item.InfraredDeviceId, &item.InfraredRecordSessionId,
		&item.EncoderSource, &item.DecoderSource, &item.SummaryReadme, &item.DetailReadme,
		&status, &item.CreatedAt, &item.DeletedAt, &item.DeletedBy,
	); err != nil {
		return err
	}
	item.Status = domainmodels.InfraredStateCoderStatus(status)
	return nil
}

func (p *postgresImpl) ReadBySessionId(ctx context.Context, sessionId uuid.UUID) (*domainmodels.InfraredStateCoder, error) {
	query, args, err := p.queryReadBySessionId(sessionId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build get infrared_state_coder query", err)
	}

	var item domainmodels.InfraredStateCoder
	if err := scanInfraredStateCoder(p.Dt.QueryRow(ctx, query, args...), &item); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("infrared_state_coder not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to get infrared_state_coder", err)
	}
	return &item, nil
}

func (p *postgresImpl) ReadById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredStateCoder, error) {
	query, args, err := p.queryReadById(id)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build get infrared_state_coder query", err)
	}

	var item domainmodels.InfraredStateCoder
	if err := scanInfraredStateCoder(p.Dt.QueryRow(ctx, query, args...), &item); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("infrared_state_coder not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to get infrared_state_coder", err)
	}
	return &item, nil
}

func (p *postgresImpl) Activate(ctx context.Context, coderId uuid.UUID, deviceId uuid.UUID) error {
	activateQuery, activateArgs, err := p.SqrD.Update("infrared_state_coder").
		Set("status", domainmodels.InfraredStateCoderStatusActive).
		Where(squirrel.Eq{"id": coderId}).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build activate infrared_state_coder query", err)
	}
	if _, err := p.Dt.Exec(ctx, activateQuery, activateArgs...); err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to activate infrared_state_coder", err)
	}

	supersedeQuery, supersedeArgs, err := p.SqrD.Update("infrared_state_coder").
		Set("status", domainmodels.InfraredStateCoderStatusSuperseded).
		Where(squirrel.Eq{"infrared_device_id": deviceId, "status": domainmodels.InfraredStateCoderStatusActive}).
		Where(squirrel.NotEq{"id": coderId}).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build supersede infrared_state_coder query", err)
	}
	if _, err := p.Dt.Exec(ctx, supersedeQuery, supersedeArgs...); err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to supersede prior infrared_state_coder", err)
	}
	return nil
}

func (p *postgresImpl) DeleteById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	query, args, err := p.queryDeleteById(id, deletedBy)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build delete infrared_state_coder query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to delete infrared_state_coder", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("infrared_state_coder not found", nil)
	}
	return nil
}
