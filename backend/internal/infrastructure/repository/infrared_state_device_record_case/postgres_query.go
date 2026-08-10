package infrastructurerepositoryinfraredstatedevicerecordcase

import (
	"github.com/Masterminds/squirrel"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

var infraredRecordCaseColumns = []string{
	"id",
	"infrared_record_session_id",
	"step",
	"description",
	"status",
}

var infraredRecordStateColumns = []string{
	"id",
	"infrared_state_device_record_case_id",
	"infrared_state_id",
	"state_value",
}

var infraredRecordRawColumns = []string{
	"id",
	"infrared_state_device_record_case_id",
	"raw_data",
	"status",
	"discarded_reason",
}

func (p *postgresImpl) queryCreateCase(sessionId uuid.UUID, step int32, description string) (query string, args []any, err error) {
	return p.SqrD.Insert("infrared_state_device_record_case").
		Columns("infrared_record_session_id", "step", "description").
		Values(sessionId, step, description).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryCreateState(caseId uuid.UUID, state domainmodels.InfraredStateDeviceRecordState) (query string, args []any, err error) {
	return p.SqrD.Insert("infrared_state_device_record_state").
		Columns("infrared_state_device_record_case_id", "infrared_state_id", "state_value").
		Values(caseId, state.InfraredStateId, state.StateValue).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryListBySessionId(sessionId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(infraredRecordCaseColumns...).
		From("infrared_state_device_record_case").
		Where(squirrel.Eq{"infrared_record_session_id": sessionId}).
		OrderBy("step").
		ToSql()
}

func (p *postgresImpl) queryGetById(id uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(infraredRecordCaseColumns...).
		From("infrared_state_device_record_case").
		Where(squirrel.Eq{"id": id}).
		ToSql()
}

func (p *postgresImpl) queryUpdateStatusById(id uuid.UUID, status domainmodels.InfraredRecordCaseStatus) (query string, args []any, err error) {
	return p.SqrD.Update("infrared_state_device_record_case").
		Where(squirrel.Eq{"id": id}).
		Set("status", status).
		ToSql()
}

func (p *postgresImpl) queryListStatesByCaseId(caseId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(infraredRecordStateColumns...).
		From("infrared_state_device_record_state").
		Where(squirrel.Eq{"infrared_state_device_record_case_id": caseId}).
		ToSql()
}

func (p *postgresImpl) queryCreateRaw(caseId uuid.UUID, rawData []byte) (query string, args []any, err error) {
	return p.SqrD.Insert("infrared_state_device_record_raw").
		Columns("infrared_state_device_record_case_id", "raw_data").
		Values(caseId, rawData).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryListRawByCaseId(caseId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(infraredRecordRawColumns...).
		From("infrared_state_device_record_raw").
		Where(squirrel.Eq{"infrared_state_device_record_case_id": caseId}).
		ToSql()
}

func (p *postgresImpl) queryUpdateRawStatusById(id uuid.UUID, status domainmodels.InfraredRecordRawStatus, discardedReason *string) (query string, args []any, err error) {
	return p.SqrD.Update("infrared_state_device_record_raw").
		Where(squirrel.Eq{"id": id}).
		Set("status", status).
		Set("discarded_reason", discardedReason).
		ToSql()
}

func (p *postgresImpl) queryGetRawById(rawId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(
		"raw.id",
		"raw.infrared_state_device_record_case_id",
		"cases.infrared_record_session_id",
		"raw.raw_data",
		"raw.status",
		"raw.discarded_reason",
	).
		From("infrared_state_device_record_raw raw").
		Join("infrared_state_device_record_case cases ON cases.id = raw.infrared_state_device_record_case_id").
		Where(squirrel.Eq{"raw.id": rawId}).
		ToSql()
}

func (p *postgresImpl) queryCountAcceptedRawByCaseId(caseId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select("COUNT(*)").
		From("infrared_state_device_record_raw").
		Where(squirrel.Eq{
			"infrared_state_device_record_case_id": caseId,
			"status":                               domainmodels.InfraredRecordRawStatusAccepted,
		}).
		ToSql()
}
