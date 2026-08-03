package infrastructurerepositorynodeclassaction

import (
	"github.com/Masterminds/squirrel"
	infrastructurerepositoryshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/shared"
	"github.com/google/uuid"
)

func (p *postgresImpl) queryCreate(
	nodeClassId uuid.UUID,
	actionId uuid.UUID,
	createdBy *uuid.UUID,
) (query string, args []any, err error) {
	return p.SqrD.Insert("node_class_action").
		Columns("node_class_id", "action_id", "created_by").
		Values(nodeClassId, actionId, createdBy).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryReadById(id uuid.UUID) (query string, args []any, err error) {
	return p.baseReadQuery().
		Where(squirrel.Eq{"nca.id": id}).
		ToSql()
}

func (p *postgresImpl) queryReadByNodeClassIdAndActionId(
	nodeClassId uuid.UUID,
	actionId uuid.UUID,
) (query string, args []any, err error) {
	return p.baseReadQuery().
		Where(squirrel.Eq{"nca.node_class_id": nodeClassId}).
		Where(squirrel.Eq{"nca.action_id": actionId}).
		Limit(1).
		ToSql()
}

func (p *postgresImpl) queryReadByPagination(
	page int,
	limit int,
	nodeClassId *uuid.UUID,
	actionId *uuid.UUID,
) (totalQuery string, totalArgs []any, query string, queryArgs []any, err error) {
	baseQ := p.baseReadQuery()
	totalQ := p.SqrD.Select("COUNT(*)").
		From("node_class_action nca").
		Join("node_classes nc ON nca.node_class_id = nc.id").
		Join("actions a ON nca.action_id = a.id").
		Where("nc.deleted_at IS NULL").
		Where("a.deleted_at IS NULL")

	if nodeClassId != nil {
		condition := squirrel.Eq{"nca.node_class_id": *nodeClassId}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if actionId != nil {
		condition := squirrel.Eq{"nca.action_id": *actionId}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}

	totalQuery, totalArgs, err = totalQ.ToSql()
	if err != nil {
		return
	}

	query, queryArgs, err = baseQ.
		OrderBy("nca.created_at DESC", "nca.id ASC").
		Limit(infrastructurerepositoryshared.NormalizeLimit(limit)).
		Offset(infrastructurerepositoryshared.NormalizeOffset(page, limit)).
		ToSql()
	return
}

func (p *postgresImpl) queryDeleteById(id uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Delete("node_class_action").
		Where(squirrel.Eq{"id": id}).
		ToSql()
}

func (p *postgresImpl) queryDeleteByNodeClassIdAndActionId(
	nodeClassId *uuid.UUID,
	actionId *uuid.UUID,
) (query string, args []any, err error) {
	q := p.SqrD.Delete("node_class_action")

	if nodeClassId != nil {
		q = q.Where(squirrel.Eq{"node_class_id": *nodeClassId})
	}
	if actionId != nil {
		q = q.Where(squirrel.Eq{"action_id": *actionId})
	}

	return q.ToSql()
}

func (p *postgresImpl) baseReadQuery() squirrel.SelectBuilder {
	return p.SqrD.Select(
		"nca.id",
		"nca.node_class_id",
		"nca.action_id",
		"nca.created_at",
		"nca.created_by",
		"nc.id",
		"nc.name",
		"nc.description",
		"nc.preferences",
		"nc.created_at",
		"nc.updated_at",
		"nc.deleted_at",
		"nc.created_by",
		"nc.updated_by",
		"nc.deleted_by",
		"a.id",
		"a.name",
		"a.description",
		"a.payload_schema_name",
		"a.payload_schema_version",
		"a.preferences",
		"a.created_at",
		"a.updated_at",
		"a.deleted_at",
		"a.created_by",
		"a.updated_by",
		"a.deleted_by",
	).
		From("node_class_action nca").
		Join("node_classes nc ON nca.node_class_id = nc.id").
		Join("actions a ON nca.action_id = a.id").
		Where("nc.deleted_at IS NULL").
		Where("a.deleted_at IS NULL")
}
