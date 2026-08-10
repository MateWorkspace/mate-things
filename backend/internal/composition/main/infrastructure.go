package compositionmain

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/MateWorkspace/mate-things/backend/internal/config"
	domaincontractsbroadcaster "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/broadcaster"
	domaincontractscache "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/cache"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsnode "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/node"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domaincontractsstorage "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/storage"
	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructurebroadcasterinfraredrecordsession "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/broadcaster/infrared_record_session"
	infrastructurebroadcastertelemetry "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/broadcaster/telemetry"
	infrastructurecacheaction "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/cache/action"
	infrastructurecacheapikey "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/cache/api_key"
	infrastructurecachefirmware "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/cache/firmware"
	infrastructurecachenode "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/cache/node"
	infrastructurecachenodeclass "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/cache/node_class"
	infrastructurecachenodeclassaction "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/cache/node_class_action"
	infrastructurecachepayloadschema "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/cache/payload_schema"
	infrastructurecachepermission "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/cache/permission"
	infrastructurecacherole "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/cache/role"
	infrastructurecacherolepermission "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/cache/role_permission"
	infrastructurecacheshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/cache/shared"
	infrastructurecacheuser "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/cache/user"
	infrastructurejsengine "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/jsengine"
	infrastructurellm "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/llm"
	infrastructureloggerleveled "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/logger/leveled"
	infrastructurenodepublish "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/node/publish"
	infrastructurenodesubscriptions "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/node/subscriptions"
	infrastructurerepositoryaction "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/action"
	infrastructurerepositoryactionlog "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/action_log"
	infrastructurerepositoryapikey "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/api_key"
	infrastructurerepositoryfirmware "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/firmware"
	infrastructurerepositoryfirmwareconfigparameter "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/firmware_config_parameter"
	infrastructurerepositoryinfrareddevice "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/infrared_device"
	infrastructurerepositoryinfrareddevicetype "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/infrared_device_type"
	infrastructurerepositoryinfraredrecordsession "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/infrared_record_session"
	infrastructurerepositoryinfraredstate "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/infrared_state"
	infrastructurerepositoryinfraredstatecoder "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/infrared_state_coder"
	infrastructurerepositoryinfraredstatedevicedefinition "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/infrared_state_device_definition"
	infrastructurerepositoryinfraredstatedevicerecordcase "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/infrared_state_device_record_case"
	infrastructurerepositoryllmconfig "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/llm_config"
	infrastructurerepositorynode "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/node"
	infrastructurerepositorynodeclass "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/node_class"
	infrastructurerepositorynodeclassaction "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/node_class_action"
	infrastructurerepositorynodeconfigvalue "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/node_config_value"
	infrastructurerepositorynodelog "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/node_log"
	infrastructurerepositorypayloadschema "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/payload_schema"
	infrastructurerepositorypermission "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/permission"
	infrastructurerepositoryrole "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/role"
	infrastructurerepositoryrolepermission "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/role_permission"
	infrastructurerepositorytelemetryrecord "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/telemetry_record"
	infrastructurerepositoryuser "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/user"
	infrastructurestoragefirmware "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/storage/firmware"
	infrastructureutilityapikey "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/utility/apikey"
	infrastructureutilityencryption "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/utility/encryption"
	infrastructureutilitypassword "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/utility/password"
	infrastructureutilitypayloadschemavalidator "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/utility/payload_schema_validator"
	infrastructureutilitytoken "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/utility/token"
	infrastructureutilitytransactor "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/utility/transactor"
)

