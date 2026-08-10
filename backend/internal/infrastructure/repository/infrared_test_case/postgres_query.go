package infrastructurerepositoryinfraredtestcase

import (
	"github.com/Masterminds/squirrel"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

var infraredTestCaseColumns = []string{
	"id",
	"infrared_state_coder_id",
	"step",
	"description",
	"status",
}

var infraredTestCaseStateColumns = []string{
	"id",
	"infrared_test_case_id",
	"infrared_state_id",
	"state_value",
}

func (p *postgresImpl) queryCreateTestCase(coderId uuid.UUID, step int32, description string) (query string, args []any, err error) {
	return p.SqrD.Insert("infrared_test_case").
		Columns("infrared_state_coder_id", "step", "description").
		Values(coderId, step, description).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryCreateTestCaseState(testCaseId uuid.UUID, state domainmodels.InfraredTestCaseState) (query string, args []any, err error) {
	return p.SqrD.Insert("infrared_test_case_state").
		Columns("infrared_test_case_id", "infrared_state_id", "state_value").
		Values(testCaseId, state.InfraredStateId, state.StateValue).
		ToSql()
}

func (p *postgresImpl) queryListByCoderId(coderId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(infraredTestCaseColumns...).
		From("infrared_test_case").
		Where(squirrel.Eq{"infrared_state_coder_id": coderId}).
		OrderBy("step").
		ToSql()
}

func (p *postgresImpl) queryGetById(id uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(infraredTestCaseColumns...).
		From("infrared_test_case").
		Where(squirrel.Eq{"id": id}).
		ToSql()
}

func (p *postgresImpl) queryUpdateStatusById(id uuid.UUID, status domainmodels.InfraredTestCaseStatus) (query string, args []any, err error) {
	return p.SqrD.Update("infrared_test_case").
		Where(squirrel.Eq{"id": id}).
		Set("status", status).
		ToSql()
}

func (p *postgresImpl) queryListStatesByTestCaseId(testCaseId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(infraredTestCaseStateColumns...).
		From("infrared_test_case_state").
		Where(squirrel.Eq{"infrared_test_case_id": testCaseId}).
		ToSql()
}
