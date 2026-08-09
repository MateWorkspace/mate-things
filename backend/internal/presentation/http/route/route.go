package presentationhttproute

import (
	"net/http"

	_ "github.com/MateWorkspace/mate-things/backend/docs/swagger"
	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	domainusecasesauth "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/auth"
	presentationhttpmiddleware "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/middleware"
	"github.com/labstack/echo/v5"
	echoSwagger "github.com/swaggo/echo-swagger/v2"
)

type PermissionMiddleware func(requiredPermissions ...string) echo.MiddlewareFunc

type AuthHandler interface {
	AuthLoginPost(c *echo.Context) error
	AuthRefreshPost(c *echo.Context) error
}

type ProfileHandler interface {
	ProfileGet(c *echo.Context) error
	ProfilePermissionsGet(c *echo.Context) error
	ProfilePatch(c *echo.Context) error
	ProfilePasswordPatch(c *echo.Context) error
}

type AdminHandler interface {
	PermissionPost(c *echo.Context) error
	PermissionGetList(c *echo.Context) error
	PermissionGetByName(c *echo.Context) error
	PermissionGetById(c *echo.Context) error
	PermissionPatch(c *echo.Context) error
	PermissionDelete(c *echo.Context) error

	LlmConfigGet(c *echo.Context) error
	LlmConfigPut(c *echo.Context) error

	RolePost(c *echo.Context) error
	RoleGetList(c *echo.Context) error
	RoleGetDefault(c *echo.Context) error
	RoleGetByName(c *echo.Context) error
	RoleGetById(c *echo.Context) error
	RolePermissionsGet(c *echo.Context) error
	RolePatch(c *echo.Context) error
	RoleSetDefaultPatch(c *echo.Context) error
	RoleDelete(c *echo.Context) error
	RolePermissionPost(c *echo.Context) error
	RolePermissionDeleteByPair(c *echo.Context) error
	RolePermissionGetList(c *echo.Context) error
	RolePermissionGetById(c *echo.Context) error
	RolePermissionGetByPair(c *echo.Context) error

	PayloadSchemaPost(c *echo.Context) error
	PayloadSchemaGetList(c *echo.Context) error
	PayloadSchemaGetLatest(c *echo.Context) error
	PayloadSchemaGetByNameAndVersion(c *echo.Context) error
	PayloadSchemaGetById(c *echo.Context) error
	PayloadSchemaPatch(c *echo.Context) error
	PayloadSchemaDelete(c *echo.Context) error

	UserPost(c *echo.Context) error
	UserGetList(c *echo.Context) error
	UserGetByUsername(c *echo.Context) error
	UserGetById(c *echo.Context) error
	UserPermissionsGet(c *echo.Context) error
	UserPatch(c *echo.Context) error
	UserPasswordPatch(c *echo.Context) error
	UserDelete(c *echo.Context) error

	ApiKeyGetList(c *echo.Context) error
	ApiKeyPost(c *echo.Context) error
	ApiKeyRegeneratePatch(c *echo.Context) error
	ApiKeyRevokePatch(c *echo.Context) error
	ApiKeyDelete(c *echo.Context) error
}

