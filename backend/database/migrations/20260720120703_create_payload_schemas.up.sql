CREATE TABLE payload_schemas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    name TEXT NOT NULL,
    version INT NOT NULL DEFAULT 1,
    definition JSONB NOT NULL,
    valid_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    valid_to TIMESTAMPTZ,
    preferences JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by UUID,
    updated_by UUID,
    deleted_by UUID,
    CONSTRAINT uq_payload_schemas_name_version UNIQUE (name, version)
);

CREATE INDEX idx_payload_schemas_definition_gin ON payload_schemas USING GIN (definition);

CREATE INDEX idx_payload_schemas_name ON payload_schemas (name);

CREATE INDEX idx_payload_schemas_name_trgm ON payload_schemas USING GIN (name gin_trgm_ops);

CREATE INDEX idx_payload_schemas_valid_from_valid_to ON payload_schemas (valid_from, valid_to);

CREATE INDEX idx_payload_schemas_deleted_at ON payload_schemas (deleted_at);