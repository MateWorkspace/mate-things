package infrastructurerepositoryuser

import (
	"encoding/json"

	"github.com/Masterminds/squirrel"
	infrastructurerepositoryshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/shared"
	"github.com/google/uuid"
)

var userColumns = []string{
	"id",
	"role_id",
	"name",
	"bio",
	"username",
	"password_hash",
	"preferences",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

var userPermissionColumns = []string{
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
	roleId uuid.UUID,
	name string,
	bio *string,
	username string,
	passwordHash string,
	createdBy *uuid.UUID,
) (query string, args []any, err error) {
	columns := []string{"role_id", "name", "username", "password_hash", "created_by"}
	values := []any{roleId, name, username, passwordHash, createdBy}

	if bio != nil {
		columns = append(columns, "bio")
		values = append(values, *bio)
	}

	return p.SqrD.Insert("users").
		Columns(columns...).
		Values(values...).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryReadById(id uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(userColumns...).
		From("users").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
}

func (p *postgresImpl) queryReadByUsername(username string) (query string, args []any, err error) {
	return p.SqrD.Select(userColumns...).
		From("users").
		Where(squirrel.Eq{"username": username}).
		Where("deleted_at IS NULL").
		Limit(1).
		ToSql()
}

func (p *postgresImpl) queryReadPermissions(userId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(userPermissionColumns...).
		From("users u").
		Join("roles r ON r.id = u.role_id").
		Join("role_permission rp ON rp.role_id = r.id").
		Join("permissions p ON p.id = rp.permission_id").
		Where(squirrel.Eq{"u.id": userId}).
		Where("u.deleted_at IS NULL").
		Where("r.deleted_at IS NULL").
		Where("p.deleted_at IS NULL").
		OrderBy("p.created_at DESC", "p.id ASC").
		ToSql()
}

func (p *postgresImpl) queryReadByPagination(
	page int,
	limit int,
	search *string,
	roleId *uuid.UUID,
) (totalQuery string, totalArgs []any, query string, queryArgs []any, err error) {
	baseQ := p.SqrD.Select(userColumns...).
		From("users").
		Where("deleted_at IS NULL")
	totalQ := p.SqrD.Select("COUNT(*)").
		From("users").
		Where("deleted_at IS NULL")

	if pattern, ok := infrastructurerepositoryshared.SearchPattern(search); ok {
		condition := squirrel.Or{
			squirrel.Expr("name ILIKE ?", pattern),
			squirrel.Expr("username ILIKE ?", pattern),
		}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if roleId != nil {
		condition := squirrel.Eq{"role_id": *roleId}
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
	roleId *uuid.UUID,
	name *string,
	bio *string,
	username *string,
	passwordHash *string,
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) (query string, args []any, err error) {
	q := p.SqrD.Update("users").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL")

	if roleId != nil {
		q = q.Set("role_id", *roleId)
	}
	if name != nil {
		q = q.Set("name", *name)
	}
	if bio != nil {
		q = q.Set("bio", *bio)
	}
	if username != nil {
		q = q.Set("username", *username)
	}
	if passwordHash != nil {
		q = q.Set("password_hash", *passwordHash)
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
	return p.SqrD.Update("users").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("deleted_by", deletedBy).
		ToSql()
}
