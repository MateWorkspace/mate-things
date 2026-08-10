package infrastructurerepositoryinfraredstatedevicerecordcase

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
) domaincontractsrepository.InfraredStateDeviceRecordCase {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func (p *postgresImpl) CreateWithStates(ctx context.Context, sessionId uuid.UUID, step int32, description string, states []domainmodels.InfraredStateDeviceRecordState) (caseId uuid.UUID, err error) {
	query, args, err := p.queryCreateCase(sessionId, step, description)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create infrared_state_device_record_case query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&caseId); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create infrared_state_device_record_case", err)
	}

	for _, state := range states {
		query, args, err := p.queryCreateState(caseId, state)
		if err != nil {
			return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create infrared_state_device_record_state query", err)
		}

		var id uuid.UUID
		if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
			return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create infrared_state_device_record_state", err)
		}
	}

	return caseId, nil
}

func scanInfraredRecordCase(row pgx.Row, item *domainmodels.InfraredStateDeviceRecordCase) error {
	return row.Scan(
		&item.Id, &item.InfraredRecordSessionId, &item.Step, &item.Description, &item.Status,
		&item.CreatedAt, &item.DeletedAt, &item.DeletedBy,
	)
}

func (p *postgresImpl) ReadListBySessionId(ctx context.Context, sessionId uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordCase, error) {
	query, args, err := p.queryReadListBySessionId(sessionId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build infrared_state_device_record_case list query", err)
	}

	rows, err := p.Dt.Query(ctx, query, args...)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to list infrared_state_device_record_case", err)
	}
	defer rows.Close()

	var result []domainmodels.InfraredStateDeviceRecordCase
	for rows.Next() {
		var item domainmodels.InfraredStateDeviceRecordCase
		if err := scanInfraredRecordCase(rows, &item); err != nil {
			return nil, infrastructurerepositoryshared.MapPgxError("failed to scan infrared_state_device_record_case", err)
		}
		result = append(result, item)
	}
	return result, nil
}

func (p *postgresImpl) ReadById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredStateDeviceRecordCase, error) {
	query, args, err := p.queryReadById(id)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build infrared_state_device_record_case get query", err)
	}

	var item domainmodels.InfraredStateDeviceRecordCase
	if err := scanInfraredRecordCase(p.Dt.QueryRow(ctx, query, args...), &item); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("infrared_state_device_record_case not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read infrared_state_device_record_case", err)
	}
	return &item, nil
}

func (p *postgresImpl) UpdateStatusById(ctx context.Context, id uuid.UUID, status domainmodels.InfraredRecordCaseStatus) error {
	query, args, err := p.queryUpdateStatusById(id, status)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build update infrared_state_device_record_case status query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to update infrared_state_device_record_case status", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("infrared_state_device_record_case not found", nil)
	}

	return nil
}

func (p *postgresImpl) DeleteById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	query, args, err := p.queryDeleteById(id, deletedBy)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build delete infrared_state_device_record_case query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to delete infrared_state_device_record_case", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("infrared_state_device_record_case not found", nil)
	}
	return nil
}

func (p *postgresImpl) ReadListStatesByCaseId(ctx context.Context, caseId uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordState, error) {
	query, args, err := p.queryReadListStatesByCaseId(caseId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build infrared_state_device_record_state list query", err)
	}

	rows, err := p.Dt.Query(ctx, query, args...)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to list infrared_state_device_record_state", err)
	}
	defer rows.Close()

	var result []domainmodels.InfraredStateDeviceRecordState
	for rows.Next() {
		var item domainmodels.InfraredStateDeviceRecordState
		if err := rows.Scan(
			&item.Id, &item.InfraredStateDeviceRecordCaseId, &item.InfraredStateId, &item.StateValue,
			&item.CreatedAt, &item.DeletedAt, &item.DeletedBy,
		); err != nil {
			return nil, infrastructurerepositoryshared.MapPgxError("failed to scan infrared_state_device_record_state", err)
		}
		result = append(result, item)
	}
	return result, nil
}

func (p *postgresImpl) DeleteStateById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	query, args, err := p.queryDeleteStateById(id, deletedBy)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build delete infrared_state_device_record_state query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to delete infrared_state_device_record_state", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("infrared_state_device_record_state not found", nil)
	}
	return nil
}

func (p *postgresImpl) CreateRaw(ctx context.Context, caseId uuid.UUID, rawData []byte) (rawId uuid.UUID, err error) {
	query, args, err := p.queryCreateRaw(caseId, rawData)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create infrared_state_device_record_raw query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&rawId); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create infrared_state_device_record_raw", err)
	}
	return rawId, nil
}

func scanInfraredRecordRaw(row pgx.Row, item *domainmodels.InfraredStateDeviceRecordRaw) error {
	return row.Scan(
		&item.Id, &item.InfraredStateDeviceRecordCaseId, &item.RawData, &item.Status, &item.DiscardedReason,
		&item.CreatedAt, &item.DeletedAt, &item.DeletedBy,
	)
}

func (p *postgresImpl) ReadListRawByCaseId(ctx context.Context, caseId uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordRaw, error) {
	query, args, err := p.queryReadListRawByCaseId(caseId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build infrared_state_device_record_raw list query", err)
	}

	rows, err := p.Dt.Query(ctx, query, args...)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to list infrared_state_device_record_raw", err)
	}
	defer rows.Close()

	var result []domainmodels.InfraredStateDeviceRecordRaw
	for rows.Next() {
		var item domainmodels.InfraredStateDeviceRecordRaw
		if err := scanInfraredRecordRaw(rows, &item); err != nil {
			return nil, infrastructurerepositoryshared.MapPgxError("failed to scan infrared_state_device_record_raw", err)
		}
		result = append(result, item)
	}
	return result, nil
}

func (p *postgresImpl) UpdateRawStatusById(ctx context.Context, id uuid.UUID, status domainmodels.InfraredRecordRawStatus, discardedReason *string) error {
	query, args, err := p.queryUpdateRawStatusById(id, status, discardedReason)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build update infrared_state_device_record_raw status query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to update infrared_state_device_record_raw status", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("infrared_state_device_record_raw not found", nil)
	}

	return nil
}

func (p *postgresImpl) DeleteRawById(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	query, args, err := p.queryDeleteRawById(id, deletedBy)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build delete infrared_state_device_record_raw query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to delete infrared_state_device_record_raw", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("infrared_state_device_record_raw not found", nil)
	}
	return nil
}

func (p *postgresImpl) ReadRawById(ctx context.Context, rawId uuid.UUID) (*domainmodels.InfraredStateDeviceRecordRaw, error) {
	query, args, err := p.queryReadRawById(rawId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build infrared_state_device_record_raw get query", err)
	}

	var item domainmodels.InfraredStateDeviceRecordRaw
	if err := scanInfraredRecordRaw(p.Dt.QueryRow(ctx, query, args...), &item); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("infrared_state_device_record_raw not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read infrared_state_device_record_raw", err)
	}
	return &item, nil
}

func (p *postgresImpl) CountAcceptedRawByCaseId(ctx context.Context, caseId uuid.UUID) (int, error) {
	query, args, err := p.queryCountAcceptedRawByCaseId(caseId)
	if err != nil {
		return 0, infrastructurerepositoryshared.QueryBuildError("failed to build count accepted infrared_state_device_record_raw query", err)
	}

	var count int
	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, infrastructurerepositoryshared.MapPgxError("failed to count accepted infrared_state_device_record_raw", err)
	}
	return count, nil
}
