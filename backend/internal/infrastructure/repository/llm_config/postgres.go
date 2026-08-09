package infrastructurerepositoryllmconfig

import (
	"context"
	"errors"

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
) domaincontractsrepository.LlmConfig {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func (p *postgresImpl) Get(ctx context.Context) (*domainmodels.LlmConfig, error) {
	query, args, err := p.queryGet()
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build llm_config get query", err)
	}

	config, err := scanPgxLlmConfig(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read llm_config", err)
	}
	return &config, nil
}

func (p *postgresImpl) Upsert(ctx context.Context, provider domainmodels.LlmProvider, model string, apiKeyEncrypted []byte, baseURL *string, updatedBy *uuid.UUID) (id uuid.UUID, err error) {
	query, args, err := p.queryUpsert(provider, model, apiKeyEncrypted, baseURL, updatedBy)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build llm_config upsert query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to upsert llm_config", err)
	}

	return id, nil
}

func scanPgxLlmConfig(row pgx.Row) (domainmodels.LlmConfig, error) {
	var config domainmodels.LlmConfig
	var provider string
	err := row.Scan(&config.Id, &provider, &config.Model, &config.ApiKeyEncrypted, &config.BaseURL, &config.UpdatedAt, &config.UpdatedBy)
	config.Provider = domainmodels.LlmProvider(provider)
	return config, err
}
