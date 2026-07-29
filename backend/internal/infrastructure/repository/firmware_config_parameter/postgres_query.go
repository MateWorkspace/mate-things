package infrastructurerepositoryfirmwareconfigparameter

import (
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

var firmwareConfigParameterColumns = []string{
	"id",
	"firmware_id",
	"key",
	"value_type",
	"preferences",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

func (p *postgresImpl) queryReadByFirmwareId(firmwareId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(firmwareConfigParameterColumns...).
		From("firmware_config_parameters").
		Where(squirrel.Eq{"firmware_id": firmwareId}).
		Where("deleted_at IS NULL").
		OrderBy("key ASC").
		ToSql()
}

func (p *postgresImpl) querySoftDeleteMissing(
	firmwareId uuid.UUID,
	keepKeys []string,
	actorId *uuid.UUID,
) (query string, args []any, err error) {
	q := p.SqrD.Update("firmware_config_parameters").
		Where(squirrel.Eq{"firmware_id": firmwareId}).
		Where("deleted_at IS NULL").
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("deleted_by", actorId)

	if len(keepKeys) > 0 {
		q = q.Where(squirrel.NotEq{"key": keepKeys})
	}

	return q.ToSql()
}

func (p *postgresImpl) queryUpsert(
	firmwareId uuid.UUID,
	key string,
	valueType string,
	actorId *uuid.UUID,
) (query string, args []any, err error) {
	return p.SqrD.Insert("firmware_config_parameters").
		Columns("firmware_id", "key", "value_type", "created_by").
		Values(firmwareId, key, valueType, actorId).
		Suffix("ON CONFLICT (firmware_id, key) DO UPDATE SET value_type = EXCLUDED.value_type, deleted_at = NULL, deleted_by = NULL, updated_at = CURRENT_TIMESTAMP, updated_by = ?", actorId).
		ToSql()
}
