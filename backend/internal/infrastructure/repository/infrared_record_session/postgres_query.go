package infrastructurerepositoryinfraredrecordsession

import (
	"time"

	"github.com/Masterminds/squirrel"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

var infraredRecordSessionColumns = []string{
	"id",
	"node_id",
	"infrared_device_id",
	"recording_state",
	"current_record_case_id",
	"is_completed",
	"checksum_clarification_used_at",
	"created_at",
	"created_by",
	"deleted_at",
	"deleted_by",
}

func (p *postgresImpl) queryCreate(nodeId uuid.UUID, infraredDeviceId uuid.UUID, createdBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Insert("infrared_record_session").
		Columns("node_id", "infrared_device_id", "created_by").
		Values(nodeId, infraredDeviceId, createdBy).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryReadById(id uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(infraredRecordSessionColumns...).
		From("infrared_record_session").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
}

func (p *postgresImpl) queryReadActiveByNodeId(nodeId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(infraredRecordSessionColumns...).
		From("infrared_record_session").
		Where(squirrel.Eq{
			"node_id":         nodeId,
			"recording_state": domainmodels.InfraredRecordingStateRecording,
		}).
		Where("deleted_at IS NULL").
		Limit(1).
		ToSql()
}

func (p *postgresImpl) queryUpdateRecordingStateById(id uuid.UUID, recordingState string, isCompleted bool) (query string, args []any, err error) {
	return p.SqrD.Update("infrared_record_session").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("recording_state", recordingState).
		Set("is_completed", isCompleted).
		ToSql()
}

func (p *postgresImpl) queryMarkChecksumClarificationUsedById(id uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Update("infrared_record_session").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("checksum_clarification_used_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		ToSql()
}

func (p *postgresImpl) queryUpdateCurrentRecordCaseIdById(id uuid.UUID, currentRecordCaseId *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Update("infrared_record_session").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("current_record_case_id", currentRecordCaseId).
		ToSql()
}

func (p *postgresImpl) queryReadByFilter(
	recordingState *string,
	infraredDeviceTypeId *uuid.UUID,
	createdAtStart *time.Time,
	createdAtEnd *time.Time,
	page int,
	limit int,
) (totalQuery string, totalArgs []any, query string, queryArgs []any, err error) {
	baseQ := p.SqrD.Select(
		"infrared_record_session.id",
		"infrared_record_session.recording_state",
		"infrared_record_session.is_completed",
		"infrared_device.id",
		"infrared_device.brand",
		"infrared_device.model",
		"infrared_device_type.id",
		"infrared_device_type.name",
		"infrared_record_session.created_at",
	).
		From("infrared_record_session").
		Join("infrared_device ON infrared_device.id = infrared_record_session.infrared_device_id").
		Join("infrared_device_type ON infrared_device_type.id = infrared_device.infrared_device_type_id")
	totalQ := p.SqrD.Select("COUNT(*)").
		From("infrared_record_session").
		Join("infrared_device ON infrared_device.id = infrared_record_session.infrared_device_id").
		Join("infrared_device_type ON infrared_device_type.id = infrared_device.infrared_device_type_id")

	baseQ = baseQ.Where("infrared_record_session.deleted_at IS NULL")
	totalQ = totalQ.Where("infrared_record_session.deleted_at IS NULL")

	baseQ, totalQ = applyInfraredRecordSessionFilters(baseQ, totalQ, recordingState, infraredDeviceTypeId, createdAtStart, createdAtEnd)

	totalQuery, totalArgs, err = totalQ.ToSql()
	if err != nil {
		return
	}

	query, queryArgs, err = baseQ.
		OrderBy("infrared_record_session.created_at DESC", "infrared_record_session.id ASC").
		Limit(uint64(limit)).
		Offset(uint64((page - 1) * limit)).
		ToSql()
	return
}

func applyInfraredRecordSessionFilters(
	baseQ squirrel.SelectBuilder,
	totalQ squirrel.SelectBuilder,
	recordingState *string,
	infraredDeviceTypeId *uuid.UUID,
	createdAtStart *time.Time,
	createdAtEnd *time.Time,
) (squirrel.SelectBuilder, squirrel.SelectBuilder) {
	if recordingState != nil {
		condition := squirrel.Eq{"infrared_record_session.recording_state": *recordingState}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if infraredDeviceTypeId != nil {
		condition := squirrel.Eq{"infrared_device_type.id": *infraredDeviceTypeId}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if createdAtStart != nil {
		condition := squirrel.GtOrEq{"infrared_record_session.created_at": *createdAtStart}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if createdAtEnd != nil {
		condition := squirrel.LtOrEq{"infrared_record_session.created_at": *createdAtEnd}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	return baseQ, totalQ
}

func (p *postgresImpl) queryDeleteById(id uuid.UUID, deletedBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Update("infrared_record_session").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("deleted_by", deletedBy).
		ToSql()
}
