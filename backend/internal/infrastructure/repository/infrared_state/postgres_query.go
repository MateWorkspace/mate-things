package infrastructurerepositoryinfraredstate

import (
	"github.com/Masterminds/squirrel"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

var infraredStateColumns = []string{
	"id",
	"infrared_device_type_id",
	"name",
	"type",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

func (p *postgresImpl) queryReadListByDeviceTypeId(infraredDeviceTypeId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(infraredStateColumns...).
		From("infrared_state").
		Where(squirrel.Eq{"infrared_device_type_id": infraredDeviceTypeId}).
		Where("deleted_at IS NULL").
		OrderBy("name").
		ToSql()
}

func (p *postgresImpl) queryReadByDeviceTypeIdAndName(infraredDeviceTypeId uuid.UUID, name string) (query string, args []any, err error) {
	return p.SqrD.Select(infraredStateColumns...).
		From("infrared_state").
		Where(squirrel.Eq{"infrared_device_type_id": infraredDeviceTypeId, "name": name}).
		Where("deleted_at IS NULL").
		ToSql()
}

func (p *postgresImpl) queryCreate(infraredDeviceTypeId uuid.UUID, name string, stateType domainmodels.InfraredStateType, createdBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Insert("infrared_state").
		Columns("infrared_device_type_id", "name", "type", "created_by").
		Values(infraredDeviceTypeId, name, string(stateType), createdBy).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryDeleteById(id uuid.UUID, deletedBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Update("infrared_state").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("deleted_by", deletedBy).
		ToSql()
}
