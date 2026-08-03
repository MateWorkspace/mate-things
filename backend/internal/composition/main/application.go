package compositionmain

import (
	"context"

	applicationactiondefinition "github.com/MateWorkspace/mate-things/backend/internal/application/action/definition"
	applicationactionexecution "github.com/MateWorkspace/mate-things/backend/internal/application/action/execution"
	applicationactionhistory "github.com/MateWorkspace/mate-things/backend/internal/application/action/history"
	applicationadminpermissionmanagement "github.com/MateWorkspace/mate-things/backend/internal/application/admin/permission_management"
	applicationadminrolemanagement "github.com/MateWorkspace/mate-things/backend/internal/application/admin/role_management"
	applicationadminschemaregistry "github.com/MateWorkspace/mate-things/backend/internal/application/admin/schema_registry"
	applicationadminusermanagement "github.com/MateWorkspace/mate-things/backend/internal/application/admin/user_management"
	applicationauthsession "github.com/MateWorkspace/mate-things/backend/internal/application/auth/session"
	applicationnodeclassmanagement "github.com/MateWorkspace/mate-things/backend/internal/application/node/class_management"
	applicationnodeconfigparameter "github.com/MateWorkspace/mate-things/backend/internal/application/node/config_parameter"
	applicationnodeconfigvalue "github.com/MateWorkspace/mate-things/backend/internal/application/node/config_value"
	applicationnodedevicemanagement "github.com/MateWorkspace/mate-things/backend/internal/application/node/device_management"
	applicationnodefirmwaremanagement "github.com/MateWorkspace/mate-things/backend/internal/application/node/firmware_management"
	applicationnodemessagingcallback "github.com/MateWorkspace/mate-things/backend/internal/application/node/messaging_callback"
	applicationnodeota "github.com/MateWorkspace/mate-things/backend/internal/application/node/ota"
	applicationnodelogquery "github.com/MateWorkspace/mate-things/backend/internal/application/node_log/query"
	applicationpreferencesupdate "github.com/MateWorkspace/mate-things/backend/internal/application/preferences/update"
	applicationprofileaccount "github.com/MateWorkspace/mate-things/backend/internal/application/profile/account"
	applicationprofileme "github.com/MateWorkspace/mate-things/backend/internal/application/profile/me"
	applicationprofilesecurity "github.com/MateWorkspace/mate-things/backend/internal/application/profile/security"
	applicationrepocacheaction "github.com/MateWorkspace/mate-things/backend/internal/application/repocache/action"
	applicationrepocachefirmware "github.com/MateWorkspace/mate-things/backend/internal/application/repocache/firmware"
	applicationrepocachenode "github.com/MateWorkspace/mate-things/backend/internal/application/repocache/node"
	applicationrepocachenodeclass "github.com/MateWorkspace/mate-things/backend/internal/application/repocache/node_class"
	applicationrepocachenodeclassaction "github.com/MateWorkspace/mate-things/backend/internal/application/repocache/node_class_action"
	applicationrepocachepayloadschema "github.com/MateWorkspace/mate-things/backend/internal/application/repocache/payload_schema"
	applicationrepocachepermission "github.com/MateWorkspace/mate-things/backend/internal/application/repocache/permission"
	applicationrepocacherole "github.com/MateWorkspace/mate-things/backend/internal/application/repocache/role"
	applicationrepocacherolepermission "github.com/MateWorkspace/mate-things/backend/internal/application/repocache/role_permission"
	applicationrepocacheuser "github.com/MateWorkspace/mate-things/backend/internal/application/repocache/user"
	applicationtelemetryingestion "github.com/MateWorkspace/mate-things/backend/internal/application/telemetry/ingestion"
	applicationtelemetryquery "github.com/MateWorkspace/mate-things/backend/internal/application/telemetry/query"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesaction "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/action"
	domainusecasesadmin "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/admin"
	domainusecasesauth "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/auth"
	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
	domainusecasesnodelog "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node_log"
	domainusecasespreferences "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/preferences"
	domainusecasesprofile "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/profile"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
	domainusecasestelemetry "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/telemetry"
)

