package infrastructurerepositoryinfraredstatedevicedefinition

import (
	"github.com/Masterminds/squirrel"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

var infraredStateDeviceDefinitionColumns = []string{
	"id",
	"infrared_device_id",
	"infrared_state_id",
	"options",
	"minimum",
	"maximum",
	"step",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

func (p *postgresImpl) queryCreate(definition domainmodels.InfraredStateDeviceDefinition, createdBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Insert("infrared_state_device_definition").
		Columns("infrared_device_id", "infrared_state_id", "options", "minimum", "maximum", "step", "created_by").
		Values(definition.InfraredDeviceId, definition.InfraredStateId, definition.Options, definition.Minimum, definition.Maximum, definition.Step, createdBy).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryListByDeviceId(infraredDeviceId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(infraredStateDeviceDefinitionColumns...).
		From("infrared_state_device_definition").
		Where(squirrel.Eq{"infrared_device_id": infraredDeviceId}).
		Where("deleted_at IS NULL").
		ToSql()
}

func (p *postgresImpl) queryDeleteById(id uuid.UUID, deletedBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Update("infrared_state_device_definition").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("deleted_by", deletedBy).
		ToSql()
}
