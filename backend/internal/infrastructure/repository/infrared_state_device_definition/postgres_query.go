package infrastructurerepositoryinfraredstatedevicedefinition

import (
	"github.com/Masterminds/squirrel"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

func (p *postgresImpl) queryCreate(definition domainmodels.InfraredStateDeviceDefinition) (query string, args []any, err error) {
	return p.SqrD.Insert("infrared_state_device_definition").
		Columns("infrared_device_id", "infrared_state_id", "options", "minimum", "maximum", "step").
		Values(definition.InfraredDeviceId, definition.InfraredStateId, definition.Options, definition.Minimum, definition.Maximum, definition.Step).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryListByDeviceId(infraredDeviceId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select("id", "infrared_device_id", "infrared_state_id", "options", "minimum", "maximum", "step").
		From("infrared_state_device_definition").
		Where(squirrel.Eq{"infrared_device_id": infraredDeviceId}).
		ToSql()
}