type application struct {
	actionRepoCache          domainusecasesrepocache.Action
	firmwareRepoCache        domainusecasesrepocache.Firmware
	nodeRepoCache            domainusecasesrepocache.Node
	nodeClassRepoCache       domainusecasesrepocache.NodeClass
	nodeClassActionRepoCache domainusecasesrepocache.NodeClassAction
	payloadSchemaRepoCache   domainusecasesrepocache.PayloadSchema
	permissionRepoCache      domainusecasesrepocache.Permission
	roleRepoCache            domainusecasesrepocache.Role
	rolePermissionRepoCache  domainusecasesrepocache.RolePermission
	userRepoCache            domainusecasesrepocache.User

	actionDefinition domainusecasesaction.Definition
	actionExecution  domainusecasesaction.Execution
	actionHistory    domainusecasesaction.History

	adminPermissionManagement domainusecasesadmin.PermissionManagement
	adminRoleManagement       domainusecasesadmin.RoleManagement
	adminSchemaRegistry       domainusecasesadmin.SchemaRegistry
	adminUserManagement       domainusecasesadmin.UserManagement

	authSession domainusecasesauth.Session

	nodeClassManagement    domainusecasesnode.ClassManagement
	nodeConfigParameter    domainusecasesnode.ConfigParameter
	nodeConfigValue        domainusecasesnode.ConfigValue
	nodeDeviceManagement   domainusecasesnode.DeviceManagement
	nodeFirmwareManagement domainusecasesnode.FirmwareManagement
	nodeMessagingCallback  domainusecasesnode.MessagingCallback
	nodeOta                domainusecasesnode.Ota

	preferencesUpdate domainusecasespreferences.Update

	profileAccount  domainusecasesprofile.Account
	profileMe       domainusecasesprofile.Me
	profileSecurity domainusecasesprofile.Security

	telemetryIngestion domainusecasestelemetry.Ingestion
	telemetryQuery     domainusecasestelemetry.Query
	nodeLogQuery       domainusecasesnodelog.Query
}

