package infrastructurerepositorynode

import (
	"encoding/json"

	infrastructurerepositoryshared "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/repository/shared"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

var nodeColumns = []string{
	"id",
	"node_class_id",
	"device_id",
	"device_info",
	"name",
	"firmware_name",
	"description",
	"is_connected",
	"preferences",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

func (p *postgresImpl) queryCreate(
	nodeClassId uuid.UUID,
	deviceId string,
	deviceInfo string,
	name string,
	firmwareName string,
	description *string,
	isConnected bool,
	createdBy *uuid.UUID,
) (query string, args []any, err error) {
	columns := []string{"node_class_id", "device_id", "device_info", "name", "firmware_name", "is_connected", "created_by"}
	values := []any{nodeClassId, deviceId, deviceInfo, name, firmwareName, isConnected, createdBy}

	if description != nil {
		columns = append(columns, "description")
		values = append(values, *description)
	}

	return p.SqrD.Insert("nodes").
		Columns(columns...).
		Values(values...).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryUpsertRegistration(
	deviceId string,
	deviceInfo string,
	firmwareName string,
) (query string, args []any, err error) {
	return `
		INSERT INTO nodes (
			node_class_id,
			device_id,
			device_info,
			name,
			firmware_name,
			is_connected
		)
		SELECT
			f.node_class_id,
			$1,
			$2,
			$3,
			f.name,
			TRUE
		FROM firmwares f
		WHERE f.name = $4
			AND f.deleted_at IS NULL
		LIMIT 1
		ON CONFLICT (device_id) DO UPDATE SET
			node_class_id = EXCLUDED.node_class_id,
			device_info = EXCLUDED.device_info,
			firmware_name = EXCLUDED.firmware_name,
			is_connected = TRUE,
			updated_at = CURRENT_TIMESTAMP
		WHERE nodes.deleted_at IS NULL
		RETURNING (xmax = 0) AS created, ` + joinNodeColumns("nodes"),
		[]any{deviceId, deviceInfo, "node_" + deviceId, firmwareName},
		nil
}

func (p *postgresImpl) queryReadById(id uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(nodeColumns...).
		From("nodes").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
}

func (p *postgresImpl) queryReadByDeviceId(deviceId string) (query string, args []any, err error) {
	return p.SqrD.Select(nodeColumns...).
		From("nodes").
		Where(squirrel.Eq{"device_id": deviceId}).
		Where("deleted_at IS NULL").
		Limit(1).
		ToSql()
}

func (p *postgresImpl) queryReadByPagination(
	page int,
	limit int,
	search *string,
	nodeClassId *uuid.UUID,
	firmwareName *string,
) (totalQuery string, totalArgs []any, query string, queryArgs []any, err error) {
	baseQ := p.SqrD.Select(nodeColumns...).
		From("nodes").
		Where("deleted_at IS NULL")
	totalQ := p.SqrD.Select("COUNT(*)").
		From("nodes").
		Where("deleted_at IS NULL")

	if pattern, ok := infrastructurerepositoryshared.SearchPattern(search); ok {
		condition := squirrel.Or{
			squirrel.Expr("device_id ILIKE ?", pattern),
			squirrel.Expr("device_info ILIKE ?", pattern),
			squirrel.Expr("name ILIKE ?", pattern),
		}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if nodeClassId != nil {
		condition := squirrel.Eq{"node_class_id": *nodeClassId}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if firmwareName != nil {
		condition := squirrel.Eq{"firmware_name": *firmwareName}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}

	totalQuery, totalArgs, err = totalQ.ToSql()
	if err != nil {
		return
	}

	query, queryArgs, err = baseQ.
		OrderBy("created_at DESC", "id ASC").
		Limit(infrastructurerepositoryshared.NormalizeLimit(limit)).
		Offset(infrastructurerepositoryshared.NormalizeOffset(page, limit)).
		ToSql()
	return
}

func (p *postgresImpl) queryUpdateById(
	id uuid.UUID,
	nodeClassId *uuid.UUID,
	deviceId *string,
	deviceInfo *string,
	name *string,
	firmwareName *string,
	description *string,
	isConnected *bool,
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) (query string, args []any, err error) {
	q := p.SqrD.Update("nodes").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL")

	if nodeClassId != nil {
		q = q.Set("node_class_id", *nodeClassId)
	}
	if deviceId != nil {
		q = q.Set("device_id", *deviceId)
	}
	if deviceInfo != nil {
		q = q.Set("device_info", *deviceInfo)
	}
	if name != nil {
		q = q.Set("name", *name)
	}
	if firmwareName != nil {
		q = q.Set("firmware_name", *firmwareName)
	}
	if description != nil {
		q = q.Set("description", *description)
	}
	if isConnected != nil {
		q = q.Set("is_connected", *isConnected)
	}
	if preferences != nil {
		q = q.Set("preferences", *preferences)
	}

	return q.
		Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("updated_by", updatedBy).
		ToSql()
}

func joinNodeColumns(prefix string) string {
	result := ""
	for i, column := range nodeColumns {
		if i > 0 {
			result += ", "
		}
		result += prefix + "." + column
	}
	return result
}

func (p *postgresImpl) queryDeleteById(
	id uuid.UUID,
	deletedBy *uuid.UUID,
) (query string, args []any, err error) {
	return p.SqrD.Update("nodes").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("deleted_by", deletedBy).
		ToSql()
}
