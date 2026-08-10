package infrastructurerepositoryinfraredtestcase

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
) domaincontractsrepository.InfraredTestCase {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func (p *postgresImpl) CreateWithStates(ctx context.Context, coderId uuid.UUID, step int32, description string, states []domainmodels.InfraredTestCaseState) (id uuid.UUID, err error) {
	query, args, err := p.queryCreateTestCase(coderId, step, description)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create infrared_test_case query", err)
	}

	var testCaseId uuid.UUID
	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&testCaseId); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create infrared_test_case", err)
	}

	for _, state := range states {
		stateQuery, stateArgs, err := p.queryCreateTestCaseState(testCaseId, state)
		if err != nil {
			return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create infrared_test_case_state query", err)
		}
		if _, err := p.Dt.Exec(ctx, stateQuery, stateArgs...); err != nil {
			return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create infrared_test_case_state", err)
		}
	}

	return testCaseId, nil
}

func (p *postgresImpl) ListByCoderId(ctx context.Context, coderId uuid.UUID) ([]domainmodels.InfraredTestCase, error) {
	query, args, err := p.queryListByCoderId(coderId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build list infrared_test_case query", err)
	}

	rows, err := p.Dt.Query(ctx, query, args...)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to list infrared_test_case", err)
	}
	defer rows.Close()

	var result []domainmodels.InfraredTestCase
	for rows.Next() {
		var item domainmodels.InfraredTestCase
		if err := rows.Scan(&item.Id, &item.InfraredStateCoderId, &item.Step, &item.Description, &item.Status); err != nil {
			return nil, infrastructurerepositoryshared.MapPgxError("failed to scan infrared_test_case", err)
		}
		result = append(result, item)
	}
	return result, nil
}

func (p *postgresImpl) GetById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredTestCase, error) {
	query, args, err := p.queryGetById(id)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build get infrared_test_case query", err)
	}

	var item domainmodels.InfraredTestCase
	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&item.Id, &item.InfraredStateCoderId, &item.Step, &item.Description, &item.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("infrared_test_case not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to get infrared_test_case", err)
	}
	return &item, nil
}

func (p *postgresImpl) UpdateStatusById(ctx context.Context, id uuid.UUID, status domainmodels.InfraredTestCaseStatus) error {
	query, args, err := p.queryUpdateStatusById(id, status)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build update infrared_test_case status query", err)
	}

	tag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to update infrared_test_case status", err)
	}
	if tag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("infrared_test_case not found", nil)
	}
	return nil
}

func (p *postgresImpl) ListStatesByTestCaseId(ctx context.Context, testCaseId uuid.UUID) ([]domainmodels.InfraredTestCaseState, error) {
	query, args, err := p.queryListStatesByTestCaseId(testCaseId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build list infrared_test_case_state query", err)
	}

	rows, err := p.Dt.Query(ctx, query, args...)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to list infrared_test_case_state", err)
	}
	defer rows.Close()

	var result []domainmodels.InfraredTestCaseState
	for rows.Next() {
		var item domainmodels.InfraredTestCaseState
		if err := rows.Scan(&item.Id, &item.InfraredTestCaseId, &item.InfraredStateId, &item.StateValue); err != nil {
			return nil, infrastructurerepositoryshared.MapPgxError("failed to scan infrared_test_case_state", err)
		}
		result = append(result, item)
	}
	return result, nil
}
