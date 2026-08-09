package compositionmain

import (
	"context"
	"net/http"

	"github.com/MateWorkspace/mate-things/backend/internal/config"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	presentationhttphandleraction "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/handler/action"
	presentationhttphandleradmin "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/handler/admin"
	presentationhttphandlerauth "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/handler/auth"
	presentationhttphandlerinfrared "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/handler/infrared"
	presentationhttphandlernode "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/handler/node"
	presentationhttphandlernodelog "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/handler/node_log"
	presentationhttphandlerpreferences "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/handler/preferences"
	presentationhttphandlerprofile "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/handler/profile"
	presentationhttphandlertelemetry "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/handler/telemetry"
	presentationhttpproxy "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/proxy"
	presentationhttproute "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/route"
	presentationmqtthandler "github.com/MateWorkspace/mate-things/backend/internal/presentation/mqtt/handler"
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
		l.app.adminApiKeyManagement,
		l.app.adminLlmConfigManagement,
	)
	nodeHandler := presentationhttphandlernode.NewHandler(
		l.app.nodeClassManagement,
		l.app.nodeDeviceManagement,
		l.app.nodeFirmwareManagement,
		l.app.nodeOta,
		l.app.nodeConfigParameter,
		l.app.nodeConfigValue,
	)
	actionHandler := presentationhttphandleraction.NewHandler(
		l.app.actionDefinition,
		l.app.actionExecution,
		l.app.actionHistory,
	)
	telemetryHandler := presentationhttphandlertelemetry.NewHandler(l.app.telemetryQuery, l.app.telemetryBroadcast, l.infra.token)
	nodeLogHandler := presentationhttphandlernodelog.NewHandler(l.app.nodeLogQuery)
	preferencesHandler := presentationhttphandlerpreferences.NewHandler(l.app.preferencesUpdate)
	infraredHandler := presentationhttphandlerinfrared.NewHandler(
		l.app.infraredRecordSessionManagement,
		l.infra.infraredRecordSessionBroadcaster,
		l.infra.token,
	)
	presentationmqtthandler.New(
		l.infra.logger,
		l.app.nodeMessagingCallback,
		l.app.infraredRecordSessionManagement,
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
		Auth:          authHandler,
		Profile:       profileHandler,
		Admin:         adminHandler,
		Node:          nodeHandler,
		Action:        actionHandler,
		Telemetry:     telemetryHandler,
		NodeLog:       nodeLogHandler,
		Preferences:   preferencesHandler,
		Infrared:      infraredHandler,
		Token:         l.infra.token,
		ApiKeyAuth:    l.app.authApiKey,
		MinioProxy:    presentationhttpproxy.NewMinioProxy(config.MinioEndpoint, config.MinioUseSsl),
		FrontendProxy: presentationhttpproxy.NewFrontendProxy("127.0.0.1:3000"),
	})

	l.pres = &presentation{}

	l.infra.logger.Info(
		ctx, tag,
		"Presentation initialized",
		domainmodels.LoggerMeta{},
	)

	return nil
}