type infrastructure struct {
	logger domaincontractslogger.Leveled

	transactor domaincontractsutility.Transactor

	actionRepository                        domaincontractsrepository.Action
	actionLogRepository                     domaincontractsrepository.ActionLog
	apiKeyRepository                        domaincontractsrepository.ApiKey
	firmwareRepository                      domaincontractsrepository.Firmware
	firmwareConfigParameterRepository       domaincontractsrepository.FirmwareConfigParameter
	infraredDeviceTypeRepository            domaincontractsrepository.InfraredDeviceType
	infraredStateRepository                 domaincontractsrepository.InfraredState
	infraredDeviceRepository                domaincontractsrepository.InfraredDevice
	infraredStateDeviceDefinitionRepository domaincontractsrepository.InfraredStateDeviceDefinition
	infraredRecordSessionRepository         domaincontractsrepository.InfraredRecordSession
	infraredStateDeviceRecordCaseRepository domaincontractsrepository.InfraredStateDeviceRecordCase
	infraredStateCoderRepository            domaincontractsrepository.InfraredStateCoder
	nodeRepository                          domaincontractsrepository.Node
	nodeConfigValueRepository               domaincontractsrepository.NodeConfigValue
	nodeLogRepository                       domaincontractsrepository.NodeLog
	nodeClassRepository                     domaincontractsrepository.NodeClass
	nodeClassActionRepository               domaincontractsrepository.NodeClassAction
	llmConfigRepository                     domaincontractsrepository.LlmConfig
	payloadSchemaRepository                 domaincontractsrepository.PayloadSchema
	permissionRepository                    domaincontractsrepository.Permission
	roleRepository                          domaincontractsrepository.Role
	rolePermissionRepository                domaincontractsrepository.RolePermission
	telemetryRecordRepository               domaincontractsrepository.TelemetryRecord
	userRepository                          domaincontractsrepository.User

	actionCache          domaincontractscache.Action
	apiKeyCache          domaincontractscache.ApiKey
	firmwareCache        domaincontractscache.Firmware
	nodeCache            domaincontractscache.Node
	nodeClassCache       domaincontractscache.NodeClass
	nodeClassActionCache domaincontractscache.NodeClassAction
	payloadSchemaCache   domaincontractscache.PayloadSchema
	permissionCache      domaincontractscache.Permission
	roleCache            domaincontractscache.Role
	rolePermissionCache  domaincontractscache.RolePermission
	userCache            domaincontractscache.User

	firmwareStorage domaincontractsstorage.Firmware

	nodePublisher     domaincontractsnode.Publish
	nodeSubscriptions domaincontractsnode.Subscriptions

	telemetryBroadcaster             domaincontractsbroadcaster.Telemetry
	infraredRecordSessionBroadcaster domaincontractsbroadcaster.InfraredRecordSession

	password               domaincontractsutility.Password
	token                  domaincontractsutility.Token
	payloadSchemaValidator domaincontractsutility.PayloadSchemaValidator
	apiKeyGenerator        domaincontractsutility.ApiKey
	llmEncryptor           domaincontractsutility.Encryptor

	llmClientFactory *infrastructurellm.ClientFactory
	encoderRunner    encoderRunnerAdapter
}

type encoderRunnerAdapter struct{}

