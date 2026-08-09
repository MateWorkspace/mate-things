CREATE TABLE llm_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    api_key_encrypted BYTEA NOT NULL,
    base_url TEXT,
    singleton BOOLEAN NOT NULL DEFAULT TRUE UNIQUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_llm_config_singleton CHECK (singleton = TRUE),
    CONSTRAINT chk_llm_config_provider CHECK (provider IN ('CLAUDE', 'OPENAI'))
);
