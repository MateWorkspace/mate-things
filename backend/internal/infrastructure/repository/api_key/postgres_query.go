package infrastructurerepositoryapikey

import (
	"time"

	"github.com/Masterminds/squirrel"
	infrastructurerepositoryshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/shared"
	"github.com/google/uuid"
)

var apiKeyColumns = []string{
	"id",
	"user_id",
	"key_hash",
	"key_last_four",
	"expires_at",
	"revoked_at",
	"created_at",
	"updated_at",
	"created_by",
	"updated_by",
}

var apiKeyWithUserColumns = []string{
	"ak.id",
	"ak.user_id",
	"ak.key_hash",
	"ak.key_last_four",
	"ak.expires_at",
	"ak.revoked_at",
	"ak.created_at",
	"ak.updated_at",
	"ak.created_by",
	"ak.updated_by",
	"u.name AS user_name",
	"u.username AS user_username",
}

func (p *postgresImpl) queryCreate(
	userId uuid.UUID,
	keyHash string,
	keyLastFour string,
	expiresAt *time.Time,
	createdBy *uuid.UUID,
) (query string, args []any, err error) {
	return p.SqrD.Insert("api_keys").
		Columns("user_id", "key_hash", "key_last_four", "expires_at", "created_by").
		Values(userId, keyHash, keyLastFour, expiresAt, createdBy).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryReadByKeyHash(keyHash string) (query string, args []any, err error) {
	return p.SqrD.Select(apiKeyColumns...).
		From("api_keys").
		Where(squirrel.Eq{"key_hash": keyHash}).
		ToSql()
}

func (p *postgresImpl) queryReadByPagination(
	page int,
	limit int,
	search *string,
	status *string,
) (totalQuery string, totalArgs []any, query string, queryArgs []any, err error) {
	baseQ := p.SqrD.Select(apiKeyWithUserColumns...).
		From("api_keys ak").
		Join("users u ON u.id = ak.user_id")
	totalQ := p.SqrD.Select("COUNT(*)").
		From("api_keys ak").
		Join("users u ON u.id = ak.user_id")

	if pattern, ok := infrastructurerepositoryshared.SearchPattern(search); ok {
		condition := squirrel.Or{
			squirrel.Expr("u.name ILIKE ?", pattern),
			squirrel.Expr("u.username ILIKE ?", pattern),
		}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}

	if status != nil {
		switch *status {
		case "active":
			condition := squirrel.And{
				squirrel.Eq{"ak.revoked_at": nil},
				squirrel.Or{
					squirrel.Eq{"ak.expires_at": nil},
					squirrel.Expr("ak.expires_at > NOW()"),
				},
			}
			baseQ = baseQ.Where(condition)
			totalQ = totalQ.Where(condition)
		case "inactive":
			condition := squirrel.Or{
				squirrel.NotEq{"ak.revoked_at": nil},
				squirrel.Expr("ak.expires_at <= NOW()"),
			}
			baseQ = baseQ.Where(condition)
			totalQ = totalQ.Where(condition)
		}
	}

	totalQuery, totalArgs, err = totalQ.ToSql()
	if err != nil {
		return
	}

	query, queryArgs, err = baseQ.
		OrderBy("ak.created_at DESC", "ak.id ASC").
		Limit(infrastructurerepositoryshared.NormalizeLimit(limit)).
		Offset(infrastructurerepositoryshared.NormalizeOffset(page, limit)).
		ToSql()
	return
}

func (p *postgresImpl) queryRegenerate(
	id uuid.UUID,
	keyHash string,
	keyLastFour string,
	expiresAt *time.Time,
	updatedBy *uuid.UUID,
) (query string, args []any, err error) {
	return p.SqrD.Update("api_keys").
		Where(squirrel.Eq{"id": id}).
		Set("key_hash", keyHash).
		Set("key_last_four", keyLastFour).
		Set("expires_at", expiresAt).
		Set("revoked_at", nil).
		Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("updated_by", updatedBy).
		ToSql()
}

func (p *postgresImpl) queryRevoke(
	id uuid.UUID,
	updatedBy *uuid.UUID,
) (query string, args []any, err error) {
	return p.SqrD.Update("api_keys").
		Where(squirrel.Eq{"id": id}).
		Set("revoked_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("updated_by", updatedBy).
		ToSql()
}

func (p *postgresImpl) queryDeleteById(id uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Delete("api_keys").
		Where(squirrel.Eq{"id": id}).
		ToSql()
}
