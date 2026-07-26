package infrastructurerepositoryrole

import (
	"context"
	"encoding/json"
	"errors"

	domaincontractsrepository "github.com/ABA-Developer/nusapala-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	infrastructurerepositoryshared "github.com/ABA-Developer/nusapala-things/backend/internal/infrastructure/repository/shared"
	"github.com/ABA-Developer/nusapala-things/backend/pkg/pgxdt"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type postgresImpl struct {
	infrastructurerepositoryshared.BasePostgres
}

func NewPostgresImpl(
	dt pgxdt.Pgxdt,
	sqrQuestion *squirrel.StatementBuilderType,
	sqrDollar *squirrel.StatementBuilderType,
) domaincontractsrepository.Role {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func (p *postgresImpl) Create(
	ctx context.Context,
	name string,
	description *string,
	isDefault *bool,
	createdBy *uuid.UUID,
) (id uuid.UUID, err error) {
	if isDefault != nil && *isDefault {
		err = p.Dt.WithTx(ctx, func(ctx context.Context) error {
			query, args, err := p.queryUnsetDefault(createdBy)
			if err != nil {
				return infrastructurerepositoryshared.QueryBuildError("failed to build unset default role query", err)
			}
			if _, err := p.Dt.Exec(ctx, query, args...); err != nil {
				return infrastructurerepositoryshared.MapPgxError("failed to unset default roles", err)
			}

			query, args, err = p.queryCreate(name, description, isDefault, createdBy)
			if err != nil {
				return infrastructurerepositoryshared.QueryBuildError("failed to build create role query", err)
			}
			if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
				return infrastructurerepositoryshared.MapPgxError("failed to create role", err)
			}
			return nil
		})
		if err != nil {
			return uuid.Nil, err
		}
		return id, nil
	}

	query, args, err := p.queryCreate(name, description, isDefault, createdBy)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create role query", err)
	}
	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create role", err)
	}

	return id, nil
}

func (p *postgresImpl) ReadById(ctx context.Context, id uuid.UUID) (role *domainmodels.Role, err error) {
	query, args, err := p.queryReadById(id)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read role query", err)
	}

	item, err := infrastructurerepositoryshared.ScanPgxRole(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("role not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read role", err)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByName(ctx context.Context, name string) (role *domainmodels.Role, err error) {
	query, args, err := p.queryReadByName(name)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read role query", err)
	}

	item, err := infrastructurerepositoryshared.ScanPgxRole(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("role not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read role", err)
	}

	return &item, nil
}

func (p *postgresImpl) ReadDefault(ctx context.Context) (role *domainmodels.Role, err error) {
	query, args, err := p.queryReadDefault()
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read default role query", err)
	}

	item, err := infrastructurerepositoryshared.ScanPgxRole(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("role not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read role", err)
	}

	return &item, nil
}

func (p *postgresImpl) ReadPermissions(ctx context.Context, roleId uuid.UUID) (permissions []domainmodels.Permission, err error) {
	query, args, err := p.queryReadPermissions(roleId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read role permissions query", err)
	}

	rows, err := p.Dt.Query(ctx, query, args...)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read role permissions", err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanPgxPermissions(rows)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to scan role permissions", err)
	}

	return items, nil
}

func (p *postgresImpl) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	search *string,
) (roles []domainmodels.Role, total int, err error) {
	totalQuery, totalArgs, query, queryArgs, err := p.queryReadByPagination(page, limit, search)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.QueryBuildError("failed to build read roles query", err)
	}

	if err := p.Dt.QueryRow(ctx, totalQuery, totalArgs...).Scan(&total); err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to count roles", err)
	}
	if total == 0 {
		return []domainmodels.Role{}, 0, nil
	}

	rows, err := p.Dt.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to read roles", err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanPgxRoles(rows)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to scan roles", err)
	}

	return items, total, nil
}

func (p *postgresImpl) UpdateById(
	ctx context.Context,
	id uuid.UUID,
	name *string,
	description *string,
	isDefault *bool,
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) (err error) {
	runUpdate := func(ctx context.Context) error {
		query, args, err := p.queryUpdateById(id, name, description, isDefault, preferences, updatedBy)
		if err != nil {
			return infrastructurerepositoryshared.QueryBuildError("failed to build update role query", err)
		}
		commandTag, err := p.Dt.Exec(ctx, query, args...)
		if err != nil {
			return infrastructurerepositoryshared.MapPgxError("failed to update role", err)
		}
		if commandTag.RowsAffected() == 0 {
			return infrastructurerepositoryshared.NotFound("role not found", nil)
		}
		return nil
	}

	if isDefault != nil && *isDefault {
		return p.Dt.WithTx(ctx, func(ctx context.Context) error {
			query, args, err := p.queryUnsetDefaultExcept(id, updatedBy)
			if err != nil {
				return infrastructurerepositoryshared.QueryBuildError("failed to build unset default role query", err)
			}
			if _, err := p.Dt.Exec(ctx, query, args...); err != nil {
				return infrastructurerepositoryshared.MapPgxError("failed to unset default roles", err)
			}
			return runUpdate(ctx)
		})
	}

	return runUpdate(ctx)
}

func (p *postgresImpl) DeleteById(
	ctx context.Context,
	id uuid.UUID,
	deletedBy *uuid.UUID,
) (err error) {
	query, args, err := p.queryDeleteById(id, deletedBy)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build delete role query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to delete role", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("role not found", nil)
	}

	return nil
}
