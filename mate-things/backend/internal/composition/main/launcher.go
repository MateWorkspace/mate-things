package compositionmain

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/MateWorkspace/mate-things/backend/internal/config"
)

const path = "main"

type launcher struct {
	drv   *driver
	infra *infrastructure
	app   *application
	pres  *presentation
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

	if err := launcher.newPresentation(ctx); err != nil {
		return 1
	}

	if err := launcher.drv.echoStartConfig.Start(ctx, launcher.drv.echo); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return 1
	}

	return 0
}