type NodeHandler interface {
	NodeClassPost(c *echo.Context) error
	NodeClassGetList(c *echo.Context) error
	NodeClassGetByName(c *echo.Context) error
	NodeClassGetById(c *echo.Context) error
	NodeClassPatch(c *echo.Context) error
	NodeClassDelete(c *echo.Context) error
	NodeClassActionsGet(c *echo.Context) error
	NodeClassActionPost(c *echo.Context) error
	NodeClassActionDeleteByPair(c *echo.Context) error
	NodeClassActionGetList(c *echo.Context) error
	NodeClassActionGetById(c *echo.Context) error
	NodeClassActionGetByPair(c *echo.Context) error

	NodeGetList(c *echo.Context) error
	NodeGetByDeviceId(c *echo.Context) error
	NodeGetById(c *echo.Context) error
	NodePatch(c *echo.Context) error
	NodeFirmwarePatch(c *echo.Context) error
	NodeDelete(c *echo.Context) error
	NodeConfigGet(c *echo.Context) error
	NodeConfigPut(c *echo.Context) error

	FirmwarePost(c *echo.Context) error
	FirmwareGetList(c *echo.Context) error
	FirmwareGetByNodeClassId(c *echo.Context) error
	FirmwareGetAvailableByNodeId(c *echo.Context) error
	FirmwareGetByName(c *echo.Context) error
	FirmwareGetById(c *echo.Context) error
	FirmwarePatch(c *echo.Context) error
	FirmwareBinaryPut(c *echo.Context) error
	FirmwareBinaryGetById(c *echo.Context) error
	FirmwareBinaryGetByName(c *echo.Context) error
	FirmwareBinaryStatByNameGet(c *echo.Context) error
	FirmwareBinaryStatByNameHead(c *echo.Context) error
	FirmwareDelete(c *echo.Context) error
	FirmwareConfigParametersGet(c *echo.Context) error

	OtaDispatchByNodeIdPost(c *echo.Context) error
	OtaDispatchByNodeDeviceIdPost(c *echo.Context) error
}

type ActionHandler interface {
	ActionPost(c *echo.Context) error
	ActionGetList(c *echo.Context) error
	ActionGetByName(c *echo.Context) error
	ActionGetById(c *echo.Context) error
	ActionPatch(c *echo.Context) error
	ActionDelete(c *echo.Context) error
	ActionDispatchPost(c *echo.Context) error
	ActionLogGetList(c *echo.Context) error
	ActionLogDelete(c *echo.Context) error
}

type TelemetryHandler interface {
	TelemetryRecordGetList(c *echo.Context) error
	TelemetryRecordGetLatest(c *echo.Context) error
	TelemetryRecordDelete(c *echo.Context) error
	TelemetryBroadcastRegister(c *echo.Context) error
	TelemetryBroadcastSessionList(c *echo.Context) error
}

type NodeLogHandler interface {
	NodeLogGetList(c *echo.Context) error
	NodeLogDelete(c *echo.Context) error
}

type PreferencesHandler interface {
	PreferencesPatch(c *echo.Context) error
}

type Args struct {
	Auth          AuthHandler
	Profile       ProfileHandler
	Admin         AdminHandler
	Node          NodeHandler
	Action        ActionHandler
	Telemetry     TelemetryHandler
	NodeLog       NodeLogHandler
	Preferences   PreferencesHandler
	Token         domaincontractsutility.Token
	ApiKeyAuth    domainusecasesauth.ApiKey
	MinioProxy    http.Handler
	FrontendProxy http.Handler
}

func Route(e *echo.Echo, args Args) {
	api := e.Group("/api")
	api.GET("/docs", func(c *echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/api/docs/index.html")
	})
	api.GET("/docs/*", echoSwagger.WrapHandler)

	publicV1 := api.Group("/v1")
	routeAuthPublic(publicV1, args.Auth)
	// The broadcast handshake authenticates itself (the browser WebSocket
	// API can't set an Authorization header, so the token travels as a
	// query param) — it can't run under the standard Auth middleware.
	routeTelemetryBroadcastRegister(publicV1, args.Telemetry)

	v1 := api.Group("/v1")
	v1.Use(presentationhttpmiddleware.Auth(args.Token, args.ApiKeyAuth))

	permission := presentationhttpmiddleware.Permission
	routeProfile(v1, args.Profile, permission)
	routeAdmin(v1, args.Admin, permission)
	routeNode(v1, args.Node, permission)
	routeAction(v1, args.Action, permission)
	routeTelemetry(v1, args.Telemetry, permission)
	routeNodeLog(v1, args.NodeLog, permission)
	routePreferences(v1, args.Preferences, permission)

	// Presigned S3-style GET/HEAD passthrough to MinIO - no auth middleware,
	// the signed query string is the auth. Not under /api: this is a raw
	// asset proxy, not a JSON API.
	minioProxyHandler := echo.WrapHandler(args.MinioProxy)
	e.GET("/minio-proxy/*", minioProxyHandler)
	e.HEAD("/minio-proxy/*", minioProxyHandler)

	// Everything else falls through to the Next.js frontend.
	e.Any("/*", echo.WrapHandler(args.FrontendProxy))
}

