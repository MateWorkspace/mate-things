package infrastructurerepositoryinfrareddevicetype

import (
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

var infraredDeviceTypeColumns = []string{
	"id",
	"name",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

func (p *postgresImpl) queryReadList() (query string, args []any, err error) {
	return p.SqrD.Select(infraredDeviceTypeColumns...).From("infrared_device_type").Where("deleted_at IS NULL").OrderBy("name").ToSql()
}

func (p *postgresImpl) queryReadByName(name string) (query string, args []any, err error) {
	return p.SqrD.Select(infraredDeviceTypeColumns...).From("infrared_device_type").Where(squirrel.Eq{"name": name}).Where("deleted_at IS NULL").ToSql()
}

func (p *postgresImpl) queryCreate(name string, createdBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Insert("infrared_device_type").Columns("name", "created_by").Values(name, createdBy).Suffix("RETURNING id").ToSql()
}

func (p *postgresImpl) queryDeleteById(id uuid.UUID, deletedBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Update("infrared_device_type").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("deleted_by", deletedBy).
		ToSql()
}
