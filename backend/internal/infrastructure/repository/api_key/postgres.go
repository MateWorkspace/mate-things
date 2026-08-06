package infrastructurerepositoryapikey

import (
	"context"
	"errors"
	"time"

	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructurerepositoryshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/shared"
	"github.com/MateWorkspace/mate-things/backend/pkg/pgxdt"
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
) domaincontractsrepository.ApiKey {
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
	userId uuid.UUID,
	keyHash string,
	keyLastFour string,
	expiresAt *time.Time,
	createdBy *uuid.UUID,
) (id uuid.UUID, err error) {
	query, args, err := p.queryCreate(userId, keyHash, keyLastFour, expiresAt, createdBy)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create api key query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError(
			"failed to create api key", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "user_id", Type: domainmodels.ErrTypeApiKeyUserExists},
		)
	}

	return id, nil
}

func (p *postgresImpl) ReadByKeyHash(ctx context.Context, keyHash string) (apiKey *domainmodels.ApiKey, err error) {
	query, args, err := p.queryReadByKeyHash(keyHash)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read api key query", err)
	}

	item, err := infrastructurerepositoryshared.ScanPgxApiKey(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read api key", err)
	}

	return &item, nil
}

func (p *postgresImpl) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	search *string,
	status *string,
) (apiKeys []domainmodels.ApiKeyWithUser, total int, err error) {
	totalQuery, totalArgs, query, queryArgs, err := p.queryReadByPagination(page, limit, search, status)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.QueryBuildError("failed to build read api keys query", err)
	}

	if err := p.Dt.QueryRow(ctx, totalQuery, totalArgs...).Scan(&total); err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to count api keys", err)
	}
	if total == 0 {
		return []domainmodels.ApiKeyWithUser{}, 0, nil
	}

	rows, err := p.Dt.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to read api keys", err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanPgxApiKeysWithUser(rows)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to scan api keys", err)
	}

	return items, total, nil
}

func (p *postgresImpl) Regenerate(
	ctx context.Context,
	id uuid.UUID,
	keyHash string,
	keyLastFour string,
	expiresAt *time.Time,
	updatedBy *uuid.UUID,
) (err error) {
	query, args, err := p.queryRegenerate(id, keyHash, keyLastFour, expiresAt, updatedBy)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build regenerate api key query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to regenerate api key", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("api key not found", nil)
	}

	return nil
}

func (p *postgresImpl) Revoke(ctx context.Context, id uuid.UUID, updatedBy *uuid.UUID) (err error) {
	query, args, err := p.queryRevoke(id, updatedBy)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build revoke api key query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to revoke api key", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("api key not found", nil)
	}

	return nil
}

func (p *postgresImpl) DeleteById(ctx context.Context, id uuid.UUID) (err error) {
	query, args, err := p.queryDeleteById(id)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build delete api key query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to delete api key", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("api key not found", nil)
	}

	return nil
}
