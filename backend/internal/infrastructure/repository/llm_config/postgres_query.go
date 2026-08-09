package infrastructurerepositoryllmconfig

import (
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

func (p *postgresImpl) queryGet() (query string, args []any, err error) {
	return p.SqrD.Select("id", "provider", "model", "api_key_encrypted", "base_url", "updated_at", "updated_by").
		From("llm_config").
		Limit(1).
		ToSql()
}

func (p *postgresImpl) queryUpsert(provider domainmodels.LlmProvider, model string, apiKeyEncrypted []byte, baseURL *string, updatedBy *uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Insert("llm_config").
		Columns("id", "provider", "model", "api_key_encrypted", "base_url", "singleton", "updated_by").
		Values(uuid.New(), string(provider), model, apiKeyEncrypted, baseURL, true, updatedBy).
		Suffix(`ON CONFLICT (singleton) DO UPDATE SET
			provider = EXCLUDED.provider,
			model = EXCLUDED.model,
			api_key_encrypted = EXCLUDED.api_key_encrypted,
			base_url = EXCLUDED.base_url,
			updated_at = NOW(),
			updated_by = EXCLUDED.updated_by
			RETURNING id`).
		ToSql()
}
