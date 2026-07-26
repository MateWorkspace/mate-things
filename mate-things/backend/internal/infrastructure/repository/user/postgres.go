package infrastructurerepositoryuser

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
) domaincontractsrepository.User {
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
	roleId uuid.UUID,
	name string,
	bio *string,
	username string,
	passwordHash string,
	createdBy *uuid.UUID,
) (id uuid.UUID, err error) {
	query, args, err := p.queryCreate(roleId, name, bio, username, passwordHash, createdBy)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create user query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create user", err)
	}

	return id, nil
}

func (p *postgresImpl) ReadById(ctx context.Context, id uuid.UUID) (user *domainmodels.User, err error) {
	query, args, err := p.queryReadById(id)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read user query", err)
	}

	item, err := infrastructurerepositoryshared.ScanPgxUser(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("user not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read user", err)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByUsername(ctx context.Context, username string) (user *domainmodels.User, err error) {
	query, args, err := p.queryReadByUsername(username)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read user query", err)
	}

	item, err := infrastructurerepositoryshared.ScanPgxUser(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("user not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read user", err)
	}

	return &item, nil
}

func (p *postgresImpl) ReadPermissions(ctx context.Context, userId uuid.UUID) (permissions []domainmodels.Permission, err error) {
	query, args, err := p.queryReadPermissions(userId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read user permissions query", err)
	}

	rows, err := p.Dt.Query(ctx, query, args...)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read user permissions", err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanPgxPermissions(rows)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to scan user permissions", err)
	}

	return items, nil
}

func (p *postgresImpl) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	search *string,
	roleId *uuid.UUID,
) (users []domainmodels.User, total int, err error) {
	totalQuery, totalArgs, query, queryArgs, err := p.queryReadByPagination(page, limit, search, roleId)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.QueryBuildError("failed to build read users query", err)
	}

	if err := p.Dt.QueryRow(ctx, totalQuery, totalArgs...).Scan(&total); err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to count users", err)
	}
	if total == 0 {
		return []domainmodels.User{}, 0, nil
	}

	rows, err := p.Dt.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to read users", err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanPgxUsers(rows)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to scan users", err)
	}

	return items, total, nil
}

func (p *postgresImpl) UpdateById(
	ctx context.Context,
	id uuid.UUID,
	roleId *uuid.UUID,
	name *string,
	bio *string,
	username *string,
	passwordHash *string,
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) (err error) {
	query, args, err := p.queryUpdateById(id, roleId, name, bio, username, passwordHash, preferences, updatedBy)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build update user query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to update user", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("user not found", nil)
	}

	return nil
}

func (p *postgresImpl) DeleteById(
	ctx context.Context,
	id uuid.UUID,
	deletedBy *uuid.UUID,
) (err error) {
	query, args, err := p.queryDeleteById(id, deletedBy)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build delete user query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to delete user", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("user not found", nil)
	}

	return nil
}
