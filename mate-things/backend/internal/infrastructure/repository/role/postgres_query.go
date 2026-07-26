package infrastructurerepositoryrole

import (
	"encoding/json"

	infrastructurerepositoryshared "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/repository/shared"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

var roleColumns = []string{
	"id",
	"name",
	"description",
	"is_default",
	"preferences",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

var rolePermissionColumns = []string{
	"p.id",
	"p.name",
	"p.description",
	"p.preferences",
	"p.created_at",
	"p.updated_at",
	"p.deleted_at",
	"p.created_by",
	"p.updated_by",
	"p.deleted_by",
}

func (p *postgresImpl) queryCreate(
	name string,
	description *string,
	isDefault *bool,
	createdBy *uuid.UUID,
) (query string, args []any, err error) {
	columns := []string{"name", "created_by"}
	values := []any{name, createdBy}

	if description != nil {
		columns = append(columns, "description")
		values = append(values, *description)
	}
	if isDefault != nil {
		columns = append(columns, "is_default")
		values = append(values, *isDefault)
	}

	return p.SqrD.Insert("roles").
		Columns(columns...).
		Values(values...).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryUnsetDefault(updatedBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Update("roles").
		Set("is_default", false).
		Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("updated_by", updatedBy).
		Where(squirrel.Eq{"is_default": true}).
		Where("deleted_at IS NULL").
		ToSql()
}

func (p *postgresImpl) queryUnsetDefaultExcept(id uuid.UUID, updatedBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Update("roles").
		Set("is_default", false).
		Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("updated_by", updatedBy).
		Where(squirrel.Eq{"is_default": true}).
		Where(squirrel.NotEq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
}

func (p *postgresImpl) queryReadById(id uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(roleColumns...).
		From("roles").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
}

func (p *postgresImpl) queryReadByName(name string) (query string, args []any, err error) {
	return p.SqrD.Select(roleColumns...).
		From("roles").
		Where(squirrel.Eq{"name": name}).
		Where("deleted_at IS NULL").
		Limit(1).
		ToSql()
}

func (p *postgresImpl) queryReadDefault() (query string, args []any, err error) {
	return p.SqrD.Select(roleColumns...).
		From("roles").
		Where(squirrel.Eq{"is_default": true}).
		Where("deleted_at IS NULL").
		Limit(1).
		ToSql()
}

func (p *postgresImpl) queryReadPermissions(roleId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(rolePermissionColumns...).
		From("role_permission rp").
		Join("permissions p ON p.id = rp.permission_id").
		Where(squirrel.Eq{"rp.role_id": roleId}).
		Where("p.deleted_at IS NULL").
		OrderBy("p.created_at DESC", "p.id ASC").
		ToSql()
}

func (p *postgresImpl) queryReadByPagination(
	page int,
	limit int,
	search *string,
) (totalQuery string, totalArgs []any, query string, queryArgs []any, err error) {
	baseQ := p.SqrD.Select(roleColumns...).
		From("roles").
		Where("deleted_at IS NULL")
	totalQ := p.SqrD.Select("COUNT(*)").
		From("roles").
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
	isDefault *bool,
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) (query string, args []any, err error) {
	q := p.SqrD.Update("roles").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL")

	if name != nil {
		q = q.Set("name", *name)
	}
	if description != nil {
		q = q.Set("description", *description)
	}
	if isDefault != nil {
		q = q.Set("is_default", *isDefault)
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
	return p.SqrD.Update("roles").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("deleted_by", deletedBy).
		ToSql()
}
