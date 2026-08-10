package infrastructurerepositoryinfraredrecordsession

import (
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

func (p *postgresImpl) queryUpdateCurrentRecordCaseIdById(id uuid.UUID, currentRecordCaseId *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Update("infrared_record_session").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("current_record_case_id", currentRecordCaseId).
		ToSql()
}

func (p *postgresImpl) queryDeleteById(id uuid.UUID, deletedBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Update("infrared_record_session").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("deleted_by", deletedBy).
		ToSql()
}