func routeAuthPublic(v1 *echo.Group, handler AuthHandler) {
	v1.POST("/auth/login", handler.AuthLoginPost)
	v1.POST("/auth/refresh", handler.AuthRefreshPost)
}

func routeTelemetryBroadcastRegister(v1 *echo.Group, handler TelemetryHandler) {
	v1.GET("/telemetry/broadcast", handler.TelemetryBroadcastRegister)
}

func routeProfile(v1 *echo.Group, handler ProfileHandler, permission PermissionMiddleware) {
	v1.GET("/profile/permissions", handler.ProfilePermissionsGet, permission("profile:get"))
	v1.GET("/profile", handler.ProfileGet, permission("profile:get"))
	v1.PATCH("/profile", handler.ProfilePatch, permission("profile:set"))
	v1.PATCH("/profile/password", handler.ProfilePasswordPatch, permission("profile_security:set"))
}

func routeAdmin(v1 *echo.Group, handler AdminHandler, permission PermissionMiddleware) {
	v1.GET("/admin/permissions", handler.PermissionGetList, permission("permission:get"))
	v1.POST("/admin/permissions", handler.PermissionPost, permission("permission:add"))
	v1.GET("/admin/permissions/by-name/:name", handler.PermissionGetByName, permission("permission:get"))
	v1.GET("/admin/permissions/:id", handler.PermissionGetById, permission("permission:get"))
	v1.PATCH("/admin/permissions/:id", handler.PermissionPatch, permission("permission:set"))
	v1.DELETE("/admin/permissions/:id", handler.PermissionDelete, permission("permission:remove"))

	v1.GET("/admin/llm-config", handler.LlmConfigGet, permission("llm_config:get"))
	v1.PUT("/admin/llm-config", handler.LlmConfigPut, permission("llm_config:set"))

	v1.GET("/admin/roles", handler.RoleGetList, permission("role:get"))
	v1.POST("/admin/roles", handler.RolePost, permission("role:add"))
	v1.GET("/admin/roles/default", handler.RoleGetDefault, permission("role:get"))
	v1.GET("/admin/roles/by-name/:name", handler.RoleGetByName, permission("role:get"))
	v1.GET("/admin/roles/:id/permissions", handler.RolePermissionsGet, permission("role_permission:get"))
	v1.PATCH("/admin/roles/:id/default", handler.RoleSetDefaultPatch, permission("role:set"))
	v1.GET("/admin/roles/:id", handler.RoleGetById, permission("role:get"))
	v1.PATCH("/admin/roles/:id", handler.RolePatch, permission("role:set"))
	v1.DELETE("/admin/roles/:id", handler.RoleDelete, permission("role:remove"))
	v1.POST("/admin/roles/:role_id/permissions/:permission_id", handler.RolePermissionPost, permission("role_permission:add"))
	v1.DELETE("/admin/roles/:role_id/permissions/:permission_id", handler.RolePermissionDeleteByPair, permission("role_permission:remove"))
	v1.GET("/admin/role-permissions", handler.RolePermissionGetList, permission("role_permission:get"))
	v1.GET("/admin/role-permissions/by-pair", handler.RolePermissionGetByPair, permission("role_permission:get"))
	v1.GET("/admin/role-permissions/:id", handler.RolePermissionGetById, permission("role_permission:get"))

	v1.GET("/admin/payload-schemas", handler.PayloadSchemaGetList, permission("payload_schema:get"))
	v1.POST("/admin/payload-schemas", handler.PayloadSchemaPost, permission("payload_schema:add"))
	v1.GET("/admin/payload-schemas/latest", handler.PayloadSchemaGetLatest, permission("payload_schema:get"))
	v1.GET("/admin/payload-schemas/by-name-version", handler.PayloadSchemaGetByNameAndVersion, permission("payload_schema:get"))
	v1.GET("/admin/payload-schemas/:id", handler.PayloadSchemaGetById, permission("payload_schema:get"))
	v1.PATCH("/admin/payload-schemas/:id", handler.PayloadSchemaPatch, permission("payload_schema:set"))
	v1.DELETE("/admin/payload-schemas/:id", handler.PayloadSchemaDelete, permission("payload_schema:remove"))

	v1.GET("/admin/users", handler.UserGetList, permission("user:get"))
	v1.POST("/admin/users", handler.UserPost, permission("user:add"))
	v1.GET("/admin/users/by-username/:username", handler.UserGetByUsername, permission("user:get"))
	v1.GET("/admin/users/:id/permissions", handler.UserPermissionsGet, permission("user_permission:get"))
	v1.PATCH("/admin/users/:id/password", handler.UserPasswordPatch, permission("user_password:set"))
	v1.GET("/admin/users/:id", handler.UserGetById, permission("user:get"))
	v1.PATCH("/admin/users/:id", handler.UserPatch, permission("user:set"))
	v1.DELETE("/admin/users/:id", handler.UserDelete, permission("user:remove"))

	v1.GET("/admin/api-keys", handler.ApiKeyGetList, permission("api_key:get"))
	v1.POST("/admin/api-keys", handler.ApiKeyPost, permission("api_key:add"))
	v1.PATCH("/admin/api-keys/:id/regenerate", handler.ApiKeyRegeneratePatch, permission("api_key:set"))
	v1.PATCH("/admin/api-keys/:id/revoke", handler.ApiKeyRevokePatch, permission("api_key:set"))
	v1.DELETE("/admin/api-keys/:id", handler.ApiKeyDelete, permission("api_key:remove"))
}

