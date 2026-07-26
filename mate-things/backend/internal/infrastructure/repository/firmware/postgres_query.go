package infrastructurerepositoryfirmware

import (
	"encoding/json"

	infrastructurerepositoryshared "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/repository/shared"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

var firmwareColumns = []string{
	"id",
	"node_class_id",
	"name",
	"size",
	"checksum",
	"binary_path",
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
	name string,
	size int32,
	checksum string,
	binaryPath string,
	createdBy *uuid.UUID,
) (query string, args []any, err error) {
	return p.SqrD.Insert("firmwares").
		Columns("node_class_id", "name", "size", "checksum", "binary_path", "created_by").
		Values(nodeClassId, name, size, checksum, binaryPath, createdBy).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryReadById(id uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(firmwareColumns...).
		From("firmwares").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
}

func (p *postgresImpl) queryReadByName(name string) (query string, args []any, err error) {
	return p.SqrD.Select(firmwareColumns...).
		From("firmwares").
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
) (totalQuery string, totalArgs []any, query string, queryArgs []any, err error) {
	baseQ := p.SqrD.Select(firmwareColumns...).
		From("firmwares").
		Where("deleted_at IS NULL")
	totalQ := p.SqrD.Select("COUNT(*)").
		From("firmwares").
		Where("deleted_at IS NULL")

	if pattern, ok := infrastructurerepositoryshared.SearchPattern(search); ok {
		condition := squirrel.Or{
			squirrel.Expr("name ILIKE ?", pattern),
			squirrel.Expr("binary_path ILIKE ?", pattern),
		}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if nodeClassId != nil {
		condition := squirrel.Eq{"node_class_id": *nodeClassId}
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
	name *string,
	size *int32,
	checksum *string,
	binaryPath *string,
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) (query string, args []any, err error) {
	q := p.SqrD.Update("firmwares").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL")

	if nodeClassId != nil {
		q = q.Set("node_class_id", *nodeClassId)
	}
	if name != nil {
		q = q.Set("name", *name)
	}
	if size != nil {
		q = q.Set("size", *size)
	}
	if checksum != nil {
		q = q.Set("checksum", *checksum)
	}
	if binaryPath != nil {
		q = q.Set("binary_path", *binaryPath)
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
	return p.SqrD.Update("firmwares").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("deleted_by", deletedBy).
		ToSql()
}
