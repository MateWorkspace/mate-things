package infrastructurerepositoryinfraredstate

import (
	"github.com/Masterminds/squirrel"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

func (p *postgresImpl) queryListByDeviceTypeId(infraredDeviceTypeId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select("id", "infrared_device_type_id", "name", "type").
		From("infrared_state").
		Where(squirrel.Eq{"infrared_device_type_id": infraredDeviceTypeId}).
		OrderBy("name").
		ToSql()
}

func (p *postgresImpl) queryReadByDeviceTypeIdAndName(infraredDeviceTypeId uuid.UUID, name string) (query string, args []any, err error) {
	return p.SqrD.Select("id", "infrared_device_type_id", "name", "type").
		From("infrared_state").
		Where(squirrel.Eq{"infrared_device_type_id": infraredDeviceTypeId, "name": name}).
		ToSql()
}

func (p *postgresImpl) queryCreate(infraredDeviceTypeId uuid.UUID, name string, stateType domainmodels.InfraredStateType) (query string, args []any, err error) {
	return p.SqrD.Insert("infrared_state").
		Columns("infrared_device_type_id", "name", "type").
		Values(infraredDeviceTypeId, name, string(stateType)).
		Suffix("RETURNING id").
		ToSql()
}