func routeNode(v1 *echo.Group, handler NodeHandler, permission PermissionMiddleware) {
	v1.GET("/node-classes", handler.NodeClassGetList, permission("node_class:get"))
	v1.POST("/node-classes", handler.NodeClassPost, permission("node_class:add"))
	v1.GET("/node-classes/by-name/:name", handler.NodeClassGetByName, permission("node_class:get"))
	v1.GET("/node-classes/:node_class_id/firmwares", handler.FirmwareGetByNodeClassId, permission("firmware:get"))
	v1.GET("/node-classes/:id/actions", handler.NodeClassActionsGet, permission("node_class_action:get"))
	v1.POST("/node-classes/:node_class_id/actions/:action_id", handler.NodeClassActionPost, permission("node_class_action:add"))
	v1.DELETE("/node-classes/:node_class_id/actions/:action_id", handler.NodeClassActionDeleteByPair, permission("node_class_action:remove"))
	v1.GET("/node-class-actions", handler.NodeClassActionGetList, permission("node_class_action:get"))
	v1.GET("/node-class-actions/by-pair", handler.NodeClassActionGetByPair, permission("node_class_action:get"))
	v1.GET("/node-class-actions/:id", handler.NodeClassActionGetById, permission("node_class_action:get"))
	v1.GET("/firmwares/:id/config-parameters", handler.FirmwareConfigParametersGet, permission("firmware:get"))
	v1.GET("/node-classes/:id", handler.NodeClassGetById, permission("node_class:get"))
	v1.PATCH("/node-classes/:id", handler.NodeClassPatch, permission("node_class:set"))
	v1.DELETE("/node-classes/:id", handler.NodeClassDelete, permission("node_class:remove"))

	v1.GET("/nodes", handler.NodeGetList, permission("node:get"))
	v1.GET("/nodes/by-device/:device_id", handler.NodeGetByDeviceId, permission("node:get"))
	v1.POST("/nodes/by-device/:device_id/ota", handler.OtaDispatchByNodeDeviceIdPost, permission("ota:dispatch"))
	v1.GET("/nodes/:node_id/firmwares/available", handler.FirmwareGetAvailableByNodeId, permission("firmware:get"))
	v1.PATCH("/nodes/:id/firmware", handler.NodeFirmwarePatch, permission("node:set"))
	v1.POST("/nodes/:id/ota", handler.OtaDispatchByNodeIdPost, permission("ota:dispatch"))
	v1.GET("/nodes/:id", handler.NodeGetById, permission("node:get"))
	v1.PATCH("/nodes/:id", handler.NodePatch, permission("node:set"))
	v1.DELETE("/nodes/:id", handler.NodeDelete, permission("node:remove"))
	v1.GET("/nodes/:id/config", handler.NodeConfigGet, permission("node_config:get"))
	v1.PUT("/nodes/:id/config", handler.NodeConfigPut, permission("node_config:set"))

	v1.GET("/firmwares", handler.FirmwareGetList, permission("firmware:get"))
	v1.POST("/firmwares", handler.FirmwarePost, permission("firmware:add"))
	v1.GET("/firmwares/by-name/:name/binary/stat", handler.FirmwareBinaryStatByNameGet, permission("firmware:get"))
	v1.HEAD("/firmwares/by-name/:name/binary", handler.FirmwareBinaryStatByNameHead, permission("firmware:get"))
	v1.GET("/firmwares/by-name/:name/binary", handler.FirmwareBinaryGetByName, permission("firmware:get"))
	v1.GET("/firmwares/by-name/:name", handler.FirmwareGetByName, permission("firmware:get"))
	v1.PUT("/firmwares/:id/binary", handler.FirmwareBinaryPut, permission("firmware:set"))
	v1.GET("/firmwares/:id/binary", handler.FirmwareBinaryGetById, permission("firmware:get"))
	v1.GET("/firmwares/:id", handler.FirmwareGetById, permission("firmware:get"))
	v1.PATCH("/firmwares/:id", handler.FirmwarePatch, permission("firmware:set"))
	v1.DELETE("/firmwares/:id", handler.FirmwareDelete, permission("firmware:remove"))
}

