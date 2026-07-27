package compositionseeder

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/MateWorkspace/mate-things/backend/internal/config"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructureloggerleveled "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/logger/leveled"
	infrastructurerepositoryaction "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/action"
	infrastructurerepositorynodeclass "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/node_class"
	infrastructurerepositorypayloadschema "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/payload_schema"
	infrastructurerepositorypermission "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/permission"
	infrastructurerepositoryrole "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/role"
	infrastructurerepositoryrolepermission "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/role_permission"
	infrastructurerepositoryuser "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/user"
	infrastructureutilitypassword "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/utility/password"
)

type infrastructure struct {
	logger domaincontractslogger.Leveled

	actionRepository         domaincontractsrepository.Action
	nodeClassRepository      domaincontractsrepository.NodeClass
	payloadSchemaRepository  domaincontractsrepository.PayloadSchema
	permissionRepository     domaincontractsrepository.Permission
	roleRepository           domaincontractsrepository.Role
	rolePermissionRepository domaincontractsrepository.RolePermission
	userRepository           domaincontractsrepository.User

	password domaincontractsutility.Password
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

	sqrQuestion := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)
	sqrDollar := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	actionRepository := infrastructurerepositoryaction.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	nodeClassRepository := infrastructurerepositorynodeclass.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	payloadSchemaRepository := infrastructurerepositorypayloadschema.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	permissionRepository := infrastructurerepositorypermission.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	roleRepository := infrastructurerepositoryrole.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	rolePermissionRepository := infrastructurerepositoryrolepermission.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	userRepository := infrastructurerepositoryuser.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)

	password := infrastructureutilitypassword.NewBcryptImpl(config.PasswordBcryptCost)

	l.infra = &infrastructure{
		logger: logger,

		actionRepository:         actionRepository,
		nodeClassRepository:      nodeClassRepository,
		payloadSchemaRepository:  payloadSchemaRepository,
		permissionRepository:     permissionRepository,
		roleRepository:           roleRepository,
		rolePermissionRepository: rolePermissionRepository,
		userRepository:           userRepository,

		password: password,
	}

	logger.Info(ctx, tag, "Infrastructure initialized", domainmodels.LoggerMeta{})

	return nil
}
