package infrastructurerepositoryinfrareddevice

import (
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

var infraredDeviceColumns = []string{
	"id",
	"infrared_device_type_id",
	"brand",
	"model",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

func (p *postgresImpl) queryCreate(infraredDeviceTypeId uuid.UUID, brand string, model string, createdBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Insert("infrared_device").
		Columns("infrared_device_type_id", "brand", "model", "created_by").
		Values(infraredDeviceTypeId, brand, model, createdBy).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryReadById(id uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(infraredDeviceColumns...).
		From("infrared_device").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
}

func (p *postgresImpl) queryDeleteById(id uuid.UUID, deletedBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Update("infrared_device").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("deleted_by", deletedBy).
		ToSql()
}
