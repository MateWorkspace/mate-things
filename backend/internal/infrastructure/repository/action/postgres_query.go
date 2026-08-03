package infrastructurerepositoryaction

import (
	"encoding/json"

	"github.com/Masterminds/squirrel"
	infrastructurerepositoryshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/shared"
	"github.com/google/uuid"
)

var actionColumns = []string{
	"id",
	"name",
	"description",
	"payload_schema_name",
	"payload_schema_version",
	"preferences",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

func (p *postgresImpl) queryCreate(
	name string,
	description *string,
	payloadSchemaName string,
	payloadSchemaVersion int32,
	createdBy *uuid.UUID,
) (query string, args []any, err error) {
	columns := []string{"name", "payload_schema_name", "payload_schema_version", "created_by"}
	values := []any{name, payloadSchemaName, payloadSchemaVersion, createdBy}

	if description != nil {
		columns = append(columns, "description")
		values = append(values, *description)
	}

	return p.SqrD.Insert("actions").
		Columns(columns...).
		Values(values...).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryReadById(id uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(actionColumns...).
		From("actions").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
}

func (p *postgresImpl) queryReadByName(name string) (query string, args []any, err error) {
	return p.SqrD.Select(actionColumns...).
		From("actions").
		Where(squirrel.Eq{"name": name}).
		Where("deleted_at IS NULL").
		Limit(1).
		ToSql()
}

func (p *postgresImpl) queryReadByPagination(
	page int,
	limit int,
	search *string,
	nodeClassId *uuid.UUID,
	payloadSchemaName *string,
	payloadSchemaVersion *int32,
) (totalQuery string, totalArgs []any, query string, queryArgs []any, err error) {
	selectColumns := append([]string{}, actionColumns...)
	selectColumns = append(selectColumns, "(SELECT COUNT(*) FROM node_class_action nca JOIN node_classes nc ON nc.id = nca.node_class_id WHERE nca.action_id = actions.id AND nc.deleted_at IS NULL) AS compatible_node_class_count")

	baseQ := p.SqrD.Select(selectColumns...).
		From("actions").
		Where("deleted_at IS NULL")
	totalQ := p.SqrD.Select("COUNT(*)").
		From("actions").
		Where("deleted_at IS NULL")

	if pattern, ok := infrastructurerepositoryshared.SearchPattern(search); ok {
		condition := squirrel.Expr("name ILIKE ?", pattern)
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if nodeClassId != nil {
		condition := squirrel.Expr(
			"EXISTS (SELECT 1 FROM node_class_action nca WHERE nca.action_id = actions.id AND nca.node_class_id = ?)",
			*nodeClassId,
		)
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if payloadSchemaName != nil {
		condition := squirrel.Eq{"payload_schema_name": *payloadSchemaName}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if payloadSchemaVersion != nil {
		condition := squirrel.Eq{"payload_schema_version": *payloadSchemaVersion}
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
	name *string,
	description *string,
	payloadSchemaName *string,
	payloadSchemaVersion *int32,
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) (query string, args []any, err error) {
	q := p.SqrD.Update("actions").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL")

	if name != nil {
		q = q.Set("name", *name)
	}
	if description != nil {
		q = q.Set("description", *description)
	}
	if payloadSchemaName != nil {
		q = q.Set("payload_schema_name", *payloadSchemaName)
	}
	if payloadSchemaVersion != nil {
		q = q.Set("payload_schema_version", *payloadSchemaVersion)
	}
	if preferences != nil {
		q = q.Set("preferences", *preferences)
	}

	return q.
		Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("updated_by", updatedBy).
		ToSql()
}

func (p *postgresImpl) queryDeleteById(
	id uuid.UUID,
	deletedBy *uuid.UUID,
) (query string, args []any, err error) {
	return p.SqrD.Update("actions").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("deleted_by", deletedBy).
		ToSql()
}
