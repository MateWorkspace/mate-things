package infrastructurerepositorynodeclass

import (
	"encoding/json"

	infrastructurerepositoryshared "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/repository/shared"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

var nodeClassColumns = []string{
	"id",
	"name",
	"description",
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
	createdBy *uuid.UUID,
) (query string, args []any, err error) {
	columns := []string{"name", "created_by"}
	values := []any{name, createdBy}

	if description != nil {
		columns = append(columns, "description")
		values = append(values, *description)
	}

	return p.SqrD.Insert("node_classes").
		Columns(columns...).
		Values(values...).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryReadById(id uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(nodeClassColumns...).
		From("node_classes").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
}

func (p *postgresImpl) queryReadByName(name string) (query string, args []any, err error) {
	return p.SqrD.Select(nodeClassColumns...).
		From("node_classes").
		Where(squirrel.Eq{"name": name}).
		Where("deleted_at IS NULL").
		Limit(1).
		ToSql()
}

func (p *postgresImpl) queryReadByPagination(
	page int,
	limit int,
	search *string,
) (totalQuery string, totalArgs []any, query string, queryArgs []any, err error) {
	baseQ := p.SqrD.Select(nodeClassColumns...).
		From("node_classes").
		Where("deleted_at IS NULL")
	totalQ := p.SqrD.Select("COUNT(*)").
		From("node_classes").
		Where("deleted_at IS NULL")

	if pattern, ok := infrastructurerepositoryshared.SearchPattern(search); ok {
		condition := squirrel.Expr("name ILIKE ?", pattern)
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
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) (query string, args []any, err error) {
	q := p.SqrD.Update("node_classes").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL")

	if name != nil {
		q = q.Set("name", *name)
	}
	if description != nil {
		q = q.Set("description", *description)
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
	return p.SqrD.Update("node_classes").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("deleted_by", deletedBy).
		ToSql()
}
