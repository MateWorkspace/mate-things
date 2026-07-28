package compositionseeder

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/MateWorkspace/mate-things/backend/internal/config"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

const path = "seeder"

type launcher struct {
	drv   *driver
	infra *infrastructure
	app   *application
}

func Launch() int {
	launcher := &launcher{}
	config.LoadEnv()
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := launcher.newDriver(ctx); err != nil {
		return 1
	}
	defer launcher.drv.cleanup()

	if err := launcher.newInfrastructure(ctx); err != nil {
		return 1
	}

	if err := launcher.newApplication(ctx); err != nil {
		return 1
	}

	if err := launcher.app.seeder.Run(ctx); err != nil {
		launcher.infra.logger.Error(ctx, path, "Seeding failed", domainmodels.LoggerMeta{"error": err.Error()})
		return 1
	}

	launcher.infra.logger.Info(ctx, path, "Seeding completed successfully", domainmodels.LoggerMeta{})

	return 0
}
