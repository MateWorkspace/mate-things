package infrastructurerepositorynodeconfigvalue

import (
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

var nodeConfigValueColumns = []string{
	"id",
	"node_id",
	"firmware_id",
	"key",
	"value",
	"preferences",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

func (p *postgresImpl) queryReadByNodeId(nodeId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(nodeConfigValueColumns...).
		From("node_config_values").
		Where(squirrel.Eq{"node_id": nodeId}).
		Where("deleted_at IS NULL").
		OrderBy("key ASC").
		ToSql()
}

func (p *postgresImpl) queryReadByNodeIdAndKey(nodeId uuid.UUID, key string) (query string, args []any, err error) {
	return p.SqrD.Select(nodeConfigValueColumns...).
		From("node_config_values").
		Where(squirrel.Eq{"node_id": nodeId, "key": key}).
		Where("deleted_at IS NULL").
		Limit(1).
		ToSql()
}

func (p *postgresImpl) queryUpsert(
	nodeId uuid.UUID,
	firmwareId uuid.UUID,
	key string,
	value string,
	actorId *uuid.UUID,
) (query string, args []any, err error) {
	return p.SqrD.Insert("node_config_values").
		Columns("node_id", "firmware_id", "key", "value", "created_by").
		Values(nodeId, firmwareId, key, value, actorId).
		Suffix("ON CONFLICT (node_id, key) DO UPDATE SET firmware_id = EXCLUDED.firmware_id, value = EXCLUDED.value, deleted_at = NULL, deleted_by = NULL, updated_at = CURRENT_TIMESTAMP, updated_by = ?", actorId).
		ToSql()
}
