package infrastructurerepositoryrolepermission

import (
	"github.com/Masterminds/squirrel"
	infrastructurerepositoryshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/shared"
	"github.com/google/uuid"
)

func (p *postgresImpl) queryCreate(
	roleId uuid.UUID,
	permissionId uuid.UUID,
	createdBy *uuid.UUID,
) (query string, args []any, err error) {
	return p.SqrD.Insert("role_permission").
		Columns("role_id", "permission_id", "created_by").
		Values(roleId, permissionId, createdBy).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryReadById(id uuid.UUID) (query string, args []any, err error) {
	return p.baseReadQuery().
		Where(squirrel.Eq{"rp.id": id}).
		ToSql()
}

func (p *postgresImpl) queryReadByRoleIdAndPermissionId(
	roleId uuid.UUID,
	permissionId uuid.UUID,
) (query string, args []any, err error) {
	return p.baseReadQuery().
		Where(squirrel.Eq{"rp.role_id": roleId}).
		Where(squirrel.Eq{"rp.permission_id": permissionId}).
		Limit(1).
		ToSql()
}

func (p *postgresImpl) queryReadByPagination(
	page int,
	limit int,
	roleId *uuid.UUID,
	permissionId *uuid.UUID,
) (totalQuery string, totalArgs []any, query string, queryArgs []any, err error) {
	baseQ := p.baseReadQuery()
	totalQ := p.SqrD.Select("COUNT(*)").
		From("role_permission rp").
		Join("roles r ON rp.role_id = r.id").
		Join("permissions p ON rp.permission_id = p.id").
		Where("r.deleted_at IS NULL").
		Where("p.deleted_at IS NULL")

	if roleId != nil {
		condition := squirrel.Eq{"rp.role_id": *roleId}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if permissionId != nil {
		condition := squirrel.Eq{"rp.permission_id": *permissionId}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}

	totalQuery, totalArgs, err = totalQ.ToSql()
	if err != nil {
		return
	}

	query, queryArgs, err = baseQ.
		OrderBy("rp.created_at DESC", "rp.id ASC").
		Limit(infrastructurerepositoryshared.NormalizeLimit(limit)).
		Offset(infrastructurerepositoryshared.NormalizeOffset(page, limit)).
		ToSql()
	return
}

func (p *postgresImpl) queryDeleteById(id uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Delete("role_permission").
		Where(squirrel.Eq{"id": id}).
		ToSql()
}

func (p *postgresImpl) queryDeleteByRoleIdAndPermissionId(
	roleId *uuid.UUID,
	permissionId *uuid.UUID,
) (query string, args []any, err error) {
	q := p.SqrD.Delete("role_permission")

	if roleId != nil {
		q = q.Where(squirrel.Eq{"role_id": *roleId})
	}
	if permissionId != nil {
		q = q.Where(squirrel.Eq{"permission_id": *permissionId})
	}

	return q.ToSql()
}

func (p *postgresImpl) baseReadQuery() squirrel.SelectBuilder {
	return p.SqrD.Select(
		"rp.id",
		"rp.role_id",
		"rp.permission_id",
		"rp.created_at",
		"rp.created_by",
		"r.id",
		"r.name",
		"r.description",
		"r.is_default",
		"r.preferences",
		"r.created_at",
		"r.updated_at",
		"r.deleted_at",
		"r.created_by",
		"r.updated_by",
		"r.deleted_by",
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
	).
		From("role_permission rp").
		Join("roles r ON rp.role_id = r.id").
		Join("permissions p ON rp.permission_id = p.id").
		Where("r.deleted_at IS NULL").
		Where("p.deleted_at IS NULL")
}
