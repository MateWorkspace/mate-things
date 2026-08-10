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
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

var infraredRecordStateColumns = []string{
	"id",
	"infrared_state_device_record_case_id",
	"infrared_state_id",
	"state_value",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

var infraredRecordRawColumns = []string{
	"id",
	"infrared_state_device_record_case_id",
	"raw_data",
	"status",
	"discarded_reason",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

func (p *postgresImpl) queryCreateCase(sessionId uuid.UUID, step int32, description string, createdBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Insert("infrared_state_device_record_case").
		Columns("infrared_record_session_id", "step", "description", "created_by").
		Values(sessionId, step, description, createdBy).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryCreateState(caseId uuid.UUID, state domainmodels.InfraredStateDeviceRecordState, createdBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Insert("infrared_state_device_record_state").
		Columns("infrared_state_device_record_case_id", "infrared_state_id", "state_value", "created_by").
		Values(caseId, state.InfraredStateId, state.StateValue, createdBy).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryListBySessionId(sessionId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(infraredRecordCaseColumns...).
		From("infrared_state_device_record_case").
		Where(squirrel.Eq{"infrared_record_session_id": sessionId}).
		Where("deleted_at IS NULL").
		OrderBy("step").
		ToSql()
}

func (p *postgresImpl) queryGetById(id uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(infraredRecordCaseColumns...).
		From("infrared_state_device_record_case").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
}

func (p *postgresImpl) queryUpdateStatusById(id uuid.UUID, status domainmodels.InfraredRecordCaseStatus, updatedBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Update("infrared_state_device_record_case").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("status", status).
		Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("updated_by", updatedBy).
		ToSql()
}

func (p *postgresImpl) queryDeleteById(id uuid.UUID, deletedBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Update("infrared_state_device_record_case").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("deleted_by", deletedBy).
		ToSql()
}

func (p *postgresImpl) queryListStatesByCaseId(caseId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(infraredRecordStateColumns...).
		From("infrared_state_device_record_state").
		Where(squirrel.Eq{"infrared_state_device_record_case_id": caseId}).
		Where("deleted_at IS NULL").
		ToSql()
}

func (p *postgresImpl) queryDeleteStateById(id uuid.UUID, deletedBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Update("infrared_state_device_record_state").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("deleted_by", deletedBy).
		ToSql()
}

func (p *postgresImpl) queryCreateRaw(caseId uuid.UUID, rawData []byte, createdBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Insert("infrared_state_device_record_raw").
		Columns("infrared_state_device_record_case_id", "raw_data", "created_by").
		Values(caseId, rawData, createdBy).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryListRawByCaseId(caseId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(infraredRecordRawColumns...).
		From("infrared_state_device_record_raw").
		Where(squirrel.Eq{"infrared_state_device_record_case_id": caseId}).
		Where("deleted_at IS NULL").
		ToSql()
}

func (p *postgresImpl) queryUpdateRawStatusById(id uuid.UUID, status domainmodels.InfraredRecordRawStatus, discardedReason *string, updatedBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Update("infrared_state_device_record_raw").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("status", status).
		Set("discarded_reason", discardedReason).
		Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("updated_by", updatedBy).
		ToSql()
}

func (p *postgresImpl) queryDeleteRawById(id uuid.UUID, deletedBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Update("infrared_state_device_record_raw").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("deleted_by", deletedBy).
		ToSql()
}

func (p *postgresImpl) queryGetRawById(rawId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(infraredRecordRawColumns...).
		From("infrared_state_device_record_raw").
		Where(squirrel.Eq{"id": rawId}).
		Where("deleted_at IS NULL").
		ToSql()
}

func (p *postgresImpl) queryCountAcceptedRawByCaseId(caseId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select("COUNT(*)").
		From("infrared_state_device_record_raw").
		Where(squirrel.Eq{
			"infrared_state_device_record_case_id": caseId,
			"status":                               domainmodels.InfraredRecordRawStatusAccepted,
		}).
		Where("deleted_at IS NULL").
		ToSql()
}
