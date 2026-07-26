package compositionmain

import (
	"context"
	"net/http"

	"github.com/ABA-Developer/nusapala-things/backend/internal/config"
	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	presentationhttphandleraction "github.com/ABA-Developer/nusapala-things/backend/internal/presentation/http/handler/action"
	presentationhttphandleradmin "github.com/ABA-Developer/nusapala-things/backend/internal/presentation/http/handler/admin"
	presentationhttphandlerauth "github.com/ABA-Developer/nusapala-things/backend/internal/presentation/http/handler/auth"
	presentationhttphandlernode "github.com/ABA-Developer/nusapala-things/backend/internal/presentation/http/handler/node"
	presentationhttphandlerpreferences "github.com/ABA-Developer/nusapala-things/backend/internal/presentation/http/handler/preferences"
	presentationhttphandlerprofile "github.com/ABA-Developer/nusapala-things/backend/internal/presentation/http/handler/profile"
	presentationhttphandlertelemetry "github.com/ABA-Developer/nusapala-things/backend/internal/presentation/http/handler/telemetry"
	presentationhttproute "github.com/ABA-Developer/nusapala-things/backend/internal/presentation/http/route"
	presentationmqtthandler "github.com/ABA-Developer/nusapala-things/backend/internal/presentation/mqtt/handler"
	echomiddleware "github.com/labstack/echo/v5/middleware"
)

type presentation struct {
}

func (l *launcher) newPresentation(ctx context.Context) error {
	const tag = path + "/presentation"

	authHandler := presentationhttphandlerauth.NewHandler(l.app.authSession)
	profileHandler := presentationhttphandlerprofile.NewHandler(
		l.app.profileMe,
		l.app.profileAccount,
		l.app.profileSecurity,
	)
	adminHandler := presentationhttphandleradmin.NewHandler(
		l.app.adminPermissionManagement,
		l.app.adminRoleManagement,
		l.app.adminSchemaRegistry,
		l.app.adminUserManagement,
	)
	nodeHandler := presentationhttphandlernode.NewHandler(
		l.app.nodeClassManagement,
		l.app.nodeDeviceManagement,
		l.app.nodeFirmwareManagement,
		l.app.nodeOta,
	)
	actionHandler := presentationhttphandleraction.NewHandler(
		l.app.actionDefinition,
		l.app.actionExecution,
		l.app.actionHistory,
	)
	telemetryHandler := presentationhttphandlertelemetry.NewHandler(l.app.telemetryQuery)
	preferencesHandler := presentationhttphandlerpreferences.NewHandler(l.app.preferencesUpdate)
	presentationmqtthandler.New(
		l.infra.logger,
		l.app.nodeMessagingCallback,
	)

	l.drv.echo.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: config.HttpCorsAllowedOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
			http.MethodPatch,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"Origin",
			"X-Requested-With",
		},
		ExposeHeaders: []string{
			"Content-Disposition",
			"Content-Length",
			"X-Firmware-Binary-Path",
			"X-Firmware-Checksum",
			"X-Firmware-Name",
		},
	}))

	presentationhttproute.Route(l.drv.echo, presentationhttproute.Args{
		Auth:        authHandler,
		Profile:     profileHandler,
		Admin:       adminHandler,
		Node:        nodeHandler,
		Action:      actionHandler,
		Telemetry:   telemetryHandler,
		Preferences: preferencesHandler,
		Token:       l.infra.token,
	})

	l.pres = &presentation{}

	l.infra.logger.Info(
		ctx, tag,
		"Presentation initialized",
		domainmodels.LoggerMeta{},
	)

	return nil
}