func routeAction(v1 *echo.Group, handler ActionHandler, permission PermissionMiddleware) {
	v1.GET("/actions", handler.ActionGetList, permission("action:get"))
	v1.POST("/actions", handler.ActionPost, permission("action:add"))
	v1.GET("/actions/by-name/:name", handler.ActionGetByName, permission("action:get"))
	v1.POST("/actions/:id/dispatch", handler.ActionDispatchPost, permission("action:dispatch"))
	v1.GET("/actions/:id", handler.ActionGetById, permission("action:get"))
	v1.PATCH("/actions/:id", handler.ActionPatch, permission("action:set"))
	v1.DELETE("/actions/:id", handler.ActionDelete, permission("action:remove"))
	v1.GET("/action-logs", handler.ActionLogGetList, permission("action_log:get"))
	v1.DELETE("/action-logs", handler.ActionLogDelete, permission("action_log:remove"))
}

func routeTelemetry(v1 *echo.Group, handler TelemetryHandler, permission PermissionMiddleware) {
	v1.GET("/telemetry", handler.TelemetryRecordGetList, permission("telemetry_record:get"))
	v1.DELETE("/telemetry", handler.TelemetryRecordDelete, permission("telemetry_record:remove"))
	v1.GET("/telemetry/latest", handler.TelemetryRecordGetLatest, permission("telemetry_record:get"))
	v1.GET("/telemetry/broadcast/sessions", handler.TelemetryBroadcastSessionList, permission("broadcast_session:get"))
}

func routeNodeLog(v1 *echo.Group, handler NodeLogHandler, permission PermissionMiddleware) {
	v1.GET("/node-logs", handler.NodeLogGetList, permission("node_log:get"))
	v1.DELETE("/node-logs", handler.NodeLogDelete, permission("node_log:remove"))
}

func routePreferences(v1 *echo.Group, handler PreferencesHandler, permission PermissionMiddleware) {
	v1.PATCH("/preferences/:resource/:id", handler.PreferencesPatch, permission("preferences:set"))
}
