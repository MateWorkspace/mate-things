package infrastructurerepositoryinfraredrecordsession

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
) domaincontractsrepository.InfraredRecordSession {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func scanInfraredRecordSession(row pgx.Row, item *domainmodels.InfraredRecordSession) error {
	return row.Scan(
		&item.Id, &item.NodeId, &item.InfraredDeviceId, &item.RecordingState,
		&item.CurrentRecordCaseId, &item.IsCompleted, &item.CreatedAt,
		&item.UpdatedAt, &item.DeletedAt, &item.CreatedBy, &item.UpdatedBy, &item.DeletedBy,
	)
}

func (p *postgresImpl) Create(ctx context.Context, nodeId uuid.UUID, infraredDeviceId uuid.UUID, createdBy *uuid.UUID) (id uuid.UUID, err error) {
	query, args, err := p.queryCreate(nodeId, infraredDeviceId, createdBy)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create infrared_record_session query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create infrared_record_session", err)
	}
	return id, nil
}

func (p *postgresImpl) GetById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredRecordSession, error) {
	query, args, err := p.queryGetById(id)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build infrared_record_session get query", err)
	}

	var item domainmodels.InfraredRecordSession
	if err := scanInfraredRecordSession(p.Dt.QueryRow(ctx, query, args...), &item); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("infrared_record_session not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read infrared_record_session", err)
	}
	return &item, nil
}

func (p *postgresImpl) GetActiveByNodeId(ctx context.Context, nodeId uuid.UUID) (*domainmodels.InfraredRecordSession, error) {
	query, args, err := p.queryGetActiveByNodeId(nodeId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build infrared_record_session get active query", err)
	}

	var item domainmodels.InfraredRecordSession
	if err := scanInfraredRecordSession(p.Dt.QueryRow(ctx, query, args...), &item); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read active infrared_record_session", err)
	}
	return &item, nil
}

func (p *postgresImpl) UpdateRecordingStateById(ctx context.Context, id uuid.UUID, recordingState string, isCompleted bool, updatedBy *uuid.UUID) error {
	query, args, err := p.queryUpdateRecordingStateById(id, recordingState, isCompleted, updatedBy)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build update infrared_record_session recording state query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to update infrared_record_session recording state", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("infrared_record_session not found", nil)
	}

	return nil
}

func (p *postgresImpl) UpdateCurrentRecordCaseIdById(ctx context.Context, id uuid.UUID, currentRecordCaseId *uuid.UUID, updatedBy *uuid.UUID) error {
	query, args, err := p.queryUpdateCurrentRecordCaseIdById(id, currentRecordCaseId, updatedBy)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build update infrared_record_session current record case id query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to update infrared_record_session current record case id", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("infrared_record_session not found", nil)
	}

	return nil
}

func (p *postgresImpl) DeleteById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	query, args, err := p.queryDeleteById(id, deletedBy)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build delete infrared_record_session query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to delete infrared_record_session", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("infrared_record_session not found", nil)
	}
	return nil
}