func (l *launcher) newApplication(ctx context.Context) error {
	const tag = path + "/application"

	actionRepoCache := applicationrepocacheaction.NewRepoCacheImpl(l.infra.actionRepository, l.infra.actionCache, l.infra.nodeClassActionCache)
	firmwareRepoCache := applicationrepocachefirmware.NewRepoCacheImpl(l.infra.firmwareRepository, l.infra.firmwareCache)
	nodeRepoCache := applicationrepocachenode.NewRepoCacheImpl(l.infra.nodeRepository, l.infra.nodeCache)
	nodeClassRepoCache := applicationrepocachenodeclass.NewRepoCacheImpl(l.infra.nodeClassRepository, l.infra.nodeClassCache, l.infra.nodeClassActionCache)
	nodeClassActionRepoCache := applicationrepocachenodeclassaction.NewRepoCacheImpl(
		l.infra.nodeClassActionRepository,
		l.infra.nodeClassActionCache,
		l.infra.nodeClassCache,
		l.infra.actionCache,
	)
	payloadSchemaRepoCache := applicationrepocachepayloadschema.NewRepoCacheImpl(l.infra.payloadSchemaRepository, l.infra.payloadSchemaCache)
	rolePermissionRepoCache := applicationrepocacherolepermission.NewRepoCacheImpl(
		l.infra.rolePermissionRepository,
		l.infra.rolePermissionCache,
		l.infra.roleCache,
		l.infra.userCache,
	)
	permissionRepoCache := applicationrepocachepermission.NewRepoCacheImpl(
		l.infra.permissionRepository,
		l.infra.permissionCache,
		l.infra.roleCache,
		l.infra.userCache,
		l.infra.rolePermissionCache,
	)
	roleRepoCache := applicationrepocacherole.NewRepoCacheImpl(
		l.infra.roleRepository,
		l.infra.roleCache,
		l.infra.rolePermissionCache,
	)
	userRepoCache := applicationrepocacheuser.NewRepoCacheImpl(l.infra.userRepository, l.infra.userCache)

	actionDefinition := applicationactiondefinition.NewUsecaseImpl(actionRepoCache, l.infra.logger)
	actionExecution := applicationactionexecution.NewUsecaseImpl(
		actionRepoCache,
		nodeRepoCache,
		nodeClassActionRepoCache,
		payloadSchemaRepoCache,
		l.infra.actionLogRepository,
		l.infra.nodePublisher,
		l.infra.payloadSchemaValidator,
		l.infra.logger,
	)
	actionHistory := applicationactionhistory.NewUsecaseImpl(l.infra.actionLogRepository, l.infra.logger)

	adminPermissionManagement := applicationadminpermissionmanagement.NewUsecaseImpl(permissionRepoCache, l.infra.logger)
	adminRoleManagement := applicationadminrolemanagement.NewUsecaseImpl(roleRepoCache, rolePermissionRepoCache, l.infra.logger)
	adminSchemaRegistry := applicationadminschemaregistry.NewUsecaseImpl(payloadSchemaRepoCache, l.infra.logger)
	adminUserManagement := applicationadminusermanagement.NewUsecaseImpl(userRepoCache, l.infra.password, l.infra.logger)

	authSession := applicationauthsession.NewUsecaseImpl(
		userRepoCache,
		roleRepoCache,
		l.infra.password,
		l.infra.token,
		l.infra.logger,
	)

	nodeClassManagement := applicationnodeclassmanagement.NewUsecaseImpl(nodeClassRepoCache, nodeClassActionRepoCache, l.infra.logger)
	nodeConfigParameter := applicationnodeconfigparameter.NewUsecaseImpl(l.infra.firmwareConfigParameterRepository, l.infra.logger)
	nodeDeviceManagement := applicationnodedevicemanagement.NewUsecaseImpl(nodeRepoCache, l.infra.logger)
	nodeFirmwareManagement := applicationnodefirmwaremanagement.NewUsecaseImpl(
		firmwareRepoCache,
		nodeRepoCache,
		l.infra.firmwareStorage,
		nodeConfigParameter,
		l.infra.logger,
	)
	nodeConfigValue := applicationnodeconfigvalue.NewUsecaseImpl(
		l.infra.nodeConfigValueRepository,
		l.infra.firmwareConfigParameterRepository,
		nodeRepoCache,
		l.infra.nodePublisher,
		l.infra.logger,
	)
	nodeMessagingCallback := applicationnodemessagingcallback.NewUsecaseImpl(
		nodeRepoCache,
		l.infra.actionLogRepository,
		l.infra.nodeLogRepository,
		l.infra.firmwareConfigParameterRepository,
		l.infra.nodeConfigValueRepository,
		l.infra.nodePublisher,
		l.infra.nodeSubscriptions,
		l.infra.logger,
	)
	nodeOta := applicationnodeota.NewUsecaseImpl(
		nodeRepoCache,
		firmwareRepoCache,
		l.infra.firmwareStorage,
		l.infra.nodePublisher,
		l.infra.logger,
	)

	preferencesUpdate := applicationpreferencesupdate.NewUsecaseImpl(
		actionRepoCache,
		firmwareRepoCache,
		nodeRepoCache,
		nodeClassRepoCache,
		payloadSchemaRepoCache,
		permissionRepoCache,
		roleRepoCache,
		userRepoCache,
		l.infra.logger,
	)

	profileAccount := applicationprofileaccount.NewUsecaseImpl(userRepoCache, l.infra.logger)
	profileMe := applicationprofileme.NewUsecaseImpl(userRepoCache, l.infra.logger)
	profileSecurity := applicationprofilesecurity.NewUsecaseImpl(userRepoCache, l.infra.password, l.infra.logger)

	telemetryIngestion := applicationtelemetryingestion.NewUsecaseImpl(
		l.infra.telemetryRecordRepository,
		payloadSchemaRepoCache,
		l.infra.payloadSchemaValidator,
		l.infra.logger,
	)
	telemetryQuery := applicationtelemetryquery.NewUsecaseImpl(l.infra.telemetryRecordRepository, l.infra.logger)
	nodeLogQuery := applicationnodelogquery.NewUsecaseImpl(l.infra.nodeLogRepository, l.infra.logger)

	l.app = &application{
		actionRepoCache:          actionRepoCache,
		firmwareRepoCache:        firmwareRepoCache,
		nodeRepoCache:            nodeRepoCache,
		nodeClassRepoCache:       nodeClassRepoCache,
		nodeClassActionRepoCache: nodeClassActionRepoCache,
		payloadSchemaRepoCache:   payloadSchemaRepoCache,
		permissionRepoCache:      permissionRepoCache,
		roleRepoCache:            roleRepoCache,
		rolePermissionRepoCache:  rolePermissionRepoCache,
		userRepoCache:            userRepoCache,

		actionDefinition: actionDefinition,
		actionExecution:  actionExecution,
		actionHistory:    actionHistory,

		adminPermissionManagement: adminPermissionManagement,
		adminRoleManagement:       adminRoleManagement,
		adminSchemaRegistry:       adminSchemaRegistry,
		adminUserManagement:       adminUserManagement,

		authSession: authSession,

		nodeClassManagement:    nodeClassManagement,
		nodeConfigParameter:    nodeConfigParameter,
		nodeConfigValue:        nodeConfigValue,
		nodeDeviceManagement:   nodeDeviceManagement,
		nodeFirmwareManagement: nodeFirmwareManagement,
		nodeMessagingCallback:  nodeMessagingCallback,
		nodeOta:                nodeOta,

		preferencesUpdate: preferencesUpdate,

		profileAccount:  profileAccount,
		profileMe:       profileMe,
		profileSecurity: profileSecurity,

		telemetryIngestion: telemetryIngestion,
		telemetryQuery:     telemetryQuery,
		nodeLogQuery:       nodeLogQuery,
	}

	l.infra.logger.Info(
		ctx,
		tag,
		"Application initialized",
		domainmodels.LoggerMeta{},
	)

	return nil
}
