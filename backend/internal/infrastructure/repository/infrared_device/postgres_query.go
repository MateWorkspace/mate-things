package infrastructurerepositoryinfrareddevice

import (
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (p *postgresImpl) queryCreate(infraredDeviceTypeId uuid.UUID, brand string, model string) (query string, args []any, err error) {
	return p.SqrD.Insert("infrared_device").
		Columns("infrared_device_type_id", "brand", "model").
		Values(infraredDeviceTypeId, brand, model).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryGetById(id uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select("id", "infrared_device_type_id", "brand", "model").
		From("infrared_device").
		Where(squirrel.Eq{"id": id}).
		ToSql()
}