func (encoderRunnerAdapter) RunEncoder(source string, state map[string]string, timeout time.Duration) ([]int32, error) {
	return infrastructurejsengine.RunEncoder(source, state, timeout)
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
	apiKeyRepository := infrastructurerepositoryapikey.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	firmwareRepository := infrastructurerepositoryfirmware.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	firmwareConfigParameterRepository := infrastructurerepositoryfirmwareconfigparameter.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	infraredDeviceTypeRepository := infrastructurerepositoryinfrareddevicetype.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	infraredStateRepository := infrastructurerepositoryinfraredstate.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	infraredDeviceRepository := infrastructurerepositoryinfrareddevice.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	infraredStateDeviceDefinitionRepository := infrastructurerepositoryinfraredstatedevicedefinition.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	infraredRecordSessionRepository := infrastructurerepositoryinfraredrecordsession.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	infraredStateDeviceRecordCaseRepository := infrastructurerepositoryinfraredstatedevicerecordcase.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	infraredStateCoderRepository := infrastructurerepositoryinfraredstatecoder.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	nodeRepository := infrastructurerepositorynode.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	nodeConfigValueRepository := infrastructurerepositorynodeconfigvalue.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	nodeLogRepository := infrastructurerepositorynodelog.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	nodeClassRepository := infrastructurerepositorynodeclass.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	nodeClassActionRepository := infrastructurerepositorynodeclassaction.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	llmConfigRepository := infrastructurerepositoryllmconfig.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
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
	apiKeyCache := infrastructurecacheapikey.NewRedisImpl(l.drv.redisClient, config.RedisCacheNamespace, cacheTtl)
	firmwareCache := infrastructurecachefirmware.NewRedisImpl(l.drv.redisClient, config.RedisCacheNamespace, cacheTtl)
	nodeCache := infrastructurecachenode.NewRedisImpl(l.drv.redisClient, config.RedisCacheNamespace, cacheTtl)
	nodeClassCache := infrastructurecachenodeclass.NewRedisImpl(l.drv.redisClient, config.RedisCacheNamespace, cacheTtl)
	nodeClassActionCache := infrastructurecachenodeclassaction.NewRedisImpl(l.drv.redisClient, config.RedisCacheNamespace, cacheTtl)
	payloadSchemaCache := infrastructurecachepayloadschema.NewRedisImpl(l.drv.redisClient, config.RedisCacheNamespace, cacheTtl)
	permissionCache := infrastructurecachepermission.NewRedisImpl(l.drv.redisClient, config.RedisCacheNamespace, cacheTtl)
	roleCache := infrastructurecacherole.NewRedisImpl(l.drv.redisClient, config.RedisCacheNamespace, cacheTtl)
	rolePermissionCache := infrastructurecacherolepermission.NewRedisImpl(l.drv.redisClient, config.RedisCacheNamespace, cacheTtl)
	userCache := infrastructurecacheuser.NewRedisImpl(l.drv.redisClient, config.RedisCacheNamespace, cacheTtl)

	firmwareStorage := infrastructurestoragefirmware.NewMinioImpl(l.drv.minioClient, config.MinioBucket, config.BaseUrl, config.MinioPresignDuration)
	nodePublisher := infrastructurenodepublish.NewMqttImpl(l.drv.mqttClient)
	nodeSubscriptions := infrastructurenodesubscriptions.NewMqttImpl(l.drv.mqttClient)
	telemetryBroadcaster := infrastructurebroadcastertelemetry.NewGorillaImpl(config.HttpCorsAllowedOrigins)
	infraredRecordSessionBroadcaster := infrastructurebroadcasterinfraredrecordsession.NewGorillaImpl(config.HttpCorsAllowedOrigins)
	password := infrastructureutilitypassword.NewBcryptImpl(config.PasswordBcryptCost)
	token := infrastructureutilitytoken.NewJwtImpl(
		config.TokenAccessSecret,
		config.TokenRefreshSecret,
		config.TokenAccessDuration,
		config.TokenRefreshDuration,
	)
	payloadSchemaValidator := infrastructureutilitypayloadschemavalidator.NewValidatorImpl()
	apiKeyGenerator := infrastructureutilityapikey.NewGeneratorImpl()

	llmEncryptor, err := infrastructureutilityencryption.NewAESGCMImpl(config.LlmEncryptionKey)
	if err != nil {
		return fmt.Errorf("failed to construct llm encryptor (check BE_LLM_ENCRYPTION_KEY is set to exactly 32 bytes): %w", err)
	}
	llmClientFactory := infrastructurellm.NewClientFactory(llmConfigRepository, llmEncryptor)
	encoderRunner := encoderRunnerAdapter{}

	l.infra = &infrastructure{
		logger: logger,

		transactor: transactor,

		actionRepository:                        actionRepository,
		actionLogRepository:                     actionLogRepository,
		apiKeyRepository:                        apiKeyRepository,
		firmwareRepository:                      firmwareRepository,
		firmwareConfigParameterRepository:       firmwareConfigParameterRepository,
		infraredDeviceTypeRepository:            infraredDeviceTypeRepository,
		infraredStateRepository:                 infraredStateRepository,
		infraredDeviceRepository:                infraredDeviceRepository,
		infraredStateDeviceDefinitionRepository: infraredStateDeviceDefinitionRepository,
		infraredRecordSessionRepository:         infraredRecordSessionRepository,
		infraredStateDeviceRecordCaseRepository: infraredStateDeviceRecordCaseRepository,
		infraredStateCoderRepository:            infraredStateCoderRepository,
		nodeRepository:                          nodeRepository,
		nodeConfigValueRepository:               nodeConfigValueRepository,
		nodeLogRepository:                       nodeLogRepository,
		nodeClassRepository:                     nodeClassRepository,
		nodeClassActionRepository:               nodeClassActionRepository,
		llmConfigRepository:                     llmConfigRepository,
		payloadSchemaRepository:                 payloadSchemaRepository,
		permissionRepository:                    permissionRepository,
		roleRepository:                          roleRepository,
		rolePermissionRepository:                rolePermissionRepository,
		telemetryRecordRepository:               telemetryRecordRepository,
		userRepository:                          userRepository,

		actionCache:          actionCache,
		apiKeyCache:          apiKeyCache,
		firmwareCache:        firmwareCache,
		nodeCache:            nodeCache,
		nodeClassCache:       nodeClassCache,
		nodeClassActionCache: nodeClassActionCache,
		payloadSchemaCache:   payloadSchemaCache,
		permissionCache:      permissionCache,
		roleCache:            roleCache,
		rolePermissionCache:  rolePermissionCache,
		userCache:            userCache,

		firmwareStorage: firmwareStorage,

		nodePublisher:     nodePublisher,
		nodeSubscriptions: nodeSubscriptions,

		telemetryBroadcaster:             telemetryBroadcaster,
		infraredRecordSessionBroadcaster: infraredRecordSessionBroadcaster,

		password:               password,
		token:                  token,
		payloadSchemaValidator: payloadSchemaValidator,
		apiKeyGenerator:        apiKeyGenerator,
		llmEncryptor:           llmEncryptor,

		llmClientFactory: llmClientFactory,
		encoderRunner:    encoderRunner,
	}

	logger.Info(ctx, tag, "Infrastructure initialized", domainmodels.LoggerMeta{})

	return nil
}
