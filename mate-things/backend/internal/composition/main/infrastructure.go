package compositionmain

import (
	"context"

	"github.com/ABA-Developer/nusapala-things/backend/internal/config"
	domaincontractscache "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/cache"
	domaincontractslogger "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/logger"
	domaincontractsnode "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/node"
	domaincontractsrepository "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/repository"
	domaincontractsstorage "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/storage"
	domaincontractsutility "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	infrastructurecacheaction "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/cache/action"
	infrastructurecachefirmware "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/cache/firmware"
	infrastructurecachenode "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/cache/node"
	infrastructurecachenodeclass "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/cache/node_class"
	infrastructurecachepayloadschema "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/cache/payload_schema"
	infrastructurecachepermission "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/cache/permission"
	infrastructurecacherole "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/cache/role"
	infrastructurecacherolepermission "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/cache/role_permission"
	infrastructurecacheshared "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/cache/shared"
	infrastructurecacheuser "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/cache/user"
	infrastructureloggerleveled "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/logger/leveled"
	infrastructurenodepublish "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/node/publish"
	infrastructurenodesubscriptions "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/node/subscriptions"
	infrastructurerepositoryaction "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/repository/action"
	infrastructurerepositoryactionlog "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/repository/action_log"
	infrastructurerepositoryfirmware "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/repository/firmware"
	infrastructurerepositorynode "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/repository/node"
	infrastructurerepositorynodeclass "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/repository/node_class"
	infrastructurerepositorypayloadschema "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/repository/payload_schema"
	infrastructurerepositorypermission "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/repository/permission"
	infrastructurerepositoryrole "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/repository/role"
	infrastructurerepositoryrolepermission "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/repository/role_permission"
	infrastructurerepositorytelemetryrecord "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/repository/telemetry_record"
	infrastructurerepositoryuser "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/repository/user"
	infrastructurestoragefirmware "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/storage/firmware"
	infrastructureutilitypassword "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/utility/password"
	infrastructureutilitypayloadschemavalidator "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/utility/payload_schema_validator"
	infrastructureutilitytoken "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/utility/token"
	infrastructureutilitytransactor "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/utility/transactor"
	"github.com/Masterminds/squirrel"
)

type infrastructure struct {
	logger domaincontractslogger.Leveled

	transactor domaincontractsutility.Transactor

	actionRepository          domaincontractsrepository.Action
	actionLogRepository       domaincontractsrepository.ActionLog
	firmwareRepository        domaincontractsrepository.Firmware
	nodeRepository            domaincontractsrepository.Node
	nodeClassRepository       domaincontractsrepository.NodeClass
	payloadSchemaRepository   domaincontractsrepository.PayloadSchema
	permissionRepository      domaincontractsrepository.Permission
	roleRepository            domaincontractsrepository.Role
	rolePermissionRepository  domaincontractsrepository.RolePermission
	telemetryRecordRepository domaincontractsrepository.TelemetryRecord
	userRepository            domaincontractsrepository.User

	actionCache         domaincontractscache.Action
	firmwareCache       domaincontractscache.Firmware
	nodeCache           domaincontractscache.Node
	nodeClassCache      domaincontractscache.NodeClass
	payloadSchemaCache  domaincontractscache.PayloadSchema
	permissionCache     domaincontractscache.Permission
	roleCache           domaincontractscache.Role
	rolePermissionCache domaincontractscache.RolePermission
	userCache           domaincontractscache.User

	firmwareStorage domaincontractsstorage.Firmware

	nodePublisher     domaincontractsnode.Publish
	nodeSubscriptions domaincontractsnode.Subscriptions

	password               domaincontractsutility.Password
	token                  domaincontractsutility.Token
	payloadSchemaValidator domaincontractsutility.PayloadSchemaValidator
}

