package compositionseeder

import (
	"context"

	seederdata "github.com/MateWorkspace/mate-things/backend/database/seeder"
	applicationseeder "github.com/MateWorkspace/mate-things/backend/internal/application/seeder"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesseeder "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/seeder"
)

type application struct {
	seeder domainusecasesseeder.Seeder
}

func (l *launcher) newApplication(ctx context.Context) error {
	const tag = path + "/application"

	data, err := seederdata.Load()
	if err != nil {
		l.infra.logger.Error(ctx, tag, "Failed to load seed data", domainmodels.LoggerMeta{"error": err.Error()})
		return err
	}

	seeder := applicationseeder.NewUsecaseImpl(
		l.infra.permissionRepository,
		l.infra.roleRepository,
		l.infra.rolePermissionRepository,
		l.infra.nodeClassRepository,
		l.infra.payloadSchemaRepository,
		l.infra.actionRepository,
		l.infra.userRepository,
		l.infra.password,
		l.infra.logger,
		data,
	)

	l.app = &application{
		seeder: seeder,
	}

	l.infra.logger.Info(ctx, tag, "Application initialized", domainmodels.LoggerMeta{})

	return nil
}
