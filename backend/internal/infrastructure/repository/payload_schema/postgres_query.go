package infrastructurerepositorypayloadschema

import (
	"encoding/json"
	"time"

	"github.com/Masterminds/squirrel"
	infrastructurerepositoryshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/shared"
	"github.com/google/uuid"
)

var payloadSchemaColumns = []string{
	"id",
	"name",
	"version",
	"definition",
	"valid_from",
	"valid_to",
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
	version int32,
	definition json.RawMessage,
	validFrom *time.Time,
	validTo *time.Time,
	createdBy *uuid.UUID,
) (query string, args []any, err error) {
	columns := []string{"name", "version", "definition", "created_by"}
	values := []any{name, version, definition, createdBy}

	if validFrom != nil {
		columns = append(columns, "valid_from")
		values = append(values, *validFrom)
	}
	if validTo != nil {
		columns = append(columns, "valid_to")
		values = append(values, *validTo)
	}

	return p.SqrD.Insert("payload_schemas").
		Columns(columns...).
		Values(values...).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryReadById(id uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(payloadSchemaColumns...).
		From("payload_schemas").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
}

func (p *postgresImpl) queryReadByNameAndVersion(
	name string,
	version int32,
) (query string, args []any, err error) {
	return p.SqrD.Select(payloadSchemaColumns...).
		From("payload_schemas").
		Where(squirrel.Eq{"name": name, "version": version}).
		Where("deleted_at IS NULL").
		Limit(1).
		ToSql()
}

func (p *postgresImpl) queryReadLatestByName(name string) (query string, args []any, err error) {
	return p.SqrD.Select(payloadSchemaColumns...).
		From("payload_schemas").
		Where(squirrel.Eq{"name": name}).
		Where("deleted_at IS NULL").
		OrderBy("version DESC", "created_at DESC", "id ASC").
		Limit(1).
		ToSql()
}

func (p *postgresImpl) queryReadByPagination(
	page int,
	limit int,
	search *string,
	validAt *time.Time,
) (totalQuery string, totalArgs []any, query string, queryArgs []any, err error) {
	baseQ := p.SqrD.Select(payloadSchemaColumns...).
		From("payload_schemas").
		Where("deleted_at IS NULL")
	totalQ := p.SqrD.Select("COUNT(*)").
		From("payload_schemas").
		Where("deleted_at IS NULL")

	if pattern, ok := infrastructurerepositoryshared.SearchPattern(search); ok {
		condition := squirrel.Expr("name ILIKE ?", pattern)
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if validAt != nil {
		condition := squirrel.Expr("valid_from <= ? AND (valid_to IS NULL OR valid_to > ?)", *validAt, *validAt)
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
	version *int32,
	definition *json.RawMessage,
	validFrom *time.Time,
	validTo *time.Time,
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) (query string, args []any, err error) {
	q := p.SqrD.Update("payload_schemas").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL")

	if name != nil {
		q = q.Set("name", *name)
	}
	if version != nil {
		q = q.Set("version", *version)
	}
	if definition != nil {
		q = q.Set("definition", *definition)
	}
	if validFrom != nil {
		q = q.Set("valid_from", *validFrom)
	}
	if validTo != nil {
		q = q.Set("valid_to", *validTo)
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
	return p.SqrD.Update("payload_schemas").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("deleted_by", deletedBy).
		ToSql()
}