func (l *launcher) newInfrastructure(ctx context.Context) error {
	const tag = path + "/infrastructure"

	var logger domaincontractslogger.Leveled
	switch config.LoggerDriver {
	case "zerolog":
		logger = infrastructureloggerleveled.NewZerologImpl(l.drv.zlogLogger, config.LoggerLevel)
	default:
		logger = infrastructureloggerleveled.NewSlogImpl(l.drv.slogLogger, config.LoggerLevel)
	}

	transactor := infrastructureutilitytransactor.NewPgxdtImpl(l.drv.tx)

	sqrQuestion := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)
	sqrDollar := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	actionRepository := infrastructurerepositoryaction.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	actionLogRepository := infrastructurerepositoryactionlog.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	firmwareRepository := infrastructurerepositoryfirmware.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	nodeRepository := infrastructurerepositorynode.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	nodeClassRepository := infrastructurerepositorynodeclass.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	payloadSchemaRepository := infrastructurerepositorypayloadschema.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	permissionRepository := infrastructurerepositorypermission.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	roleRepository := infrastructurerepositoryrole.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	rolePermissionRepository := infrastructurerepositoryrolepermission.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	telemetryRecordRepository := infrastructurerepositorytelemetryrecord.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	userRepository := infrastructurerepositoryuser.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)

	cacheTtl := infrastructurecacheshared.TtlConfig{
		Identity:   config.RedisTtlIdentity,
		Pagination: config.RedisTtlPagination,
		Relation:   config.RedisTtlRelation,
	}
	actionCache := infrastructurecacheaction.NewRedisImpl(l.drv.redisClient, config.RedisCacheNamespace, cacheTtl)
	firmwareCache := infrastructurecachefirmware.NewRedisImpl(l.drv.redisClient, config.RedisCacheNamespace, cacheTtl)
	nodeCache := infrastructurecachenode.NewRedisImpl(l.drv.redisClient, config.RedisCacheNamespace, cacheTtl)
	nodeClassCache := infrastructurecachenodeclass.NewRedisImpl(l.drv.redisClient, config.RedisCacheNamespace, cacheTtl)
	payloadSchemaCache := infrastructurecachepayloadschema.NewRedisImpl(l.drv.redisClient, config.RedisCacheNamespace, cacheTtl)
	permissionCache := infrastructurecachepermission.NewRedisImpl(l.drv.redisClient, config.RedisCacheNamespace, cacheTtl)
	roleCache := infrastructurecacherole.NewRedisImpl(l.drv.redisClient, config.RedisCacheNamespace, cacheTtl)
	rolePermissionCache := infrastructurecacherolepermission.NewRedisImpl(l.drv.redisClient, config.RedisCacheNamespace, cacheTtl)
	userCache := infrastructurecacheuser.NewRedisImpl(l.drv.redisClient, config.RedisCacheNamespace, cacheTtl)

	firmwareStorage := infrastructurestoragefirmware.NewMinioImpl(l.drv.minioClient, config.MinioBucket)
	nodePublisher := infrastructurenodepublish.NewMqttImpl(l.drv.mqttClient)
	nodeSubscriptions := infrastructurenodesubscriptions.NewMqttImpl(l.drv.mqttClient)
	password := infrastructureutilitypassword.NewBcryptImpl(config.PasswordBcryptCost)
	token := infrastructureutilitytoken.NewJwtImpl(
		config.TokenAccessSecret,
		config.TokenRefreshSecret,
		config.TokenAccessDuration,
		config.TokenRefreshDuration,
	)
	payloadSchemaValidator := infrastructureutilitypayloadschemavalidator.NewValidatorImpl()

	l.infra = &infrastructure{
		logger: logger,

		transactor: transactor,

		actionRepository:          actionRepository,
		actionLogRepository:       actionLogRepository,
		firmwareRepository:        firmwareRepository,
		nodeRepository:            nodeRepository,
		nodeClassRepository:       nodeClassRepository,
		payloadSchemaRepository:   payloadSchemaRepository,
		permissionRepository:      permissionRepository,
		roleRepository:            roleRepository,
		rolePermissionRepository:  rolePermissionRepository,
		telemetryRecordRepository: telemetryRecordRepository,
		userRepository:            userRepository,

		actionCache:         actionCache,
		firmwareCache:       firmwareCache,
		nodeCache:           nodeCache,
		nodeClassCache:      nodeClassCache,
		payloadSchemaCache:  payloadSchemaCache,
		permissionCache:     permissionCache,
		roleCache:           roleCache,
		rolePermissionCache: rolePermissionCache,
		userCache:           userCache,

		firmwareStorage: firmwareStorage,

		nodePublisher:     nodePublisher,
		nodeSubscriptions: nodeSubscriptions,

		password:               password,
		token:                  token,
		payloadSchemaValidator: payloadSchemaValidator,
	}

	logger.Info(ctx, tag, "Infrastructure initialized", domainmodels.LoggerMeta{})

	return nil
}
