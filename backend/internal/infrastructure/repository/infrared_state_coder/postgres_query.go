package infrastructurerepositoryinfraredstatecoder

import (
	"github.com/Masterminds/squirrel"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

var infraredStateCoderColumns = []string{
	"id",
	"infrared_device_id",
	"infrared_record_session_id",
	"encoder_source",
	"decoder_source",
	"summary_readme",
	"detail_readme",
	"status",
	"created_at",
	"deleted_at",
	"deleted_by",
}

func (p *postgresImpl) queryCreate(coder domainmodels.InfraredStateCoder) (query string, args []any, err error) {
	return p.SqrD.Insert("infrared_state_coder").
		Columns("infrared_device_id", "infrared_record_session_id", "encoder_source", "decoder_source", "summary_readme", "detail_readme").
		Values(coder.InfraredDeviceId, coder.InfraredRecordSessionId, coder.EncoderSource, coder.DecoderSource, coder.SummaryReadme, coder.DetailReadme).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryReadById(id uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(infraredStateCoderColumns...).
		From("infrared_state_coder").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
}

func (p *postgresImpl) queryReadBySessionId(sessionId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(infraredStateCoderColumns...).
		From("infrared_state_coder").
		Where(squirrel.Eq{"infrared_record_session_id": sessionId}).
		Where("deleted_at IS NULL").
		OrderBy("created_at DESC").
		Limit(1).
		ToSql()
}

func (p *postgresImpl) queryDeleteById(id uuid.UUID, deletedBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Update("infrared_state_coder").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("deleted_by", deletedBy).
		ToSql()
}
